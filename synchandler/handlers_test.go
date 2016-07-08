package synchandler

// TODO(doug4j@gmail.com): Transition handler testing to use a mock/stub syncdao rather than syncdaopq
import (
	"bytes"
	"database/sql"
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncdao"
	"data-sync-tools-go/syncdao/syncdaopq"
	"data-sync-tools-go/syncmsg"
	"data-sync-tools-go/syncutil"
	"data-sync-tools-go/testhelper"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	//_ "github.com/gorilla/mux"
	"io"
	"io/ioutil"

	"github.com/twinj/uuid"
	//"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
	//"time"
)

type delegateHandlerFunc struct {
	call http.HandlerFunc
}

func (caller delegateHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	caller.call(w, r)
}

var (
	server *httptest.Server
	//reader   io.Reader //Ignore this for now
	//usersUrl string
	getSyncConfigURLOK        string
	getSyncConfigURLNotPaired string
	changeQueueURLOK          string
	sessionBaseURL            string
	syncPairStateURL          string
	syncPairConfigURL         string
	syncDataBaseURL           string

	testDbUser     = testhelper.TestDbUser
	testDbPassword = testhelper.TestDbPassword
	testDbName     = testhelper.TestDbName
	testDbHost     = testhelper.TestDbHost
	testDbPort     = testhelper.TestDbPort
)

func setupDatabaseObj() {
	teardownDatabaseObj()
	dbFactory, err := syncdaopq.NewPostgresSQLDaosFactory(testDbUser, testDbPassword, testDbHost, testDbName, testDbPort)
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

type TestItem struct {
	ExpectedValue interface{}
}

var serviceAnswersExpecations = map[string]TestItem{
	"GetSyncConfigOK": TestItem{
		ExpectedValue: PairConfigResponse{
			PairID:            "*pair-1",
			PairName:          "A <-> Z",
			MaxSesDurValue:    10,
			MaxSesDurUnit:     "minutes",
			SyncDataTransForm: "json:V1",
			SyncMsgTransForm:  "json:V1",
			SyncMsgSecPol:     "none",
			SyncConflictURI:   "none",
			Node1: syncdao.NodePairItem{
				NodeID:                    "*node-spoke1",
				NodeName:                  "A",
				Enabled:                   true,
				DataMsgConsumerURI:        "",
				DataMsgProducerURI:        "",
				MgmtMsgConsumerURI:        "",
				MgmtMsgProducerURI:        "",
				SyncDataPersistanceFormat: "",
				InMsgBatchSize:            100,
				MaxOutMsgBatchSize:        200,
				InChanDepthSize:           16,
				MaxOutChanDeptSize:        16,
				DataVersionName:           "-not implemented-",
				Entities: []syncdao.EntityPairItem{
					syncdao.EntityPairItem{
						EntitySingularName:    "Entity 1",
						EntityPluralName:      "Entity 1",
						ProcessOrderAddUpdate: 1,
						ProcessOrderDelete:    4,
						EntityHandlerURI:      "none",
					},
					syncdao.EntityPairItem{
						EntitySingularName:    "Entity 2",
						EntityPluralName:      "Entity 2",
						ProcessOrderAddUpdate: 2,
						ProcessOrderDelete:    3,
						EntityHandlerURI:      "none",
					},
					syncdao.EntityPairItem{
						EntitySingularName:    "Entity 3",
						EntityPluralName:      "Entity 3",
						ProcessOrderAddUpdate: 2,
						ProcessOrderDelete:    3,
						EntityHandlerURI:      "none",
					},
					syncdao.EntityPairItem{
						EntitySingularName:    "Entity 4",
						EntityPluralName:      "Entity 4",
						ProcessOrderAddUpdate: 3,
						ProcessOrderDelete:    2,
						EntityHandlerURI:      "none",
					},
					syncdao.EntityPairItem{
						EntitySingularName:    "Entity 5",
						EntityPluralName:      "Entity 5",
						ProcessOrderAddUpdate: 4,
						ProcessOrderDelete:    1,
						EntityHandlerURI:      "none",
					},
					syncdao.EntityPairItem{
						EntitySingularName:    "Contact",
						EntityPluralName:      "Contacts",
						ProcessOrderAddUpdate: 4,
						ProcessOrderDelete:    1,
						EntityHandlerURI:      "none",
					},
				},
			},
			Node2: syncdao.NodePairItem{
				NodeID:                    "*node-hub",
				NodeName:                  "Z",
				Enabled:                   true,
				DataMsgConsumerURI:        "",
				DataMsgProducerURI:        "",
				MgmtMsgConsumerURI:        "",
				MgmtMsgProducerURI:        "",
				SyncDataPersistanceFormat: "",
				InMsgBatchSize:            100,
				MaxOutMsgBatchSize:        200,
				InChanDepthSize:           16,
				MaxOutChanDeptSize:        16,
				DataVersionName:           "-not implemented-",
				Entities: []syncdao.EntityPairItem{
					syncdao.EntityPairItem{
						EntitySingularName:    "Entity 1",
						EntityPluralName:      "Entity 1",
						ProcessOrderAddUpdate: 1,
						ProcessOrderDelete:    4,
						EntityHandlerURI:      "none",
					},
					syncdao.EntityPairItem{
						EntitySingularName:    "Entity 2",
						EntityPluralName:      "Entity 2",
						ProcessOrderAddUpdate: 2,
						ProcessOrderDelete:    3,
						EntityHandlerURI:      "none",
					},
					syncdao.EntityPairItem{
						EntitySingularName:    "Entity 3",
						EntityPluralName:      "Entity 3",
						ProcessOrderAddUpdate: 2,
						ProcessOrderDelete:    3,
						EntityHandlerURI:      "none",
					},
					syncdao.EntityPairItem{
						EntitySingularName:    "Entity 4",
						EntityPluralName:      "Entity 4",
						ProcessOrderAddUpdate: 3,
						ProcessOrderDelete:    2,
						EntityHandlerURI:      "none",
					},
					syncdao.EntityPairItem{
						EntitySingularName:    "Entity 5",
						EntityPluralName:      "Entity 5",
						ProcessOrderAddUpdate: 4,
						ProcessOrderDelete:    1,
						EntityHandlerURI:      "none",
					},
					syncdao.EntityPairItem{
						EntitySingularName:    "Contact",
						EntityPluralName:      "Contacts",
						ProcessOrderAddUpdate: 4,
						ProcessOrderDelete:    1,
						EntityHandlerURI:      "none",
					},
				},
			},
			Response:    "OK",
			ResponseMsg: "",
		}},
	"GetSyncConfigNotPaired": TestItem{
		ExpectedValue: PairConfigResponse{
			Response:    "RequestNodeAndOtherPeerExistsButNotPaired",
			ResponseMsg: "",
		}},
	"GetSyncConfigDatabaseNotSet": TestItem{
		ExpectedValue: "Server Configuration Error: Database not set",
	},
	"QueueSyncChangesOK": TestItem{
		ExpectedValue: QueueSyncChangesResponse{
			RecordsQueuedCount: 6,
			SessionID:          "mySession1",
			NodeID:             "*node-spoke1",
			MsgID:              "myMsgId1",
			Result:             "OK",
			ResultMsg:          "",
		},
	},
	"CreateSession": TestItem{
		ExpectedValue: CreateSyncSessionResponse{
			PairID:             "*pair-1",
			RequestedSessionID: "placeholder",
			ActualSessionID:    "placeholder",
			Result:             "OK",
			ResultMsg:          "",
		},
	},
	"QuerySession": TestItem{
		ExpectedValue: QueryPairStateResponse{
			PairID:       "*pair-1",
			State:        "placeholder",
			SessionID:    "placeholder",
			SessionStart: 0,
			LastUpdated:  0,
			Result:       "OK",
			ResultMsg:    "",
		},
	},
	"UpdateSession": TestItem{
		ExpectedValue: UpdateSyncSessionStateResponse{
			SessionID:      "placeholder",
			RequestedState: "Queuing",
			Result:         "OK",
			ResultMsg:      "",
		},
	},
	"CloseSession": TestItem{
		ExpectedValue: CloseSyncSessionResponse{
			PairID:             "*pair-1",
			RequestedSessionID: "placeholder",
			ActualSessionID:    "placeholder",
			Result:             "OK",
			ResultMsg:          "",
		},
	},
	"NodeResponseOK": TestItem{
		ExpectedValue: NodeResponse{
			NodeID:          "placeholder",
			NodeName:        "some-new-node",
			DataVersionName: "Demo Model 1",

			MsgID:     "placeholder",
			Result:    "OK",
			ResultMsg: "",
		},
	},
	"DeleteNodeResponseOK": TestItem{
		ExpectedValue: DeleteNodeResponse{
			NodeID:    "placeholder",
			MsgID:     "placeholder",
			Result:    "OK",
			ResultMsg: "",
		},
	},
}

// TODO(doug4j@gmail.com): This will eventually be abstracted out from handler testing.
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

func setupHTTPEnv() {
	db, err := createAndVerifyDBConn(testDbUser, testDbPassword, testDbHost, testDbName, testDbPort)
	if err != nil {
		syncutil.Fatal(db)
		panic("Cannot create database connection")
	}
	handlers := Handlers{
		Repository: syncapi.Repository{
			DataRepo: syncdaopq.NewDataRepository(db),
		},
	}
	server = httptest.NewServer(NewRouter(handlers)) //Creating new server with the user handlers
	getSyncConfigURLOK = fmt.Sprintf("%s/syncPairConfig/node1Name/A/node2Name/Z", server.URL)
	getSyncConfigURLNotPaired = fmt.Sprintf("%s/syncPairConfig/node1Name/A/node2Name/C", server.URL)
	changeQueueURLOK = fmt.Sprintf("%s/changeQueue/sessionId/mySession1/nodeId/*node-spoke1/msgId/myMsgId1", server.URL)
	sessionBaseURL = fmt.Sprintf("%s/syncSession/", server.URL)
	syncPairStateURL = fmt.Sprintf("%s/syncPairState/", server.URL)
	syncPairConfigURL = fmt.Sprintf("%s/syncPairConfig/", server.URL)
	syncDataBaseURL = fmt.Sprintf("%s/syncData/", server.URL)
	setupDatabaseObj()
}

func teardownHTTPEnv() {
	teardownDatabaseObj()
}

func TestHandlers_Index(t *testing.T) {
	testName := "TestHandlers_Index"
	testhelper.StartTest(testName)
	setupHTTPEnv()
	defer teardownHTTPEnv()

	var reader io.Reader
	request, err := http.NewRequest("GET", server.URL, reader)

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		syncutil.Fatal(err)
		panic(err.Error())
	}

	if res.StatusCode != http.StatusOK {
		t.Error("Result incorrect. Actual Response=" + res.Status + ". Using url='" + server.URL + "'")
		return
	}
	response, err := ioutil.ReadAll(res.Body)
	fmt.Printf("response=%s", response)
	res.Body.Close()
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	testhelper.EndTest(testName)
}

func TestHandlers_Node(t *testing.T) {
	testName := "TestHandlers_Node"
	testhelper.StartTest(testName)

	setupHTTPEnv()
	defer teardownHTTPEnv()

	var nilReader io.Reader

	//Create Node
	nodeUUID := uuid.NewV4()
	nodeID := uuid.Formatter(nodeUUID, uuid.CleanHyphen)

	node := syncdao.SyncNode{
		NodeID:          nodeID,
		NodeName:        "some-new-node",
		DataVersionName: "Demo Model 1",
	}

	buf, err := json.Marshal(node)
	if err != nil {
		t.Error("Sending Json is not correct")
		return
	}

	request, err := http.NewRequest("POST", server.URL+"/syncNode", bytes.NewBuffer(buf))
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		syncutil.Fatal(err)
		t.Error("Sending Json is not correct")
	}

	if res.StatusCode != http.StatusOK {
		t.Error("Result incorrect. Actual Response=" + res.Status + ". Using url='" + server.URL + "'")
		return
	}
	response, err := ioutil.ReadAll(res.Body)
	fmt.Printf("response=%s", response)
	res.Body.Close()
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	//Read by Node Id
	request, err = http.NewRequest("GET", server.URL+"/syncNode/nodeId/"+nodeID, nilReader)
	res, err = http.DefaultClient.Do(request)
	if err != nil {
		syncutil.Fatal(err)
		t.Errorf("%s", err.Error())
	}

	if res.StatusCode != http.StatusOK {
		t.Error("Result incorrect. Actual Response=" + res.Status + ". Using url='" + server.URL + "'")
		return
	}
	decoder := json.NewDecoder(res.Body)
	var actual NodeResponse
	err = decoder.Decode(&actual)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
		return
	}

	expected := serviceAnswersExpecations["NodeResponseOK"].ExpectedValue.(NodeResponse)
	expected.NodeID = nodeID
	expected.MsgID = nodeID
	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("Result incorrect. Expected=%s. Actual=%s.", expected, actual)
	}

	//Read by Node Name
	request, err = http.NewRequest("GET", server.URL+"/syncNode/nodeName/some-new-node", nilReader)
	res, err = http.DefaultClient.Do(request)
	if err != nil {
		syncutil.Fatal(err)
		t.Error(err.Error())
	}

	if res.StatusCode != http.StatusOK {
		t.Error("Result incorrect. Actual Response=" + res.Status + ". Using url='" + server.URL + "'")
		return
	}
	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&actual)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
		return
	}
	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("Result incorrect. Expected=%s. Actual=%s.", expected, actual)
	}

	//Delete Node
	request, err = http.NewRequest("DELETE", server.URL+"/syncNode/nodeId/"+nodeID, nilReader)
	res, err = http.DefaultClient.Do(request)
	if err != nil {
		syncutil.Fatal(err)
		t.Error(err.Error())
	}

	if res.StatusCode != http.StatusOK {
		t.Error("Result incorrect. Actual Response=" + res.Status + ". Using url='" + server.URL + "'")
		return
	}
	decoder = json.NewDecoder(res.Body)
	var actualDelete DeleteNodeResponse
	err = decoder.Decode(&actualDelete)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
		return
	}
	expectedDelete := serviceAnswersExpecations["DeleteNodeResponseOK"].ExpectedValue.(DeleteNodeResponse)
	expectedDelete.NodeID = nodeID
	expectedDelete.MsgID = nodeID
	if reflect.DeepEqual(actualDelete, expectedDelete) == false {
		t.Errorf("Result incorrect. Expected=%s. Actual=%s.", expectedDelete, actualDelete)
	}
	testhelper.EndTest(testName)
}

func TestHandlers_GetSyncConfigOK(t *testing.T) {
	testName := "TestHandlers_GetSyncConfigOK"
	testhelper.StartTest(testName)

	setupHTTPEnv()
	defer teardownHTTPEnv()

	var reader io.Reader
	request, err := http.NewRequest("GET", getSyncConfigURLOK, reader)

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		t.Errorf("Http client failed, error: %s", err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Result incorrect. Actual Status=%s. Using url:'%s'", res.Status, getSyncConfigURLOK)
		return
	}

	decoder := json.NewDecoder(res.Body)
	var actual PairConfigResponse
	err = decoder.Decode(&actual)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
	}
	expected := serviceAnswersExpecations["GetSyncConfigOK"].ExpectedValue
	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("Result incorrect. Expected=%v. Actual=%v.", expected, actual)
	}
	testhelper.EndTest(testName)
}

func TestHandlers_GetSyncConfigNotPaired(t *testing.T) {
	testName := "TestHandlers_GetSyncConfigNotPaired"
	testhelper.StartTest(testName)

	setupHTTPEnv()
	defer teardownHTTPEnv()

	var reader io.Reader
	request, err := http.NewRequest("GET", getSyncConfigURLNotPaired, reader)

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		t.Errorf("Http client failed, error: %s", err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Result incorrect. Actual Status=%s. Using url:'%s'", res.Status, getSyncConfigURLNotPaired)
		return
	}

	decoder := json.NewDecoder(res.Body)
	var actual PairConfigResponse
	err = decoder.Decode(&actual)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
	}
	expected := serviceAnswersExpecations["GetSyncConfigNotPaired"].ExpectedValue
	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("Result incorrect. Expected=%v. Actual=%v.", expected, actual)
	}
	testhelper.EndTest(testName)
}

func TestHandlers_GetSyncConfigDatabaseNotSet(t *testing.T) {
	testName := "TestHandlers_GetSyncConfigDatabaseNotSet"
	testhelper.StartTest(testName)

	setupHTTPEnv()
	//Purposefully have a configuration error
	syncdao.DefaultDaos = nil
	defer teardownHTTPEnv()

	var reader io.Reader
	request, err := http.NewRequest("GET", getSyncConfigURLOK, reader)

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		t.Errorf("Http client failed, error: %s", err.Error())
	}
	if res.StatusCode != http.StatusInternalServerError {
		response, err := ioutil.ReadAll(res.Body)
		fmt.Printf("response=%s", response)
		res.Body.Close()
		if err != nil {
			t.Errorf("%s", err.Error())
		}

		t.Errorf("Result incorrect. Actual Status=%s. Using url:'%s'", res.Status, getSyncConfigURLOK)
		return
	}
	response, err := ioutil.ReadAll(res.Body)
	var actual = string(response)
	actual = strings.TrimSpace(actual)
	res.Body.Close()
	if err != nil {
		syncutil.Error(err.Error())
		t.Errorf("%s", err.Error())
	}
	//var actual = res.Body
	expected := serviceAnswersExpecations["GetSyncConfigDatabaseNotSet"].ExpectedValue
	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("Result incorrect. Expected=%s. Actual=%s.", expected, actual)
	}
	testhelper.EndTest(testName)
}

func TestHandlers_Session(t *testing.T) {
	testName := "TestHandlers_Session"
	testhelper.StartTest(testName)

	setupHTTPEnv()
	defer teardownHTTPEnv()

	//Create New Sesssion
	sessionUUID := uuid.NewV4()
	sessionID := uuid.Formatter(sessionUUID, uuid.CleanHyphen)
	var reader io.Reader
	sessionURL := fmt.Sprintf(sessionBaseURL+"sessionId/%s/pairId/%s", sessionID, "*pair-1")
	request, err := http.NewRequest("POST", sessionURL, reader)
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Errorf("Http client failed, error: %s", err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Result incorrect. Actual Status=%s. Using url:'%s'", res.Status, sessionURL)
		return
	}
	decoder := json.NewDecoder(res.Body)
	var actual CreateSyncSessionResponse
	err = decoder.Decode(&actual)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
		return
	}
	expected, ok := serviceAnswersExpecations["CreateSession"].ExpectedValue.(CreateSyncSessionResponse)
	if !ok {
		t.Errorf("Failed to cast CreateSyncSessionResponse")
		return
	}

	expected.RequestedSessionID = sessionID
	expected.ActualSessionID = sessionID
	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("Result incorrect. Expected=%s. Actual=%s.", expected, actual)
	}

	//Create New Session with CreateSessionThisSessionIdAlreadyActive
	sessionURL = fmt.Sprintf(sessionBaseURL+"sessionId/%s/pairId/%s", sessionID, "*pair-1")
	request, err = http.NewRequest("POST", sessionURL, reader)
	res, err = http.DefaultClient.Do(request)

	if err != nil {
		t.Errorf("Http client failed, error: %s", err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Result incorrect. Actual Status=%s. Using url:'%s'", res.Status, sessionURL)
		return
	}
	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&actual)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
		return
	}
	expected.Result = "ThisSessionIdAlreadyActive"
	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("Result incorrect. Expected=%s. Actual=%s.", expected, actual)
	}

	//Create New Session with DifferentSessionIdAlreadyExists
	diffSessionUUID := uuid.NewV4()
	diffSessionID := uuid.Formatter(diffSessionUUID, uuid.CleanHyphen)

	sessionURL = fmt.Sprintf(sessionBaseURL+"sessionId/%s/pairId/%s", diffSessionID, "*pair-1")
	request, err = http.NewRequest("POST", sessionURL, reader)
	res, err = http.DefaultClient.Do(request)

	if err != nil {
		t.Errorf("Http client failed, error: %s", err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Result incorrect. Actual Status=%s. Using url:'%s'", res.Status, sessionURL)
		return
	}
	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&actual)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
		return
	}
	expected.Result = "DifferentSessionIdAlreadyActive"
	expected.RequestedSessionID = diffSessionID
	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("Result incorrect. Expected=%s. Actual=%s.", expected, actual)
	}

	//Create New Session Not Paired
	diffSessionUUID = uuid.NewV4()
	diffSessionID = uuid.Formatter(diffSessionUUID, uuid.CleanHyphen)
	var badPairID = "Non-existent-pair-id"
	sessionURL = fmt.Sprintf(sessionBaseURL+"sessionId/%s/pairId/%s", diffSessionID, badPairID)
	request, err = http.NewRequest("POST", sessionURL, reader)
	res, err = http.DefaultClient.Do(request)

	if err != nil {
		t.Errorf("Http client failed, error: %s", err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Result incorrect. Actual Status=%s. Using url:'%s'", res.Status, sessionURL)
		return
	}
	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&actual)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
		return
	}
	expected.PairID = badPairID
	expected.Result = "CreateSyncSessionUnknownError"
	expected.ActualSessionID = ""
	expected.RequestedSessionID = diffSessionID
	expected.ResultMsg = "Could not persist data to database:sql: no rows in result set"
	if reflect.DeepEqual(actual, expected) == false {
		t.Errorf("Result incorrect. Expected=%s. Actual=%s.", expected, actual)
	}

	//Query session state 1 of 3
	sessionURL = fmt.Sprintf(syncPairStateURL+"pairId/%s", "*pair-1")
	request, err = http.NewRequest("GET", sessionURL, reader)
	res, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Errorf("Http client failed, error: %s", err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Result incorrect. Actual Status=%s. Using url:'%s'", res.Status, sessionURL)
		return
	}
	decoder = json.NewDecoder(res.Body)

	var actualQuery QueryPairStateResponse
	err = decoder.Decode(&actualQuery)
	//log.Println("TestHandlers_Session.actualQuery", actualQuery)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
		return
	}
	expectedQuery, ok := serviceAnswersExpecations["QuerySession"].ExpectedValue.(QueryPairStateResponse)
	if !ok {
		t.Errorf("Failed to cast CreateSyncSessionResponse")
		return
	}
	expectedQuery.State = "Initializing"
	expectedQuery.SessionID = sessionID
	expectedQuery.SessionStart = 0
	expectedQuery.LastUpdated = 0
	expectedQuery.SessionStart = 0
	if reflect.DeepEqual(actualQuery, expectedQuery) == false {
		t.Errorf("Result incorrect. Expected=%s. Actual=%s.", expected, actual)
	}

	//Update session state
	sessionURL = fmt.Sprintf(sessionBaseURL+"sessionId/%s/pairId/%s/state/Queuing", sessionID, "*pair-1")
	request, err = http.NewRequest("PUT", sessionURL, reader)
	res, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Errorf("Http client failed, error: %s", err.Error())
		return
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Result incorrect. Actual Status=%s. Using url:'%s'", res.Status, sessionURL)
		return
	}
	decoder = json.NewDecoder(res.Body)
	var actualUpdate UpdateSyncSessionStateResponse
	err = decoder.Decode(&actualUpdate)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
		return
	}
	expectedUpdate, ok := serviceAnswersExpecations["UpdateSession"].ExpectedValue.(UpdateSyncSessionStateResponse)
	if !ok {
		t.Errorf("Failed to cast CreateSyncSessionResponse")
		return
	}
	expectedUpdate.SessionID = sessionID
	if reflect.DeepEqual(actualUpdate, expectedUpdate) == false {
		t.Errorf("Result incorrect. Expected=%s. Actual=%s.", expected, actual)
		return
	}
	//log.Println("TestHandlers_Session.actualUpdate", actualUpdate)

	//Query session state 2 of 3
	sessionURL = fmt.Sprintf(syncPairStateURL+"pairId/%s", "*pair-1")
	request, err = http.NewRequest("GET", sessionURL, reader)
	res, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Errorf("Http client failed, error: %s", err.Error())
		return
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Result incorrect. Actual Status=%s. Using url:'%s'", res.Status, sessionURL)
		return
	}
	decoder = json.NewDecoder(res.Body)

	err = decoder.Decode(&actualQuery)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
		return
	}
	//log.Println("TestHandlers_Session.actualQuery", actualQuery)
	expectedQuery.State = "Queuing"
	expectedQuery.SessionID = sessionID
	expectedQuery.SessionStart = 0
	expectedQuery.LastUpdated = 0
	expectedQuery.SessionStart = 0
	if reflect.DeepEqual(actualQuery, expectedQuery) == false {
		t.Errorf("Result incorrect. Expected=%s. Actual=%s.", expected, actual)
	}

	//Close session state
	sessionURL = fmt.Sprintf(sessionBaseURL+"sessionId/%s/pairId/%s", sessionID, "*pair-1")
	request, err = http.NewRequest("DELETE", sessionURL, reader)
	res, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Errorf("Http client failed, error: %s", err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Result incorrect. Actual Status=%s. Using url:'%s'", res.Status, sessionURL)
		return
	}
	decoder = json.NewDecoder(res.Body)
	var actualDelete CloseSyncSessionResponse
	err = decoder.Decode(&actualDelete)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
		return
	}
	//log.Println("TestHandlers_Session.actualDelete", actualDelete)
	expectedDelete, ok := serviceAnswersExpecations["CloseSession"].ExpectedValue.(CloseSyncSessionResponse)
	if !ok {
		t.Errorf("Failed to cast CreateSyncSessionResponse")
		return
	}
	expectedDelete.RequestedSessionID = sessionID
	expectedDelete.ActualSessionID = sessionID
	if reflect.DeepEqual(actualDelete, expectedDelete) == false {
		t.Errorf("Result incorrect. Expected=%s. Actual=%s.", expectedDelete, expectedDelete)
	}

	//Query session state 3 of 3
	sessionURL = fmt.Sprintf(syncPairStateURL+"pairId/%s", "*pair-1")
	request, err = http.NewRequest("GET", sessionURL, reader)
	res, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Errorf("Http client failed, error: %s", err.Error())
		return
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Result incorrect. Actual Status=%s. Using url:'%s'", res.Status, sessionURL)
		return
	}
	decoder = json.NewDecoder(res.Body)

	err = decoder.Decode(&actualQuery)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
		return
	}
	//log.Println("TestHandlers_Session.actualQuery", actualQuery)
	expectedQuery.State = "Inactive"
	expectedQuery.SessionID = ""
	expectedQuery.SessionStart = 0
	expectedQuery.LastUpdated = 0
	expectedQuery.SessionStart = 0
	if reflect.DeepEqual(actualQuery, expectedQuery) == false {
		t.Errorf("Result incorrect. Expected=%s. Actual=%s.", expected, actual)
	}
	testhelper.EndTest(testName)
}

func Test_DataHex1(t *testing.T) {

	creator := syncmsg.NewCreator()
	contact := testhelper.Contact{"0934A378-DEDB-4207-B99C-DD0D61DC59BC", creator.FormatTimeFromString("1987-05-28 00:00:00.000"), "Mindy", 5, 1.0, "Johnson", 2}
	contactSyncPackage, err := testhelper.CreateRecordAndSupport(contact, false, syncmsg.SentSyncStateEnum_PersistedStandardSentToPeer, "")
	if err != nil {
		syncutil.Fatal("Marshaling error: ", err)
		return //This return will never get called as the above panics
	}
	expected := "a413aa0a942e01f6c2f1606f4eb1799733df1a21af46e9f1218b775455c990a7"

	if contactSyncPackage.RecordSha256Hex != expected {
		t.Errorf("resulting hash is not correct. Should be %v. Actual: '%v'", expected, contactSyncPackage.RecordSha256Hex)
	}

	syncutil.Debug("RecordSha256Hex:", contactSyncPackage.RecordSha256Hex)
}

func Test_DataHex2(t *testing.T) {

	creator := syncmsg.NewCreator()
	contact := testhelper.Contact{"0934A378-DEDB-4207-B99C-DD0D61DC59BC", creator.FormatTimeFromString("1987-05-28 00:00:00.000"), "Mindy", 5, 2.0, "Johnson", 1}
	contactSyncPackage, err := testhelper.CreateRecordAndSupport(contact, false, syncmsg.SentSyncStateEnum_PersistedStandardSentToPeer, "")
	if err != nil {
		syncutil.Fatal("Marshaling error: ", err)
		return //This return will never get called as the above panics
	}
	//expected := "a413aa0a942e01f6c2f1606f4eb1799733df1a21af46e9f1218b775455c990a7"
	expected := "c7f5967ca5f9c0c1b6fd4b9cca1d3c2de5764f57ab875554bfef0794bb78f2c9"

	if contactSyncPackage.RecordSha256Hex != expected {
		t.Errorf("resulting hash is not correct. Should be %v. Actual: '%v'", expected, contactSyncPackage.RecordSha256Hex)
	}

	syncutil.Debug("RecordSha256Hex:", contactSyncPackage.RecordSha256Hex)

}
