package synchandler

import (
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncmsg"
	"data-sync-tools-go/syncutil"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
)

// TODO(doug4j@gmail.com): Finish http sample

//ProcessSyncData handles data for syncing on an active session sync. Invoked performed via the following http command:
//	curl -H "Content-Type: application/json" --request PUT http://localhost:8080/syncdata/sessionId/{sessionId}/pairId/{pairId}/nodeId/{nodeId}/isDelete/{isDelete}/transactionBindId/{transactionBindId} -d '{"syncEntityName":"Entity1", "msgs": [{"recordId":"uuid1", "recordHash":"theHash1", "lastKnownPeerHash":"theHash1", "sentSyncState":"0", recordBytesSize:"2049" },{"recordId":"uuid2", "recordHash":"theHash2", "lastKnownPeerHash":"theHash2", "sentSyncState":"0", recordBytesSize:"2049" } ]}'
//	curl -H "Content-Type: application/json" --request PUT http://localhost:8080/changeQueue/sessionId/my-session-id/nodeId/*node-spoke1/msgId/my-msg-id

//ProcessSyncData acccepts data to be processed for syncing.
func (handlers Handlers) ProcessSyncData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var validArgs processSyncDataArgs
	var err error

	validArgs, err = processProcessSyncDataArgs(vars)
	if err != nil {
		msg := err.Error()
		syncutil.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	inputBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		syncutil.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	requestData := &syncmsg.ProtoSyncEntityMessageRequest{}
	err = proto.Unmarshal(inputBytes, requestData)
	if err != nil {
		syncutil.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if validArgs.transactionBindID != *requestData.TransactionBindId {
		msg := fmt.Sprintf("Transaction Id '%s' in request URL is different than that in the data '%s'", validArgs.transactionBindID, *requestData.TransactionBindId)
		//msg := ""
		syncutil.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	msgProcessor, err := createFetcherAndProcessorForProcessSyncData(validArgs, handlers.Repository, requestData)
	if err != nil {
		msg := err.Error()
		syncutil.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	var answer *syncmsg.ProtoSyncEntityMessageResponse
	answer = msgProcessor.Process(requestData)
	syncutil.Debug("processed ", len(answer.Items), " items")
	//syncutil.Debug("answer", answer)

	//syncutil.Info("Result:", *answer.Result, "ResultMsg:", *answer.ResultMsg)
	data, err := proto.Marshal(answer)
	if err != nil {
		syncutil.Error("Error readying response: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//syncutil.Debug(answer)

	_, err = w.Write(data)
	if err != nil {
		syncutil.NotImplementedMsg(err.Error())
	}
	return
}

func processProcessSyncDataArgs(vars map[string]string) (processSyncDataArgs, error) {

	answer := processSyncDataArgs{}
	var hasValue bool

	answer.sessionID, hasValue = vars["sessionId"]
	if !hasValue {
		return answer, errors.New(MissingParamSessionID)
	}
	answer.nodeIDToProcess, hasValue = vars["nodeId"]
	if !hasValue {
		return answer, errors.New(MissingParamNodeID)
	}
	answer.transactionBindID, hasValue = vars["transactionBindId"]
	if !hasValue {
		return answer, errors.New(MissingParamTransBindId)
	}
	return answer, nil
}

type processSyncDataArgs struct {
	sessionID         string
	nodeIDToProcess   string
	transactionBindID string
}

func createFetcherAndProcessorForProcessSyncData(validArgs processSyncDataArgs, repo syncapi.Repository, requestData *syncmsg.ProtoSyncEntityMessageRequest) (syncapi.MessageProcessing, error) {
	var msgProcessor syncapi.MessageProcessing
	var err error

	if repo.ConfigRepo == nil {
		msg := "ConfigRepo is nil"
		syncutil.Error(msg)
		return msgProcessor, errors.New(msg)
	}

	entityFetcher, err := repo.ConfigRepo.CreateEntityFetcher(validArgs.sessionID, validArgs.nodeIDToProcess)
	if err != nil {
		ctxMsg := "Could not create entityFetcher."
		syncutil.Error(ctxMsg, err)
		msg := ctxMsg + err.Error()
		return msgProcessor, errors.New(msg)
	}
	var entitiesByPluralName map[string]syncapi.EntityNameItem

	entitiesByPluralName, err = entityFetcher.FindPluralEntityNamesByID(validArgs.sessionID, validArgs.nodeIDToProcess)
	if err != nil {
		msg := err.Error()
		syncutil.Error(msg)
		return msgProcessor, errors.New(msg)
	}

	msgProcessor, err = repo.DataRepo.CreateMessageProcessor(validArgs.sessionID, validArgs.nodeIDToProcess, entitiesByPluralName)
	if err != nil {
		ctxMsg := "Could not create msgFetcher."
		syncutil.Error(ctxMsg, err)
		msg := ctxMsg + err.Error()
		return msgProcessor, errors.New(msg)
	}
	return msgProcessor, nil
}
