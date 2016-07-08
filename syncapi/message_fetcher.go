package syncapi

// TODO(doug4j@gmail.com): Add an abstraction to the Protocol Buffers intefaces in 'messages.pb.go' in the syncmsg package.
import "data-sync-tools-go/syncmsg"

//MessageFetchingData defines the typical re-usable data for MessageFetching.
type MessageFetchingData struct {
	//SessionId is the already active session for syncing.
	SessionID string
	//NodeId is the node whose data is to be fetched.
	NodeID string
	//MaxGroupBytesSize is the max number of bytes that should be part of a given group of messages.
	MaxGroupBytesSize int
	//MaxMsgs is the max number of messages that should be part of a given group of messages.
	MaxMsgs int
}

//MessageFetching provides services for retrieving a group of sync messages for downstream processes.
type MessageFetching interface {
	//Fetch retrieves a group of sync messages for downstream processes returning a value with no request items
	//in it should it be (Copied from SyncMessagesFetcher.Fetch() interface)
	Fetch(entities []EntityNameItem, changeType ProcessSyncChangeEnum) (*syncmsg.ProtoRequestSyncEntityMessageResponse, error)

	//FindEntitiesToFetch retrieves entity names to be processed for a given sync session by orderNumber
	//FindEntitiesToFetch(orderNum int, sessionID string, changeType ProcessSyncChangeEnum) ([]EntityNameItem, error)

	//MarkProcessed(entities []string, sessionID string, changeType ProcessSyncChangeEnum) error
}
