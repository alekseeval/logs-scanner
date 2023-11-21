package dao

import (
	"github.com/sirupsen/logrus"
	"scan_project/internal/model"
	"sync"
)

type ScansDao struct {
	jobsScans     map[daoKey][]model.JobScan
	jMutex        sync.RWMutex
	servicesScans map[daoKey][]model.ServiceScan
	sMutex        sync.RWMutex
	logger        *logrus.Entry
}

func NewScansDao(logger *logrus.Entry) ScansDao {
	return ScansDao{
		jobsScans:     make(map[daoKey][]model.JobScan),
		jMutex:        sync.RWMutex{},
		servicesScans: make(map[daoKey][]model.ServiceScan),
		sMutex:        sync.RWMutex{},
		logger:        logger,
	}
}

type daoKey struct {
	clusterName string
	namespace   string
}

func (sd *ScansDao) GetJobsScans(clusterName string, namespace string) []model.JobScan {
	sd.logger.
		WithField("params", []string{clusterName, namespace}).
		Debug("Get jobs scans")
	sd.jMutex.RLock()
	scans := sd.jobsScans[daoKey{
		clusterName: clusterName,
		namespace:   namespace,
	}]
	sd.jMutex.RUnlock()
	if scans != nil {
		return scans
	} else {
		return make([]model.JobScan, 0)
	}
}

func (sd *ScansDao) UpdateJobsScans(clusterName string, namespace string, jobsScans []model.JobScan) error {
	sd.logger.
		WithField("params", []string{clusterName, namespace}).
		WithField("rows", len(jobsScans)).
		Debug("Update jobs scans")
	key := daoKey{
		clusterName: clusterName,
		namespace:   namespace,
	}
	sd.jMutex.Lock()
	sd.jobsScans[key] = jobsScans
	sd.jMutex.Unlock()
	return nil
}

func (sd *ScansDao) GetServicesScans(clusterName string, namespace string) []model.ServiceScan {
	sd.logger.
		WithField("params", []string{clusterName, namespace}).
		Debug("Get services scans")
	sd.sMutex.RLock()
	scans := sd.servicesScans[daoKey{
		clusterName: clusterName,
		namespace:   namespace,
	}]
	sd.sMutex.RUnlock()
	if scans != nil {
		return scans
	} else {
		return make([]model.ServiceScan, 0)
	}
}

func (sd *ScansDao) UpdateServicesScans(clusterName string, namespace string, servicesScans []model.ServiceScan) error {
	sd.logger.
		WithField("params", []string{clusterName, namespace}).
		WithField("rows", len(servicesScans)).
		Debug("Update services scans")
	key := daoKey{
		clusterName: clusterName,
		namespace:   namespace,
	}
	sd.sMutex.Lock()
	sd.servicesScans[key] = servicesScans
	sd.sMutex.Unlock()
	return nil
}
