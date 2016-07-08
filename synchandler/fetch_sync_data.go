package synchandler

import (
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncmsg"
	"data-sync-tools-go/syncutil"
	"errors"
	"fmt"
	"strconv"
	//"encoding/json"
	//"fmt"
	"github.com/golang/protobuf/proto"
	//"io"

	"net/http"
	//"strconv"
)

//FetchSyncData retrieves data to be processed for syncing.
//func FetchSyncData(w http.ResponseWriter, r *http.Request) {

//FetchSyncData retrieves data to be processed for syncing.
//	{httpbase}/fetchData/sessionId/{sessionId}/nodeId/{nodeId}/orderItem/{orderItem}/{changeType:AddOrUpdate|Delete}/
//	curl -H "Content-Type: application/json" --request GET http://localhost:8080/fetchData/sessionId/8AD65ECB-5826-4CC9-B54D-0723A7FC99B9/nodeId/B616BEB5-341E-4701-862F-006EEA0230C0/orderItem/{orderItem}/changeType/AddOrUpdate
func (handlers Handlers) FetchSyncData(w http.ResponseWriter, r *http.Request) {
	//syncutil.Info("requestURI:", r.RequestURI)
	vars := handlers.VarsHandler(r) //mux.Vars(r)
	//syncutil.Info("vars:", vars)

	var validArgs fetchSyncDataArgs
	validArgs, err := processFetchSyncDataArgs(vars)
	if err != nil {
		msg := err.Error()
		syncutil.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	entityFetcher, msgFetcher, err := createFetchersForProcessFetchSyncData(validArgs, handlers.Repository)
	if err != nil {
		msg := err.Error()
		syncutil.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	entities, err := entityFetcher.FindEntitiesForFetch(int(validArgs.orderNum), validArgs.sessionID, validArgs.nodeIDToProcess, validArgs.changeType)
	if err != nil {
		errMsg := err.Error()
		syncutil.Error(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
	}

	var answer *syncmsg.ProtoRequestSyncEntityMessageResponse

	answer, err = msgFetcher.Fetch(entities, validArgs.changeType)
	if err != nil {
		syncutil.Error(err)
		answer.Result = syncmsg.SyncRequestEntityMessageResponseResult_ErrorCreatingMsgs.Enum()
		answer.ResultMsg = proto.String(err.Error())
		data, err := proto.Marshal(answer)
		if err != nil {
			syncutil.Error("Error readying response: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			syncutil.NotImplementedMsg(err.Error())
		}
		return
	}

	// answer.Request = request
	//
	// if len(request.Items) > 0 {
	// 	answer.Result = syncmsg.SyncRequestEntityMessageResponseResult_HasMsgs.Enum()
	// } else {
	// 	answer.Result = syncmsg.SyncRequestEntityMessageResponseResult_NoMsgs.Enum()
	// }
	// answer.ResultMsg = proto.String("")

	data, err := proto.Marshal(answer)
	if err != nil {
		syncutil.Error("Error readying response: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		syncutil.NotImplementedMsg(err.Error())
	}
	//syncutil.Debug(request)
}

func processFetchSyncDataArgs(vars map[string]string) (fetchSyncDataArgs, error) {

	var orderNumStr, changeTypeStr string

	answer := fetchSyncDataArgs{}

	var hasValue bool
	answer.sessionID, hasValue = vars["sessionId"]
	if !hasValue {
		return answer, errors.New(MissingParamSessionID)
	}
	answer.nodeIDToProcess, hasValue = vars["nodeId"]
	if !hasValue {
		return answer, errors.New(MissingParamNodeID)
	}
	orderNumStr, hasValue = vars["orderNum"]
	if !hasValue {
		return answer, errors.New(MissingParamOrderNum)
	}
	orderNum, err := strconv.ParseInt(orderNumStr, 10, 32)
	if err != nil {
		errMsg := fmt.Sprintf(BadOrderNumConversion, orderNumStr)
		return answer, errors.New(errMsg)
	}
	answer.orderNum = orderNum
	changeTypeStr, hasValue = vars["changeType"]
	if !hasValue {
		return answer, errors.New(MissingParamChangeType)
	}
	if changeTypeStr == "AddOrUpdate" {
		answer.changeType = syncapi.ProcessSyncChangeEnumAddOrUpdate
	} else if changeTypeStr == "Delete" {
		answer.changeType = syncapi.ProcessSyncChangeEnumDelete
	} else {
		errMsg := fmt.Sprintf(BadChangeTypeConversion, changeTypeStr)
		return answer, errors.New(errMsg)
	}
	return answer, nil
}

func createFetchersForProcessFetchSyncData(validArgs fetchSyncDataArgs, repo syncapi.Repository) (syncapi.EntityFetching, syncapi.MessageFetching, error) {
	var entityFetcher syncapi.EntityFetching
	var msgFetcher syncapi.MessageFetching
	var err error
	entityFetcher, err = repo.ConfigRepo.CreateEntityFetcher(validArgs.sessionID, validArgs.nodeIDToProcess)
	if err != nil {
		ctxMsg := "Could not create entityFetcher."
		syncutil.Error(ctxMsg, err)
		msg := ctxMsg + err.Error()
		return entityFetcher, msgFetcher, errors.New(msg)
	}
	msgFetcher, err = repo.DataRepo.CreateMessageFetcher(validArgs.sessionID, validArgs.nodeIDToProcess)
	if err != nil {
		ctxMsg := "Could not create msgFetcher."
		syncutil.Error(ctxMsg, err)
		msg := ctxMsg + err.Error()
		return entityFetcher, msgFetcher, errors.New(msg)
	}
	return entityFetcher, msgFetcher, nil
}

type fetchSyncDataArgs struct {
	sessionID       string
	nodeIDToProcess string
	orderNum        int64
	changeType      syncapi.ProcessSyncChangeEnum
}
