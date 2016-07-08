package synchandler

import (
	"data-sync-tools-go/syncutil"
	"encoding/json"

	"github.com/gorilla/mux"
	//"log"
	"net/http"
)

// TODO(doug4j@gmail.com): Finish http sample

//QueueSyncChanges obtains sync config information by node pair names. Invoked performed via the following http command:
//	curl -H "Content-Type: application/json" --request PUT http://localhost:8080/changeQueue/sessionId/my-session-id/nodeId/*node-spoke1
//	curl -H "Content-Type: application/json" --request PUT http://localhost:8080/changeQueue/sessionId/my-session-id/nodeId/*node-spoke1/msgId/my-msg-id
func (handlers Handlers) QueueSyncChanges(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var sessionID string
	var nodeIDToQueue string
	var msgID string

	var hasValue bool
	sessionID, hasValue = vars["sessionId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	nodeIDToQueue, hasValue = vars["nodeId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	msgID, hasValue = vars["msgId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	//log.Printf("SessionId: '%s', NodeId: '%s'", sessionId, nodeId)

	queuer, err := handlers.Repository.DataRepo.CreateMessageQueuer(sessionID, nodeIDToQueue)

	recordsQueuedCount, err := queuer.Queue(sessionID, nodeIDToQueue)
	var response QueueSyncChangesResponse
	if err != nil {
		response = QueueSyncChangesResponse{
			RecordsQueuedCount: recordsQueuedCount,
			SessionID:          sessionID,
			NodeID:             nodeIDToQueue,
			MsgID:              msgID,
			Result:             "NotQueuedUnknownError",
			ResultMsg:          "Failed to queue changes: " + err.Error(),
		}
	} else {
		response = QueueSyncChangesResponse{
			RecordsQueuedCount: recordsQueuedCount,
			SessionID:          sessionID,
			NodeID:             nodeIDToQueue,
			MsgID:              msgID,
			Result:             "OK",
			ResultMsg:          "",
		}
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		syncutil.NotImplementedMsg(err.Error())
	}
	return
}

//QueueSyncChangesResponse represents the response from a queue sync changes request.
type QueueSyncChangesResponse struct {
	RecordsQueuedCount int    `json:"recordsQueuedCount"`
	SessionID          string `json:"sessionId"`
	NodeID             string `json:"nodeId"`
	MsgID              string `json:"msgId"`
	Result             string `json:"result"`
	ResultMsg          string `json:"resultMsg"`
}
