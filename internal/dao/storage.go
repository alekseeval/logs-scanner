package dao

import "scan_project/internal/kube"

type Storage struct {
	kube.ClusterDAOI
	kube.ScansDAOI
}

func NewStorage(clusterDAO kube.ClusterDAOI, scansDAO kube.ScansDAOI) *Storage {
	return &Storage{
		ClusterDAOI: clusterDAO,
		ScansDAOI:   scansDAO,
	}
}
