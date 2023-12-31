package dao

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"scan_project/configuration"
	"scan_project/internal/model"
	"strconv"
	"time"
)

// PostgresDB is struct which implements kube.ClusterDAOI interface and provides access to PostgresSQL DB
type PostgresDB struct {
	db     *sqlx.DB
	logger *logrus.Entry
}

type clusterView struct {
	Config     string         `db:"config_str"`
	Name       string         `db:"name"`
	NameSpaces pq.StringArray `db:"namespaces"`
}

func (kcv *clusterView) convertToCluster() *model.Cluster {
	return &model.Cluster{
		Config:     kcv.Config,
		Name:       kcv.Name,
		Namespaces: kcv.NameSpaces,
	}
}

// NewPostgresDB initialize PostgresDB struct
//
//	Error can be occurred by initial ping to db
func NewPostgresDB(config *configuration.Config, logger *logrus.Entry) (*PostgresDB, error) {
	dbConfig := config.System.Postgres
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s connect_timeout=%d",
		dbConfig.User, dbConfig.Password, dbConfig.Ip, dbConfig.Port, dbConfig.DbName, dbConfig.Timeout)
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Duration(config.System.Postgres.Timeout) * time.Second)
	return &PostgresDB{
		db:     db,
		logger: logger,
	}, err
}

// AddCluster saves model.Cluster, except Namespaces field
//
//	To save Namespaces should be used PostgresDB.AddNamespaceToCluster method
func (p *PostgresDB) AddCluster(cluster *model.Cluster) (*model.Cluster, error) {
	queryRow := `SELECT * FROM create_cluster($1, $2)`
	queryParams := []interface{}{cluster.Name, cluster.Config}
	row := p.db.QueryRowx(queryRow, queryParams...)
	var kcv clusterView
	err := row.StructScan(&kcv)
	p.logDBRequest(queryRow, queryParams)
	return kcv.convertToCluster(), p.convertDbErrorToInternal(err)
}

func (p *PostgresDB) GetClusterByName(clusterName string) (*model.Cluster, error) {
	queryRow := `SELECT * FROM get_cluster_by_name($1)`
	queryParams := []interface{}{clusterName}
	row := p.db.QueryRowx(queryRow, queryParams...)
	var kcv clusterView
	err := row.StructScan(&kcv)
	p.logDBRequest(queryRow, queryParams)
	return kcv.convertToCluster(), p.convertDbErrorToInternal(err)
}

// EditClusterConfig change cluster config only
//
//	To change namespaces list where AddNamespaceToCluster and DeleteNamespaceFromCluster methods
func (p *PostgresDB) EditClusterConfig(clusterName string, clusterConfig string) (*model.Cluster, error) {
	queryRow := `SELECT * FROM edit_cluster($1, $2)`
	queryParams := []interface{}{clusterName, clusterConfig}
	p.logDBRequest(queryRow, queryParams)
	row := p.db.QueryRowx(queryRow, queryParams...)
	var kcv clusterView
	err := row.StructScan(&kcv)
	if err != nil {
		return nil, p.convertDbErrorToInternal(err)
	}
	return kcv.convertToCluster(), nil
}

func (p *PostgresDB) DeleteCluster(clusterName string) error {
	queryRow := `SELECT * FROM delete_cluster($1)`
	queryParams := []interface{}{clusterName}
	_, err := p.db.Exec(queryRow, queryParams...)
	p.logDBRequest(queryRow, queryParams)
	return p.convertDbErrorToInternal(err)
}

func (p *PostgresDB) GetAllClusters() ([]model.Cluster, error) {
	queryRow := `SELECT * FROM get_clusters()`
	rows, err := p.db.Queryx(queryRow)
	if err != nil {
		return nil, err
	}
	allConfigs := make([]model.Cluster, 0)
	p.logDBRequest(queryRow, nil)
	for rows.Next() {
		var kcv clusterView
		err = rows.StructScan(&kcv)
		if err != nil {
			return nil, err
		}
		allConfigs = append(allConfigs, *kcv.convertToCluster())
	}
	return allConfigs, p.convertDbErrorToInternal(err)
}

func (p *PostgresDB) AddNamespaceToCluster(clusterName string, namespaceName string) error {
	queryRow := `SELECT * FROM add_namespace($1, $2)`
	queryParams := []interface{}{namespaceName, clusterName}
	_, err := p.db.Exec(queryRow, queryParams...)
	p.logDBRequest(queryRow, queryParams)
	return p.convertDbErrorToInternal(err)
}

func (p *PostgresDB) DeleteNamespaceFromCluster(clusterName string, namespaceName string) error {
	queryRow := `SELECT * FROM delete_namespace($1, $2)`
	queryParams := []interface{}{namespaceName, clusterName}
	_, err := p.db.Exec(queryRow, queryParams...)
	p.logDBRequest(queryRow, queryParams)
	return p.convertDbErrorToInternal(err)
}

// logDBRequest write to log information about request. Method uses slog entry from PostgresDB struct
func (p *PostgresDB) logDBRequest(queryRow string, queryParams interface{}) {
	p.logger.WithFields(logrus.Fields{
		"params": queryParams,
		"query":  queryRow,
	}).Info("db query")
}

func (p *PostgresDB) convertDbErrorToInternal(dbError error) error {
	if dbError == nil {
		return dbError
	}
	if pqErr, ok := dbError.(*pq.Error); ok {
		errCode, err := strconv.Atoi(pqErr.SQLState())
		if err == nil {
			return &model.ServerError{
				Code:        errCode,
				Description: dbError.Error(),
			}
		}
	}
	return &model.ServerError{
		Code:        model.UnknownDBError,
		Description: dbError.Error(),
	}
}
