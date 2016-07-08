package syncapi

//SyncFieldDefinition defines a field definition for syncing
// type SyncFieldDefinition struct {
// 	FieldName    string
// 	FieldType    SyncFieldTypeEnum
// 	IsPrimaryKey bool
// }

//EntityPairItem represents the data for a given node in a SyncPair.
// type EntityPairItem struct {
// 	EntitySingularName    string `json:"entitySingularName"`
// 	EntityPluralName      string `json:"entityPluralName"`
// 	ProcessOrderAddUpdate int    `json:"processOrderAddUpdate"`
// 	ProcessOrderDelete    int    `json:"processOrderDelete"`
// 	EntityHandlerURI      string `json:"entityHandlerUri"`
// }

//EntityNameItem provides the various named forms of the entity (singular and plural).
type EntityNameItem struct {
	SingularName string
	PluralName   string
}

// type EntityFetcherType struct {
//
// }

//EntityFetching retrieves entities in from different contexts
type EntityFetching interface {
	//FindEntitiesToFetch retrieves entity names to be processed for a given sync session by orderNumber
	FindEntitiesForFetch(orderNum int, sessionID string, nodeID string, changeType ProcessSyncChangeEnum) ([]EntityNameItem, error)

	//FindEntitiesToProcess retrieves entity names to be processed by plural name
	//FindEntitiesForProcess(request syncmsg.ProtoSyncEntityMessageRequest) (map[string]EntityNameItem, error)

	// TODO(doug4j@gmail.com): Implement FindPluralEntityNamesById for client implementation support.
	FindPluralEntityNamesByID(sessionID string, nodeID string) (map[string]EntityNameItem, error)
}
