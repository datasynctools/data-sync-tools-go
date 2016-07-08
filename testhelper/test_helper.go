//Package testhelper provides data and helper utilities for testing purposes.
package testhelper

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"data-sync-tools-go/syncmsg"
	"data-sync-tools-go/syncutil"
	"encoding/hex"
	"log"
	"runtime"
	"strings"
	"text/template"
	"time"

	proto "github.com/golang/protobuf/proto"
	//"time"
)

var (
	//TestDbUser is the default database user for testing
	TestDbUser = "doug"
	//TestDbPassword is the default database password for testing
	TestDbPassword = "postgres"
	//TestDbName is the default database name for testing
	TestDbName = "threads"
	//TestDbHost is the default database host for testing
	TestDbHost = "localhost"
	//TestDbPort is the default database port for testing
	TestDbPort = 5432
)

//RemoveSyncTablesIfNeeded removes the existing data and tables from the sync model (tables starting with sync_*)
//usingn the given database connection. If the tables do not exist than a console messages is reported to that effect.
func RemoveSyncTablesIfNeeded(db *sql.DB) {
	_, err := db.Exec(dropSyncModelTablesSQL)
	if err != nil {
		msg := "Error getting results from database. It may simply be that the tables do not exist, please confirm. Error:%s"
		syncutil.Info(msg, err.Error())
		//log.Println("Error inserting to database, inputData=", item)
	}
	syncutil.Info("Existing sync_* tables were dropped if available")
}

//AddProfile1SampleSyncData adds the following records counts to the following tables:
//	Table               Record Count  Comment
//	-----------------   ------------  ---------------------------------------------------------------------
//	sync_data_version         2       1 record is an orphaned node associated with a record
//	sync_data_entity          0
//	sync_state                0
//	sync_node                 4       Nodes A, B, C, & Z. Z is the Hub. The rest are spokes.
//	sync_data_field           0
//	sync_peer_state           0
//	sync_pair                 5       A<->Z, B<->Z, A<->B (Partial), C<->Z(Dup 1 of 2), C<->Z(Dup 2 of 2)
//	sync_pair_nodes           9       1 orphaned node associated with a record
func AddProfile1SampleSyncData(db *sql.DB) {
	_, err := db.Exec(profile1SQL)
	if err != nil {
		msg := "Error getting results from database. Error:%s"
		syncutil.Error(msg, err.Error())
		panic(err)
	}
	syncutil.Info("Created profile 1 sample data for sync_*")
}

//AddProfile2SampleSyncData adds the following records counts to the following tables:
//	|Header->                                                                                          Copied From Profile
//	Table               Record Count  Comment                                                                1
//	-----------------   ------------  --------------------------------------------------------------------- ---
//	sync_data_version         2       1 record is an orphaned node associated with a record                 Yes
//	sync_data_entity          6       5 records ()'Entity 1' - 'Entity 5') and 1 'Contact' record            -
//	sync_state                0                                                                              -
//	sync_node                 4       Nodes A, B, C, & Z. Z is the Hub. The rest are spokes.                Yes
//	sync_data_field           9       2 records for 'Entity 1'. 7 records are for 'Contact'.                 -
//	sync_peer_state           0                                                                              -
//	sync_pair                 5       A<->Z, B<->Z, A<->B (Partial), C<->Z(Dup 1 of 2), C<->Z(Dup 2 of 2)   Yes
//	sync_pair_nodes           9       1 orphaned node associated with a record                              Yes
func AddProfile2SampleSyncData(db *sql.DB) {
	AddProfile1SampleSyncData(db)
	_, err := db.Exec(profile2SQL)
	if err != nil {
		msg := "Error getting results from database. Error:%s"
		syncutil.Error(msg, err.Error())
		panic(err)
	}
	syncutil.Info("Created profile 2 sample data for sync_*")
}

//AddProfile3SampleSyncData adds the following records counts to the following tables:
//	|Header->                                                                                          Copied From Profile
//	Table               Record Count  Comment                                                                1   2
//	-----------------   ------------  --------------------------------------------------------------------- --- ---
//	sync_data_version         2       1 record is an orphaned node associated with a record                 Yes  -
//	sync_data_entity          6       5 records ()'Entity 1' - 'Entity 5') and 1 'Contact' record            -  Yes
//	sync_state                3       3 records for 'Entity 1'                                               -   -
//	sync_node                 4       Nodes A, B, C, & Z. Z is the Hub. The rest are spokes.                Yes  -
//	sync_data_field           9       2 records for 'Entity 1'. 7 records are for 'Contact'.                 -  Yes
//	sync_peer_state           3       3 records for 'Entity 1'                                               -   -
//	sync_pair                 5       A<->Z, B<->Z, A<->B (Partial), C<->Z(Dup 1 of 2), C<->Z(Dup 2 of 2)   Yes  -
//	sync_pair_nodes           9       1 orphaned node associated with a record                              Yes  -
func AddProfile3SampleSyncData(db *sql.DB) {
	AddProfile2SampleSyncData(db)
	_, err := db.Exec(profile3SQL)
	if err != nil {
		msg := "Error creating profile 3 data. Error:"
		syncutil.Error(msg, err.Error())
		panic(err)
	}
	syncutil.Info("Created profile 3 sample data for sync_*")
}

//AddProfile4SampleSyncData adds the following records counts to the following tables:
//	|Header->                                                                                          Copied From Profile
//	Table               Record Count  Comment                                                                1
//	-----------------   ------------  --------------------------------------------------------------------- ---
//	sync_data_version         2       1 record is an orphaned node associated with a record                 Yes
//	sync_data_entity          1       2 records (Entity 'A' and 'B')                                         -
//	sync_state                3       3 records for 'Entity 1'                                               -
//	sync_node                 4       Nodes A, B, C, & Z. Z is the Hub. The rest are spokes.                Yes
//	sync_data_field           9       2 records for 'Entity 1'. 7 records are for 'Contact'.                 -
//	sync_peer_state           3       3 records for 'Entity 1'                                               -
//	sync_pair                 5       A<->Z, B<->Z, A<->B (Partial), C<->Z(Dup 1 of 2), C<->Z(Dup 2 of 2)   Yes
//	sync_pair_nodes           9       1 orphaned node associated with a record                              Yes
func AddProfile4SampleSyncData(db *sql.DB) {
	AddProfile1SampleSyncData(db)
	_, err := db.Exec(profile4SQL)
	if err != nil {
		msg := "Error getting results from database. Error:%s"
		syncutil.Error(msg, err.Error())
		panic(err)
	}
	syncutil.Info("Created profile 4 sample data for sync_*")
}

//AddProfile5SampleSyncData adds the following records counts to the following tables:
//	|Header->                                                                                          Copied From Profile
//	Table               Record Count  Comment                                                                1   2
//	-----------------   ------------  --------------------------------------------------------------------- --- ---
//	sync_data_version         2       1 record is an orphaned node associated with a record                 Yes  -
//	sync_data_entity          6       5 records ()'Entity 1' - 'Entity 5') and 1 'Contact' record            -  Yes
//	sync_state                3       3 records for 'Entity 1'                                               -   -
//	sync_node                 4       Nodes A, B, C, & Z. Z is the Hub. The rest are spokes.                Yes  -
//	sync_data_field           9       2 records for 'Entity 1'. 7 records are for 'Contact'.                 -  Yes
//	sync_peer_state           3       6 records for 'Entity 1' (3 for 2 nodes)                               -   -
//	sync_pair                 5       A<->Z, B<->Z, A<->B (Partial), C<->Z(Dup 1 of 2), C<->Z(Dup 2 of 2)   Yes  -
//	sync_pair_nodes           9       1 orphaned node associated with a record                              Yes  -
func AddProfile5SampleSyncData(db *sql.DB) {
	AddProfile2SampleSyncData(db)
	_, err := db.Exec(profile5SQL)
	if err != nil {
		msg := "Error creating profile 5 data. Error:"
		syncutil.Error(msg, err.Error())
		panic(err)
	}
	syncutil.Info("Created profile 5 sample data for sync_*")
}

//CreateSyncTables adds the sync model (tables starting with sync_*) to the given database connection.
func CreateSyncTables(db *sql.DB) {
	_, err := db.Exec(createSyncModelTablesSQL)
	if err != nil {
		msg := "Error getting results from database. Error:%s"
		syncutil.Error(msg, err.Error())
		panic(err)
	}
	syncutil.Info("Created sync_* tables")
}

//SetupData sets up the database for as layed out in (base-dir)/src/sql/common/Linkmeup.org-Sample-App-Data.xlsx. It uses AddProfile3SampleSyncData and then adds to it with it's own data.
func SetupData(db *sql.DB, profile string) {

	RemoveSyncTablesIfNeeded(db)
	CreateSyncTables(db)
	//AddProfile2SampleSyncData(db)
	profile = strings.ToLower(profile)
	if profile == "profile3" {
		AddProfile3SampleSyncData(db)
	} else if profile == "profile5" {
		AddProfile5SampleSyncData(db)
	}

	sqlStr4 := `
	drop table if exists contacts;
`
	sqlStr5 := `
--1: Contacts
CREATE TABLE contacts (
	ContactId		varchar(36)		NOT NULL,
	DateOfBirth		timestamp		NULL,
	FirstName		varchar(36)		NULL,
	LastName		varchar(36)		NULL,
	HeightFt		int				NULL,
	HeightInch		real			NULL,
	PreferredHeight	smallint		NOT NULL	default(1),
	PRIMARY KEY (ContactId)
);
`
	creator := syncmsg.NewCreator()

	contact1 := Contact{"B6581A36-804D-45AC-B2E2-F6DA265AF7DE", creator.FormatTimeFromString("1990-04-29 00:00:00.000"), "Jack", 6, 1.0, "Smith", 1}
	contactSyncPackage1, err := CreateRecordAndSupport(contact1, false, syncmsg.SentSyncStateEnum_PersistedNeverSentToPeer, "")
	if err != nil {
		syncutil.Fatal("Marshaling error: ", err)
		return //This return will never get called as the above panics
	}

	previousContact2 := Contact{"911DD745-8916-41C4-9973-F8B38A501602", creator.FormatTimeFromString("1994-07-10 00:00:00.000"), "Henry", 5, 5.5, "Adins", 2}
	previousContactSyncPackage2, err := CreateRecordAndSupport(previousContact2, false, syncmsg.SentSyncStateEnum_PersistedStandardSentToPeer, "")
	if err != nil {
		syncutil.Fatal("Marshaling error: ", err)
		return //This return will never get called as the above panics
	}

	contact2 := Contact{"911DD745-8916-41C4-9973-F8B38A501602", creator.FormatTimeFromString("1994-07-10 00:00:00.000"), "Henry", 5, 5.5, "Adkins", 2}
	contactSyncPackage2, err := CreateRecordAndSupport(contact2, false, syncmsg.SentSyncStateEnum_PersistedStandardSentToPeer, previousContactSyncPackage2.RecordSha256Hex)
	if err != nil {
		syncutil.Fatal("Marshaling error: ", err)
		return //This return will never get called as the above panics
	}

	previousContact4 := Contact{"0934A378-DEDB-4207-B99C-DD0D61DC59BC", creator.FormatTimeFromString("1987-05-28 00:00:00.000"), "Mindy", 5, 1.0, "Johnson", 2}
	previousContactSyncPackage4, err := CreateRecordAndSupport(previousContact4, false, syncmsg.SentSyncStateEnum_PersistedStandardSentToPeer, "")
	if err != nil {
		syncutil.Fatal("Marshaling error: ", err)
		return //This return will never get called as the above panics
	}

	// syncutil.Debug("\n\npreviousContactSyncPackage4.Contact.ContactID: " + previousContactSyncPackage4.Contact.ContactID + "previousContactSyncPackage4.RecordSha256Hex: " + previousContactSyncPackage4.RecordSha256Hex + " and previousContactSyncPackage4.PeerLastKnownHash:" + previousContactSyncPackage4.PeerLastKnownHash + " \n\n")

	// contact4 := Contact{"0934A378-DEDB-4207-B99C-DD0D61DC59BC", creator.FormatTimeFromString("1987-05-28 00:00:00.000"), "Mindy", 5, 2.0, "Johnson", 1}
	// contactSyncPackage4, err := CreateRecordAndSupport(contact4, false, syncmsg.SentSyncStateEnum_PersistedStandardSentToPeer, previousContactSyncPackage4.RecordSha256Hex)
	// if err != nil {
	// 	syncutil.Fatal("Marshaling error: ", err)
	// 	return //This return will never get called as the above panics
	// }

	// contact4 := Contact{"0934A378-DEDB-4207-B99C-DD0D61DC59BC", creator.FormatTimeFromString("1987-05-28 00:00:00.000"), "Mindy", 5, 2.0, "Johnson", 1}
	// contactSyncPackage4, err := CreateRecordAndSupport(contact4, false, syncmsg.SentSyncStateEnum_PersistedStandardSentToPeer, previousContactSyncPackage4.RecordSha256Hex)
	// if err != nil {
	// 	syncutil.Fatal("Marshaling error: ", err)
	// 	return //This return will never get called as the above panics
	// }

	sqlStr6 := ""
	contactSyncPackages := []ContactSyncPackage{contactSyncPackage1, contactSyncPackage2, previousContactSyncPackage4}
	for index, contactSyncPackage := range contactSyncPackages {

		//contactDateStr := contactSyncPackage.Contact.DateOfBirthAsUTC
		//dateOfBirthStr := contactSyncPackage.Contact.DateOfBirthAsUTC.Format("2006-01-02 15:04:05.000")

		sqlTemplate := template.Must(template.New("sqlTemplateContactAndSyncState").Parse(sqlTemplateContactAndSyncState))
		var substitutedSQLBytes bytes.Buffer
		err := sqlTemplate.Execute(&substitutedSQLBytes, struct {
			ContactPackage ContactSyncPackage
			Index          int
		}{contactSyncPackage, index})
		if err != nil {
			panic(err)
		}
		sqlStr6 = sqlStr6 + substitutedSQLBytes.String()

		// sqlTemplate = template.Must(template.New("sqlTemplateContactSyncPeerState").Parse(sqlTemplateContactSyncPeerState))
		// var substitutedSQLBytes2 bytes.Buffer
		// err = sqlTemplate.Execute(&substitutedSQLBytes2, struct {
		// 	ContactPackage ContactSyncPackage
		// 	Index          int
		// }{contactSyncPackage, index})
		// if err != nil {
		// 	panic(err)
		// }
		// sqlStr6 = sqlStr6 + substitutedSQLBytes2.String()

	}

	strs := make([]string, 6)
	strs[0] = sqlStr4
	strs[1] = sqlStr5
	strs[2] = sqlStr6

	syncutil.Debug("sqlStr6:", sqlStr6)

	for _, sqlStr := range strs {
		//log.Printf("Running sql for index [%v]", index)
		_, err := db.Exec(sqlStr)
		if err != nil {
			msg := "Error getting results from database. Error:%s"
			syncutil.Fatal(msg, err.Error())
			//log.Println("Error inserting to database, inputData=", item)
			panic(err)
		}
		if err != nil {
			msg := "Error getting results from database. Error:"
			syncutil.Fatal(msg, err.Error())
			panic(err)
		}
		//log.Printf("Wrote %v database objects on index [%v]", affectedCount, index)
	}
}

//StartTest prints out a friendly name for tests at their start.
func StartTest(testName string) {
	log.Println("")
	log.Println("")
	pc, file, line, _ := runtime.Caller(1)

	fullPCName := runtime.FuncForPC(pc).Name()
	lastIndexOfPc := strings.LastIndex(fullPCName, "/") + 1
	justPcName := fullPCName[lastIndexOfPc:len(fullPCName)]

	lastIndexOfFile := strings.LastIndex(file, "/") + 1
	justFileName := file[lastIndexOfFile:len(file)]

	//log.Printf("INFO [%s:%d] [%s] %v", justFileName, line, justPcName, msg)
	log.Printf("***START [%s:%d] [%s] %v", justFileName, line, justPcName, testName)

	//log.Printf("***START " + testName + " [%s:%d] [%s] %v", justFileName, line, justPcName, msg))
	log.Println("")
}

//EndTest prints out a friendly name for tests at their end.
func EndTest(testName string) {
	log.Println("")
	pc, file, line, _ := runtime.Caller(1)

	fullPCName := runtime.FuncForPC(pc).Name()
	lastIndexOfPc := strings.LastIndex(fullPCName, "/") + 1
	justPcName := fullPCName[lastIndexOfPc:len(fullPCName)]

	lastIndexOfFile := strings.LastIndex(file, "/") + 1
	justFileName := file[lastIndexOfFile:len(file)]
	log.Printf("***END [%s:%d] [%s] %v", justFileName, line, justPcName, testName)
	//log.Println("***END " + testName)
	log.Println("")
	log.Println("")
}

//Hash256Bytes is a helper method that turns bytes into a sha256Hex string value
func Hash256Bytes(bytes []byte) string {
	hasher := sha256.New()
	hasher.Write(bytes)
	msgDigest := hasher.Sum(nil)
	sha256Hex := hex.EncodeToString(msgDigest)
	return sha256Hex
}

//CreateSyncData creates test data to send to the service for processing sync messages as laid out in (base-dir)/src/sql/common/Linkmeup.org-Sample-App-Data.xlsx.
func CreateSyncData() (*syncmsg.ProtoSyncEntityMessageRequest, error) {
	creator := syncmsg.NewCreator()
	record1 := &syncmsg.ProtoRecord{
		Fields: []*syncmsg.ProtoField{
			//Important: these field names need to be in alphabetical order
			//contactIdField, dateOfBirthField, firstNameField, heightFtField, heightInchField, lastNameField,
			//preferredHeightField,
			creator.CreateStringProtoField("contactId", "C441B7F6-D3E8-4E7F-9A3B-1BD4BAB6FA48"),
			creator.CreateTimeProtoField("dateOfBirth", creator.FormatTimeFromString("1990-04-29 00:00:00.000")),
			creator.CreateStringProtoField("firstName", "John"),
			creator.CreateInt64ProtoField("heightFt", 6),
			creator.CreateDoubleProtoField("heightInch", 2.5),
			creator.CreateStringProtoField("lastName", "Doe"),
			creator.CreateInt64ProtoField("preferredHeight", 1),
		},
	}
	if len(creator.Errors) != 0 {
		request := &syncmsg.ProtoSyncEntityMessageRequest{}
		return request, syncmsg.NewCreatorError(creator.Errors)
	}
	record2LastKnown := &syncmsg.ProtoRecord{
		Fields: []*syncmsg.ProtoField{
			//Important: these field names need to be in alphabetical order
			creator.CreateStringProtoField("contactId", "911DD745-8916-41C4-9973-F8B38A501602"),
			creator.CreateTimeProtoField("dateOfBirth", creator.FormatTimeFromString("1994-07-10 00:00:00.000")),
			creator.CreateStringProtoField("firstName", "Henry"),
			creator.CreateInt64ProtoField("heightFt", 5),
			creator.CreateDoubleProtoField("heightInch", 5.5),
			creator.CreateStringProtoField("lastName", "Adins"),
			creator.CreateInt64ProtoField("preferredHeight", 1),
		},
	}
	if len(creator.Errors) != 0 {
		request := &syncmsg.ProtoSyncEntityMessageRequest{}
		return request, syncmsg.NewCreatorError(creator.Errors)
	}
	record2 := &syncmsg.ProtoRecord{
		Fields: []*syncmsg.ProtoField{
			//Important: these field names need to be in alphabetical order
			creator.CreateStringProtoField("contactId", "911DD745-8916-41C4-9973-F8B38A501602"),
			creator.CreateTimeProtoField("dateOfBirth", creator.FormatTimeFromString("1994-07-10 00:00:00.000")),
			creator.CreateStringProtoField("firstName", "Henry"),
			creator.CreateInt64ProtoField("heightFt", 5),
			creator.CreateDoubleProtoField("heightInch", 5.5),
			creator.CreateStringProtoField("lastName", "Adkins"),
			creator.CreateInt64ProtoField("preferredHeight", 1),
		},
	}
	if len(creator.Errors) != 0 {
		request := &syncmsg.ProtoSyncEntityMessageRequest{}
		return request, syncmsg.NewCreatorError(creator.Errors)
	}

	/*
		record3 := &syncmsg.ProtoRecord{
			Fields: []*syncmsg.ProtoField{
				//Important: these field names need to be in alphabetical order
				createStringProtoField("contactId", "42FC5EBC-088D-4DC2-BDBD-D89FA170E5C6"),
				createTimeProtoField("dateOfBirth", createDateFromString("1992-01-23 00:00:00.000")),
				createStringProtoField("firstName", "John"),
				createInt64ProtoField("heightFt", 5),
				createDoubleProtoField("heightInch", 9),
				createStringProtoField("lastName", "Doe"),
				createInt64ProtoField("preferredHeight", 1),
			},
		}
	*/
	record1Bytes, err := proto.Marshal(record1)
	if err != nil {
		syncutil.Error("marshaling error: ", err)
		return &syncmsg.ProtoSyncEntityMessageRequest{}, err
	}
	record2Bytes, err := proto.Marshal(record2)
	if err != nil {
		syncutil.Error("marshaling error: ", err)
		return &syncmsg.ProtoSyncEntityMessageRequest{}, err
	}
	record2LastKnownBytes, err := proto.Marshal(record2LastKnown)
	if err != nil {
		syncutil.Error("marshaling error: ", err)
		return &syncmsg.ProtoSyncEntityMessageRequest{}, err
	}
	requestData := &syncmsg.ProtoSyncEntityMessageRequest{
		IsDelete:          proto.Bool(false),
		TransactionBindId: proto.String("0F16AEED-E4B4-483E-A7A3-CCABA831FE6E"),
		Items: []*syncmsg.ProtoSyncDataMessagesRequest{
			&syncmsg.ProtoSyncDataMessagesRequest{
				EntityPluralName: proto.String("Contacts"),
				Msgs: []*syncmsg.ProtoSyncDataMessageRequest{
					&syncmsg.ProtoSyncDataMessageRequest{
						RecordId:          proto.String("B6581A36-804D-45AC-B2E2-F6DA265AF7DE"),
						RecordHash:        proto.String(Hash256Bytes(record1Bytes)),
						LastKnownPeerHash: nil,
						SentSyncState:     syncmsg.SentSyncStateEnum(syncmsg.SentSyncStateEnum_PersistedFirstTimeSentToPeer).Enum(),
						RecordBytesSize:   proto.Uint32(uint32(len(record1Bytes))),
						RecordData:        record1Bytes,
					},
					&syncmsg.ProtoSyncDataMessageRequest{
						RecordId:          proto.String("911DD745-8916-41C4-9973-F8B38A501602"),
						RecordHash:        proto.String(Hash256Bytes(record2Bytes)),
						LastKnownPeerHash: proto.String(Hash256Bytes(record2LastKnownBytes)),
						SentSyncState:     syncmsg.SentSyncStateEnum(syncmsg.SentSyncStateEnum_PersistedStandardSentToPeer).Enum(),
						RecordBytesSize:   proto.Uint32(uint32(len(record2Bytes))),
						RecordData:        record2Bytes,
					},
				},
			},
			/*
				&ProtoSyncDataMessagesRequest{
					SyncEntityName: proto.String("Contact"),
					Msgs: []*ProtoSyncDataMessageRequest{
						&ProtoSyncDataMessageRequest{
							RecordId:          proto.String("RecordId2"),
							RecordHash:        proto.String("RecordHash2"),
							LastKnownPeerHash: nil,
							SentSyncState:     SentSyncStateEnum(SentSyncStateEnum_FirstTimeSentToPeer).Enum(),
							RecordBytesSize:   proto.Uint32(uint32(len(record2Bytes))),
							RecordData:        record2Bytes,
						},
					},
				},
			*/
		},
	}
	return requestData, nil
}

//CreateSyncDataAck creates test data to send to the service for processing sync messages as laid out in (base-dir)/src/sql/common/Linkmeup.org-Sample-App-Data.xlsx.
func CreateSyncDataAck() (*syncmsg.ProtoSyncEntityMessageResponse, error) {
	creator := syncmsg.NewCreator()
	record1 := &syncmsg.ProtoRecord{
		Fields: []*syncmsg.ProtoField{
			//Important: these field names need to be in alphabetical order
			//contactIdField, dateOfBirthField, firstNameField, heightFtField, heightInchField, lastNameField,
			//preferredHeightField,
			creator.CreateStringProtoField("contactId", "C441B7F6-D3E8-4E7F-9A3B-1BD4BAB6FA48"),
			creator.CreateTimeProtoField("dateOfBirth", creator.FormatTimeFromString("1990-04-29 00:00:00.000")),
			creator.CreateStringProtoField("firstName", "John"),
			creator.CreateInt64ProtoField("heightFt", 6),
			creator.CreateDoubleProtoField("heightInch", 2.5),
			creator.CreateStringProtoField("lastName", "Doe"),
			creator.CreateInt64ProtoField("preferredHeight", 1),
		},
	}
	if len(creator.Errors) != 0 {
		request := &syncmsg.ProtoSyncEntityMessageResponse{}
		return request, syncmsg.NewCreatorError(creator.Errors)
	}
	// record2LastKnown := &syncmsg.ProtoRecord{
	// 	Fields: []*syncmsg.ProtoField{
	// 		//Important: these field names need to be in alphabetical order
	// 		creator.CreateStringProtoField("contactId", "911DD745-8916-41C4-9973-F8B38A501602"),
	// 		creator.CreateTimeProtoField("dateOfBirth", creator.FormatTimeFromString("1994-07-10 00:00:00.000")),
	// 		creator.CreateStringProtoField("firstName", "Henry"),
	// 		creator.CreateInt64ProtoField("heightFt", 5),
	// 		creator.CreateDoubleProtoField("heightInch", 5.5),
	// 		creator.CreateStringProtoField("lastName", "Adins"),
	// 		creator.CreateInt64ProtoField("preferredHeight", 1),
	// 	},
	// }
	if len(creator.Errors) != 0 {
		request := &syncmsg.ProtoSyncEntityMessageResponse{}
		return request, syncmsg.NewCreatorError(creator.Errors)
	}
	record2 := &syncmsg.ProtoRecord{
		Fields: []*syncmsg.ProtoField{
			//Important: these field names need to be in alphabetical order
			creator.CreateStringProtoField("contactId", "911DD745-8916-41C4-9973-F8B38A501602"),
			creator.CreateTimeProtoField("dateOfBirth", creator.FormatTimeFromString("1994-07-10 00:00:00.000")),
			creator.CreateStringProtoField("firstName", "Henry"),
			creator.CreateInt64ProtoField("heightFt", 5),
			creator.CreateDoubleProtoField("heightInch", 5.5),
			creator.CreateStringProtoField("lastName", "Adkins"),
			creator.CreateInt64ProtoField("preferredHeight", 1),
		},
	}
	if len(creator.Errors) != 0 {
		request := &syncmsg.ProtoSyncEntityMessageResponse{}
		return request, syncmsg.NewCreatorError(creator.Errors)
	}

	/*
		record3 := &syncmsg.ProtoRecord{
			Fields: []*syncmsg.ProtoField{
				//Important: these field names need to be in alphabetical order
				createStringProtoField("contactId", "42FC5EBC-088D-4DC2-BDBD-D89FA170E5C6"),
				createTimeProtoField("dateOfBirth", createDateFromString("1992-01-23 00:00:00.000")),
				createStringProtoField("firstName", "John"),
				createInt64ProtoField("heightFt", 5),
				createDoubleProtoField("heightInch", 9),
				createStringProtoField("lastName", "Doe"),
				createInt64ProtoField("preferredHeight", 1),
			},
		}
	*/
	record1Bytes, err := proto.Marshal(record1)
	if err != nil {
		syncutil.Error("marshaling error: ", err)
		return &syncmsg.ProtoSyncEntityMessageResponse{}, err
	}
	record2Bytes, err := proto.Marshal(record2)
	if err != nil {
		syncutil.Error("marshaling error: ", err)
		return &syncmsg.ProtoSyncEntityMessageResponse{}, err
	}
	// record2LastKnownBytes, err := proto.Marshal(record2LastKnown)
	// if err != nil {
	// 	syncutil.Error("marshaling error: ", err)
	// 	return &syncmsg.ProtoSyncEntityMessageResponse{}, err
	// }

	// TransactionBindId *string                          `protobuf:"bytes,1,req,name=transaction_bind_id" json:"transaction_bind_id,omitempty"`
	// Items             []*ProtoSyncDataMessagesResponse `protobuf:"bytes,2,rep,name=items" json:"items,omitempty"`
	// Result            *SyncEntityMessageResponseResult `protobuf:"varint,3,req,name=result,enum=SyncEntityMessageResponseResult" json:"result,omitempty"`
	// ResultMsg         *string                          `protobuf:"bytes,4,req,name=result_msg" json:"result_msg,omitempty"`
	// XXX_unrecognized  []byte                           `json:"-"`

	// EntityPluralName *string                         `protobuf:"bytes,1,req,name=entity_plural_name" json:"entity_plural_name,omitempty"`
	// Msgs             []*ProtoSyncDataMessageResponse `protobuf:"bytes,2,rep,name=msgs" json:"msgs,omitempty"`

	// RecordId *string `protobuf:"bytes,1,req,name=record_id" json:"record_id,omitempty"`
	// RequestHash  *string           `protobuf:"bytes,2,opt,name=request_hash" json:"request_hash,omitempty"`
	// ResponseHash *string           `protobuf:"bytes,3,req,name=response_hash" json:"response_hash,omitempty"`
	// SyncState    *AckSyncStateEnum `protobuf:"varint,4,req,name=sync_state,enum=AckSyncStateEnum" json:"sync_state,omitempty"`
	// RecordBytesSize *uint32 `protobuf:"varint,5,opt,name=record_bytes_size" json:"record_bytes_size,omitempty"`
	// RecordData       []byte `protobuf:"bytes,6,opt,name=record_data" json:"record_data,omitempty"`

	requestData := &syncmsg.ProtoSyncEntityMessageResponse{
		TransactionBindId: proto.String("0F16AEED-E4B4-483E-A7A3-CCABA831FE6E"),
		Items: []*syncmsg.ProtoSyncDataMessagesResponse{
			&syncmsg.ProtoSyncDataMessagesResponse{
				EntityPluralName: proto.String("Contacts"),
				Msgs: []*syncmsg.ProtoSyncDataMessageResponse{
					&syncmsg.ProtoSyncDataMessageResponse{
						RecordId:        proto.String("B6581A36-804D-45AC-B2E2-F6DA265AF7DE"),
						RequestHash:     nil,
						ResponseHash:    proto.String(Hash256Bytes(record1Bytes)),
						SyncState:       syncmsg.AckSyncStateEnum(syncmsg.AckSyncStateEnum_AckFastBatch).Enum(),
						RecordBytesSize: proto.Uint32(uint32(len(record1Bytes))),
						RecordData:      nil,
					},
					&syncmsg.ProtoSyncDataMessageResponse{
						RecordId:        proto.String("911DD745-8916-41C4-9973-F8B38A501602"),
						RequestHash:     nil,
						ResponseHash:    proto.String(Hash256Bytes(record2Bytes)),
						SyncState:       syncmsg.AckSyncStateEnum(syncmsg.AckSyncStateEnum_AckFastBatch).Enum(),
						RecordBytesSize: proto.Uint32(uint32(len(record2Bytes))),
						RecordData:      nil,
					},
				},
			},
		},
		Result:    syncmsg.SyncEntityMessageResponseResult(syncmsg.SyncEntityMessageResponseResult_OK).Enum(),
		ResultMsg: proto.String(""),
	}
	return requestData, nil
}

//Contact aids in testing, providing a sample record.
type Contact struct {
	ContactID        string
	DateOfBirthAsUTC time.Time
	FirstName        string
	HeightFt         int
	HeightInch       float64
	LastName         string
	PreferredHeight  int
}

//DateOfBurthAsString returns the DateOfBirthAsUTC in the standard format for sql and transmission.
func (contact Contact) DateOfBurthAsString() string {
	return contact.DateOfBirthAsUTC.Format("2006-01-02 15:04:05.000")
}

//ContactSyncPackage is the calculated values for processing
type ContactSyncPackage struct {
	Contact           Contact
	Record            *syncmsg.ProtoRecord
	RecordBytes       []byte
	RecordSha256Hex   string
	PeerLastKnownHash string
	RecordHex         string
	ChangedByClient   bool
	SentState         syncmsg.SentSyncStateEnum
	//RecordBytesStr  string
}

//RecordBytesLen returns the length of RecordBytes
func (contactPackage ContactSyncPackage) RecordBytesLen() int {
	return len(contactPackage.RecordBytes)
}

//CreateRecordAndSupport provides the syncmsg.ProtoRecord, recordBytes, recordSha256Hex, and recordHex of a ProtoRecord.
func CreateRecordAndSupport(contact Contact, changedByClient bool, sentState syncmsg.SentSyncStateEnum, peerLastKnownHash string) (ContactSyncPackage, error) {
	var record syncmsg.ProtoRecord
	creator := syncmsg.NewCreator()
	if contact.DateOfBirthAsUTC.IsZero() {
		record = syncmsg.ProtoRecord{
			Fields: []*syncmsg.ProtoField{
				//Important: these field names need to be in alphabetical order
				creator.CreateStringProtoField("contactId", contact.ContactID),
				//creator.CreateTimeProtoField("dateOfBirth", contact.DateOfBirthAsUTC),
				creator.CreateStringProtoField("firstName", contact.FirstName),
				creator.CreateInt64ProtoField("heightFt", contact.HeightFt),
				creator.CreateDoubleProtoField("heightInch", contact.HeightInch),
				creator.CreateStringProtoField("lastName", contact.LastName),
				creator.CreateInt64ProtoField("preferredHeight", contact.PreferredHeight),
			},
		}
		//syncutil.Debug("CreateRecordAndSupport Is Zero")
	} else {
		record = syncmsg.ProtoRecord{
			Fields: []*syncmsg.ProtoField{
				//Important: these field names need to be in alphabetical order
				creator.CreateStringProtoField("contactId", contact.ContactID),
				creator.CreateTimeProtoField("dateOfBirth", contact.DateOfBirthAsUTC),
				creator.CreateStringProtoField("firstName", contact.FirstName),
				creator.CreateInt64ProtoField("heightFt", contact.HeightFt),
				creator.CreateDoubleProtoField("heightInch", contact.HeightInch),
				creator.CreateStringProtoField("lastName", contact.LastName),
				creator.CreateInt64ProtoField("preferredHeight", contact.PreferredHeight),
			},
		}
		//syncutil.Debug("CreateRecordAndSupport Is NOT Zero")
	}

	var recordBytes []byte
	var recordSha256Hex string
	var recordHex string
	var err error
	recordBytes, err = proto.Marshal(&record)

	if err != nil {
		syncutil.Fatal("Marshaling error: ", err)
		return ContactSyncPackage{contact, &record, recordBytes, recordSha256Hex, peerLastKnownHash, recordHex, changedByClient, sentState}, err
	}
	recordSha256Hex = Hash256Bytes(recordBytes)
	//syncutil.Debug("record2Sha256Hex:", record2Sha256Hex)
	recordHex = hex.EncodeToString(recordBytes)
	//recordBytesStr := fmt.Sprintf("%v", len(contactSyncPackage.RecordBytes))

	if peerLastKnownHash == "" && sentState.IsNotFirstTimeSend() {
		peerLastKnownHash = recordSha256Hex
	}

	return ContactSyncPackage{contact, &record, recordBytes, recordSha256Hex, peerLastKnownHash, recordHex, changedByClient, sentState}, err
}

var createSyncModelTablesSQL = `
--1:
CREATE TABLE sync_data_version (
DataVersionName		varchar(36)	NOT NULL,
RecordCreated			timestamp		NOT NULL	default(now()),
PRIMARY KEY (DataVersionName)
);

--2:
CREATE TABLE sync_data_entity (
EntitySingularName	varchar(50)		NOT NULL,
EntityPluralName		varchar(50)		NOT NULL,
DataVersionName			varchar(36)		NOT NULL,
ProcOrderAddUpdate	integer				NOT NULL,
ProcOrderDelete			integer				NOT NULL,
EntityHandlerUri		varchar(2048) NULL,
RecordCreated				timestamp			NOT NULL	default(now()),
PRIMARY KEY (EntitySingularName)
);

--3:
CREATE TABLE sync_state (
EntitySingularName	varchar(50)		NOT NULL,
RecordId						varchar(112)	NOT NULL,
DataVersionName			varchar(36)		NOT NULL,
RecordHash					varchar(100)	NOT NULL,
RecordData					bytea 				NOT NULL,
RecordBytesSize			int						NOT NULL,
IsDelete						boolean				NOT NULL 	default('false'),
GroupName						varchar(256)	NULL, --RESERVED: Not currently in use
GroupType						varchar(256)	NULL, --RESERVED: Not currently in use
GroupSeqNum					int						NULL, --RESERVED: Not currently in use
RecordCreated				timestamp			NOT NULL	default(now()),
PRIMARY KEY (EntitySingularName, RecordId)
);

--4:
CREATE TABLE sync_node(
NodeId							varchar(36) 	NOT NULL,
NodeName						varchar(256) 	NOT NULL,
DataVersionName			varchar(36)		NOT NULL,
Enabled							boolean 			NOT NULL	default('true'),
DataMsgConsUri			varchar(2048) NULL,
DataMsgProdUri			varchar(2048) NULL,
MgmtMsgConsUri			varchar(2048) NULL,
MgmtMsgProdUri			varchar(2048) NULL,
SyncDataPersistForm	varchar(36) 	NULL,
InMsgBatchSize			int 					NULL 			default(100),
MaxOutMsgBatchSize  int 					NULL 			default(200),
InChanDepthSize			int 					NULL 			default(16),
MaxOutChanDepthSize int 					NULL 			default(16),
RecordCreated				timestamp			NOT NULL	default(now()),
UNIQUE(NodeName),
PRIMARY KEY (NodeId)
);

--5:
CREATE TABLE sync_peer_state (
NodeId										varchar(36)		NOT NULL,
EntitySingularName				varchar(50)		NOT NULL,
RecordId									varchar(112)	NOT NULL,
SessionBindId							varchar(36)		NULL,
TransactionBindReceiveId	varchar(36)		NULL,
TransactionBindSendId			varchar(36)		NULL,
QueueBindSendId						varchar(36)		NULL,
SentLastKnownHash					varchar(100)	NULL,
PeerLastKnownHash					varchar(100)	NULL,
IsConflict 								boolean				NOT NULL 	default('false'),
IsDelete 									boolean				NOT NULL 	default('false'),
SentSyncState							integer 			NOT NULL 	default(1),
-- 1-never sent to peer (default on add record), 2-first time sent to peer, 3-standard sent to peer,
-- 4-fast delete
RecordBytesSize						int						NOT NULL,
ChangedByClient						boolean				NOT NULL 	default('false'),
LastUpdated								date					NULL,
RecordCreated							timestamp			NOT NULL	default(now()),
PRIMARY KEY (NodeId, EntitySingularName, RecordId),
CONSTRAINT valid_sent_sync_state CHECK (SentSyncState >= 1 AND SentSyncState <= 4)
);

--6:
CREATE TABLE sync_data_field (
EntitySingularName	varchar(50)		NOT NULL,
FieldName						varchar(100)	NOT NULL, -- this is a canonical name regardless of local naming convensions using camel case starting with Upper Case
DataVersionName			varchar(36) 	NOT NULL,
DataTypeName				varchar(6)		NOT NULL,
IsPrimaryKey				boolean			NOT NULL,
RecordCreated				timestamp		NOT NULL	default(now()),
PRIMARY KEY (EntitySingularName, FieldName),
CONSTRAINT valid_date_type_name CHECK (DataTypeName = 'String' OR DataTypeName = 'Date' OR
										DataTypeName = 'Bool'	   OR DataTypeName = 'Int'  OR
										DataTypeName = 'Float'   OR DataTypeName = 'Binary')
);

--7:
CREATE TABLE sync_pair(
PairId						varchar(36) 	NOT NULL,
PairName					varchar(72) 	NOT NULL,
MaxSesDurValue		int 			NULL 		default(10),
MaxSesDurUnit			varchar(36) 	NOT NULL 	default('minutes'),
SyncDataTransForm	varchar(36) 	NOT NULL 	default('json:V1'),
SyncMsgTransForm	varchar(36) 	NOT NULL 	default('json:V1'),
SyncMsgSecPol			varchar(36) 	NOT NULL 	default('none'),
SyncSessionId			varchar(36)		NULL,
SyncSessionState	varchar(36)		NOT NULL	default('Inactive'),
SyncSessionStart	timestamp 		NULL,
SyncConflictUri	  varchar(2048) 	NOT NULL 	default('none'),
RecordCreated			timestamp		NOT NULL	default(now()),
UNIQUE(PairName, SyncSessionId),
PRIMARY KEY (PairId),
CONSTRAINT valid_session_state CHECK (SyncSessionState = 'Inactive' OR SyncSessionState = 'Initializing' OR
										SyncSessionState = 'Seeding'  OR SyncSessionState = 'Queuing'      OR
										SyncSessionState = 'Syncing'  OR SyncSessionState = 'Canceling')
);

--8:
CREATE TABLE sync_pair_nodes(
PairId						varchar(36) 	NOT NULL,
NodeId						varchar(36) 	NOT NULL,
TargetNodeId			varchar(36) 	NOT NULL,
LastSeededDate		timestamp		NULL,
SeededDataVersion	varchar(36)		NULL,
SyncConflictUri	  varchar(2048) 	NOT NULL,
TotalSessRecBytes	int				NOT NULL	default(0),
TotalSessRecCount	int				NOT NULL	default(0),
ProceSessRecBytes	int				NOT NULL	default(0),
ProceSessRecCount	int				NOT NULL	default(0),
RecordCreated			timestamp		NOT NULL	default(now()),
PRIMARY KEY (PairId, NodeId, TargetNodeId)
);

--GRANT SELECT, INSERT, UPDATE, DELETE ON sync_pair_nodes TO doug;
--GRANT SELECT, INSERT, UPDATE, DELETE ON sync_pair TO doug;
--GRANT SELECT, INSERT, UPDATE, DELETE ON sync_node TO doug;

/*
sync_pair_nodes			>---*:1--- sync_node
|-- NodeId  			>------- NodeId
|-- NodeId  			>------- TargetNodeId
*/
ALTER TABLE sync_pair_nodes ADD CONSTRAINT FK_sync_pair_nodes_source_source_node
FOREIGN KEY(NodeId) REFERENCES sync_node (NodeId);

ALTER TABLE sync_pair_nodes ADD CONSTRAINT FK_sync_pair_nodes_source_target_node
FOREIGN KEY(TargetNodeId) REFERENCES sync_node (NodeId);

/*
sync_pair_nodes			>---*:1--- sync_pair
|-- PairId  			>------- PairId
*/
ALTER TABLE sync_pair_nodes ADD CONSTRAINT FK_sync_pair_nodes_sync_pair
FOREIGN KEY(PairId) REFERENCES sync_pair (PairId);

/*
sync_pair_nodes			>---*:1--- sync_pair
|-- PairId  			>------- PairId
*/
ALTER TABLE sync_pair_nodes ADD CONSTRAINT FK_sync_pair_nodes_sync_data_version
FOREIGN KEY(SeededDataVersion) REFERENCES sync_data_version (DataVersionName);

/*
sync_peer_state		>---*:1--- sync_node
|--	NodeId			>------- NodeId
*/
ALTER TABLE sync_peer_state ADD CONSTRAINT FK_sync_peer_state_sync_node
FOREIGN KEY(NodeId) REFERENCES sync_node (NodeId);

/*
sync_peer_state 	>---*:1--- sync_state
|--	EntityId		>------- EntityId
|--	RecordId  		>------- RecordId
*/
ALTER TABLE sync_peer_state ADD CONSTRAINT FK_sync_peer_state_sync_state
FOREIGN KEY(EntitySingularName, RecordId) REFERENCES sync_state (EntitySingularName, RecordId);

/*
sync_node			>---*:1--- sync_data_version
|--	DataVersionId	>------- DataVersionId
*/
ALTER TABLE sync_node ADD CONSTRAINT FK_sync_node_sync_data_version
FOREIGN KEY(DataVersionName) REFERENCES sync_data_version (DataVersionName);

/*
sync_state				>---*:1--- sync_data_entity
|-- EntityId  			>------- EntityId
*/
ALTER TABLE sync_state ADD CONSTRAINT FK_sync_state_sync_entity
FOREIGN KEY(EntitySingularName) REFERENCES sync_data_entity (EntitySingularName);

/*
sync_state				>---*:1--- sync_data_version
|-- DataVersionName  			>------- DataVersionName
*/
ALTER TABLE sync_state ADD CONSTRAINT FK_sync_state_sync_data_version
FOREIGN KEY(DataVersionName) REFERENCES sync_data_version (DataVersionName);

/*
sync_data_field			>---*:1--- sync_data_entity
|-- EntityId  			>------- EntityId
*/
ALTER TABLE sync_data_entity ADD CONSTRAINT FK_sync_data_entity_sync_data_version
FOREIGN KEY(DataVersionName) REFERENCES sync_data_version (DataVersionName);

/*
sync_data_field			>---*:1--- sync_data_entity
|-- EntityId  			>------- EntityId
*/
ALTER TABLE sync_data_field ADD CONSTRAINT FK_sync_data_field_sync_data_entity
FOREIGN KEY(EntitySingularName) REFERENCES sync_data_entity (EntitySingularName);

/*
sync_data_field			>---*:1--- sync_data_version
|-- DataVersion  			>------- DataVersion
*/
ALTER TABLE sync_data_field ADD CONSTRAINT FK_sync_data_field_sync_data_version
FOREIGN KEY(DataVersionName) REFERENCES sync_data_version (DataVersionName);
`

var dropSyncModelTablesSQL = `
drop table sync_pair_nodes;
drop table sync_pair;
drop table sync_data_field;
drop table sync_peer_state;
drop table sync_node;
drop table sync_state;
drop table sync_data_entity;
drop table sync_data_version;
`

var profile1SQL = `
--1: sync_data_version
insert into sync_data_version (DataVersionName) values ('Demo Model 1');
insert into sync_data_version (DataVersionName) values ('Demo Model 2 (orphand node)');

--2: sync_entity (IGNORED)

--3: sync_state (IGNORED)

--4: sync_node
INSERT INTO sync_node(nodeId,nodeName, DataVersionName) VALUES('*node-spoke1','A', 'Demo Model 1');
INSERT INTO sync_node(nodeId,nodeName, DataVersionName) VALUES('*node-spoke2','B', 'Demo Model 2 (orphand node)'); --this one has orphaned data profile
INSERT INTO sync_node(nodeId,nodeName, DataVersionName) VALUES('*node-spoke3','C', 'Demo Model 1');
INSERT INTO sync_node(nodeId,nodeName, DataVersionName) VALUES('*node-hub','Z', 'Demo Model 1');

--5: sync_peer_state (IGNORED)

--6: sync_data_field (IGNORED)

--7: sync_pair
INSERT INTO sync_pair(PairId,PairName) VALUES('*pair-1','A <-> Z');
INSERT INTO sync_pair(PairId,PairName) VALUES('*pair-2','B <-> Z');
INSERT INTO sync_pair(PairId,PairName) VALUES('*pair-3','A <-> B (Partial)');
INSERT INTO sync_pair(PairId,PairName) VALUES('*pair-4','C <-> Z (Dup 1 of 2)');
INSERT INTO sync_pair(PairId,PairName) VALUES('*pair-5','C <-> Z (Dup 2 of 2)');

--8: sync_pair_nodes
INSERT INTO sync_pair_nodes(pairId,NodeId,targetNodeId,SeededDataVersion,SyncConflictUri) VALUES('*pair-1','*node-hub','*node-spoke1','Demo Model 1','none');
INSERT INTO sync_pair_nodes(pairId,NodeId,targetNodeId,SeededDataVersion,SyncConflictUri) VALUES('*pair-1','*node-spoke1','*node-hub','Demo Model 1','none');
INSERT INTO sync_pair_nodes(pairId,NodeId,targetNodeId,SeededDataVersion,SyncConflictUri) VALUES('*pair-2','*node-hub','*node-spoke2','Demo Model 1','none');
INSERT INTO sync_pair_nodes(pairId,NodeId,targetNodeId,SeededDataVersion,SyncConflictUri) VALUES('*pair-2','*node-spoke2','*node-hub','Demo Model 1','none');
INSERT INTO sync_pair_nodes(pairId,NodeId,targetNodeId,SeededDataVersion,SyncConflictUri) VALUES('*pair-3','*node-spoke2','*node-spoke1','Demo Model 1','none');
INSERT INTO sync_pair_nodes(pairId,NodeId,targetNodeId,SeededDataVersion,SyncConflictUri) VALUES('*pair-4','*node-hub','*node-spoke3','Demo Model 1','none');
INSERT INTO sync_pair_nodes(pairId,NodeId,targetNodeId,SeededDataVersion,SyncConflictUri) VALUES('*pair-4','*node-spoke3','*node-hub','Demo Model 1','none');
INSERT INTO sync_pair_nodes(pairId,NodeId,targetNodeId,SeededDataVersion,SyncConflictUri) VALUES('*pair-5','*node-hub','*node-spoke3','Demo Model 1','none');
INSERT INTO sync_pair_nodes(pairId,NodeId,targetNodeId,SeededDataVersion,SyncConflictUri) VALUES('*pair-5','*node-spoke3','*node-hub','Demo Model 1','none');
`
var profile2SQL = `
--1: sync_data_version (IGNORED)

--2: sync_entity
INSERT INTO sync_data_entity (EntitySingularName,EntityPluralName,DataVersionName,ProcOrderAddUpdate,ProcOrderDelete,EntityHandlerUri,RecordCreated) values ('Entity 1','Entity 1','Demo Model 1',1,4,'none','2015-10-14 08:40:05.000');
INSERT INTO sync_data_entity (EntitySingularName,EntityPluralName,DataVersionName,ProcOrderAddUpdate,ProcOrderDelete,EntityHandlerUri,RecordCreated) values ('Entity 2','Entity 2','Demo Model 1',2,3,'none','2015-10-14 08:40:05.000');
INSERT INTO sync_data_entity (EntitySingularName,EntityPluralName,DataVersionName,ProcOrderAddUpdate,ProcOrderDelete,EntityHandlerUri,RecordCreated) values ('Entity 3','Entity 3','Demo Model 1',2,3,'none','2015-10-14 08:40:05.000');
INSERT INTO sync_data_entity (EntitySingularName,EntityPluralName,DataVersionName,ProcOrderAddUpdate,ProcOrderDelete,EntityHandlerUri,RecordCreated) values ('Entity 4','Entity 4','Demo Model 1',3,2,'none','2015-10-14 08:40:05.000');
INSERT INTO sync_data_entity (EntitySingularName,EntityPluralName,DataVersionName,ProcOrderAddUpdate,ProcOrderDelete,EntityHandlerUri,RecordCreated) values ('Entity 5','Entity 5','Demo Model 1',4,1,'none','2015-10-14 08:40:05.000');

INSERT INTO sync_data_entity  (EntitySingularName,EntityPluralName,DataVersionName,ProcOrderAddUpdate,ProcOrderDelete,EntityHandlerUri,RecordCreated) values ('Contact','Contacts','Demo Model 1',4,1,'none','2015-10-14 08:40:05.000');

--3: sync_state (IGNORED)

--4: sync_node (IGNORED)

--5: sync_peer_state (IGNORED)

--6: sync_data_field
INSERT INTO sync_data_field (EntitySingularName, FieldName, DataTypeName, DataVersionName, IsPrimaryKey) VALUES ('Entity 1','FirstName','String','Demo Model 1', false);
INSERT INTO sync_data_field (EntitySingularName, FieldName, DataTypeName, DataVersionName, IsPrimaryKey) VALUES ('Entity 1','LastName','String','Demo Model 1', false);

INSERT INTO sync_data_field (EntitySingularName, FieldName, DataTypeName, DataVersionName, IsPrimaryKey) VALUES ('Contact','contactId','String','Demo Model 1', true);
INSERT INTO sync_data_field (EntitySingularName, FieldName, DataTypeName, DataVersionName, IsPrimaryKey) VALUES ('Contact','dateOfBirth','Date','Demo Model 1', false);
INSERT INTO sync_data_field (EntitySingularName, FieldName, DataTypeName, DataVersionName, IsPrimaryKey) VALUES ('Contact','firstName','String','Demo Model 1', false);
INSERT INTO sync_data_field (EntitySingularName, FieldName, DataTypeName, DataVersionName, IsPrimaryKey) VALUES ('Contact','heightFt','Int','Demo Model 1', false);
INSERT INTO sync_data_field (EntitySingularName, FieldName, DataTypeName, DataVersionName, IsPrimaryKey) VALUES ('Contact','heightInch','Float','Demo Model 1', false);
INSERT INTO sync_data_field (EntitySingularName, FieldName, DataTypeName, DataVersionName, IsPrimaryKey) VALUES ('Contact','lastName','String','Demo Model 1', false);
INSERT INTO sync_data_field (EntitySingularName, FieldName, DataTypeName, DataVersionName, IsPrimaryKey) VALUES ('Contact','preferredHeight','Int','Demo Model 1', false);

--7: sync_pair (IGNORED)

--8: sync_pair_nodes (IGNORED)
`

var profile3SQL = `
--1: sync_data_version (IGNORED)

--2: sync_entity (IGNORED)

--3: sync_state
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('Entity 1','*record-1','Demo Model 1','hash-1','<record data>',10,'false','2015-10-14 08:40:05.000');
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('Entity 1','*record-2','Demo Model 1','hash-2','<record data>',11,'false','2015-10-14 08:40:05.000');
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('Entity 1','*record-3','Demo Model 1','hash-3','<record data>',12,'false','2015-10-14 08:40:05.000');

--4: sync_node (IGNORED)

--5: sync_peer_state
--Entries for *node-hub
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','Entity 1','*record-1',NULL,NULL,NULL,NULL,NULL,'false',1,'true',0,'2015-10-14 08:40:05.000');
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','Entity 1','*record-2',NULL,NULL,NULL,NULL,NULL,'false',1,'true',0,'2015-10-14 08:40:05.000');
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','Entity 1','*record-3',NULL,NULL,NULL,NULL,NULL,'false',1,'true',0,'2015-10-14 08:40:05.000');

--6: sync_data_field (IGNORED)

--7: sync_pair (IGNORED)

--8: sync_pair_nodes (IGNORED)
`
var profile4SQL = `
--1: sync_data_version (IGNORED)

--2: sync_entity
INSERT INTO sync_data_entity (EntitySingularName,EntityPluralName,DataVersionName,ProcOrderAddUpdate,ProcOrderDelete,EntityHandlerUri,RecordCreated) values ('A','A','Demo Model 1',1,4,'none','2015-10-14 08:40:05.000');
INSERT INTO sync_data_entity (EntitySingularName,EntityPluralName,DataVersionName,ProcOrderAddUpdate,ProcOrderDelete,EntityHandlerUri,RecordCreated) values ('B','B','Demo Model 1',1,4,'none','2015-10-14 08:40:05.000');
INSERT INTO sync_data_entity (EntitySingularName,EntityPluralName,DataVersionName,ProcOrderAddUpdate,ProcOrderDelete,EntityHandlerUri,RecordCreated) values ('C','C','Demo Model 1',1,4,'none','2015-10-14 08:40:05.000');

--3: sync_state (IGNORED)

--4: sync_node (IGNORED)

--5: sync_peer_state (IGNORED)

--6: sync_data_field (IGNORED)

--7: sync_pair (IGNORED)

--8: sync_pair_nodes (IGNORED)
`

var profile5SQL = `
--1: sync_data_version (IGNORED)

--2: sync_entity (IGNORED)

--3: sync_state
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('Entity 1','*record-1','Demo Model 1','hash-1','<record data>',10,'false','2015-10-14 08:40:05.000');
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('Entity 1','*record-2','Demo Model 1','hash-2','<record data>',11,'false','2015-10-14 08:40:05.000');
INSERT INTO sync_state (EntitySingularName,RecordId,DataVersionName,RecordHash,RecordData,RecordBytesSize,IsDelete,RecordCreated) VALUES ('Entity 1','*record-3','Demo Model 1','hash-3','<record data>',12,'false','2015-10-14 08:40:05.000');

--4: sync_node (IGNORED)

--5: sync_peer_state
--Entries for *node-hub
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','Entity 1','*record-1',NULL,NULL,NULL,NULL,NULL,'false',1,'true',0,'2015-10-14 08:40:05.000');
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','Entity 1','*record-2',NULL,NULL,NULL,NULL,NULL,'false',1,'true',0,'2015-10-14 08:40:05.000');
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-hub','Entity 1','*record-3',NULL,NULL,NULL,NULL,NULL,'false',1,'true',0,'2015-10-14 08:40:05.000');
--Entries for *node-spoke1
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-spoke1','Entity 1','*record-1',NULL,NULL,NULL,NULL,NULL,'false',1,'false',0,'2015-10-14 08:40:05.000');
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-spoke1','Entity 1','*record-2',NULL,NULL,NULL,NULL,NULL,'false',1,'false',0,'2015-10-14 08:40:05.000');
INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-spoke1','Entity 1','*record-3',NULL,NULL,NULL,NULL,NULL,'false',1,'false',0,'2015-10-14 08:40:05.000');

--6: sync_data_field (IGNORED)

--7: sync_pair (IGNORED)

--8: sync_pair_nodes (IGNORED)
`

var sqlTemplateContactAndSyncState = `
--Start of a new entry {{.Index}}
INSERT into sync_state (EntitySingularName, RecordId, DataVersionName, RecordHash, RecordData, RecordBytesSize) values ('Contact', '{{.ContactPackage.Contact.ContactID}}', 'Demo Model 1', '{{.ContactPackage.RecordSha256Hex}}', '{{.ContactPackage.RecordHex}}', {{.ContactPackage.RecordBytesLen}});
INSERT into Contacts (ContactId, DateofBirth, FirstName, LastName, HeightFt, HeightInch, PreferredHeight) values('{{.ContactPackage.Contact.ContactID}}', '{{.ContactPackage.Contact.DateOfBurthAsString}}', '{{.ContactPackage.Contact.FirstName}}', '{{.ContactPackage.Contact.LastName}}', {{.ContactPackage.Contact.HeightFt}}, {{.ContactPackage.Contact.HeightInch}}, {{.ContactPackage.Contact.PreferredHeight}});
{{if .ContactPackage.SentState.IsNotFirstTimeSend}}INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-spoke1','Contact','{{.ContactPackage.Contact.ContactID}}',NULL,NULL,NULL,'{{.ContactPackage.PeerLastKnownHash}}','{{.ContactPackage.PeerLastKnownHash}}','false',{{.ContactPackage.SentState.Value}},{{.ContactPackage.ChangedByClient}},{{.ContactPackage.RecordBytesLen}}, '2015-10-14 08:40:05.000');{{end}}`

// var sqlTemplateContactAndSyncState = `
// --Start of a new entry {{.Index}}
// INSERT into sync_state (EntitySingularName, RecordId, DataVersionName, RecordHash, RecordData, RecordBytesSize) values ('Contact', '{{.ContactPackage.Contact.ContactID}}', 'Demo Model 1', '{{.ContactPackage.RecordSha256Hex}}', '{{.ContactPackage.RecordHex}}', {{.ContactPackage.RecordBytesLen}});
// INSERT into Contacts (ContactId, DateofBirth, FirstName, LastName, HeightFt, HeightInch, PreferredHeight) values('{{.ContactPackage.Contact.ContactID}}', '{{.ContactPackage.Contact.DateOfBurthAsString}}', '{{.ContactPackage.Contact.FirstName}}', '{{.ContactPackage.Contact.LastName}}', {{.ContactPackage.Contact.HeightFt}}, {{.ContactPackage.Contact.HeightInch}}, {{.ContactPackage.Contact.PreferredHeight}});
// {{if .ContactPackage.SentState.IsNotFirstTimeSend}}INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId, TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-spoke1','Contact','{{.ContactPackage.Contact.ContactID}}',NULL,NULL,NULL,'{{.ContactPackage.RecordSha256Hex}}','{{.ContactPackage.PeerLastKnownHash}}','false',{{.ContactPackage.SentState.Value}},{{.ContactPackage.ChangedByClient}},{{.ContactPackage.RecordBytesLen}}, '2015-10-14 08:40:05.000');{{end}}`

// var sqlTemplateContactSyncPeerState = `
// INSERT INTO sync_peer_state (NodeId,EntitySingularName,RecordId,SessionBindId,TransactionBindReceiveId,TransactionBindSendId,SentLastKnownHash,PeerLastKnownHash,IsConflict,SentSyncState,ChangedByClient,RecordBytesSize,LastUpdated) VALUES ('*node-spoke1','Contact','{{.ContactPackage.Contact.ContactID}}',NULL,NULL,NULL,'{{.ContactPackage.RecordSha256Hex}}','{{.ContactPackage.PeerLastKnownHash}}','false',{{.ContactPackage.SentState.Value}},{{.ContactPackage.ChangedByClient}},{{.ContactPackage.RecordBytesLen}},'2015-10-14 08:40:05.000');
// `

// --Start of a new entry
// INSERT into sync_state (EntitySingularName, RecordId, DataVersionName, RecordHash, RecordData, RecordBytesSize) values ('Contact', '{{.Contact.ContactID}}', 'Demo Model 1', '{{.RecordSha256Hex}}, '{{.RecordHex}}', {{.RecordBytes}});
// INSERT into Contacts (ContactId, DateofBirth, FirstName, LastName, HeightFt, HeightInch, PreferredHeight) values({{.Contact.ContactID}}', '{{.Date}}', '{{.Contact.FirstName}}', '{{.Contact.LastName}}', {{.Contact.HeightFt}}, {{.Contact.HeightInch}}, {{.Contact.PreferredHeight}});
// `
