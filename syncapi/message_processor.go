package syncapi

import "data-sync-tools-go/syncmsg"

//MessageProcessingData defines the typical re-usable data for MessageProcesing.
type MessageProcessingData struct {
	//SessionId is the already active session for syncing.
	SessionID string
	//NodeId is the node whose data is to be fetched.
	NodeID string
	//EntitiesByPluralName allows access to EntityNameItems by plural name.
	EntitiesByPluralName map[string]EntityNameItem
}

//MessageProcessing provides services for processing a group of sync messages for downstream processes.
type MessageProcessing interface {
	//Process takes messages to sync and processes them to the underlying datastore and gives a report in the form of a
	//response message in it should it be (Copied from SyncMessagesFetcher.Fetch() interface)
	Process(request *syncmsg.ProtoSyncEntityMessageRequest) *syncmsg.ProtoSyncEntityMessageResponse

	//FindEntitiesToProcess retrieves entity names to be processed by plural name
	//FindEntitiesToProcess(request syncmsg.ProtoSyncEntityMessageRequest) (map[string][]EntityNameItem, error)

	//MarkProcessed(entities []string, sessionID string, changeType ProcessSyncChangeEnum) error
}
