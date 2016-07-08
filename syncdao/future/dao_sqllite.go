package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteDaosFactory struct {
	db          *sql.DB
	syncNodeDao SyncNodeDao
	syncPairDao SyncPairDao
}

func NewSqliteDaosFactory(dbUser string, dbPassword string, dbName string, dbPort int) (*SqliteDaosFactory, error) {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		var nilDaosFactory *SqliteDaosFactory
		return nilDaosFactory, err
	}

	if err != nil {
		var nilDaosFactory *SqliteDaosFactory
		return nilDaosFactory, err
	}
	err = db.Ping()
	if err != nil {
		var nilDaosFactory *SqliteDaosFactory
		return nilDaosFactory, err
	}
	syncNodeDao := new(SyncNodeSqliteDao)
	syncNodeDao.db = db
	syncPairDao := new(SyncPairSqliteDao)
	syncPairDao.db = db
	factory := new(SqliteDaosFactory)
	factory.db = db
	factory.syncNodeDao = syncNodeDao
	factory.syncPairDao = syncPairDao
	return factory, nil
}

func (self SqliteDaosFactory) SyncNodeDao() SyncNodeDao {
	return self.syncNodeDao
}

func (self SqliteDaosFactory) SyncPairDao() SyncPairDao {
	return self.syncPairDao
}

func (self SqliteDaosFactory) Close() {
	self.db.Close()
}
