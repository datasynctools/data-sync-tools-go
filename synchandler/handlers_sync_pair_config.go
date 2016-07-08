package synchandler

import (
	"data-sync-tools-go/syncdao"
	"data-sync-tools-go/syncutil"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

//MARK: GetSyncConfig Processing START
func emptyResponse() PairConfigResponse {
	answer := PairConfigResponse{
		PairID:            "",
		PairName:          "",
		MaxSesDurValue:    0,
		MaxSesDurUnit:     "",
		SyncDataTransForm: "",
		SyncMsgTransForm:  "",
		SyncMsgSecPol:     "",
		SyncConflictURI:   "",
		Response:          "",
		ResponseMsg:       "",
	}
	return answer
}

func covertToPairConfigResponse(syncPair syncdao.SyncPair, node1 syncdao.NodePairItem, node2 syncdao.NodePairItem) PairConfigResponse {
	answer := PairConfigResponse{
		PairID:            syncPair.PairID,
		PairName:          syncPair.PairName,
		MaxSesDurValue:    syncPair.MaxSesDurValue,
		MaxSesDurUnit:     syncPair.MaxSesDurUnit,
		SyncDataTransForm: syncPair.SyncDataTransForm,
		SyncMsgTransForm:  syncPair.SyncMsgTransForm,
		SyncMsgSecPol:     syncPair.SyncMsgSecPol,
		SyncConflictURI:   syncPair.SyncConflictURI,
		Node1:             node1,
		Node2:             node2,
		Response:          "OK",
		ResponseMsg:       "",
	}
	return answer
}

//GetSyncConfig obtains sync config information by node pair names. Invoked performed via the following http command:
//	curl http://localhost:8080/getSyncConfig/requestingNodeName/A/toPairNodeName/B
func GetSyncConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var node1Name string
	var node2Name string
	var hasValue bool
	node1Name, hasValue = vars["node1Name"]
	if !hasValue {
		syncutil.Error("node1Name parameter not present")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	node2Name, hasValue = vars["node2Name"]
	if !hasValue {
		syncutil.Error("node2Name parameter not present")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if dbSetupError(w) {
		return
	}
	syncPair, err := syncdao.DefaultDaos.SyncPairDao().GetPairByNames(node1Name, node2Name)
	if err != nil {
		switch {
		case err == syncdao.ErrDaoNoDataFound:
			answer := emptyResponse()
			answer.Response = "RequestNodeAndOtherPeerExistsButNotPaired"
			err = json.NewEncoder(w).Encode(answer)
			if err != nil {
				syncutil.NotImplementedMsg(err.Error())
			}
			return
		default:
			answer := PairConfigResponse{
				Response:    "CannotGetPairConfigUnknownError",
				ResponseMsg: "Could get data from database:" + err.Error(),
			}
			err = json.NewEncoder(w).Encode(answer)
			if err != nil {
				syncutil.NotImplementedMsg(err.Error())
			}
			return
		}
	}
	node1, err := syncdao.DefaultDaos.SyncPairDao().GetNodePairItem(syncPair.PairID, node1Name)
	if err != nil {
		answer := PairConfigResponse{
			Response:    "CannotGetPairConfigUnknownError",
			ResponseMsg: "Could get data from database:" + err.Error(),
		}
		err = json.NewEncoder(w).Encode(answer)
		if err != nil {
			syncutil.NotImplementedMsg(err.Error())
		}
		return
	}
	node1Entities, err := syncdao.DefaultDaos.SyncPairDao().GetEntityPairItem(syncPair.PairID, node1Name)
	if err != nil {
		answer := PairConfigResponse{
			Response:    "CannotGetPairConfigUnknownError",
			ResponseMsg: "Could get data from database:" + err.Error(),
		}
		err = json.NewEncoder(w).Encode(answer)
		if err != nil {
			syncutil.NotImplementedMsg(err.Error())
		}
		return
	}
	node1.Entities = node1Entities

	node2, err := syncdao.DefaultDaos.SyncPairDao().GetNodePairItem(syncPair.PairID, node2Name)
	if err != nil {
		answer := PairConfigResponse{
			Response:    "CannotGetPairConfigUnknownError",
			ResponseMsg: "Could get data from database:" + err.Error(),
		}
		err = json.NewEncoder(w).Encode(answer)
		if err != nil {
			syncutil.NotImplementedMsg(err.Error())
		}
		return
	}

	node2Entities, err := syncdao.DefaultDaos.SyncPairDao().GetEntityPairItem(syncPair.PairID, node2Name)
	if err != nil {
		answer := PairConfigResponse{
			Response:    "CannotGetPairConfigUnknownError",
			ResponseMsg: "Could get data from database:" + err.Error(),
		}
		err = json.NewEncoder(w).Encode(answer)
		if err != nil {
			syncutil.NotImplementedMsg(err.Error())
		}
		return
	}
	node2.Entities = node2Entities

	answer := covertToPairConfigResponse(syncPair, node1, node2)
	err = json.NewEncoder(w).Encode(answer)
	if err != nil {
		syncutil.NotImplementedMsg(err.Error())
	}
	return
}

//PairConfigResponse is the reponse for GetSyncConfig.
type PairConfigResponse struct {
	PairID            string               `json:"pairId"`
	PairName          string               `json:"pairName"`
	MaxSesDurValue    int                  `json:"maxSesDurValue"`
	MaxSesDurUnit     string               `json:"maxSesDurUnit"`
	SyncDataTransForm string               `json:"syncDataTransForm"`
	SyncMsgTransForm  string               `json:"syncMsgTransForm"`
	SyncMsgSecPol     string               `json:"syncMsgSecPol"`
	SyncConflictURI   string               `json:"syncConflictUri"`
	Node1             syncdao.NodePairItem `json:"node1"`
	Node2             syncdao.NodePairItem `json:"node2"`
	Response          string               `json:"response"`
	ResponseMsg       string               `json:"responseMsg"`
}

//MARK: GetSyncConfig Processing END
