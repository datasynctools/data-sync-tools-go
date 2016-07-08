package synchandler

import (
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncmsg"
	"data-sync-tools-go/syncutil"
	"data-sync-tools-go/testhelper"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/proto"
)

func readyQueuingDataOK() *httptest.Server {
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
		queuer: mockMessageQueuer{
			queuerAnswer: 3,
		},
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

func TestHandlers_ChangeQueueOK(t *testing.T) {
	testName := syncutil.GetCallingName()
	testhelper.StartTest(testName)

	server = readyQueuingDataOK()

	// Now that you have a form, you can submit it to your handler.
	sessionID := "A0DF65CC-6632-4465-B52D-98C362C687C5"
	nodeID := "*node-spoke1" //*node-spoke1 == "A" "FD6E8797-E0D4-4222-91F5-535E1BFC9AD5"
	transactionBindID := "EE037DB1-9984-4CDC-A164-B244FF534CA1"

	sessionURL := fmt.Sprintf(server.URL+
		"/changeQueue/sessionId/%s/nodeId/%s/msgId/%s", sessionID, nodeID, transactionBindID)
	syncutil.Debug(sessionURL)
	var reader io.Reader
	request, err := http.NewRequest("PUT", sessionURL, reader)
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		syncutil.Error(err)
		t.Error(err.Error())
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		responseBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			syncutil.Error("Error: " + err.Error())
			t.Error(err.Error())
			return
		}
		err = fmt.Errorf("bad status: %s. msg: %v", res.Status, string(responseBytes))
		syncutil.Error(err)
		t.Error(err.Error())
		return
	}

	decoder := json.NewDecoder(res.Body)
	var actual QueueSyncChangesResponse
	err = decoder.Decode(&actual)
	if err != nil {
		t.Errorf("Error decoding json: %s", err.Error())
	}

	if actual.RecordsQueuedCount != 3 {
		t.Errorf("Response should be 3. But, it is %v", actual.RecordsQueuedCount)
	}

	syncutil.Debug(actual)

	testhelper.EndTest(testName)
}
