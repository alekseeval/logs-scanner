package kube

import "scan_project/internal/model"

type ClusterDAOI interface {
	KubeConfigDAOI
	NamespaceDAOI
}

type ScansDAOI interface {
	JobsScanDAOI
	ServicesScanDAOI
}

type KubeConfigDAOI interface {
	AddKubeConfig(kubeConfig *model.KubeConfig) (*model.KubeConfig, error)
	GetKubeConfigByName(kubeConfigName string) (*model.KubeConfig, error)
	EditKubeConfig(kubeConfig *model.KubeConfig) (*model.KubeConfig, error)
	DeleteKubeConfig(kubeConfigName string) error
	GetAllConfigs() ([]model.KubeConfig, error)
}

type NamespaceDAOI interface {
	AddNamespaceToCubeConfig(kubeConfigName string, NamespaceName string) error
	DeleteNamespaceFromKubeconfig(namespaceName string) error
}

type JobsScanDAOI interface {
	GetJobsScans(configName string, namespace string) []model.JobScan
	UpdateJobsScans(configName string, namespace string, jobsScans []model.JobScan) error
}

type ServicesScanDAOI interface {
	GetServicesScans(configName string, namespace string) []model.ServiceScan
	UpdateServicesScans(configName string, namespace string, servicesScans []model.ServiceScan) error
}
