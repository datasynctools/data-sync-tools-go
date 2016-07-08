package synchandler

import (
	"data-sync-tools-go/syncmsg"
	"data-sync-tools-go/syncutil"
	"errors"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
)

import "net/http"

//AcknowledgeSyncData marks the final results of a processed sync data batch.
func (handlers Handlers) AcknowledgeSyncData(w http.ResponseWriter, r *http.Request) {
	syncutil.Info("Entering AcknowledgeSyncData")

	vars := mux.Vars(r)

	var validArgs ackSyncDataArgs
	var err error

	validArgs, err = processAckSyncDataArgs(vars)
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

	requestData := &syncmsg.ProtoSyncEntityMessageResponse{}
	err = proto.Unmarshal(inputBytes, requestData)
	if err != nil {
		syncutil.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	syncutil.Info("Hi from AcknowledgeSyncData, validArgs:", validArgs, ", requestData:", requestData)
}

func processAckSyncDataArgs(vars map[string]string) (ackSyncDataArgs, error) {

	answer := ackSyncDataArgs{}
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

type ackSyncDataArgs struct {
	sessionID         string
	nodeIDToProcess   string
	transactionBindID string
}
