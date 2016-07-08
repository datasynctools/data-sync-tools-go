package synchandler

// TODO(doug4j@gmail.com): Transition handler testing to use a mock/stub syncdao rather than syncdaopq
import (
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncmsg"
	"data-sync-tools-go/syncutil"
	"data-sync-tools-go/testhelper"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func readyFetchSyncDataHasData() *httptest.Server {
	dataRepo := mockDataRepository{
		fetcher: mockSyncMessageFetcher{
			fetchAnswer: &syncmsg.ProtoRequestSyncEntityMessageResponse{
				Result:    syncmsg.SyncRequestEntityMessageResponseResult_HasMsgs.Enum(),
				ResultMsg: proto.String(""),
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
		createFetcherError: nil,
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
	server = httptest.NewServer(NewRouter(handlers)) //Creating new server with the user handlers
	return server
}

func TestHandlers_FetchSyncDataHasData(t *testing.T) {
	testName := syncutil.GetCallingName()
	testhelper.StartTest(testName)

	server = readyFetchSyncDataHasData()
	var reader io.Reader

	request, err := http.NewRequest("GET", server.URL+"/fetchData/sessionId/F32A9BB9-7691-4F20-B337-55414E45B1D1/nodeId/9702F991-F1E7-4186-A97B-8B804F723F87/orderNum/1/changeType/AddOrUpdate", reader)
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		syncutil.Fatal(err)
		t.Error(err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Error("Result incorrect. Actual Response=" + res.Status + ". Using url='" + server.URL + "'")
		return
	}
	responseBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		syncutil.Fatal("Error: " + err.Error())
		t.Error(err.Error())
	}

	responseData := &syncmsg.ProtoRequestSyncEntityMessageResponse{}
	err = proto.Unmarshal(responseBytes, responseData)
	if err != nil {
		t.Errorf("Error decoding response: %s", err.Error())
	}
	//syncutil.Info("responseData:", responseData)
	expectedResult := syncmsg.SyncRequestEntityMessageResponseResult_HasMsgs.String()
	if responseData.Result.String() != expectedResult {
		t.Errorf("Result should be: %v. Not: %v", expectedResult, responseData.Result)
		//return
	}

	// TODO(doug4j@gmail.com): Add more test validation

	testhelper.EndTest(testName)
}

func readyFetchSyncDataNoMsgs() *httptest.Server {
	dataRepo := mockDataRepository{
		fetcher: mockSyncMessageFetcher{
			fetchAnswer: &syncmsg.ProtoRequestSyncEntityMessageResponse{
				Result:    syncmsg.SyncRequestEntityMessageResponseResult_NoMsgs.Enum(),
				ResultMsg: proto.String(""),
				Request: &syncmsg.ProtoSyncEntityMessageRequest{
					IsDelete:          proto.Bool(false),
					TransactionBindId: proto.String("some-bind-id"),
					Items:             []*syncmsg.ProtoSyncDataMessagesRequest{},
				},
			},
			fetchError: nil,
		},
		createFetcherError: nil,
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
	server = httptest.NewServer(NewRouter(handlers)) //Creating new server with the user handlers
	return server
}
func TestHandlers_FetchSyncDataNoMsgs(t *testing.T) {
	testName := syncutil.GetCallingName()
	testhelper.StartTest(testName)

	server = readyFetchSyncDataNoMsgs()
	var reader io.Reader

	request, err := http.NewRequest("GET", server.URL+"/fetchData/sessionId/F32A9BB9-7691-4F20-B337-55414E45B1D1/nodeId/9702F991-F1E7-4186-A97B-8B804F723F87/orderNum/1/changeType/AddOrUpdate", reader)
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		syncutil.Fatal(err)
		t.Error(err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Error("Result incorrect. Actual Response=" + res.Status + ". Using url='" + server.URL + "'")
		return
	}
	responseBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		syncutil.Fatal("Error: " + err.Error())
		t.Error(err.Error())
	}

	responseData := &syncmsg.ProtoRequestSyncEntityMessageResponse{}
	err = proto.Unmarshal(responseBytes, responseData)
	if err != nil {
		t.Errorf("Error decoding response: %s", err.Error())
	}
	syncutil.Info("responseData:", responseData)
	expectedResult := syncmsg.SyncRequestEntityMessageResponseResult_NoMsgs.String()
	if responseData.Result.String() != expectedResult {
		t.Errorf("Result should be: %v. Not: %v", expectedResult, responseData.Result)
		//return
	}

	// TODO(doug4j@gmail.com): Add more test validation

	testhelper.EndTest(testName)
}

func readyFetchSyncDataErrorCreatingMsgs(expectedErrorMsg string) *httptest.Server {

	dataRepo := mockDataRepository{
		fetcher: mockSyncMessageFetcher{
			fetchAnswer: &syncmsg.ProtoRequestSyncEntityMessageResponse{
				Request: &syncmsg.ProtoSyncEntityMessageRequest{
					IsDelete:          proto.Bool(false),
					TransactionBindId: proto.String("some-bind-id"),
					Items:             []*syncmsg.ProtoSyncDataMessagesRequest{},
				},
			},
			fetchError: errors.New(expectedErrorMsg),
		},
		createFetcherError: nil,
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
	server = httptest.NewServer(NewRouter(handlers)) //Creating new server with the user handlers
	return server
}

func TestHandlers_FetchSyncDataErrorCreatingMsgs(t *testing.T) {
	testName := syncutil.GetCallingName()
	testhelper.StartTest(testName)

	expectedErrorMsg := "Some Error"
	server = readyFetchSyncDataErrorCreatingMsgs(expectedErrorMsg)

	var reader io.Reader

	request, err := http.NewRequest("GET", server.URL+"/fetchData/sessionId/F32A9BB9-7691-4F20-B337-55414E45B1D1/nodeId/9702F991-F1E7-4186-A97B-8B804F723F87/orderNum/1/changeType/AddOrUpdate", reader)
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		syncutil.Fatal(err)
		t.Error(err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Error("Result incorrect. Actual Response=" + res.Status + ". Using url='" + server.URL + "'")
		return
	}
	responseBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		syncutil.Fatal("Error: " + err.Error())
		t.Error(err.Error())
	}

	responseData := &syncmsg.ProtoRequestSyncEntityMessageResponse{}
	err = proto.Unmarshal(responseBytes, responseData)
	if err != nil {
		t.Errorf("Error decoding response: %s", err.Error())
	}
	syncutil.Info("responseData:", responseData)
	expectedResult := syncmsg.SyncRequestEntityMessageResponseResult_ErrorCreatingMsgs.String()
	if responseData.Result.String() != expectedResult {
		t.Errorf("Result should be: %v. Not: %v", expectedResult, responseData.Result)
		//return
	}
	expectedResultMsg := expectedErrorMsg
	actualResponseData := *responseData.ResultMsg
	if actualResponseData != expectedResultMsg {
		t.Errorf("Result should be: %v. Not: %v", expectedResultMsg, actualResponseData)
		//return
	}

	// TODO(doug4j@gmail.com): Add more test validation

	testhelper.EndTest(testName)
}

func TestHandlers_FetchSyncDataMissingOrderNumParm(t *testing.T) {
	testName := syncutil.GetCallingName()
	testhelper.StartTest(testName)

	var server *httptest.Server
	handlers := Handlers{}
	handlers.VarsHandler = func(r *http.Request) map[string]string {
		return map[string]string{
			"sessionId": "F32A9BB9-7691-4F20-B337-55414E45B1D1",
			"nodeId":    "F32A9BB9-7691-4F20-B337-55414E45B1D1",
		}
	}
	server = httptest.NewServer(NewRouter(handlers)) //Creating new server with the user handlers
	var reader io.Reader

	request, err := http.NewRequest("GET", server.URL+"/fetchData/sessionId/F32A9BB9-7691-4F20-B337-55414E45B1D1/nodeId/9702F991-F1E7-4186-A97B-8B804F723F87/orderNum/1/changeType/AddOrUpdate", reader)

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		syncutil.Fatal(err)
		t.Error(err.Error())
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Error("Result incorrect. Actual Response=" + res.Status + ". Using url='" + request.URL.String() + "'")
		return
	}
	responseBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		syncutil.Fatal("Error: " + err.Error())
		t.Error(err.Error())
	}
	actualResponse := string(responseBytes)
	expectedResponse := MissingParamOrderNum + "\n"
	assert.Equal(t, expectedResponse, actualResponse, "body response should be equal")

	testhelper.EndTest(testName)
}

func TestHandlers_FetchSyncDataMissingChangeTypeParm(t *testing.T) {
	testName := syncutil.GetCallingName()
	testhelper.StartTest(testName)

	var server *httptest.Server
	handlers := Handlers{}
	handlers.VarsHandler = func(r *http.Request) map[string]string {
		return map[string]string{
			"sessionId": "F32A9BB9-7691-4F20-B337-55414E45B1D1",
			"nodeId":    "F32A9BB9-7691-4F20-B337-55414E45B1D1",
			"orderNum":  "3",
		}
	}
	server = httptest.NewServer(NewRouter(handlers)) //Creating new server with the user handlers
	var reader io.Reader

	request, err := http.NewRequest("GET", server.URL+"/fetchData/sessionId/F32A9BB9-7691-4F20-B337-55414E45B1D1/nodeId/9702F991-F1E7-4186-A97B-8B804F723F87/orderNum/1/changeType/AddOrUpdate", reader)

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		syncutil.Fatal(err)
		t.Error(err.Error())
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Error("Result incorrect. Actual Response=" + res.Status + ". Using url='" + request.URL.String() + "'")
		return
	}
	responseBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		syncutil.Fatal("Error: " + err.Error())
		t.Error(err.Error())
	}
	actualResponse := string(responseBytes)
	expectedResponse := MissingParamChangeType + "\n"
	assert.Equal(t, expectedResponse, actualResponse, "body response should be equal")

	testhelper.EndTest(testName)
}

func TestHandlers_FetchSyncDataMissingNodeIdParm(t *testing.T) {
	testName := syncutil.GetCallingName()
	testhelper.StartTest(testName)

	var server *httptest.Server
	handlers := Handlers{}
	handlers.VarsHandler = func(r *http.Request) map[string]string {
		return map[string]string{
			"sessionId": "F32A9BB9-7691-4F20-B337-55414E45B1D1",
		}
	}
	server = httptest.NewServer(NewRouter(handlers)) //Creating new server with the user handlers
	var reader io.Reader

	request, err := http.NewRequest("GET", server.URL+"/fetchData/sessionId/F32A9BB9-7691-4F20-B337-55414E45B1D1/nodeId/9702F991-F1E7-4186-A97B-8B804F723F87/orderNum/1/changeType/AddOrUpdate", reader)

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		syncutil.Fatal(err)
		t.Error(err.Error())
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Error("Result incorrect. Actual Response=" + res.Status + ". Using url='" + request.URL.String() + "'")
		return
	}
	responseBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		syncutil.Fatal("Error: " + err.Error())
		t.Error(err.Error())
	}
	actualResponse := string(responseBytes)
	expectedResponse := MissingParamNodeID + "\n"
	assert.Equal(t, expectedResponse, actualResponse, "body response should be equal")

	testhelper.EndTest(testName)
}

func TestHandlers_FetchSyncDataMissingSessionIdParm(t *testing.T) {
	testName := syncutil.GetCallingName()
	testhelper.StartTest(testName)

	var server *httptest.Server
	handlers := Handlers{}
	handlers.VarsHandler = func(r *http.Request) map[string]string {
		return map[string]string{}
	}
	server = httptest.NewServer(NewRouter(handlers)) //Creating new server with the user handlers
	var reader io.Reader

	request, err := http.NewRequest("GET", server.URL+"/fetchData/sessionId/F32A9BB9-7691-4F20-B337-55414E45B1D1/nodeId/9702F991-F1E7-4186-A97B-8B804F723F87/orderNum/1/changeType/AddOrUpdate", reader)

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		syncutil.Fatal(err)
		t.Error(err.Error())
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Error("Result incorrect. Actual Response=" + res.Status + ". Using url='" + request.URL.String() + "'")
		return
	}
	responseBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		syncutil.Fatal("Error: " + err.Error())
		t.Error(err.Error())
	}
	actualResponse := string(responseBytes)
	expectedResponse := MissingParamSessionID + "\n"
	assert.Equal(t, expectedResponse, actualResponse, "body response should be equal")

	testhelper.EndTest(testName)
}

func TestHandlers_FetchSyncDataBadOrderNumConversion(t *testing.T) {
	testName := syncutil.GetCallingName()
	testhelper.StartTest(testName)

	var server *httptest.Server
	handlers := Handlers{}

	orderNumStr := "Not-a-Number"

	handlers.VarsHandler = func(r *http.Request) map[string]string {
		return map[string]string{
			"sessionId":  "F32A9BB9-7691-4F20-B337-55414E45B1D1",
			"nodeId":     "F32A9BB9-7691-4F20-B337-55414E45B1D1",
			"orderNum":   orderNumStr,
			"changetype": "AddOrUpdate",
		}
	}
	server = httptest.NewServer(NewRouter(handlers)) //Creating new server with the user handlers
	var reader io.Reader

	request, err := http.NewRequest("GET", server.URL+"/fetchData/sessionId/F32A9BB9-7691-4F20-B337-55414E45B1D1/nodeId/9702F991-F1E7-4186-A97B-8B804F723F87/orderNum/1/changeType/AddOrUpdate", reader)

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		syncutil.Fatal(err)
		t.Error(err.Error())
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Error("Result incorrect. Actual Response=" + res.Status + ". Using url='" + request.URL.String() + "'")
		return
	}
	responseBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		syncutil.Fatal("Error: " + err.Error())
		t.Error(err.Error())
	}
	actualResponse := string(responseBytes)
	//expectedResponse := BadOrderNumConversion + "\n"
	expectedResponse := fmt.Sprintf(BadOrderNumConversion, orderNumStr) + "\n"

	assert.Equal(t, expectedResponse, actualResponse, "body response should be equal")

	testhelper.EndTest(testName)
}
