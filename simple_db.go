package manager

import (
	"database/sql"
	"github.com/joaosoft/logger"

	"sync"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	_ "github.com/lib/pq"              // postgres driver
)

// SimpleDB ...
type SimpleDB struct {
	*sql.DB
	logger logger.ILogger
	config  *DBConfig
	started bool
}

// NewSimpleDB ...
func (manager *Manager) NewSimpleDB(config *DBConfig) IDB {
	return &SimpleDB{
		config: config,
		logger: manager.logger,
	}
}

// Get ...
func (db *SimpleDB) Get() *sql.DB {
	return db.DB
}

// Start ...
func (db *SimpleDB) Start(wg *sync.WaitGroup) error {
	if wg == nil {
		wg = &sync.WaitGroup{}
		wg.Add(1)
	}

	defer wg.Done()

	if db.started {
		return nil
	}

	if conn, err := db.config.Connect(); err != nil {
		return err
	} else {
		db.DB = conn
		db.started = true
	}

	return nil
}

// Stop ...
func (db *SimpleDB) Stop(wg *sync.WaitGroup) error {
	if wg == nil {
		wg = &sync.WaitGroup{}
		wg.Add(1)
	}

	defer wg.Done()

	if !db.started {
		return nil
	}

	if err := db.Close(); err != nil {
		return err
	}
	db.started = false

	return nil
}

// Started ...
func (db *SimpleDB) Started() bool {
	return db.started
}
