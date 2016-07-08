package syncapi

//Repository defines a repository for dependency injection for sync servicing.
type Repository struct {
	DataRepo   DataRepositoryable
	ConfigRepo ConfigRepositoryable
	//SessionRepo SessionRepository
	//ConflictRepo ConflictRepository
}

//DataRepositoryable acts as a factory to access a store for servicing local sync data.
type DataRepositoryable interface {
	CreateMessageFetcher(sessionID string, nodeID string) (MessageFetching, error)
	CreateMessageProcessor(sessionID string, nodeID string, entitiesByPluralName map[string]EntityNameItem) (MessageProcessing, error)
	CreateMessageQueuer(sessionID string, nodeID string) (MessageQueuing, error)
	// TODO(doug4j@gmail.com): Implement CreateMessageAcknowledger for client implementation support.

	//CreateMessageAcknowledger(sessionID string, nodeID string) (MessageAcknowledging, error)

	/*
		QueueChanges(sessionID string, nodeIDToQueue string) (int, error)
	*/
}

/*
//SessionRepository establishes how to service session informaiton regarding data sync.
type SessionRepository interface {
	CreateSyncSession(syncSession CreateSyncSessionRequest) (CreateSyncSessionResult, error)
	UpdateSyncSessionState(request UpdateSyncSessionStateRequest) (UpdateSyncSessionStateResult, error)
	CloseSyncSession(syncSession CloseSyncSessionRequest) (CloseSyncSessionResult, error)
	QueryPairState(request QueryPairStateRequest) (QueryPairStateResult, error)
}
*/

//ConfigRepositoryable defines the configurations supporting a data sync cluster.
type ConfigRepositoryable interface {
	CreateEntityFetcher(sessionID string, nodeID string) (EntityFetching, error)

	// AddNode(item SyncNode) error
	// GetOneNodeByNodeName(nodeName string) (SyncNode, error)
	// GetOneNodeByNodeID(nodeID string) (SyncNode, error)
	// DeleteNodeByNodeID(nodeID string) error

	//Note: There is no significance 'requestingNodeName' and 'toPairWithNodeName' in terms of query results. This
	//is more of a logical construct

	// GetPairByNames(requestingNodeName string, toPairWithNodeName string) (SyncPair, error)
	// GetNodePairItem(pairID string, nodeName string) (NodePairItem, error)
	// GetEntityPairItem(pairID string, nodeName string) ([]EntityPairItem, error)

	// BUG(doug4j@gmail.com): Add UpdateSyncingWithTotals(request UpdateSyncingRequestWithTotals) (UpdateSyncingResult, error)
	// BUG(doug4j@gmail.com): Add UpdateSyncingWithProcessed(request UpdateSyncingRequestWithProcessed) (UpdateSyncingResult, error)

}

/*
//CreateSyncSessionResult represents the results from creating a SyncSession.
type CreateSyncSessionResult struct {
	//Valid values: 'OK', 'ThisSessionIdAlreadyActive', or 'DifferentSessionIdAlreadyActive'
	Result          string
	ActualSessionID string
}

//UpdateSyncSessionStateResult represents the results from updating the state of a SyncSession.
type UpdateSyncSessionStateResult struct {
	Result             string
	ResultMsg          string
	RequestedSessionID string
	ActualSessionID    string
	RequestedState     string
	ResultingState     string
}

//SyncNode represents a device or instance on a running device that synchronizes data.
type SyncNode struct {
	//MsgId string
	NodeID          string
	NodeName        string
	DataVersionName string
}

//SyncPair represents a pair of devices or instance on a running devices that synchronizes data.
type SyncPair struct {
	PairID   string
	PairName string
	//DataVersionMapId string
	MaxSesDurValue    int
	MaxSesDurUnit     string
	SyncDataTransForm string
	SyncMsgTransForm  string
	SyncMsgSecPol     string
	SyncSessionID     string
	SyncSessionState  string
	SyncSessionStart  time.Time
	SyncConflictURI   string
	RecordCreated     time.Time
}

//NodePairItem represents the configuration for a given node in a SyncPair.
type NodePairItem struct {
	NodeID                    string           `json:"nodeId"`
	NodeName                  string           `json:"nodeName"`
	Enabled                   bool             `json:"enabled"`
	DataMsgConsumerURI        string           `json:"dataMsgConsumerUri"`
	DataMsgProducerURI        string           `json:"dataMsgProducerUri"`
	MgmtMsgConsumerURI        string           `json:"mgmtMsgConsumerUri"`
	MgmtMsgProducerURI        string           `json:"mgmtMsgProducerUri"`
	SyncDataPersistanceFormat string           `json:"syncDataPersistanceFormat"`
	InMsgBatchSize            int              `json:"inMsgBatchSize"`
	MaxOutMsgBatchSize        int              `json:"maxOutMsgBatchSize"`
	InChanDepthSize           int              `json:"inChanDepthSize"`
	MaxOutChanDeptSize        int              `json:"maxOutChanDeptSize"`
	DataVersionName           string           `json:"dataVersionName"`
	Entities                  []EntityPairItem `json:"entities"`
}

//EntityPairItem represents the data for a given node in a SyncPair.
type EntityPairItem struct {
	EntitySingularName    string `json:"entitySingularName"`
	EntityPluralName      string `json:"entityPluralName"`
	ProcessOrderAddUpdate int    `json:"processOrderAddUpdate"`
	ProcessOrderDelete    int    `json:"processOrderDelete"`
	EntityHandlerURI      string `json:"entityHandlerUri"`
}

//CreateSyncSessionRequest represents a request to create to sync session for a given SyncPair.
type CreateSyncSessionRequest struct {
	PairID    string `json:"pairId"`
	SessionID string `json:"sessionId"`
}

//QueryPairStateRequest represents a request for the current state of a given SyncPair.
type QueryPairStateRequest struct {
	PairID string `json:"pairId"`
}

//UpdateSyncSessionStateRequest represents the data for a given node in a SyncPair.
type UpdateSyncSessionStateRequest struct {
	State     string `json:"state"`
	PairID    string `json:"pairId"`
	SessionID string `json:"sessionId"`
}

//CloseSyncSessionRequest represents a request to close to sync session for a given SyncPair.
type CloseSyncSessionRequest struct {
	// BUG(doug4j@gmail.com): Remove PairId from this call. It is superfluous.
	PairID    string `json:"pairId"`
	SessionID string `json:"sessionId"`
}

//QueryPairStateResult represents the results from querying the state of a SyncPair.
type QueryPairStateResult struct {
	// TODO(doug4j@gmail.com): Valid Values: 'Inactive','Initializing', 'Seeding', 'Queuing', 'Syncing', 'Canceling'
	SessionID    string
	State        string
	LastUpdated  time.Time
	SessionStart time.Time
}

//CloseSyncSessionResult represents the results from closing a SyncSession.
type CloseSyncSessionResult struct {
	// TODO(doug4j@gmail.com): Valid Values: 'OK', 'ThisSessionIdAlreadyInActive', or 'DifferentSessionIdAlreadyActive'
	Result          string
	ActualSessionID string
}

//QuerySessionConfigResult represents the results from querying the configuration for a SyncPair.
type QuerySessionConfigResult struct {
	PairID            string       `json:"pairId"`
	PairName          string       `json:"pairName"`
	MaxSesDurValue    int          `json:"maxSesDurValue"`
	MaxSesDurUnit     string       `json:"maxSesDurUnit"`
	SyncDataTransForm string       `json:"syncDataTransForm"`
	SyncMsgTransForm  string       `json:"syncMsgTransForm"`
	SyncMsgSecPol     string       `json:"syncMsgSecPol"`
	SyncConflictURI   string       `json:"syncConflictUri"`
	Node1             NodePairItem `json:"node1"`
	Node2             NodePairItem `json:"node2"`
}

*/

// TODO(doug4j@gmail.com): Add type UpdateSyncingResult struct
/*
type UpdateSyncingResult struct {
	Result    string
	ResultMsg string
}
*/

// BUG(doug4j@gmail.com): Create UpdateSyncingRequestWithTotals
/*
type UpdateSyncingRequestWithTotals struct {
	State        string `json:"state"`
	PairId       string `json:"pairId"`
	SessionId    string `json:"sessionId"`
	RecordBytes1 string `json:"recordBytes1"`
	RecordCount1 string `json:"recordCount1"`
	RecordBytes2 string `json:"recordBytes2"`
	RecordCount2 string `json:"recordCount2"`
}

type UpdateSyncingRequestWithProcessed struct {
	State        string `json:"state"`
	PairId       string `json:"pairId"`
	SessionId    string `json:"sessionId"`
	RecordBytes1 string `json:"recordBytes1"`
	RecordCount1 string `json:"recordCount1"`
	RecordBytes2 string `json:"recordBytes2"`
	RecordCount2 string `json:"recordCount2"`
}
*/
