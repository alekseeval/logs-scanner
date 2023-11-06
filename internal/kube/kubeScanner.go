package kube

import (
	"context"
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
}

func NewKubeScanner(storage StorageI, cfg *configuration.Config, logger *logrus.Entry) *KubeScanner {
	// TODO: Проброс из кофига ключевого слова для грепа? error использовать как default?
	return &KubeScanner{
		storage:           storage,
		kubernetesTimeout: cfg.System.Kubernetes.Timeout,
		jobsRegexp:        regexp.MustCompile("(?i)error"),
		startProcessWg:    sync.WaitGroup{},
		logger:            logger,
		stopChan:          make(chan struct{}, 1),
		isRunning:         false,
	}
}

// Start will do scan all kubernetes clusters gotten from ClusterDAOI every intervalSec seconds and save result in ScansDAOI
//
// The first scan will take place immediately
func (ks *KubeScanner) Start(intervalSec int) {
	ks.isRunning = true
	ks.startProcessWg.Add(1)
	defer ks.startProcessWg.Done()
	ks.ScanAll()
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	for {
		select {
		case <-ks.stopChan:
			ticker.Stop()
			return
		case <-ticker.C:
			ks.ScanAll()
			ticker.Reset(time.Duration(intervalSec) * time.Second)
		}
	}
}

func (ks *KubeScanner) Shutdown() {
	ks.stopChan <- struct{}{}
	ks.isRunning = false
	ks.logger.Info("Stopping KubeScanner")
	ks.startProcessWg.Wait()
	ks.logger.Info("KubeScanner successfully stopped")
}

// ScanAll scans all configs and namespaces from model.ClusterDAOI and saved them into model.ScansDAOI
func (ks *KubeScanner) ScanAll() {
	clusters, err := ks.storage.GetAllClusters()
	if err != nil {
		ks.logger.
			WithField("error", err).
			Error("Failed to get configs from DB")
		return
	}
	for _, cluster := range clusters {
		kubeRest, err := clientcmd.RESTConfigFromKubeConfig([]byte(cluster.Config))
		if err != nil {
			ks.logger.
				WithField("error", err).
				Error("Failed to initialize kubernetes config from DB string")
			continue
		}
		kubeRest.Timeout = time.Duration(*ks.kubernetesTimeout) * time.Second
		kubeClientSet, err := kubernetes.NewForConfig(kubeRest)
		if err != nil {
			ks.logger.
				WithField("error", err).
				Error("Failed to initialize kubernetes config client set")
		}
		for _, ns := range cluster.Namespaces {
			ks.logger.Debugf("Start scan namespace %s from cluster %s", ns, cluster.Name)
			servicesScans, jobsScans, err := ks.ScanNamespace(kubeClientSet, ns)
			if err != nil {
				ks.logger.
					WithField("error", err).
					Error(fmt.Sprintf("Failed to scan namespace %s", ns))
				continue
			}
			ks.logger.Debugf("Namespace %s from cluster %s was successfully scanned", ns, cluster.Name)
			err = ks.storage.UpdateJobsScans(cluster.Name, ns, jobsScans)
			if err != nil {
				ks.logger.
					WithField("error", err).
					Error("Failed to save Jobs scans")
			}
			err = ks.storage.UpdateServicesScans(cluster.Name, ns, servicesScans)
			if err != nil {
				ks.logger.
					WithField("error", err).
					Error("Failed to save Services scans")
			}
		}
	}
}

// ScanNamespace return scans for jobs and services into specific Namespace for cluster
func (ks *KubeScanner) ScanNamespace(kubeClient *kubernetes.Clientset, namespace string) (servicesScans []model.ServiceScan, jobsScans []model.JobScan, err error) {
	// Get all pods

	pods, err := kubeClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	for _, pod := range pods.Items {
		// Get pod logs
		if !ks.isRunning {
			return nil, nil, fmt.Errorf("scanner was stopped")
		}
		switch pod.Status.Phase {
		case v1.PodRunning:
			serviceScan, err := ks.scanServiceLog(kubeClient, &pod)
			if err != nil {
				ks.logger.
					WithField("error", err).
					Error("Error occured while getting pod logs")
				continue
			}
			servicesScans = append(servicesScans, *serviceScan)
		case v1.PodFailed, v1.PodSucceeded:
			jobScan, err := ks.scanJobLog(kubeClient, &pod)
			if err != nil {
				ks.logger.
					WithField("error", err).
					Error("Error occured while getting pod logs")
				continue
			}
			jobsScans = append(jobsScans, *jobScan)
		}
	}
	return servicesScans, jobsScans, nil
}
