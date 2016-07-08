package synchandler

import (
	"bytes"
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncmsg"
	"data-sync-tools-go/syncutil"
	"data-sync-tools-go/testhelper"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/proto"
)

func readyProcessSyncDataOK() *httptest.Server {
	dataRepo := mockDataRepository{
		fetcher: mockSyncMessageFetcher{
			fetchAnswer: &syncmsg.ProtoRequestSyncEntityMessageResponse{
				Request: &syncmsg.ProtoSyncEntityMessageRequest{
					IsDelete:          proto.Bool(false),
					TransactionBindId: proto.String("some-bind-id"),
					Items: []*syncmsg.ProtoSyncDataMessagesRequest{
						&syncmsg.ProtoSyncDataMessagesRequest{
							EntityPluralName: proto.String("SomeEntityName"),
							Msgs: []*syncmsg.ProtoSyncDataMessageRequest{
								&syncmsg.ProtoSyncDataMessageRequest{
									RecordId:          proto.String("some-record-id"),
									RecordHash:        proto.String("some-record-hash"),
									LastKnownPeerHash: nil,
									SentSyncState:     syncmsg.SentSyncStateEnum_PersistedFirstTimeSentToPeer.Enum(),
									RecordBytesSize:   proto.Uint32(123),
									RecordData:        []byte("some-record-data"),
								},
							},
						},
					},
				},
			},
			fetchError: nil,
		},
		processor: mockMessageProcessor{
			processorAnswer: &syncmsg.ProtoSyncEntityMessageResponse{
				TransactionBindId: proto.String("0F16AEED-E4B4-483E-A7A3-CCABA831FE6E"),
				Result:            syncmsg.SyncEntityMessageResponseResult_OK.Enum(),
				ResultMsg:         proto.String(""),
				Items: []*syncmsg.ProtoSyncDataMessagesResponse{
					&syncmsg.ProtoSyncDataMessagesResponse{
						EntityPluralName: proto.String("Tests"),
						Msgs: []*syncmsg.ProtoSyncDataMessageResponse{&syncmsg.ProtoSyncDataMessageResponse{
							RecordId:        proto.String("recordId1"),
							RequestHash:     proto.String("recordHash1"),
							ResponseHash:    proto.String("recordResponseHash1"),
							SyncState:       syncmsg.AckSyncStateEnum_AckFastBatch.Enum(),
							RecordBytesSize: proto.Uint32(100),
							RecordData:      []byte{},
						},
						},
					},
				},
			},
		},
		//createFetcherError: nil,
	}

	configRepo := mockConfigRepository{
		fetcher: mockEntityFetcher{
			findForFetchAnswer:   []syncapi.EntityNameItem{},
			findForProcessAnswer: map[string]syncapi.EntityNameItem{},
			findForFetchError:    nil,
			findForProcessError:  nil,
		},
		createFetcherError: nil,
	}

	var server *httptest.Server
	handlers := Handlers{
		Repository: syncapi.Repository{
			DataRepo:   dataRepo,
			ConfigRepo: configRepo,
		},
	}
	syncutil.Debug("handlers", handlers.Repository.ConfigRepo)
	server = httptest.NewServer(NewRouter(handlers)) //Creating new server with the user handlers
	return server
}

func TestHandlers_ProcessSyncDataOK(t *testing.T) {
	testName := syncutil.GetCallingName()
	testhelper.StartTest(testName)
	// setupHTTPEnv()
	// defer teardownHTTPEnv()
	// syncutil.Debug("\nReady\n")
	server = readyProcessSyncDataOK()
	//var reader io.Reader
	//Prepare a data for URL.
	requestData, err := testhelper.CreateSyncData()
	if err != nil {
		syncutil.Fatal("Cannot create request data: ", err)
		t.Error(err.Error())
		return
	}
	//log.Println("TestHandlers_DataSync.requestData", requestData)
	data, err := proto.Marshal(requestData)
	if err != nil {
		syncutil.Fatal("Cannot marshal request data: ", err)
		t.Error(err.Error())
		return
	}
	var b bytes.Buffer
	b.Write(data)

	// Now that you have a form, you can submit it to your handler.
	sessionID := "A0DF65CC-6632-4465-B52D-98C362C687C5"
	nodeID := "*node-spoke1" //*node-spoke1 == "A" "FD6E8797-E0D4-4222-91F5-535E1BFC9AD5"
	//transactionBindID := "EE037DB1-9984-4CDC-A164-B244FF534CA1"
	transactionBindID := *requestData.TransactionBindId
	sessionURL := fmt.Sprintf(server.URL+
		"/syncData/sessionId/%s/nodeId/%s/transactionBindId/%s", sessionID, nodeID, transactionBindID)

	//log.Println("TestHandlers_DataSync", "Calling [" + sessionUrl + "]")
	req, err := http.NewRequest("PUT", sessionURL, &b)
	if err != nil {
		syncutil.Error(err)
		t.Error(err.Error())
		return
	}

	// Submit the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		syncutil.Error(err)
		t.Error(err.Error())
		return
	}

	responseBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		syncutil.Error("Error: " + err.Error())
		t.Error(err.Error())
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s. msg: %v", res.Status, string(responseBytes))
		syncutil.Error(err)
		t.Error(err.Error())
		return
	}

	responseData := &syncmsg.ProtoSyncEntityMessageResponse{}
	err = proto.Unmarshal(responseBytes, responseData)
	if err != nil {
		t.Errorf("Error decoding response: %s", err.Error())
		return
	}
	//syncutil.Debug("responseData: ", responseData)

	testhelper.EndTest(testName)
}
