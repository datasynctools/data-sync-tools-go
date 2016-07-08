package syncdaopq

import (
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncmsg"
	"data-sync-tools-go/syncutil"
	"data-sync-tools-go/testhelper"
	"testing"

	"github.com/golang/protobuf/proto"
)

func TestProcessor_ProcessorFastBatchOK(t *testing.T) {
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
	testhelper.SetupData(db, "profile5")

	sessionID := "my-session-id-1"
	nodeIDToProcess := "*node-spoke1"

	entitiesByPluralName := map[string]syncapi.EntityNameItem{
		"As":       syncapi.EntityNameItem{SingularName: "A", PluralName: "As"},
		"Bs":       syncapi.EntityNameItem{SingularName: "B", PluralName: "Bs"},
		"Contacts": syncapi.EntityNameItem{SingularName: "Contact", PluralName: "Contacts"},
	}

	var msgProcessor syncapi.MessageProcessing
	msgProcessor, err = newMessageProcessor(sessionID, nodeIDToProcess, entitiesByPluralName, db)
	if err != nil {
		msg := "Failed to Create Msg Processor : " + err.Error()
		syncutil.Error(msg)
		t.Error(msg)
		return
	}

	creator := syncmsg.NewCreator()

	lastContact4 := testhelper.Contact{"0934A378-DEDB-4207-B99C-DD0D61DC59BC", creator.FormatTimeFromString("1987-05-28 00:00:00.000"), "Mindy", 5, 1.0, "Johnson", 2}
	lastContactSyncPackage4, err := testhelper.CreateRecordAndSupport(lastContact4, false, syncmsg.SentSyncStateEnum_PersistedStandardSentToPeer, "")
	if err != nil {
		syncutil.Fatal("Marshaling error: ", err)
		return //This return will never get called as the above panics
	}

	contact4 := testhelper.Contact{"0934A378-DEDB-4207-B99C-DD0D61DC59BC", creator.FormatTimeFromString("1987-05-28 00:00:00.000"), "Mindy", 5, 2.0, "Johnson", 1}
	contactSyncPackage4, err := testhelper.CreateRecordAndSupport(contact4, true, syncmsg.SentSyncStateEnum_PersistedStandardSentToPeer, lastContactSyncPackage4.RecordSha256Hex)
	if err != nil {
		syncutil.Fatal("Marshaling error: ", err)
		return //This return will never get called as the above panics
	}

	syncutil.Debug("contactSyncPackage4.RecordSha256Hex:", contactSyncPackage4.RecordSha256Hex, "contactSyncPackage4.PeerLastKnownHash", contactSyncPackage4.PeerLastKnownHash)

	contact6 := testhelper.Contact{"151EFA13-A3AD-4C18-A2CE-9D66D0AED112", creator.FormatTimeFromString("1988-01-23 00:00:00.000"), "Jill", 5, 3.0, "Anderson", 1}
	contactSyncPackage6, err := testhelper.CreateRecordAndSupport(contact6, true, syncmsg.SentSyncStateEnum_PersistedFirstTimeSentToPeer, "")
	if err != nil {
		syncutil.Fatal("Marshaling error: ", err)
		return //This return will never get called as the above panics
	}

	request1 := &syncmsg.ProtoSyncEntityMessageRequest{
		IsDelete:          proto.Bool(false),
		TransactionBindId: proto.String("some-bind-id"),
		Items: []*syncmsg.ProtoSyncDataMessagesRequest{
			&syncmsg.ProtoSyncDataMessagesRequest{
				EntityPluralName: proto.String("Contacts"),
				Msgs: []*syncmsg.ProtoSyncDataMessageRequest{
					&syncmsg.ProtoSyncDataMessageRequest{
						RecordId:          proto.String(contact4.ContactID),
						RecordHash:        proto.String(contactSyncPackage4.RecordSha256Hex),
						LastKnownPeerHash: proto.String(contactSyncPackage4.PeerLastKnownHash),
						SentSyncState:     syncmsg.SentSyncStateEnum_PersistedStandardSentToPeer.Enum(),
						RecordBytesSize:   proto.Uint32(123),
						RecordData:        contactSyncPackage4.RecordBytes,
					},
					&syncmsg.ProtoSyncDataMessageRequest{
						RecordId:          proto.String(contact6.ContactID),
						RecordHash:        proto.String(contactSyncPackage6.RecordSha256Hex),
						LastKnownPeerHash: nil,
						SentSyncState:     syncmsg.SentSyncStateEnum_PersistedFirstTimeSentToPeer.Enum(),
						RecordBytesSize:   proto.Uint32(123),
						RecordData:        contactSyncPackage6.RecordBytes,
					},
				},
			},
		},
	}
	syncutil.Debug("recordId", "0934A378-DEDB-4207-B99C-DD0D61DC59BC", "contactSyncPackage4.RecordSha256Hex", contactSyncPackage4.RecordSha256Hex, "lastContactSyncPackage4.RecordSha256Hex", lastContactSyncPackage4.RecordSha256Hex, "")

	var response1 *syncmsg.ProtoSyncEntityMessageResponse
	response1 = msgProcessor.Process(request1)
	syncutil.Debug("response1", response1)

	expectedMsg := "All records are fast batch"
	if *response1.ResultMsg != expectedMsg {
		t.Errorf("Result not expected '%v'. it's '%v'", expectedMsg, response1)
	}

	testhelper.EndTest(testName)
}
