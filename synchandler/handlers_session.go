package synchandler

import (
	"data-sync-tools-go/syncdao"
	"data-sync-tools-go/syncutil"
	"encoding/json"
	//"fmt"
	"github.com/gorilla/mux"
	//"io/ioutil"

	"net/http"
)

//MARK: CreateSyncSession Processing START

//CreateSyncSession creates a session for syncing. Invoked performed via the following http command:
//	curl -i --header "Content-Type: application/json" --request POST http://localhost:8080/syncSession/sessionId/8E0E2E15-6505-429F-8269-7336D8421D89/pairId/08751B44-289E-45C9-A4F9-F02F14A5D490
func CreateSyncSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		sessionID string
		pairID    string
		hasValue  bool
		err       error
	)
	sessionID, hasValue = vars["sessionId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	pairID, hasValue = vars["pairId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	//see http://stackoverflow.com/questions/15672556/handling-json-post-request-in-go
	var request = syncdao.CreateSyncSessionRequest{
		PairID:    pairID,
		SessionID: sessionID,
	}
	var response CreateSyncSessionResponse
	var result syncdao.CreateSyncSessionDaoResult
	if dbSetupError(w) {
		return
	}
	result, err = syncdao.DefaultDaos.SyncPairDao().CreateSyncSession(request)
	if err != nil {
		errMsg := "Could not persist data to database:" + err.Error()
		response = CreateSyncSessionResponse{
			PairID:             request.PairID,
			RequestedSessionID: request.SessionID,
			ActualSessionID:    "",
			Result:             "CreateSyncSessionUnknownError",
			ResultMsg:          errMsg,
		}
	} else {
		response = CreateSyncSessionResponse{
			PairID:             request.PairID,
			RequestedSessionID: request.SessionID,
			ActualSessionID:    result.ActualSessionID,
			Result:             result.Result,
			ResultMsg:          "",
		}
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		syncutil.NotImplementedMsg(err.Error())
	}
	return
}

//CreateSyncSessionResponse represents the response to a create sync session request.
type CreateSyncSessionResponse struct {
	PairID             string `json:"pairId"`
	RequestedSessionID string `json:"requestedSessionId"`
	ActualSessionID    string `json:"actualSessionId"`
	// BUG(doug4j@gmail.com): Create enum CreateNodeResponseType:
	//case OK = "OK"
	//case ThisSessionIdAlreadyActive = "ThisSessionIdAlreadyActive"
	//case DifferentSessionIdAlreadyActive = "DifferentSessionIdAlreadyActive"
	//case CreateSyncSessionUnknownError = "CreateSyncSessionUnknownError"
	Result    string `json:"result"`
	ResultMsg string `json:"resultMsg"`
}

//MARK: CreateSyncSession Processing END

//MARK: UpdateSyncSessionState Processing START

// UpdateSyncSessionState updates an active sync session. Valid states include: 'Inactive', 'Initializing', 'Seeding', 'Queuing', 'Syncing', or 'Canceling'.
// Invoked performed via the following http command:
//	curl -i --header "Content-Type: application/json" --request PUT http://localhost:8080/SyncSession/sessionId/66091BCC-E470-41BC-9025-9686514CD4B1/state/queuing
func UpdateSyncSessionState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		sessionID string
		pairID    string
		state     string
		hasValue  bool
		err       error
	)
	sessionID, hasValue = vars["sessionId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	state, hasValue = vars["state"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	pairID, hasValue = vars["pairId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var request = syncdao.UpdateSyncSessionStateRequest{
		SessionID: sessionID,
		State:     state,
		PairID:    pairID,
	}
	var answer UpdateSyncSessionStateResponse
	if dbSetupError(w) {
		return
	}
	response, err := syncdao.DefaultDaos.SyncPairDao().UpdateSyncSessionState(request)
	if err != nil {
		errMsg := "Could not persist data to database:" + err.Error()
		answer = UpdateSyncSessionStateResponse{
			SessionID:      sessionID,
			RequestedState: state,
			Result:         "UpdateSyncSessionUnknownError",
			ResultMsg:      errMsg,
		}
	} else {
		answer = UpdateSyncSessionStateResponse{
			SessionID:      sessionID,
			RequestedState: state,
			Result:         response.Result,
			ResultMsg:      response.ResultMsg,
		}
	}
	err = json.NewEncoder(w).Encode(answer)
	if err != nil {
		syncutil.NotImplementedMsg(err.Error())
	}
	return
}

//UpdateSyncSessionStateResponse represents the answer to an UpdateSyncSessionStateRequest in dao.go.
type UpdateSyncSessionStateResponse struct {
	//PairId             string `json:"pairId"`
	SessionID      string `json:"sessionId"`
	RequestedState string `json:"requestedState"`
	Result         string `json:"result"`
	ResultMsg      string `json:"resultMsg"`
}

//MARK: UpdateSyncSession Processing END

//MARK: UpdateSyncSessionStateWithTotals Processing START
/*
curl -i --header "Content-Type: application/json" --request PUT http://localhost:8080/SyncSession/sessionId/66091BCC-E470-41BC-9025-9686514CD4B1/state/queuing
*/

// BUG(doug4j@gmail.com): Create UpdateSyncSessionStateWithTotals:
/*
func UpdateSyncSessionStateWithTotals(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		sessionId string
		pairId    string
		state     string
		hasValue  bool
		err       error
	)
	sessionId, hasValue = vars["sessionId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	state, hasValue = vars["state"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	pairId, hasValue = vars["pairId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var request = UpdateSyncSessionStateRequest{
		SessionId: sessionId,
		State:     state,
		PairId:    pairId,
	}
	var responseJson UpdateSyncSessionStateResponse
	if dbSetupError(w) {
		return
	}
	response, err := DefaultDaos.SyncPairDao().UpdateSyncSessionState(request)
	if err != nil {
		errMsg := "Could not persist data to database:" + err.Error()
		responseJson = UpdateSyncSessionStateResponse{
			SessionId:      sessionId,
			RequestedState: state,
			Result:         "UpdateSyncSessionUnknownError",
			ResultMsg:      errMsg,
		}
		//log.Println(errMsg)
	} else {
		responseJson = UpdateSyncSessionStateResponse{
			SessionId:      sessionId,
			RequestedState: state,
			Result:         response.Result,
			ResultMsg:      response.ResultMsg,
		}
	}
	json.NewEncoder(w).Encode(responseJson)
	return
}
*/

//MARK: UpdateSyncSessionStateWithTotals Processing END

//MARK: UpdateSyncSessionStateWithProcessed Processing START
/*
curl -i --header "Content-Type: application/json" --request PUT http://localhost:8080/SyncSession/sessionId/66091BCC-E470-41BC-9025-9686514CD4B1/state/queuing
*/

// BUG(doug4j@gmail.com): Create UpdateSyncSessionStateWithProcessed:
/*
func UpdateSyncSessionStateWithProcessed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		sessionId    string
		pairId       string
		recordBytes1 string
		recordBytes2 string
		recordCount1 string
		recordCount2 string
		hasValue     bool
		err          error
	)
	sessionId, hasValue = vars["sessionId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	pairId, hasValue = vars["pairId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	recordBytes1, hasValue = vars["recordBytes1"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	recordBytes2, hasValue = vars["recordBytes2"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	recordCount1, hasValue = vars["recordCount1"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	recordCount2, hasValue = vars["recordCount2"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var request = UpdateSyncingRequestWithProcessed{
		SessionId:    sessionId,
		PairId:       pairId,
		RecordBytes1: recordBytes1,
		RecordCount1: recordCount1,
		RecordBytes2: recordBytes2,
		RecordCount2: recordCount2,
	}
	var responseJson UpdateSyncSessionStateResponse
	if dbSetupError(w) {
		return
	}
	response, err := DefaultDaos.SyncPairDao().UpdateSyncingWithProcessed(request)
	if err != nil {
		errMsg := "Could not persist data to database:" + err.Error()
		responseJson = UpdateSyncSessionStateResponse{
			SessionId:      sessionId,
			RequestedState: "Syncing",
			Result:         "UpdateSyncSessionUnknownError",
			ResultMsg:      errMsg,
		}
		//log.Println(errMsg)
	} else {
		responseJson = UpdateSyncSessionStateResponse{
			SessionId:      sessionId,
			RequestedState: "Syncing",
			Result:         response.Result,
			ResultMsg:      response.ResultMsg,
		}
	}
	json.NewEncoder(w).Encode(responseJson)
	return
}
*/

//MARK: UpdateSyncSessionStateWithProcessed Processing END

//MARK: QueryPairState Processing START

//GetPairState obtains the current state of a syncpair.
func GetPairState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		pairID   string
		hasValue bool
		err      error
	)
	pairID, hasValue = vars["pairId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var result syncdao.QueryPairStateDaoResult
	if dbSetupError(w) {
		return
	}
	var request = syncdao.QueryPairStateRequest{
		PairID: pairID,
	}

	result, err = syncdao.DefaultDaos.SyncPairDao().QueryPairState(request)
	var answer QueryPairStateResponse
	if err != nil {
		errMsg := "Could not execute query to database:" + err.Error()
		answer = QueryPairStateResponse{
			PairID:    request.PairID,
			State:     "",
			SessionID: "",
			Result:    "Error",
			ResultMsg: errMsg,
		}
		//log.Println(errMsg)
	} else {
		answer = QueryPairStateResponse{
			PairID:    request.PairID,
			State:     result.State,
			SessionID: result.SessionID,
			Result:    "OK",
			ResultMsg: "",
		}
	}
	err = json.NewEncoder(w).Encode(answer)
	if err != nil {
		syncutil.NotImplementedMsg(err.Error())
	}
	return
}

//QueryPairStateResponse represents the answer to a QueryPairRequest in dao.go.
type QueryPairStateResponse struct {
	PairID       string `json:"pairId"`
	State        string `json:"state"`
	SessionID    string `json:"sessionId"`
	SessionStart int    `json:"sessionStart"`
	LastUpdated  int    `json:"lastUpdated"`

	// BUG(doug4j@gmail.com): Create enum CreateNodeResponseType:
	//case OK = "OK"
	//case ThisSessionIdAlreadyActive = "ThisSessionIdAlreadyInactive"
	//case DifferentSessionIdAlreadyExists = "DifferentSessionIdAlreadyExists"
	//case CreateSyncSessionUnknownError = "CreateSyncSessionUnknownError"
	Result    string `json:"result"`
	ResultMsg string `json:"resultMsg"`
}

//MARK: QueryPairState Processing END

//MARK: CloseSyncSession Processing START

//CloseSyncSession closes a session for syncing. Invoked performed via the following http command:
//	curl -i --header "Content-Type: application/json" --request PUT --data '{"pairId":"79EC415D-FC24-416B-A5D7-73F185B1B4DA","sessionId":"66091BCC-E470-41BC-9025-9686514CD4B1"}' http://localhost:8080/closeSyncSession
func CloseSyncSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		sessionID string
		pairID    string
		hasValue  bool
		err       error
	)
	sessionID, hasValue = vars["sessionId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	pairID, hasValue = vars["pairId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	//see http://stackoverflow.com/questions/15672556/handling-json-post-request-in-go
	var request = syncdao.CloseSyncSessionRequest{
		PairID:    pairID,
		SessionID: sessionID,
	}
	var answer CloseSyncSessionResponse
	var result syncdao.CloseSyncSessionDaoResult
	if dbSetupError(w) {
		return
	}
	result, err = syncdao.DefaultDaos.SyncPairDao().CloseSyncSession(request)
	if err != nil {
		errMsg := "Could not persist data to database:" + err.Error()
		answer = CloseSyncSessionResponse{
			PairID:             request.PairID,
			RequestedSessionID: request.SessionID,
			ActualSessionID:    "",
			Result:             "CloseSyncSessionUnknownError",
			ResultMsg:          errMsg,
		}
		syncutil.Error(errMsg)
	} else {
		answer = CloseSyncSessionResponse{
			PairID:             request.PairID,
			RequestedSessionID: request.SessionID,
			ActualSessionID:    result.ActualSessionID,
			Result:             result.Result,
			ResultMsg:          "",
		}
	}
	err = json.NewEncoder(w).Encode(answer)
	if err != nil {
		syncutil.NotImplementedMsg(err.Error())
	}
	return
}

//CloseSyncSessionResponse represents the response to CloseSyncSessionRequest in dao.go.
type CloseSyncSessionResponse struct {
	PairID             string `json:"pairId"`
	RequestedSessionID string `json:"requestedSessionId"`
	ActualSessionID    string `json:"actualSessionId"`
	// BUG(doug4j@gmail.com): Create enum CreateNodeResponseType:
	//case OK = "OK"
	//case ThisSessionIdAlreadyActive = "ThisSessionIdAlreadyInactive"
	//case DifferentSessionIdAlreadyExists = "DifferentSessionIdAlreadyExists"
	//case CreateSyncSessionUnknownError = "CreateSyncSessionUnknownError"
	Result    string `json:"result"`
	ResultMsg string `json:"resultMsg"`
}

//MARK: CloseSyncSession Processing END
