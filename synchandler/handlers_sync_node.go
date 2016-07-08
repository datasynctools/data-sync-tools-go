package synchandler

import (
	"data-sync-tools-go/syncdao"
	"data-sync-tools-go/syncutil"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

//CreateNewNode creates a sync node in the sync cluster. Invoked performed via the following http command:
//	curl -H "Content-Type: application/json" -d '{"msgId":"39709F79-2036-44CD-B75F-D97AB6050872", "node":"node-BE6A1019-BC9F-49A0-82FC-EF03D06B54CB"}' http://localhost:8080/createNode
func CreateNewNode(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	//see http://stackoverflow.com/questions/15672556/handling-json-post-request-in-go
	var request syncdao.SyncNode
	err := decoder.Decode(&request)
	if err != nil {
		errMsg := "Could not parse json from body"
		syncutil.Error(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	if dbSetupError(w) {
		return
	}
	err = syncdao.DefaultDaos.SyncNodeDao().AddNode(request)
	var answer CreateNodeResponse
	if err != nil {

		answer = CreateNodeResponse{
			NodeID:          request.NodeID,
			NodeName:        request.NodeName,
			DataVersionName: request.DataVersionName,
			MsgID:           request.NodeID,
			Result:          "NotCreatedUnknownError",
			ResultMsg:       err.Error(),
		}
	} else {
		answer = CreateNodeResponse{
			NodeID:          request.NodeID,
			NodeName:        request.NodeName,
			DataVersionName: request.DataVersionName,
			MsgID:           request.NodeID,
			Result:          "OK",
			ResultMsg:       "",
		}
	}
	err = json.NewEncoder(w).Encode(answer)
	if err != nil {
		syncutil.NotImplementedMsg(err.Error())
	}
	return
}

//CreateNodeResponse is the response structure for CreateNewNode.
type CreateNodeResponse struct {
	NodeID          string `json:"nodeId"`
	NodeName        string `json:"nodeName"`
	DataVersionName string `json:"dataVersionName"`

	MsgID     string `json:"msgId"`
	Result    string `json:"result"` // BUG(doug4j@gmail.com): Create enum CreateNodeResponseType
	ResultMsg string `json:"resultMsg"`
}

//NodeResponse is used for as the response for GetNodeByNodeName and GetNodeByNodeId
type NodeResponse struct {
	NodeID          string `json:"nodeId"`
	NodeName        string `json:"nodeName"`
	DataVersionName string `json:"dataVersionName"`

	MsgID     string `json:"msgId"`
	Result    string `json:"result"` // BUG(doug4j@gmail.com): Create enum CreateNodeResponseType
	ResultMsg string `json:"resultMsg"`
}

//DeleteNodeResponse is the response structure for DeleteNodeByNodeId
type DeleteNodeResponse struct {
	NodeID    string `json:"nodeId"`
	MsgID     string `json:"msgId"`
	Result    string `json:"result"` // BUG(doug4j@gmail.com): Create enum CreateNodeResponseType
	ResultMsg string `json:"resultMsg"`
}

//GetNodeByNodeName obtains node information by node name. Invoked performed via the following http command:
//	curl -H "Content-Type: application/json" -d '{"msgId":"39709F79-2036-44CD-B75F-D97AB6050872", "node":"node-BE6A1019-BC9F-49A0-82FC-EF03D06B54CB"}' http://localhost:8080/syncNode/
func GetNodeByNodeName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		nodeName string
		hasValue bool
		err      error
	)
	nodeName, hasValue = vars["nodeName"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if dbSetupError(w) {
		return
	}
	syncNode, err := syncdao.DefaultDaos.SyncNodeDao().GetOneNodeByNodeName(nodeName)
	var answer NodeResponse
	if err != nil {
		answer = NodeResponse{
			NodeID:          "",
			NodeName:        nodeName,
			DataVersionName: "",

			//MsgId:     nodeId,
			Result:    "Error",
			ResultMsg: err.Error(),
		}
	} else {
		answer = NodeResponse{
			NodeID:          syncNode.NodeID,
			NodeName:        syncNode.NodeName,
			DataVersionName: syncNode.DataVersionName,

			MsgID:     syncNode.NodeID,
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

//GetNodeByNodeID obtains node information by node id. Invoked performed via the following http command:
//	curl -H "Content-Type: application/json" -d '{"msgId":"39709F79-2036-44CD-B75F-D97AB6050872", "node":"node-BE6A1019-BC9F-49A0-82FC-EF03D06B54CB"}' http://localhost:8080/syncNode/
func GetNodeByNodeID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		nodeID   string
		hasValue bool
		err      error
	)
	nodeID, hasValue = vars["nodeId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if dbSetupError(w) {
		return
	}
	syncNode, err := syncdao.DefaultDaos.SyncNodeDao().GetOneNodeByNodeID(nodeID)
	var answer NodeResponse
	if err != nil {
		answer = NodeResponse{
			NodeID:          nodeID,
			NodeName:        "",
			DataVersionName: "",

			MsgID:     nodeID,
			Result:    "Error",
			ResultMsg: err.Error(),
		}
	} else {
		answer = NodeResponse{
			NodeID:          syncNode.NodeID,
			NodeName:        syncNode.NodeName,
			DataVersionName: syncNode.DataVersionName,

			MsgID:     syncNode.NodeID,
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

//DeleteNodeByNodeID removes the sync node in the cluster node by node id. Invoked performed via the following http command:
//	curl -H "Content-Type: application/json" -d '{"msgId":"39709F79-2036-44CD-B75F-D97AB6050872", "node":"node-BE6A1019-BC9F-49A0-82FC-EF03D06B54CB"}' http://localhost:8080/syncNode/
func DeleteNodeByNodeID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		nodeID   string
		hasValue bool
		err      error
	)
	nodeID, hasValue = vars["nodeId"]
	if !hasValue {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if dbSetupError(w) {
		return
	}
	err = syncdao.DefaultDaos.SyncNodeDao().DeleteNodeByNodeID(nodeID)
	var answer DeleteNodeResponse
	if err != nil {
		answer = DeleteNodeResponse{
			NodeID:    nodeID,
			MsgID:     nodeID,
			Result:    "Error",
			ResultMsg: err.Error(),
		}
	} else {
		answer = DeleteNodeResponse{
			NodeID:    nodeID,
			MsgID:     nodeID,
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
