package syncdaopq

import (
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncutil"
	"data-sync-tools-go/testhelper"
	"fmt"
	"log"
	"testing"
)

func TestQueuer_QueueOK(t *testing.T) {
	testName := syncutil.GetCallingName()
	testhelper.StartTest(testName)

	db, err := createAndVerifyDBConn(testDbUser, testDbPassword, testDbHost, testDbName, testDbPort)
	if err != nil {
		t.Error("Failed to connect to database: " + err.Error())
		return
	}
	closeQuietly := func() {
		err := db.Close()
		if err != nil {
			syncutil.Error("Quietly handling of db close error. Error: " + err.Error())
		}
	}
	defer closeQuietly()
	testhelper.RemoveSyncTablesIfNeeded(db)
	testhelper.CreateSyncTables(db)
	testhelper.SetupData(db, "profile3")

	nodeIDToQueue := "*node-spoke1"

	//syncNodeDao := syncdao.DefaultDaos.SyncNodeDao()

	sessionID := "my-session-id-1"

	var msgQueuer syncapi.MessageQueuing
	msgQueuer, err = newMessageQueuer(db)
	if err != nil {
		msg := "Failed to Create Msg Processor : " + err.Error()
		syncutil.Error(msg)
		t.Error(msg)
		return
	}
	recordsQueued, err := msgQueuer.Queue(sessionID, nodeIDToQueue)
	if err != nil {
		t.Error("Failed to get pair names: " + err.Error())
		fmt.Println("Error")
		return
	}
	var expectedCount int
	// TODO(doug4j@gmail.com): Double checkt that this is the correct expected result
	expectedCount = 5
	if recordsQueued != expectedCount {
		log.Printf("Actual Count: %v. Expected count: %v", recordsQueued, expectedCount)
		t.Error("Records queue count incorrect. Expected ")
	}

	sqlStr := `
update sync_peer_state set SessionBindId=null, ChangedByClient ='0'  where NodeId='*node-spoke1';
update sync_state set RecordHash='hash-2b' where RecordId='*record-2';
update sync_state set RecordHash='hash-3b' where RecordId='*record-3';`

	_, err = db.Exec(sqlStr)
	if err != nil {
		t.Error("Error changing a record: " + err.Error())
		return
	}

	recordsQueued, err = msgQueuer.Queue(sessionID, nodeIDToQueue)
	if err != nil {
		t.Error("Failed to get pair names: " + err.Error())
		fmt.Println("Error")
		return
	}

	expectedCount = 3
	if recordsQueued != expectedCount {
		log.Printf("Actual Count: %v. Expected %v", recordsQueued, expectedCount)
		t.Error("Records queue count incorrect")
	}

	testhelper.EndTest(testName)
}
