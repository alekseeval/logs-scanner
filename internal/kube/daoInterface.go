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
	AddCluster(cluster *model.Cluster) (*model.Cluster, error)
	GetClusterByName(clusterName string) (*model.Cluster, error)
	EditClusterConfig(clusterName string, kubeConfig string) (*model.Cluster, error)
	DeleteCluster(clusterName string) error
	GetAllClusters() ([]model.Cluster, error)
}

type namespaceDAOI interface {
	AddNamespaceToCluster(kubeConfigName string, NamespaceName string) error
	DeleteNamespaceFromCluster(clusterName string, namespaceName string) error
}

type jobsScanDAOI interface {
	GetJobsScans(clusterName string, namespace string) []model.JobScan
	ClearJobsScans(clusterName string, namespace string)
	UpdateJobsScans(clusterName string, namespace string, jobsScans []model.JobScan) error
}

type servicesScanDAOI interface {
	GetServicesScans(clusterName string, namespace string) []model.ServiceScan
	ClearServicesScans(clusterName string, namespace string)
	UpdateServicesScans(clusterName string, namespace string, servicesScans []model.ServiceScan) error
}
