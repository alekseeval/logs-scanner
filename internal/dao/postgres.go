package dao

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"scan_project/configuration"
	"scan_project/internal/model"
)

// PostgresDB is struct which implements kube.ClusterDAOI interface and provides access to PostgresSQL DB
type PostgresDB struct {
	db     *sqlx.DB
	logger *logrus.Entry
}

// NewPostgresDB initialize PostgresDB struct
//
//	Error can be occurred by initial ping to db
func NewPostgresDB(config *configuration.Config, logger *logrus.Entry) (*PostgresDB, error) {
	dbConfig := config.System.Postgres
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s connect_timeout=%d",
		dbConfig.User, dbConfig.Password, dbConfig.Ip, dbConfig.Port, dbConfig.DbName, dbConfig.Timeout)
	db, err := sqlx.Connect("postgres", connStr)
	return &PostgresDB{
		db:     db,
		logger: logger,
	}, err
}

// AddKubeConfig saves model.KubeConfig, except NameSpaces field
//
//	To save NameSpaces should be used PostgresDB.AddNamespaceToCubeConfig method
func (p *PostgresDB) AddKubeConfig(kubeConfig *model.KubeConfig) (*model.KubeConfig, error) {
	queryRow := `SELECT * FROM tools_api.create_kubeconfig($1, $2)`
	queryParams := []interface{}{kubeConfig.Name, kubeConfig.Config}
	row := p.db.QueryRowx(queryRow, queryParams...)
	var kc model.KubeConfig
	err := row.StructScan(&kc)
	p.logDBRequest(queryRow, queryParams)
	return &kc, err
}

func (p *PostgresDB) GetKubeConfigByName(kubeConfigName string) (*model.KubeConfig, error) {
	queryRow := `SELECT * FROM tools_api.get_kubeconfig_by_name($1)`
	queryParams := []interface{}{kubeConfigName}
	row := p.db.QueryRowx(queryRow, queryParams...)
	var kc model.KubeConfig
	err := row.StructScan(&kc)
	p.logDBRequest(queryRow, queryParams)
	return &kc, err
}

// EditKubeConfig change kubeconfig access string only
//
//	To change namespaces list where AddNamespaceToCubeConfig and DeleteNamespaceFromKubeconfig methods
func (p *PostgresDB) EditKubeConfig(kubeConfig *model.KubeConfig) (*model.KubeConfig, error) {
	queryRow := `SELECT * FROM tools_api.edit_kubeconfig($1, $2)`
	queryParams := []interface{}{kubeConfig.Name, kubeConfig.Config}
	row := p.db.QueryRowx(queryRow, queryParams...)
	var kc model.KubeConfig
	err := row.StructScan(&kc)
	p.logDBRequest(queryRow, queryParams)
	return &kc, err
}

func (p *PostgresDB) DeleteKubeConfig(kubeConfigName string) error {
	queryRow := `SELECT * FROM tools_api.delete_kubeconfig($1)`
	queryParams := []interface{}{kubeConfigName}
	_, err := p.db.Exec(queryRow, queryParams...)
	p.logDBRequest(queryRow, queryParams)
	return err
}

func (p *PostgresDB) GetAllConfigs() ([]model.KubeConfig, error) {
	queryRow := `SELECT * FROM tools_api.get_kubeconfigs()`
	rows, err := p.db.Queryx(queryRow)
	allConfigs := make([]model.KubeConfig, 0)
	for rows.Next() {
		var ck model.KubeConfig
		err = rows.StructScan(&ck)
		if err != nil {
			return nil, err
		}
		allConfigs = append(allConfigs, ck)
	}
	p.logDBRequest(queryRow, nil)
	return allConfigs, err
}

func (p *PostgresDB) AddNamespaceToCubeConfig(kubeConfigName string, namespaceName string) error {
	queryRow := `SELECT * FROM tools_api.add_namespace($1, $2)`
	queryParams := []interface{}{namespaceName, kubeConfigName}
	_, err := p.db.Exec(queryRow, queryParams...)
	p.logDBRequest(queryRow, queryParams)
	return err
}

func (p *PostgresDB) DeleteNamespaceFromKubeconfig(namespaceName string) error {
	queryRow := `SELECT * FROM tools_api.delete_namespace($1, $2)`
	queryParams := []interface{}{namespaceName}
	_, err := p.db.Exec(queryRow, queryParams...)
	p.logDBRequest(queryRow, queryParams)
	return err
}

// logDBRequest write to log information about request. Method uses slog entry from PostgresDB struct
func (p *PostgresDB) logDBRequest(queryRow string, queryParams interface{}) {
	p.logger.WithFields(logrus.Fields{
		"params": fmt.Sprintf("%v", queryParams),
		"query":  queryRow,
	}).Info("db query")
}
