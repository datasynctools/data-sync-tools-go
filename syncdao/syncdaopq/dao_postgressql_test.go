package syncdaopq

// BUG(doug4j@gmail.com): Move this test file into a separate package, see https://medium.com/@matryer/5-simple-tips-and-tricks-for-writing-unit-tests-in-golang-619653f90742#.o8nxf7z53
import (
	"data-sync-tools-go/syncdao"
	"data-sync-tools-go/testhelper"
	"fmt"
	"log"
	"testing"
	"time"
)

var (
	testDbUser     = testhelper.TestDbUser
	testDbPassword = testhelper.TestDbPassword
	testDbName     = testhelper.TestDbName
	testDbHost     = testhelper.TestDbHost
	testDbPort     = testhelper.TestDbPort
	dbFactory      syncdao.DaosFactory
	//db             *sql.DB
)

func setupDatabaseObj() {
	teardownDatabaseObj()
	dbFactory, err := NewPostgresSQLDaosFactory(testDbUser, testDbPassword, testDbHost, testDbName, testDbPort)
	if err != nil {
		panic(err)
	}
	syncdao.DefaultDaos = dbFactory
	db := dbFactory.SQLDb()
	testhelper.SetupData(db, "profile3")
}

func teardownDatabaseObj() {
	if syncdao.DefaultDaos != nil {
		syncdao.DefaultDaos.Close()
	}
}

func TestSyncPairPostgresSqlDao_GetPairByNames(t *testing.T) {
	log.Println("")
	log.Println("")
	log.Println("***START TestSyncPairPostgresSqlDao_GetPairByNames")
	log.Println("")

	setupDatabaseObj()
	defer teardownDatabaseObj()

	requestingNodeName := "A"
	toPairWithNodeName := "Z"

	syncPairDao := syncdao.DefaultDaos.SyncPairDao()

	syncPair, err := syncPairDao.GetPairByNames(requestingNodeName, toPairWithNodeName)
	if err != nil {
		t.Error("Failed to get pair names: " + err.Error())
		fmt.Println("Error")
		return
	}
	if syncPair.PairID != "*pair-1" {
		t.Error("Pair Id incorrect")
	}
	if syncPair.PairName != "A <-> Z" {
		t.Error("Pair Name incorrect")
	}
	if syncPair.MaxSesDurValue != 10 {
		t.Error("Max Session Duration Value incorrect")
	}
	if syncPair.MaxSesDurUnit != "minutes" {
		t.Error("Max Session Duration Unit incorrect")
	}
	if syncPair.SyncDataTransForm != "json:V1" {
		t.Error("SyncDataTransForm incorrect")
	}
	if syncPair.SyncMsgTransForm != "json:V1" {
		t.Error("SyncMsgTransForm incorrect")
	}
	if syncPair.SyncMsgSecPol != "none" {
		t.Error("SyncMsgSecPol incorrect")
	}
	if syncPair.SyncConflictURI != "none" {
		t.Error("SyncConflictUri incorrect")
	}
	var nilString string
	if syncPair.SyncSessionID != nilString {
		t.Error("SyncSessionId incorrect")
	}
	if syncPair.SyncSessionState != "Inactive" {
		t.Error("SyncSessionState incorrect, expected 'inactive' but was '" + syncPair.SyncSessionState + "'")
	}
	var nilTime time.Time
	if syncPair.SyncSessionStart != nilTime {
		t.Error("SyncSessionStart incorrect")
	}
	if syncPair.RecordCreated.Equal(nilTime) {
		t.Error("RecordCreated incorrect")
	}
	log.Println("")
	log.Println("***END TestSyncPairPostgresSqlDao_GetPairByNames")
	log.Println("")
	log.Println("")
}

func TestSyncPairPostgresSqlDao_GetNodePairItem(t *testing.T) {
	log.Println("")
	log.Println("")
	log.Println("***START TestSyncPairPostgresSqlDao_GetNodePairItem")
	log.Println("")

	setupDatabaseObj()
	defer teardownDatabaseObj()

	requestingNodeName := "A"
	//toPairWithNodeName := "B"

	syncPairDao := syncdao.DefaultDaos.SyncPairDao()

	item, err := syncPairDao.GetNodePairItem("*pair-1", requestingNodeName)
	if err != nil {
		t.Error("Failed to get pair names: " + err.Error())
		return
	}
	if item.NodeID != "*node-spoke1" {
		t.Error("Node Id incorrect")
	}
	if item.NodeName != "A" {
		t.Error("Node Name incorrect")
	}
	log.Println("")
	log.Println("***END TestSyncPairPostgresSqlDao_GetNodePairItem")
	log.Println("")
	log.Println("")
}

func TestSyncPairPostgresSqlDao_GetEntityPairItem(t *testing.T) {
	log.Println("")
	log.Println("")
	log.Println("***START TestSyncPairPostgresSqlDao_GetEntityPairItem")
	log.Println("")

	setupDatabaseObj()
	defer teardownDatabaseObj()

	requestingNodeName := "A"
	//toPairWithNodeName := "B"

	syncPairDao := syncdao.DefaultDaos.SyncPairDao()

	items, err := syncPairDao.GetEntityPairItem("*pair-1", requestingNodeName)
	if err != nil {
		t.Error("Failed to get entities: " + err.Error())
		return
	}
	//log.Println("items", items)
	if len(items) != 6 {
		t.Error("Missing items")
	}
	log.Println("")
	log.Println("***END TestSyncPairPostgresSqlDao_GetEntityPairItem")
	log.Println("")
	log.Println("")
}

func TestSyncPairPostgresSqlDao_CreateSyncSession_CloseSyncSession(t *testing.T) {
	log.Println("")
	log.Println("")
	log.Println("***START TestSyncPairPostgresSqlDao_CreateSyncSession_CloseSyncSession")
	log.Println("")

	setupDatabaseObj()
	defer teardownDatabaseObj()

	syncPairDao := syncdao.DefaultDaos.SyncPairDao()

	pairID := "*pair-1"
	sessionID := "*session-id-1"

	item := syncdao.CreateSyncSessionRequest{
		PairID:    pairID,
		SessionID: sessionID,
	}
	var createSyncSessionDaoResult syncdao.CreateSyncSessionDaoResult
	createSyncSessionDaoResult, err := syncPairDao.CreateSyncSession(item)
	if err != nil {
		t.Error("Failed to get pair names: " + err.Error())
		fmt.Println("Error")
		return
	}
	if createSyncSessionDaoResult.Result != "OK" {
		t.Error("Result incorrect")
	}
	if createSyncSessionDaoResult.ActualSessionID != sessionID {
		t.Error("ActualSessionId incorrect")
	}

	var syncPair syncdao.SyncPair
	requestingNodeName := "A"
	toPairWithNodeName := "Z"
	syncPair, err = syncPairDao.GetPairByNames(requestingNodeName, toPairWithNodeName)
	if err != nil {
		t.Error("Failed to get pair names: " + err.Error())
		fmt.Println("Error")
		return
	}
	if syncPair.SyncSessionID != sessionID {
		t.Error("SyncSessionId incorrect")
	}
	if syncPair.SyncSessionState != "Initializing" {
		t.Error("SyncSessionState incorrect, expected 'inactive' but was '" + syncPair.SyncSessionState + "'")
	}
	var nilTime time.Time
	if syncPair.SyncSessionStart.Equal(nilTime) {
		t.Error("SyncSessionStart incorrect")
	}
	if syncPair.RecordCreated.Equal(nilTime) {
		t.Error("RecordCreated incorrect")
	}

	closeItem := syncdao.CloseSyncSessionRequest{
		PairID:    pairID,
		SessionID: sessionID,
	}
	var closeSyncSessionDaoResult syncdao.CloseSyncSessionDaoResult
	closeSyncSessionDaoResult, err = syncPairDao.CloseSyncSession(closeItem)
	if err != nil {
		t.Error("Failed to get pair names: " + err.Error())
		fmt.Println("Error")
		return
	}
	if closeSyncSessionDaoResult.Result != "OK" {
		t.Error("Result incorrect")
	}
	if closeSyncSessionDaoResult.ActualSessionID != sessionID {
		t.Error("ActualSessionId incorrect")
	}

	syncPair, err = syncPairDao.GetPairByNames(requestingNodeName, toPairWithNodeName)
	if err != nil {
		t.Error("Failed to get pair names: " + err.Error())
		fmt.Println("Error")
		return
	}
	var nilString string
	if syncPair.SyncSessionID != nilString {
		t.Error("SyncSessionId incorrect")
	}
	if syncPair.SyncSessionState != "Inactive" {
		t.Error("SyncSessionState incorrect, expected 'inactive' but was '" + syncPair.SyncSessionState + "'")
	}
	if syncPair.SyncSessionStart != nilTime {
		t.Error("SyncSessionStart incorrect")
	}
	if syncPair.RecordCreated.Equal(nilTime) {
		t.Error("RecordCreated incorrect")
	}
	log.Println("")
	log.Println("***END TestSyncPairPostgresSqlDao_CreateSyncSession_CloseSyncSession")
	log.Println("")
	log.Println("")
}
