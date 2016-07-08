package synchandler

import (
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncmsg"
)

type mockDataRepository struct {
	fetcher              syncapi.MessageFetching
	processor            syncapi.MessageProcessing
	queuer               syncapi.MessageQueuing
	createFetcherError   error
	createProcessorError error
	createQueuerError    error
	//findForProcessAnswer map[string]syncapi.EntityNameItem
}

func (repo mockDataRepository) CreateMessageFetcher(SessionID string, NodeID string) (syncapi.MessageFetching, error) {
	return repo.fetcher, repo.createFetcherError
}

func (repo mockDataRepository) CreateMessageProcessor(sessionID string, nodeID string, entitiesByPluralName map[string]syncapi.EntityNameItem) (syncapi.MessageProcessing, error) {
	return repo.processor, repo.createProcessorError
}

func (repo mockDataRepository) CreateMessageQueuer(sessionID string, nodeID string) (syncapi.MessageQueuing, error) {
	return repo.queuer, repo.createQueuerError
}

type mockSyncMessageFetcher struct {
	fetchAnswer      *syncmsg.ProtoRequestSyncEntityMessageResponse
	fetchError       error
	findEntityAnswer []syncapi.EntityNameItem
	findEntityError  error
}

func (fetcher mockSyncMessageFetcher) Fetch(entities []syncapi.EntityNameItem, changeType syncapi.ProcessSyncChangeEnum) (*syncmsg.ProtoRequestSyncEntityMessageResponse, error) {
	return fetcher.fetchAnswer, fetcher.fetchError
}

func (fetcher mockSyncMessageFetcher) FindEntitiesToFetch(orderNum int, sessionID string, changeType syncapi.ProcessSyncChangeEnum) ([]syncapi.EntityNameItem, error) {
	return fetcher.findEntityAnswer, fetcher.findEntityError
}

type mockConfigRepository struct {
	fetcher            syncapi.EntityFetching
	createFetcherError error
}

func (repo mockConfigRepository) CreateEntityFetcher(sessionID string, nodeID string) (syncapi.EntityFetching, error) {
	return repo.fetcher, repo.createFetcherError
}

type mockEntityFetcher struct {
	findForFetchAnswer   []syncapi.EntityNameItem
	findForFetchError    error
	findForProcessAnswer map[string]syncapi.EntityNameItem
	findForProcessError  error
}

func (fetcher mockEntityFetcher) FindEntitiesForFetch(orderNum int, sessionID string, nodeID string, changeType syncapi.ProcessSyncChangeEnum) ([]syncapi.EntityNameItem, error) {
	return fetcher.findForFetchAnswer, fetcher.findForFetchError
}

func (fetcher mockEntityFetcher) FindPluralEntityNamesByID(sessionID string, nodeID string) (map[string]syncapi.EntityNameItem, error) {
	return fetcher.findForProcessAnswer, fetcher.findForProcessError
}

type mockMessageProcessor struct {
	processorAnswer *syncmsg.ProtoSyncEntityMessageResponse
}

func (processor mockMessageProcessor) Process(request *syncmsg.ProtoSyncEntityMessageRequest) *syncmsg.ProtoSyncEntityMessageResponse {
	return processor.processorAnswer
}

type mockMessageQueuer struct {
	queuerAnswer int
	queueError   error
}

func (queuer mockMessageQueuer) Queue(sessionID string, nodeIDToQueue string) (int, error) {
	return queuer.queuerAnswer, queuer.queueError
}
