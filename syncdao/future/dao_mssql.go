package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/denisenkom/go-mssqldb"
)

type MsSqlDaosFactory struct {
	db          *sql.DB
	syncNodeDao SyncNodeDao
	syncPairDao SyncPairDao
}

func NewMsSqlDaosFactory(dbUser string, dbPassword string, dbServer string, dbName string, dbPort int) (*MsSqlDaosFactory, error) {
	log.Println("Using MS SQL Server Mode.")
	//user=%s password=%s dbname=%s sslmode=disable
	connString := fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%d", dbServer, dbName, dbUser, dbPassword, dbPort)

	//dbinfo := fmt.Sprintf("server=%s;user id=%s;password=%s", dbName dbUser, dbPassword, )
	db, err := sql.Open("mssql", connString)
	if err != nil {
		var nilMsSqlDaosFactory *MsSqlDaosFactory
		return nilMsSqlDaosFactory, err
	}
	err = db.Ping()
	if err != nil {
		var nilMsSqlDaosFactory *MsSqlDaosFactory
		return nilMsSqlDaosFactory, err
	}
	syncNodeDao := new(SyncNodeMsSqlDao)
	syncNodeDao.db = db
	syncPairDao := new(SyncPairMsSqlDao)
	syncPairDao.db = db
	factory := new(MsSqlDaosFactory)
	factory.db = db
	factory.syncNodeDao = syncNodeDao
	factory.syncPairDao = syncPairDao
	log.Println("Database Ready")
	return factory, nil
}

func (self MsSqlDaosFactory) SyncNodeDao() SyncNodeDao {
	return self.syncNodeDao
}

func (self MsSqlDaosFactory) SyncPairDao() SyncPairDao {
	return self.syncPairDao
}

func (self MsSqlDaosFactory) Close() {
	self.db.Close()
}
