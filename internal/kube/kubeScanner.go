package kube

// TODO:
//		- Распараллеливание скана подов? Инициализировать WG внтури скана неймспейса и параллелить скан джоб и сервисов

import (
	"bufio"
	"context"
	"fmt"
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
	stopChan          chan struct{}
	startProcessWg    sync.WaitGroup
	jobsRegexp        *regexp.Regexp
}

func NewKubeScanner(kubeConfigsDAO ClusterDAOI, scansDAO ScansDAOI, KubernetesTimeout *int) *KubeScanner {
	// TODO: Проброс из кофига ключевого слова для грепа? error использовать как default?
	return &KubeScanner{
		kubeConfigsDAO:    kubeConfigsDAO,
		scansDAO:          scansDAO,
		kubernetesTimeout: KubernetesTimeout,
		jobsRegexp:        regexp.MustCompile("(?i)error"),
		startProcessWg:    sync.WaitGroup{},
	}
}

// Start will do scan all kubernetes clusters gotten from ClusterDAOI every intervalSec seconds and save result in ScansDAOI
//
// The first scan will take place immediately
func (ks *KubeScanner) Start(intervalSec int, ctx context.Context) {
	ks.startProcessWg.Add(1)
	defer ks.startProcessWg.Done()
	ks.ScanAll()
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
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
	ks.startProcessWg.Wait()
	fmt.Println("Scanner successfully stopped") // TODO: log info
}

// ScanAll scans all configs and namespaces from model.ClusterDAOI and saved them into model.ScansDAOI
func (ks *KubeScanner) ScanAll() {
	kubeConfigs, err := ks.kubeConfigsDAO.GetAllConfigs()
	if err != nil {
		fmt.Println("Failed to get configs from DB -- ", err) // TODO: log error
	}
	for _, cfg := range kubeConfigs {
		kubeConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(cfg.Config))
		if err != nil {
			fmt.Println("Failed to initialize kubernetes config from DB string -- ", err) // TODO: log error
		}
		kubeConfig.Timeout = time.Duration(*ks.kubernetesTimeout) * time.Second
		clientSet, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			fmt.Println("Failed to initialize kubernetes config client set -- ", err) // TODO: log error
		}
		for _, ns := range cfg.NameSpaces {
			servicesScans, jobsScans, err := ks.ScanNamespace(clientSet, ns)
			if err != nil {
				fmt.Println("Failed to scan ", ns) // TODO: log error
			}
			err = ks.scansDAO.UpdateJobsScans(cfg.Name, ns, jobsScans)
			if err != nil {
				fmt.Println("Failed to save Jobs scans") // TODO: log error
			}
			err = ks.scansDAO.UpdateServicesScans(cfg.Name, ns, servicesScans)
			if err != nil {
				fmt.Println("Failed to save Services scans") // TODO: log error
			}
		}
	}
}

// ScanNamespace return scans for jobs and services into specific Namespace for cluster
func (ks *KubeScanner) ScanNamespace(kubeClient *kubernetes.Clientset, namespace string) (servicesScans []*model.ServiceScan, jobsScans []*model.JobScan, err error) {
	// Get all pods
	pods, err := kubeClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		// TODO: вывод в лог ??? Или ловить уже выше и там выводить ошибку?
		return nil, nil, err
	}
	for _, pod := range pods.Items {
		// Get some pod logs
		switch pod.Status.Phase {
		case v1.PodRunning:
			serviceScan, err := ks.ScanServiceLog(kubeClient, &pod)
			if err != nil {
				fmt.Print("Error occured while getting pod logs -- ", pod.Name) // TODO: log error
			}
			servicesScans = append(servicesScans, serviceScan)
		case v1.PodFailed, v1.PodSucceeded:
			jobScan, err := ks.ScanJobLog(kubeClient, &pod)
			if err != nil {
				fmt.Print("Error occured while getting pod logs -- ", pod.Name) // TODO: log error
			}
			jobsScans = append(jobsScans, jobScan)
		}
	}
	return servicesScans, jobsScans, nil
}

func (ks *KubeScanner) ScanServiceLog(kubeClient *kubernetes.Clientset, pod *v1.Pod) (serviceScan *model.ServiceScan, err error) {
	serviceScan.ServiceName = pod.Name
	serviceScan.RestartsCount = 0                              // TODO: нужно выцеплять из состояния контейнера внутри пода.. Надо ли оно?
	serviceScan.Uptime = pod.CreationTimestamp.Sub(time.Now()) // TODO: is that correct?
	podLogOpts := &v1.PodLogOptions{}
	req := kubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, podLogOpts)
	podLogsStream, err := req.Stream(context.TODO())
	if err != nil {
		return nil, err
	}
	defer podLogsStream.Close() // TODO: нужно ли закрывать???
	scanner := bufio.NewScanner(podLogsStream)
	for scanner.Scan() {
		log := &model.CommonServiceLog{}
		err := json.Unmarshal(scanner.Bytes(), log)
		if err != nil {
			serviceScan.NoneJsonLinesCount++
		}
		switch *log.Level {
		case model.Trace, model.Debug, model.Info, model.Warning, model.Error, model.Fatal:
			serviceScan.LogTypeCountMap[*log.Level] += 1
		default:
			fmt.Println("Unknown log level -- ", *log.Level) // TODO: log warning
		}
		fmt.Println(*log.Level)
	}
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
		JobName:     pod.Name,
		Age:         pod.CreationTimestamp.Sub(time.Now()), // TODO: is that correct?
		FullLog:     sb.String(),
		GrepPattern: *ks.jobsRegexp,
		GrepLog:     matchedLogRows,
	}, nil
}
