package dal

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
	sd.jMutex.RLock()
	defer sd.jMutex.RUnlock()
	sd.logRequest([]string{clusterName, namespace}, "Get jobs scans", nil)
	scans := sd.jobsScans[daoKey{
		clusterName: clusterName,
		namespace:   namespace,
	}]
	if scans != nil {
		return scans
	} else {
		return make([]model.JobScan, 0)
	}
}

func (sd *ScansDao) ClearJobsScans(clusterName string, namespace string) {
	sd.jMutex.Lock()
	defer sd.jMutex.Unlock()
	sd.logRequest([]string{clusterName, namespace}, "Clear jobs scans", nil)
	key := daoKey{
		clusterName: clusterName,
		namespace:   namespace,
	}
	delete(sd.jobsScans, key)
}

func (sd *ScansDao) UpdateJobsScans(clusterName string, namespace string, jobsScans []model.JobScan) error {
	sd.jMutex.Lock()
	defer sd.jMutex.Unlock()
	rowsNum := len(jobsScans)
	sd.logRequest([]string{clusterName, namespace}, "Update jobs scans", &rowsNum)
	key := daoKey{
		clusterName: clusterName,
		namespace:   namespace,
	}
	sd.jobsScans[key] = jobsScans
	return nil
}

func (sd *ScansDao) GetServicesScans(clusterName string, namespace string) []model.ServiceScan {
	sd.sMutex.RLock()
	defer sd.sMutex.RUnlock()
	sd.logRequest([]string{clusterName, namespace}, "Get services scans", nil)
	key := daoKey{
		clusterName: clusterName,
		namespace:   namespace,
	}
	scans := sd.servicesScans[key]
	if scans != nil {
		return scans
	} else {
		return make([]model.ServiceScan, 0)
	}
}

func (sd *ScansDao) ClearServicesScans(clusterName string, namespace string) {
	sd.sMutex.Lock()
	defer sd.sMutex.Unlock()
	sd.logRequest([]string{clusterName, namespace}, "Clear services scans", nil)
	key := daoKey{
		clusterName: clusterName,
		namespace:   namespace,
	}
	delete(sd.servicesScans, key)
}

func (sd *ScansDao) UpdateServicesScans(clusterName string, namespace string, servicesScans []model.ServiceScan) error {
	sd.sMutex.Lock()
	defer sd.sMutex.Unlock()
	rowsNum := len(servicesScans)
	sd.logRequest([]string{clusterName, namespace}, "Update services scans", &rowsNum)
	key := daoKey{
		clusterName: clusterName,
		namespace:   namespace,
	}
	sd.servicesScans[key] = servicesScans
	return nil
}

func (sd *ScansDao) logRequest(params []string, msg string, rowsNum *int) {
	entry := sd.logger.WithField("params", params)
	if rowsNum != nil {
		entry = entry.WithField("rows", *rowsNum)
	}
	entry.Debug(msg)
}
