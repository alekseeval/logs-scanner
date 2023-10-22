package kube

import (
	"bufio"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"regexp"
	"scan_project/internal/model"
	"strings"
	"sync"
	"time"
)

type KubeScanner struct {
	kubeConfigsDAO    ClusterDAOI
	scansDAO          ScansDAOI
	kubernetesTimeout *int
	logger            *logrus.Entry
	stopChan          chan struct{}
	startProcessWg    sync.WaitGroup
	jobsRegexp        *regexp.Regexp
}

func NewKubeScanner(kubeConfigsDAO ClusterDAOI, scansDAO ScansDAOI, KubernetesTimeout *int, logger *logrus.Entry) *KubeScanner {
	// TODO: Проброс из кофига ключевого слова для грепа? error использовать как default?
	return &KubeScanner{
		kubeConfigsDAO:    kubeConfigsDAO,
		scansDAO:          scansDAO,
		kubernetesTimeout: KubernetesTimeout,
		jobsRegexp:        regexp.MustCompile("(?i)error"),
		startProcessWg:    sync.WaitGroup{},
		logger:            logger,
		stopChan:          make(chan struct{}, 1),
	}
}

// Start will do scan all kubernetes clusters gotten from ClusterDAOI every intervalSec seconds and save result in ScansDAOI
//
// The first scan will take place immediately
func (ks *KubeScanner) Start(intervalSec int) {
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
	ks.logger.Info("Stopping KubeScanner")
	ks.startProcessWg.Wait()
	ks.logger.Info("KubeScanner successfully stopped")
}

// ScanAll scans all configs and namespaces from model.ClusterDAOI and saved them into model.ScansDAOI
func (ks *KubeScanner) ScanAll() {
	kubeConfigs, err := ks.kubeConfigsDAO.GetAllConfigs()
	if err != nil {
		ks.logger.
			WithField("error", err).
			Error("Failed to get configs from DB")
	}
	for _, cfg := range kubeConfigs {
		kubeConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(cfg.Config))
		if err != nil {
			ks.logger.
				WithField("error", err).
				Error("Failed to initialize kubernetes config from DB string")
		}
		kubeConfig.Timeout = time.Duration(*ks.kubernetesTimeout) * time.Second
		clientSet, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			ks.logger.
				WithField("error", err).
				Error("Failed to initialize kubernetes config client set")
		}
		for _, ns := range cfg.NameSpaces {
			servicesScans, jobsScans, err := ks.ScanNamespace(clientSet, ns)
			if err != nil {
				ks.logger.
					WithField("error", err).
					Error(fmt.Sprintf("Failed to scan namespace %s", ns))
			}
			err = ks.scansDAO.UpdateJobsScans(cfg.Name, ns, jobsScans)
			if err != nil {
				ks.logger.
					WithField("error", err).
					Error("Failed to save Jobs scans")
			}
			err = ks.scansDAO.UpdateServicesScans(cfg.Name, ns, servicesScans)
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
		// Get some pod logs
		switch pod.Status.Phase {
		case v1.PodRunning:
			serviceScan, err := ks.ScanServiceLog(kubeClient, &pod)
			if err != nil {
				ks.logger.
					WithField("error", err).
					Error("Error occured while getting pod logs")
			}
			servicesScans = append(servicesScans, *serviceScan)
		case v1.PodFailed, v1.PodSucceeded:
			jobScan, err := ks.ScanJobLog(kubeClient, &pod)
			if err != nil {
				ks.logger.
					WithField("error", err).
					Error("Error occured while getting pod logs")
			}
			jobsScans = append(jobsScans, *jobScan)
		}
	}
	return servicesScans, jobsScans, nil
}

func (ks *KubeScanner) ScanServiceLog(kubeClient *kubernetes.Clientset, pod *v1.Pod) (serviceScan *model.ServiceScan, err error) {
	serviceScan = &model.ServiceScan{
		ServiceName:     pod.Name,
		LogTypeCountMap: make(map[model.LogLevelType]int),
	}
	serviceScan.ServiceName = pod.Name
	serviceScan.RestartsCount = 0                                   // TODO: нужно выцеплять из состояния контейнера внутри пода.. Надо ли оно?
	serviceScan.Uptime = time.Now().Sub(pod.CreationTimestamp.Time) // TODO: is that correct?
	podLogOpts := &v1.PodLogOptions{}
	req := kubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, podLogOpts)
	podLogsStream, err := req.Stream(context.TODO())
	if err != nil {
		return nil, err
	}
	defer podLogsStream.Close() // TODO: нужно ли закрывать???
	scanner := bufio.NewScanner(podLogsStream)
	linesCount := 0
	for scanner.Scan() {
		linesCount++
		log := &model.CommonServiceLog{}
		err := json.Unmarshal(scanner.Bytes(), log)
		if err != nil {
			serviceScan.NoneJsonLinesCount++
			continue
		}
		switch log.Level {
		case model.Trace, model.Debug, model.Info, model.Warning, model.Error, model.Fatal:
			serviceScan.LogTypeCountMap[log.Level] += 1
		default:
			ks.logger.Warning(fmt.Sprintf("Unknown log level -- %s", log.Level))
		}
	}
	serviceScan.TotalLines = linesCount
	serviceScan.ScanFinishTime = time.Now()
	return serviceScan, nil
}

func (ks *KubeScanner) ScanJobLog(kubeClient *kubernetes.Clientset, pod *v1.Pod) (*model.JobScan, error) {
	podLogOpts := &v1.PodLogOptions{}
	req := kubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, podLogOpts)
	podLogsStream, err := req.Stream(context.TODO())
	if err != nil {
		return nil, err
	}
	defer podLogsStream.Close() // TODO: нужно ли закрывать???
	var sb strings.Builder
	matchedLogRows := make([]string, 0)
	scanner := bufio.NewScanner(podLogsStream)
	for scanner.Scan() {
		strokeText := scanner.Text()
		if ks.jobsRegexp.MatchString(strokeText) {
			matchedLogRows = append(matchedLogRows, strokeText)
		}
		sb.WriteString(scanner.Text())
		sb.WriteRune('\n')
	}
	return &model.JobScan{
		JobName:        pod.Name,
		Age:            time.Now().Sub(pod.CreationTimestamp.Time), // TODO: is that correct?
		FullLog:        sb.String(),
		GrepPattern:    *ks.jobsRegexp,
		GrepLog:        matchedLogRows,
		ScanFinishTime: time.Now(),
	}, nil
}
