package dao

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

type kubeConfigView struct {
	Config     string         `db:"config_str"`
	Name       string         `db:"name"`
	NameSpaces pq.StringArray `db:"namespaces"`
}

func (kcv *kubeConfigView) convertToKubeConfig() *model.KubeConfig {
	return &model.KubeConfig{
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
	return &PostgresDB{
		db:     db,
		logger: logger,
	}, err
}

// AddKubeConfig saves model.KubeConfig, except Namespaces field
//
//	To save Namespaces should be used PostgresDB.AddNamespaceToCubeConfig method
func (p *PostgresDB) AddKubeConfig(kubeConfig *model.KubeConfig) (*model.KubeConfig, error) {
	queryRow := `SELECT * FROM create_kubeconfig($1, $2)`
	queryParams := []interface{}{kubeConfig.Name, kubeConfig.Config}
	row := p.db.QueryRowx(queryRow, queryParams...)
	var kcv kubeConfigView
	err := row.StructScan(&kcv)
	p.logDBRequest(queryRow, queryParams)
	return kcv.convertToKubeConfig(), err
}

func (p *PostgresDB) GetKubeConfigByName(kubeConfigName string) (*model.KubeConfig, error) {
	queryRow := `SELECT * FROM get_kubeconfig_by_name($1)`
	queryParams := []interface{}{kubeConfigName}
	row := p.db.QueryRowx(queryRow, queryParams...)
	var kcv kubeConfigView
	err := row.StructScan(&kcv)
	p.logDBRequest(queryRow, queryParams)
	return kcv.convertToKubeConfig(), err
}

// EditKubeConfig change kubeconfig access string only
//
//	To change namespaces list where AddNamespaceToCubeConfig and DeleteNamespaceFromKubeconfig methods
func (p *PostgresDB) EditKubeConfig(clusterName string, kubeconfig string) (*model.KubeConfig, error) {
	queryRow := `SELECT * FROM edit_kubeconfig($1, $2)`
	queryParams := []interface{}{clusterName, kubeconfig}
	row := p.db.QueryRowx(queryRow, queryParams...)
	var kcv kubeConfigView
	err := row.StructScan(&kcv)
	p.logDBRequest(queryRow, queryParams)
	return kcv.convertToKubeConfig(), err
}

func (p *PostgresDB) DeleteKubeConfig(kubeConfigName string) error {
	queryRow := `SELECT * FROM delete_kubeconfig($1)`
	queryParams := []interface{}{kubeConfigName}
	_, err := p.db.Exec(queryRow, queryParams...)
	p.logDBRequest(queryRow, queryParams)
	return err
}

func (p *PostgresDB) GetAllConfigs() ([]model.KubeConfig, error) {
	queryRow := `SELECT * FROM get_kubeconfigs()`
	rows, err := p.db.Queryx(queryRow)
	if err != nil {
		return nil, err
	}
	allConfigs := make([]model.KubeConfig, 0)
	p.logDBRequest(queryRow, nil)
	for rows.Next() {
		var kcv kubeConfigView
		err = rows.StructScan(&kcv)
		if err != nil {
			return nil, err
		}
		allConfigs = append(allConfigs, *kcv.convertToKubeConfig())
	}
	return allConfigs, err
}

func (p *PostgresDB) AddNamespaceToCubeConfig(kubeConfigName string, namespaceName string) error {
	queryRow := `SELECT * FROM add_namespace($1, $2)`
	queryParams := []interface{}{namespaceName, kubeConfigName}
	_, err := p.db.Exec(queryRow, queryParams...)
	p.logDBRequest(queryRow, queryParams)
	return err
}

func (p *PostgresDB) DeleteNamespaceFromKubeconfig(clusterName string, namespaceName string) error {
	queryRow := `SELECT * FROM delete_namespace($1, $2)`
	queryParams := []interface{}{namespaceName, clusterName}
	_, err := p.db.Exec(queryRow, queryParams...)
	p.logDBRequest(queryRow, queryParams)
	return err
}

// logDBRequest write to log information about request. Method uses slog entry from PostgresDB struct
func (p *PostgresDB) logDBRequest(queryRow string, queryParams interface{}) {
	p.logger.WithFields(logrus.Fields{
		"params": queryParams,
		"query":  queryRow,
	}).Info("db query")
}
