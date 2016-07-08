package synchandler

import (
	"net/http"

	"github.com/gorilla/mux"
)

//NewRouter instantiates instances of mux.Router for processing HTTP requests and sets the handlers 'VarsHandler' with the mux.Vars if it is not already set.
func NewRouter(handlers Handlers) *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	if handlers.VarsHandler == nil {
		handlers.VarsHandler = mux.Vars
	}
	routes := createRoutes(handlers)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)

	}

	return router
}

//Route defines a path and executable destination for mux.Router.
type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

func createRoutes(handlers Handlers) []route {
	var routes = []route{
		route{
			"Index",
			"GET",
			"/",
			Index,
		},
		//curl -H "Content-Type: application/json" -d '{"msgId":"39709F79-2036-44CD-B75F-D97AB6050872", "node":"node-BE6A1019-BC9F-49A0-82FC-EF03D06B54CB"}' http://localhost:8080/createNode
		route{
			"CreateNewNode",
			"POST",
			"/syncNode",
			CreateNewNode,
		},
		route{
			"GetNodeByNodeName",
			"GET",
			"/syncNode/nodeName/{nodeName}",
			GetNodeByNodeName,
		},
		route{
			"GetNodeByNodeId",
			"GET",
			"/syncNode/nodeId/{nodeId}",
			GetNodeByNodeID,
		},
		route{
			"DeleteNodeByNodeId",
			"DELETE",
			"/syncNode/nodeId/{nodeId}",
			DeleteNodeByNodeID,
		},
		route{
			"GetSyncConfig",
			"GET",
			"/syncPairConfig/node1Name/{node1Name}/node2Name/{node2Name}",
			GetSyncConfig,
		},
		route{
			"QueueSyncChanges",
			"PUT",
			"/changeQueue/sessionId/{sessionId}/nodeId/{nodeId}/msgId/{msgId}",
			handlers.QueueSyncChanges,
		},
		route{
			"CreateSyncSession",
			"POST",
			"/syncSession/sessionId/{sessionId}/pairId/{pairId}",
			CreateSyncSession,
		},
		route{
			"UpdateSyncSessionState",
			"PUT",
			"/syncSession/sessionId/{sessionId}/pairId/{pairId}/state/{state:Seeding|Queuing|Syncing|Canceling}",
			//"/syncSession/sessionId/{sessionId}/pairId/{pairId}/state/{state:Seeding|Queuing}",
			UpdateSyncSessionState,
		},
		route{
			"ProcessSyncData",
			"PUT",
			"/syncData/sessionId/{sessionId}/nodeId/{nodeId}/transactionBindId/{transactionBindId}",
			handlers.ProcessSyncData,
		},
		route{
			"FetchSyncData",
			"GET",
			"/fetchData/sessionId/{sessionId}/nodeId/{nodeId}/orderNum/{orderNum:[0-9]+}/changeType/{changeType:AddOrUpdate|Delete}/",
			handlers.FetchSyncData,
		},
		route{
			"AcknowledgeSyncData",
			"PUT",
			"/ackData/sessionId/{sessionId}/nodeId/{nodeId}/transactionBindId/{transactionBindId}",
			handlers.AcknowledgeSyncData,
		},
		route{
			"CloseSyncSession",
			"DELETE",
			"/syncSession/sessionId/{sessionId}/pairId/{pairId}",
			CloseSyncSession,
		},
		route{
			"GetPairState",
			"GET",
			"/syncPairState/pairId/{pairId}",
			GetPairState,
		},
		route{
			"IntegrationTestReset",
			"GET",
			"/integrationTest/testName/{testName}",
			IntegrationTestReset,
		},

		// BUG(doug4j@gmail.com): Make calls more restful where more of the signaling is done in the uri rather than json AND nouns are
		//favored over verbs
		/*
			Route{
				"CreateNewNode",
				"POST",
				"/node/{nodeId}",
				CreateNewNode,
			},
			Route{
				"GetSyncConfig",
				"GET",
				"/syncConfig/requestingNodeName/{requestingNodeName}/toPairNodeName/{toPairNodeName}",
				GetSyncConfig,
			},
			//curl -H "Content-Type: application/json" -d '{"msgId":"39709F79-2036-44CD-B75F-D97AB6050872"}' http://localhost:8080/node/{nodeId}/msgId/{msgId}
			Route{
				"CreateSyncSession",
				"POST",
				"/syncSession/{sesssionId}/pairId/{pairId}",
				CreateSyncSession,
			},
			//curl -H "Content-Type: application/json" -d '{"msgId":"39709F79-2036-44CD-B75F-D97AB6050872"}' http://localhost:8080/node/{nodeId}
			Route{
				"CreateSyncSession",
				"POST",
				"/syncSession/{sesssionId}/pairId/{pairId}[/msgId/{msgId}]",
				CreateSyncSession,
			},
			Route{
				"QueueSyncChanges",
				"PUT",
				"/changeQueue/sessionId/{sessionId}/nodeId/{nodeId}",
				QueueSyncChanges,
			},
			Route{
				"CloseSyncSession",
				"DELETE",
				"/syncSession/{sessionId}/pairId/{pairId}",
				CloseSyncSession,
			},

		*/

		/*
				Route{
				"ProcessSyncData",
				"POST",
				"/syncData/sessionId/{sessionId}/pairId/{pairId}/nodeId/{nodeId}/isDelete/{isDelete}/transactionBindId/{transactionBindId}",
				ProcessSyncData,
			},

				curl -H "Content-Type: application/json" --request PUT http://localhost:8080/syncdata/sessionId/{sessionId}/pairId/{pairId}/nodeId/{nodeId}/isDelete/{isDelete}/transactionBindId/{transactionBindId} -d
					Route{
						"GetSyncDataMsgsInGroup",
						"GET",
						"/syncdata/sessionId/{sessionId}/pairId/{pairId}/nodeId/{nodeId}/changeType/{changeType:AddOrUpdate|Delete}/group/{groupNo}/maxGroupDataSize/{maxGroupDataSize}", //maxMsgs/{maxMsgNum}/
						GetSyncDataMsgsInGroup,
					},
					Route{
						"UpdateSyncDataState",
						"PUT",
						"/syncdata/sessionId/{sessionId}/pairId/{pairId}/updateState",
						UpdateSyncDataState,
					},


					Route{
						"UpdateSyncSessionStateWithTotals",
						"PUT",
						"/syncSession/sessionId/{sessionId}/pairId/{pairId}/state/Syncing/Totals/recordBytes1/{recordBytes1}/recordCount1/{recordCount1}/recordBytes2/{recordBytes2}/recordCount2/{recordCount2}",
						UpdateSyncSessionStateWithTotals,
					},
					Route{
						"UpdateSyncSessionStateWithProcessed",
						"PUT",
						"/syncSession/sessionId/{sessionId}/pairId/{pairId}/state/Syncing/Processed/recordBytes1/{recordBytes1}/recordCount1/{recordCount1}/recordBytes2/{recordBytes2}/recordCount2/{recordCount2}",
						UpdateSyncSessionStateWithProcessed,
					},

		*/

	}
	return routes
}

//Routes provides the list of available HTTP routes.
//type routes []route
