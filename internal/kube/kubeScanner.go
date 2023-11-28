package kube

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"regexp"
	"scan_project/configuration"
	"scan_project/internal/model"
	"sync"
	"time"
)

type KubeScanner struct {
	storage           StorageI
	kubernetesTimeout *int
	logger            *logrus.Entry
	stopChan          chan struct{}
	startProcessWg    sync.WaitGroup
	jobsRegexp        *regexp.Regexp
	isRunning         bool
	servicesRegexp    *regexp.Regexp
}

func NewKubeScanner(storage StorageI, cfg *configuration.Config, logger *logrus.Entry) *KubeScanner {
	return &KubeScanner{
		storage:           storage,
		kubernetesTimeout: cfg.System.Kubernetes.Timeout,
		jobsRegexp:        regexp.MustCompile(cfg.JobsGrepPattern),
		startProcessWg:    sync.WaitGroup{},
		logger:            logger,
		stopChan:          make(chan struct{}, 1),
		isRunning:         false,
		servicesRegexp:    regexp.MustCompile("\"level\":\"\\w+\""),
	}
}

// Start will do scan all kubernetes clusters gotten from ClusterDAOI every intervalSec seconds and save result in ScansDAOI
//
// The first scan will take place immediately
func (ks *KubeScanner) Start(intervalSec int) error {
	if ks.isRunning {
		return fmt.Errorf("kube-scanner are already running")
	}
	ks.isRunning = true
	ks.startProcessWg.Add(1)
	defer ks.startProcessWg.Done()
	ks.ScanAll()
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	for {
		select {
		case <-ks.stopChan:
			ticker.Stop()
			return nil
		case <-ticker.C:
			ks.ScanAll()
			ticker.Reset(time.Duration(intervalSec) * time.Second)
		}
	}
}

func (ks *KubeScanner) Shutdown(ctx context.Context) error {
	if !ks.isRunning {
		return fmt.Errorf("kube-scanner is already down")
	}
	ks.stopChan <- struct{}{}
	ks.isRunning = false
	ks.logger.Info("Stopping KubeScanner")
	shutdownWg := make(chan struct{}, 1)
	go func() {
		ks.startProcessWg.Wait()
		shutdownWg <- struct{}{}
	}()
	select {
	case <-shutdownWg:
		ks.logger.Info("KubeScanner successfully stopped")
	case <-ctx.Done():
		ks.logger.Info("KubeScanner is forced to stop")
		return fmt.Errorf("kube-scanner is forced to stop. Some processes may be not done yet")
	}
	return nil
}

// ScanAll scans all configs and namespaces from model.ClusterDAOI and saved them into model.ScansDAOI
func (ks *KubeScanner) ScanAll() {
	clusters, err := ks.storage.GetAllClusters()
	if err != nil {
		ks.logger.
			WithField("error", err).
			Error("failed to get clusters from DB")
		return
	}
	for _, cluster := range clusters {
		ks.ScanCluster(cluster)
	}
}

func (ks *KubeScanner) ScanCluster(cluster model.Cluster) {
	ks.logger.Tracef("Start scanning cluster %s", cluster.Name)
	wg := sync.WaitGroup{}
	for _, namespace := range cluster.Namespaces {
		wg.Add(1)
		go func(ns string) {
			defer wg.Done()
			ks.logger.Tracef("Start scanning namespace %s in cluster %s", ns, cluster.Name)
			err := ks.ScanNamespace(cluster, ns)
			if err != nil {
				ks.logger.
					WithField("error", err).
					Errorf("Failed to scan namespace %s in cluster %s", ns, cluster.Name)
			} else {
				ks.logger.Tracef("Successfully scanned namespace %s in cluster %s", ns, cluster.Name)
			}
		}(namespace)
	}
	wg.Wait()
	ks.logger.Tracef("%s cluster scan completed", cluster.Name)
}

// ScanNamespace return scans for jobs and services into specific Namespace for cluster
func (ks *KubeScanner) ScanNamespace(cluster model.Cluster, namespace string) error {
	// Stop scanning if app are shutting down
	if !ks.isRunning {
		return fmt.Errorf("service was stopped, abort all scans")
	}
	// Init kubernetes REST
	kubeConfigBytes, err := base64.StdEncoding.DecodeString(cluster.Config)
	if err != nil {
		ks.logger.
			WithField("error", err).
			WithField("kube-config", cluster.Config).
			Error("Failed to decode kube-config")
		return err
	}
	kubeRest, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigBytes)
	if err != nil {
		ks.logger.
			WithField("error", err).
			Error("Failed to initialize kubernetes config from DB string")
		return err
	}
	kubeRest.Timeout = time.Duration(*ks.kubernetesTimeout) * time.Second
	kubeClient, err := kubernetes.NewForConfig(kubeRest)
	if err != nil {
		ks.logger.
			WithField("error", err).
			Errorf("Failed to initialize kubernetes config client set for cluster %s", cluster.Name)
		return err
	}
	// List all pods
	pods, err := kubeClient.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		ks.logger.
			WithField("error", err).
			Errorf("Failed to list all pods using kubernetes ClientSet for cluster %s", cluster.Name)
		return err
	}
	var (
		servicesScans = make([]model.ServiceScan, 0)
		jobsScans     = make([]model.JobScan, 0)
	)
	// Scan gotten pods
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	for _, pod := range pods.Items {
		wg.Add(1)
		time.Sleep(time.Second / 5) //  Else stack on cluster limits and get Error from time to time
		go func(p v1.Pod) {
			defer wg.Done()
			switch p.Status.Phase {
			case v1.PodRunning:
				serviceScan, err := ks.scanServiceLog(kubeClient, &p)
				if err != nil {
					ks.logger.
						WithField("error", err).
						Errorf("Error occured while scanning pod %s logs", p.Name)
					return
				}
				mutex.Lock()
				servicesScans = append(servicesScans, *serviceScan)
				mutex.Unlock()
			case v1.PodFailed, v1.PodSucceeded:
				jobScan, err := ks.scanJobLog(kubeClient, &p)
				if err != nil {
					ks.logger.
						WithField("error", err).
						Errorf("Error occured while getting pod %s logs", p.Name)
					return
				}
				mutex.Lock()
				jobsScans = append(jobsScans, *jobScan)
				mutex.Unlock()
			}
		}(pod)
	}
	wg.Wait()
	// Save all scans result
	err = ks.storage.UpdateServicesScans(cluster.Name, namespace, servicesScans)
	if err != nil {
		ks.logger.
			WithField("error", err).
			Error("failed to save services scans")
		return err
	}
	err = ks.storage.UpdateJobsScans(cluster.Name, namespace, jobsScans)
	if err != nil {
		ks.logger.
			WithField("error", err).
			Error("failed to save jobs scans")
		return err
	}
	return nil
}
