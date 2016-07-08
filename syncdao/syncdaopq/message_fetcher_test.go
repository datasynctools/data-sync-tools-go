package syncdaopq

import (
	"bytes"
	"database/sql"
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncmsg"
	"data-sync-tools-go/syncutil"
	"data-sync-tools-go/testhelper"
	"encoding/hex"
	"fmt"
	"testing"
	"text/template"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/twinj/uuid"
)

func setupSyncFetcherTestFetcherOKMsgSpread(db *sql.DB, sessionBindID string) {
	// teardownDatabaseObj()
	// dbFactory, err := NewPostgresSQLDaosFactory(testDbUser, testDbPassword, testDbHost, testDbName, testDbPort)
	// if err != nil {
	// 	panic(err)
	// }
	// syncdao.DefaultDaos = dbFactory

	now := time.Now()
	timeString := now.Format(SQLDateFormat)

	//Custom
	sqlTemplate := template.Must(template.New("sqlTemplate").Parse(sqlTemplateFetcherOKMsgSpread))
	var substitutedSQLBytes bytes.Buffer
	err := sqlTemplate.Execute(&substitutedSQLBytes, struct {
		NowString     string
		SessionBindID string
	}{timeString, sessionBindID})
	if err != nil {
		panic(err)
	}
	substitutedSQL := substitutedSQLBytes.String()
	// sqlDbable, ok := syncdao.DefaultDaos.(syncdao.SQLDbable)
	// if !ok {
	// 	syncutil.Error("dao factory does not contain sql SQLDbable interface")
	// 	return
	// }
	// db := sqlDbable.SQLDb()

	testhelper.RemoveSyncTablesIfNeeded(db)
	testhelper.CreateSyncTables(db)
	testhelper.AddProfile4SampleSyncData(db)
	//syncutil.Info(substitutedSQL)
	result, err := db.Exec(substitutedSQL)
	if err != nil {
		msg := "Error getting results from database. Error:%s"
		syncutil.Fatal(msg, err.Error())
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		msg := "Error getting results from database. Error:%s"
		syncutil.Fatal(msg, err.Error())
	}

	creator := syncmsg.NewCreator()
	recordID := "record 06"
	record := &syncmsg.ProtoRecord{
		Fields: []*syncmsg.ProtoField{
			//Important: these field names need to be in alphabetical order
			creator.CreateStringProtoField("contactId", recordID),
			//creator.CreateTimeProtoField("dateOfBirth", creator.FormatTimeFromString("1994-07-10 00:00:00.000")),
			creator.CreateStringProtoField("firstName", "Jennifer"),
			creator.CreateInt64ProtoField("heightFt", 5),
			creator.CreateDoubleProtoField("heightInch", 4.0),
			creator.CreateStringProtoField("lastName", "Smith"),
			creator.CreateInt64ProtoField("preferredHeight", 2),
		},
	}
	recordBytes, err := proto.Marshal(record)
	if err != nil {
		syncutil.Fatal("marshaling error: ", err)
		return
		//return err
	}
	//recordSha256Hex := Hash256Bytes(recordBytes)
	//syncutil.Debug("record7Sha256Hex:", record7Sha256Hex)
	recordAsHex := hex.EncodeToString(recordBytes)
	// recordSQL := `
	// UPDATE sync_state SET RecordData=$1 WHERE RecordId=$2;
	// `
	result, err = db.Exec("UPDATE sync_state SET RecordData='" + recordAsHex + "' WHERE RecordId='" + recordID + "';")
	//result, err := db.Exec("UPDATE sync_state SET RecordData='" + recordAsHex + "' WHERE RecordId='" + recordID + "';")
	//result, err := db.Exec(recordSQL, recordAsHex, recordID)
	// recordSQL := `
	// UPDATE sync_state SET RecordData=$1 WHERE EntitySingularName=$2 AND RecordId=$3;
	// `
	// result, err := db.Exec(recordSQL, recordAsHex, "B", recordID)
	if err != nil {
		msg := "Error getting results from database. Error:%s"
		syncutil.Fatal(msg, err.Error())
		//log.Println("Error inserting to database, inputData=", item)
		panic(err)
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		msg := "Error getting results from database. Error:%s"
		syncutil.Fatal(msg, err.Error())
		panic(err)
	}
	if rowsAffected != 1 {
		syncutil.Info("rowsAffected:", rowsAffected)
		panic("sample data record was not written")
	}
	syncutil.Info("Created setupSyncFetcherTestFetcherOKMsgSpread sample data.")
}

func TestSyncFetcher_TestBinary(t *testing.T) {
	creator := syncmsg.NewCreator()
	recordID := "record 06"
	expectedRecord := &syncmsg.ProtoRecord{
		Fields: []*syncmsg.ProtoField{
			//Important: these field names need to be in alphabetical order
			creator.CreateStringProtoField("contactId", recordID),
			//creator.CreateTimeProtoField("dateOfBirth", creator.FormatTimeFromString("1994-07-10 00:00:00.000")),
			creator.CreateStringProtoField("firstName", "Jennifer"),
			creator.CreateInt64ProtoField("heightFt", 5),
			creator.CreateDoubleProtoField("heightInch", 4.0),
			creator.CreateStringProtoField("lastName", "Smith"),
			creator.CreateInt64ProtoField("preferredHeight", 2),
		},
	}
	recordBytes, err := proto.Marshal(expectedRecord)
	if err != nil {
		syncutil.Fatal("marshaling error: ", err)
		return
		//return err
	}
	//recordSha256Hex := Hash256Bytes(recordBytes)
	//syncutil.Debug("record7Sha256Hex:", record7Sha256Hex)
	recordSha256Hex := testhelper.Hash256Bytes(recordBytes)
	recordBytesAsHex := hex.EncodeToString(recordBytes)
	recordSQL := `
		INSERT into sync_state (EntitySingularName, RecordId, DataVersionName, RecordHash, RecordData, RecordBytesSize)
		values('A', $1, 'Demo Model 1', $2, $3, $4);
`
	/*
			recordSQL := `
		UPDATE sync_state SET RecordData=$1 WHERE EntitySingularName=$2 AND RecordId=$3;
		`
	*/
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
	testhelper.AddProfile4SampleSyncData(db)

	//fmt.Sprintf("%v", len(recordBytes))
	_, err = db.Exec(recordSQL, recordID, recordSha256Hex, recordBytesAsHex, len(recordBytes))
	if err != nil {
		msg := "Error getting results from database. Error:"
		syncutil.Fatal(msg, err.Error())
		//log.Println("Error inserting to database, inputData=", item)
		panic(err)
	}

	//syncutil.Info(substitutedSQL)
	result := db.QueryRow("select RecordData from sync_state where RecordId=$1;", recordID)

	var binData string
	//var binData []byte
	err = result.Scan(&binData)
	if err != nil {
		msg := "Error getting results from database. Error:"
		syncutil.Fatal(msg, err.Error())
	}
	//syncutil.Info(binData)
	bin, err := hex.DecodeString(binData)
	if err != nil {
		msg := "Error decoding hex bytes from database. Error:"
		syncutil.Fatal(msg, err.Error())
	}
	actualRecord := &syncmsg.ProtoRecord{}
	err = proto.Unmarshal(bin, actualRecord)
	if err != nil {
		msg := "Error parsing actual record bytes. Error:"
		syncutil.Fatal(msg, err.Error())
		//log.Println("Error inserting to database, inputData=", item)
		panic(err)
	}
	//syncutil.Info(actualRecord)
	//syncutil.Info("Created setupSyncFetcherTestFetcherOKMsgSpread sample data. processed=", rowsAffected, ",recordHex=", recordHex)
}

func TestSyncFetcher_TestFetcherOKMsgSpread(t *testing.T) {
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
	sessionUUID := uuid.Formatter(uuid.NewV4(), uuid.CleanHyphen)
	//pairUUID := uuid.Formatter(uuid.NewV4(), uuid.CleanHyphen)
	//nodeUUID := uuid.Formatter(uuid.NewV4(), uuid.CleanHyphen)
	nodeID := "*node-hub"

	setupSyncFetcherTestFetcherOKMsgSpread(db, sessionUUID)

	//var fetcher syncapi.SyncMessagesFetcher
	fetcher, err := newMessageFetcher(sessionUUID, nodeID, db)
	if err != nil {
		syncutil.Error(err.Error())
		t.Error(err.Error())
	}
	/*
		fetcher := PostgresSQLSyncMessagesFetcher{
			SyncMessagesFetcherType: syncapi.SyncMessagesFetcherType{
				SessionID:         sessionUUID,
				PairID:            pairUUID,
				NodeID:            "*node-hub",
				MaxGroupBytesSize: 50,
				MaxMsgs:           1,
			},
			db: db,
		}
	*/
	var changeType = syncapi.ProcessSyncChangeEnumAddOrUpdate
	var answer *syncmsg.ProtoRequestSyncEntityMessageResponse
	var msgsCount int

	//msgs 1
	entities := []syncapi.EntityNameItem{
		syncapi.EntityNameItem{
			SingularName: "A",
			PluralName:   "A",
		},
		syncapi.EntityNameItem{
			SingularName: "B",
			PluralName:   "B",
		},
		syncapi.EntityNameItem{
			SingularName: "C",
			PluralName:   "C",
		},
	}
	answer, err = fetcher.Fetch(entities, changeType)
	if err != nil {
		syncutil.Error(err.Error())
		t.Error(err.Error())
	}
	//syncutil.Debug(entityMsgRequest)
	if len(answer.Request.Items) != 1 {
		msg := fmt.Sprintf("Should have found exactly 1 item, instead found %v items.", len(answer.Request.Items))
		t.Error(msg)
		return
	}
	msgsCount = (len(answer.Request.Items[0].Msgs))
	if msgsCount != 1 {
		msg := fmt.Sprintf("Should have found 1 msg, instead found %v msgs", msgsCount)
		t.Error(msg)
		return
	}
	//msgs 2
	answer, err = fetcher.Fetch(entities, changeType)
	if err != nil {
		syncutil.Error(err.Error())
		t.Error(err.Error())
	}
	if len(answer.Request.Items) != 2 {
		msg := fmt.Sprintf("Should have found exactly 2 item, instead found %v items.", len(answer.Request.Items))
		t.Error(msg)
		return
	}
	msgsCount = (len(answer.Request.Items[0].Msgs) + len(answer.Request.Items[1].Msgs))
	if msgsCount != 4 {
		msg := fmt.Sprintf("Should have found 4 msg, instead found %v msgs", msgsCount)
		t.Error(msg)
		return
	}
	//msgs 3
	answer, err = fetcher.Fetch(entities, changeType)
	if err != nil {
		syncutil.Error(err.Error())
		t.Error(err.Error())
	}
	if len(answer.Request.Items) != 1 {
		msg := fmt.Sprintf("Should have found exactly 1 item, instead found %v items.", len(answer.Request.Items))
		t.Error(msg)
		return
	}

	//Sample a record to make sure the response msgs parse OK.
	actualMsg3Record := &syncmsg.ProtoRecord{}
	//syncutil.Info(entityMsgRequest)
	//syncutil.Debug(entityMsgRequest.Items[0].Msgs[0].RecordData)
	err = proto.Unmarshal(answer.Request.Items[0].Msgs[0].RecordData, actualMsg3Record)
	if err != nil {
		t.Errorf("Error decoding response: %s", err.Error())
	}

	msgsCount = len(answer.Request.Items[0].Msgs)
	if msgsCount != 2 {
		msg := fmt.Sprintf("Should have found 2 msg, instead found %v msgs", msgsCount)
		t.Error(msg)
		return
	}
	//msgs 4
	answer, err = fetcher.Fetch(entities, changeType)
	if err != nil {
		syncutil.Error(err.Error())
		t.Error(err.Error())
	}
	if len(answer.Request.Items) != 0 {
		msg := fmt.Sprintf("Should have found exactly 0 item, instead found %v items.", len(answer.Request.Items))
		t.Error(msg)
		return
	}
	//msgs 5
	changeType = syncapi.ProcessSyncChangeEnumDelete
	answer, err = fetcher.Fetch(entities, changeType)
	if err != nil {
		syncutil.Error(err.Error())
		t.Error(err.Error())
	}
	if len(answer.Request.Items) != 1 {
		msg := fmt.Sprintf("Should have found exactly 1 item, instead found %v items.", len(answer.Request.Items))
		t.Error(msg)
		return
	}
	msgsCount = len(answer.Request.Items[0].Msgs)
	if msgsCount != 1 {
		msg := fmt.Sprintf("Should have found 1 msg, instead found %v msgs", msgsCount)
		t.Error(msg)
		return
	}
	//msgs 6
	changeType = syncapi.ProcessSyncChangeEnumDelete
	answer, err = fetcher.Fetch(entities, changeType)
	if err != nil {
		syncutil.Error(err.Error())
		t.Error(err.Error())
	}
	if len(answer.Request.Items) != 1 {
		msg := fmt.Sprintf("Should have found exactly 1 item, instead found %v items.", len(answer.Request.Items))
		t.Error(msg)
		return
	}
	msgsCount = len(answer.Request.Items[0].Msgs)
	if msgsCount != 1 {
		msg := fmt.Sprintf("Should have found 1 msg, instead found %v msgs", msgsCount)
		t.Error(msg)
		return
	}
	//msgs 7
	answer, err = fetcher.Fetch(entities, changeType)
	if err != nil {
		syncutil.Error(err.Error())
		t.Error(err.Error())
	}
	if len(answer.Request.Items) != 0 {
		msg := fmt.Sprintf("Should have found exactly 0 item, instead found %v items.", len(answer.Request.Items))
		t.Error(msg)
		return
	}

	testhelper.EndTest(testName)
}

func TestSyncFetcher_TestFetcherOKNoMsgs(t *testing.T) {
	testName := syncutil.GetCallingName()
	testhelper.StartTest(testName)
	syncutil.Info("\n\nFINISH ME\n\n")
	testhelper.EndTest(testName)
}

var sqlTemplateFetcherOKMsgSpread = `
--sync_state-1
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('A','record 01','Demo Model 1','some-hash-A','` + dummyData + `',50,'false','{{.NowString}}');
--sync_state-2
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('A','record 02','Demo Model 1','some-hash-B','` + dummyData + `',10,'false','{{.NowString}}');
--sync_state-3
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('A','record 03','Demo Model 1','some-hash-C','` + dummyData + `',10,'false','{{.NowString}}');
--sync_state-4
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('A','record 04','Demo Model 1','some-hash-D','` + dummyData + `',29,'false','{{.NowString}}');
--sync_state-5
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('B','record 05','Demo Model 1','some-hash-E','` + dummyData + `',2,'false','{{.NowString}}');
--sync_state-6
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('B','record 06','Demo Model 1','some-hash-F','` + dummyData + `',20,'false','{{.NowString}}');
--sync_state-7
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('B','record 07','Demo Model 1','some-hash-G','` + dummyData + `',30,'false','{{.NowString}}');
--sync_state-8
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('A','record 08','Demo Model 1','some-hash-H','` + dummyData + `',50,'true','{{.NowString}}');
--sync_state-9
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('B','record 09','Demo Model 1','some-hash-I','` + dummyData + `',20,'true','{{.NowString}}');

--sync_peer_state
--//nodeToFetchId add/update
--SyncPeerState-1 recordId1
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','A','record 01','{{.SessionBindID}}',NULL,NULL,'some-hash-A',NULL,'false',1,'true',50,'{{.NowString}}');
--SyncPeerState-2 record2
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','A','record 02','{{.SessionBindID}}',NULL,NULL,'some-hash-B',NULL,'false',1,'true',10,'{{.NowString}}');
--SyncPeerState-3 record3
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','A','record 03','{{.SessionBindID}}',NULL,NULL,'some-hash-C',NULL,'false',1,'true',10,'{{.NowString}}');
--SyncPeerState-4 record4
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','A','record 04','{{.SessionBindID}}',NULL,NULL,'some-hash-D',NULL,'false',1,'true',29,'{{.NowString}}');
--SyncPeerState-5 record5
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','B','record 05','{{.SessionBindID}}',NULL,NULL,'some-hash-E',NULL,'false',1,'true',2,'{{.NowString}}');
--SyncPeerState-6 record6
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','B','record 06','{{.SessionBindID}}',NULL,NULL,'some-hash-F',NULL,'false',1,'true',20,'{{.NowString}}');
--SyncPeerState-7 record7
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','B','record 07','{{.SessionBindID}}',NULL,NULL,'some-hash-G',NULL,'false',1,'true',30,'{{.NowString}}');
--nodeToFetchId delete
--SyncPeerState-8 record8
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,IsDelete,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','A','record 08','{{.SessionBindID}}',NULL,NULL,'some-hash-H',NULL,'false','true',1,'true',50,'{{.NowString}}');
--SyncPeerState-9 record9
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,IsDelete,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','B','record 09','{{.SessionBindID}}',NULL,NULL,'some-hash-I',NULL,'false','true',1,'true',20,'{{.NowString}}');
--nodeToIgnoreId add/update
--SyncPeerState-10 record1
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-spoke2','A','record 01','{{.SessionBindID}}',NULL,NULL,'some-hash-A',NULL,'false',1,'true',50,'{{.NowString}}');
--SyncPeerState-11 recordId6
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-spoke2','B','record 06','{{.SessionBindID}}',NULL,NULL,'some-hash-F',NULL,'false',1,'true',20,'{{.NowString}}');
--nodeToIgnoreId delete
--SyncPeerState-12 recordId8
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,IsDelete,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-spoke2','A','record 08','{{.SessionBindID}}',NULL,NULL,'some-hash-H',NULL,'false','true',1,'true',50,'{{.NowString}}');
--SyncPeerState-13 recordId9
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,IsDelete,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-spoke2','B','record 09','{{.SessionBindID}}',NULL,NULL,'some-hash-I',NULL,'false','true',1,'true',20,'{{.NowString}}');
`
