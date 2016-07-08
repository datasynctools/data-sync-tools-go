package syncdaopq

import (
	"database/sql"
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncutil"
)

//NewDataRepository provides postgressql database access for a DataRepository.
func NewDataRepository(db *sql.DB) syncapi.DataRepositoryable {
	return dataRepositoryType{
		db: db,
	}
}

type dataRepositoryType struct {
	db *sql.DB
}

func (dataRepository dataRepositoryType) CreateMessageFetcher(sessionID string, nodeID string) (syncapi.MessageFetching, error) {
	fetcher, err := newMessageFetcher(sessionID, nodeID, dataRepository.db)
	if err != nil {
		syncutil.Error(err)
		return fetcher, err
	}
	return fetcher, nil
}

func (dataRepository dataRepositoryType) CreateMessageProcessor(sessionID string, nodeID string, entitiesByPluralName map[string]syncapi.EntityNameItem) (syncapi.MessageProcessing, error) {
	processor, err := newMessageProcessor(sessionID, nodeID, entitiesByPluralName, dataRepository.db)
	if err != nil {
		syncutil.Error(err)
		return processor, err
	}
	return processor, nil
}

func (dataRepository dataRepositoryType) CreateMessageQueuer(sessionID string, nodeID string) (syncapi.MessageQueuing, error) {
	return newMessageQueuer(dataRepository.db)
}

//NewConfigRepository provides postgressql database access for ConfigurationRepository.
func NewConfigRepository(db *sql.DB) syncapi.ConfigRepositoryable {
	return configRepositoryType{
		db: db,
	}
}

type configRepositoryType struct {
	db *sql.DB
}

func (configRepository configRepositoryType) CreateEntityFetcher(sessionID string, nodeID string) (syncapi.EntityFetching, error) {
	fetcher, err := newEntityFetcher(sessionID, nodeID, configRepository.db)
	if err != nil {
		syncutil.Error(err)
		return fetcher, err
	}
	return fetcher, nil
}
