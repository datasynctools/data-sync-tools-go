//Data Sync Server
package main

import (
	"database/sql"
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncdao"
	"data-sync-tools-go/syncdao/syncdaopq"
	"data-sync-tools-go/synchandler"
	"data-sync-tools-go/syncutil"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var dbType = flag.String("dbty", "postgressql", "The database to use: 'postgressql' or 'mssql'.")
var dbUser = flag.String("dbusr", "doug", "The name of the database database to use: 'postgressql', 'sqlite' or 'mssql'.")
var dbPass = flag.String("dbpw", "", "The database password.")
var dbServer = flag.String("dbsv", "localhost", "The database server.")
var dbName = flag.String("dbnm", "threads", "The database name.")
var dbPort = flag.Int("dbpt", 0, "The database port.")
var httpAddress = flag.String("httpad", "", "The http address.")
var httpPort = flag.Int("httppt", 8080, "The http port.")

func main() {

	flag.Parse()

	var router *mux.Router
	var hasPassword string
	if dbPass == nil || *dbPass == "" {
		hasPassword = "false"
	} else {
		hasPassword = "true"
	}
	log.Printf("dbty:'%s', dbusr:'%s', dbnm='%s', dbpt=%d, passwordSupplied=%s, dbpw='%s'", *dbType, *dbUser,
		*dbName, *dbPort, hasPassword, *dbPass)

	var db *sql.DB
	var dbFactory syncdao.DaosFactory
	var err error
	//Setup Database
	if *dbType == "postgressql" {
		db, err = createAndVerifyDBConn(*dbUser, *dbPass, *dbServer, *dbName, *dbPort)
		if err != nil {
			log.Fatal("Bad argument for 'dbty'")
			return
		}
		handlers := synchandler.Handlers{
			Repository: syncapi.Repository{
				DataRepo:   syncdaopq.NewDataRepository(db),
				ConfigRepo: syncdaopq.NewConfigRepository(db),
			},
		}
		router = synchandler.NewRouter(handlers)

		dbFactory, err = syncdaopq.NewPostgresSQLDaosFactory(*dbUser, *dbPass, *dbServer, *dbName, *dbPort)
		//} else if *dbType == "sqlite" {
		//	dbFactory, err = NewSqliteDaosFactory(*dbUser, *dbPass, *dbName, *dbPort)
		//} else if *dbType == "mssql" {
		//} else if *dbType == "mssql" {
		//dbFactory, err = NewPostgresSqlDaosFactory("doug", "postgres", "threads")
		//	dbFactory, err = NewMsSqlDaosFactory(*dbUser, *dbPass, *dbServer, *dbName, *dbPort)
		//dbFactory, err = NewMsSqlDaosFactory("doug", "postgres", "threads")
	} else {
		log.Fatal("Bad argument for 'dbty'")
		return
	}
	if err != nil {
		panic(err)
	}
	syncdao.DefaultDaos = dbFactory
	if syncdao.DefaultDaos == nil {
		log.Println("No Data object Found")
	}
	defer dbFactory.Close()
	var address = fmt.Sprintf("%v:%v", *httpAddress, *httpPort)
	log.Println("Listening on " + address)
	log.Fatal(http.ListenAndServe(address, router))
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
