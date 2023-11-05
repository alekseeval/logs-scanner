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
	AddCluster(kubeConfig *model.Cluster) (*model.Cluster, error)
	GetClusterByName(kubeConfigName string) (*model.Cluster, error)
	EditClusterConfig(clusterName string, kubeconfig string) (*model.Cluster, error)
	DeleteCluster(kubeConfigName string) error
	GetAllClusters() ([]model.Cluster, error)
}

type namespaceDAOI interface {
	AddNamespaceToCluster(kubeConfigName string, NamespaceName string) error
	DeleteNamespaceFromCluster(clusterName string, namespaceName string) error
}

type jobsScanDAOI interface {
	GetJobsScans(clusterName string, namespace string) []model.JobScan
	UpdateJobsScans(clusterName string, namespace string, jobsScans []model.JobScan) error
}

type servicesScanDAOI interface {
	GetServicesScans(clusterName string, namespace string) []model.ServiceScan
	UpdateServicesScans(clusterName string, namespace string, servicesScans []model.ServiceScan) error
}
