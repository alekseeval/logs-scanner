package kube

import "scan_project/internal/model"

type StorageI interface {
	ClusterDAOI
	ScansDAOI
}

type ClusterDAOI interface {
	kubeConfigDAOI
	namespaceDAOI
}

type ScansDAOI interface {
	jobsScanDAOI
	servicesScanDAOI
}

type kubeConfigDAOI interface {
	AddKubeConfig(kubeConfig *model.KubeConfig) (*model.KubeConfig, error)
	GetKubeConfigByName(kubeConfigName string) (*model.KubeConfig, error)
	EditKubeConfig(clusterName string, kubeconfig string) (*model.KubeConfig, error)
	DeleteKubeConfig(kubeConfigName string) error
	GetAllConfigs() ([]model.KubeConfig, error)
}

type namespaceDAOI interface {
	AddNamespaceToCubeConfig(kubeConfigName string, NamespaceName string) error
	DeleteNamespaceFromKubeconfig(clusterName string, namespaceName string) error
}

type jobsScanDAOI interface {
	GetJobsScans(configName string, namespace string) []model.JobScan
	UpdateJobsScans(configName string, namespace string, jobsScans []model.JobScan) error
}

type servicesScanDAOI interface {
	GetServicesScans(configName string, namespace string) []model.ServiceScan
	UpdateServicesScans(configName string, namespace string, servicesScans []model.ServiceScan) error
}
