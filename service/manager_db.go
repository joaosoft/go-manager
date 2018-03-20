package gomanager

import "database/sql"

type IDB interface {
	Get() *sql.DB
	Start() error
	Stop() error
	Started() bool
}

// DBConfig ...
type DBConfig struct {
	Driver     string `json:"driver"`
	DataSource string `json:"datasource"`
}

// NewDBConfig...
func NewDBConfig(driver, datasource string) *DBConfig {
	return &DBConfig{
		Driver:     driver,
		DataSource: datasource,
	}
}

// AddWeb ...
func (manager *GoManager) AddDB(key string, db IDB) error {
	manager.dbs[key] = db
	log.Infof("database %s added", key)

	return nil
}

// RemoveWeb ...
func (manager *GoManager) RemoveDB(key string) (IDB, error) {
	db := manager.dbs[key]

	delete(manager.dbs, key)
	log.Infof("database %s removed", key)

	return db, nil
}

// GetDB ...
func (manager *GoManager) GetDB(key string) IDB {
	if db, exists := manager.dbs[key]; exists {
		return db
	}
	log.Infof("database %s doesn't exist", key)
	return nil
}
