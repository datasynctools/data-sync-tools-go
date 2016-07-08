//Package syncdaopq implements core syncdao interfaces for postgressql
//(http://www.postgresql.org/).
package syncdaopq

import (
	"database/sql"
	"data-sync-tools-go/syncdao"
	"data-sync-tools-go/syncutil"
	"fmt"
	//"log"
)

//PostgresSQLDaosFactory creates the data access objects for data synchronization to postgressql. It implements syncdao.DaosFactory.
type PostgresSQLDaosFactory struct {
	db          *sql.DB
	syncNodeDao syncdao.SyncNodeDao
	syncPairDao syncdao.SyncPairDao
}

//NewPostgresSQLDaosFactory creates a PostgresSqlDaosFactory instance.
func NewPostgresSQLDaosFactory(dbUser string, dbPassword string, dbHost string, dbName string, dbPort int) (*PostgresSQLDaosFactory, error) {
	syncutil.Info("Using Postgressql Mode.")
	db, err := createAndVerifyDBConn(dbUser, dbPassword, dbHost, dbName, dbPort)
	if err != nil {
		syncutil.Error(err, ". Cannot establish a database connection.")
		var nilDaosFactory *PostgresSQLDaosFactory
		return nilDaosFactory, err
	}
	syncNodeDao := new(SyncNodePostgresSQLDao)
	syncNodeDao.db = db
	syncPairDao := new(SyncPairPostgresSQLDao)
	syncPairDao.db = db
	factory := new(PostgresSQLDaosFactory)
	factory.db = db
	factory.syncNodeDao = syncNodeDao
	factory.syncPairDao = syncPairDao
	//log.Println("Database Ready")
	return factory, nil
}

func createAndVerifyDBConn(dbUser string, dbPassword string, dbHost string, dbName string, dbPort int) (*sql.DB, error) {
	dbinfo := fmt.Sprintf("user=%s password=%s host=%s dbname=%s port=%v sslmode=disable", dbUser, dbPassword, dbHost, dbName, dbPort)
	//dbinfo := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable", "doug", "postgres", "localhost", "threads")
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		syncutil.Error(err, ". Using connection:", dbinfo)
		return db, err
	}
	err = db.Ping()
	if err != nil {
		syncutil.Error("Cannot ping db. error:", err, ". Using connection:", dbinfo)
		return db, err
	}
	return db, nil
}

//SQLDb supplies the underlying database instance.
func (factory PostgresSQLDaosFactory) SQLDb() *sql.DB {
	return factory.db
}

//SyncNodeDao provides the syncdao.SyncNodeDao instance.
func (factory PostgresSQLDaosFactory) SyncNodeDao() syncdao.SyncNodeDao {
	return factory.syncNodeDao
}

//SyncPairDao provides the syncdao.SyncPairDao instance.
func (factory PostgresSQLDaosFactory) SyncPairDao() syncdao.SyncPairDao {
	return factory.syncPairDao
}

//Close closes the underlying database connection.
func (factory PostgresSQLDaosFactory) Close() {
	err := factory.db.Close()
	if err != nil {
		syncutil.Error("Quietly handling of close error. Error: " + err.Error())
	}
}

//SQLDateFormat is the date format for postgressql
var SQLDateFormat = "2006-01-02 15:04:05.000"

/*
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
*/
