package syncdaopq

import (
	"database/sql"
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncutil"
)

//NewSyncMessagesFetcher creates an instance of the struct SyncMessagesFetcherType
func newEntityFetcher(sessionID string, nodeID string, db *sql.DB) (syncapi.EntityFetching, error) {
	fetcher := postgresSQLEntityFetcher{
		db: db,
	}
	return fetcher, nil
}

//postgresSQLEntityFetcher logically implements EntityFetcher for the PostgressSQL database.
type postgresSQLEntityFetcher struct {
	db *sql.DB
}

func (fetcher postgresSQLEntityFetcher) FindEntitiesForFetch(orderNum int, sessionID string, nodeID string, changeType syncapi.ProcessSyncChangeEnum) ([]syncapi.EntityNameItem, error) {
	// TODO(doug4j@gmail.com): Add caching on a per session basis so we only hit the database once per orderNum

	var answer = []syncapi.EntityNameItem{}
	var rows *sql.Rows
	var err error

	if changeType == syncapi.ProcessSyncChangeEnumAddOrUpdate {
		rows, err = fetcher.db.Query(sqlFindAddOrUpdateEntities, orderNum, nodeID)
	} else {
		rows, err = fetcher.db.Query(sqlFindDeleteEntities, orderNum, nodeID)
	}
	if err != nil {
		syncutil.Error(err)
		return answer, err
	}

	closeRowQuietly := func() {
		err := rows.Close()
		if err != nil {
			syncutil.Error("Quietly handling of row close error. Error: " + err.Error())
		}
	}
	defer closeRowQuietly()

	var entitySingularName, entityPluralName string

	for rows.Next() {
		err = rows.Scan(&entitySingularName, &entityPluralName)
		if err != nil {
			syncutil.Error(err.Error())
			return answer, err
		}
		item := syncapi.EntityNameItem{
			SingularName: entitySingularName,
			PluralName:   entityPluralName,
		}
		answer = append(answer, item)
	}
	return answer, nil
}

func (fetcher postgresSQLEntityFetcher) FindPluralEntityNamesByID(sessionID string, nodeID string) (map[string]syncapi.EntityNameItem, error) {
	// TODO(doug4j@gmail.com): FindEntitiesForProcess
	var answer = map[string]syncapi.EntityNameItem{}

	var rows *sql.Rows
	var err error

	rows, err = fetcher.db.Query(sqlFindEntityNamesByNodeID, nodeID)
	if err != nil {
		msg := "Cannot get the entities using  nodeID '" + nodeID + "'. " + err.Error()
		syncutil.Error(msg)
		return answer, err
	}

	closeRowQuietly := func() {
		err := rows.Close()
		if err != nil {
			syncutil.Error("Quietly handling of row close error. Error: " + err.Error())
		}
	}
	defer closeRowQuietly()

	var entitySingularName, entityPluralName string

	for rows.Next() {
		err = rows.Scan(&entitySingularName, &entityPluralName)
		if err != nil {
			syncutil.Error(err.Error())
			return answer, err
		}
		item := syncapi.EntityNameItem{
			SingularName: entitySingularName,
			PluralName:   entityPluralName,
		}
		answer[entityPluralName] = item
	}
	return answer, nil
}

const (
	sqlFindEntityNamesByNodeID = `
select EntitySingularName, EntityPluralName from sync_data_entity where DataVersionName in (select DataVersionName from sync_node where NodeID=$1);
`
	// 	sqlFindEntityNamesByNodeID = `
	// select EntitySingularName, EntityPluralName from sync_data_entity where ProcOrderAddUpdate=$1
	// 	AND (DataVersionName in (select DataVersionName from sync_node where NodeID=$1));
	// `

	sqlFindAddOrUpdateEntities = `
select EntitySingularName, EntityPluralName from sync_data_entity where ProcOrderAddUpdate=$1
	AND (DataVersionName in (select DataVersionName from sync_node where NodeID=$2));
`

	sqlFindDeleteEntities = `
select EntitySingularName, EntityPluralName from sync_data_entity where ProcOrderDelete=$1
AND (DataVersionName in (select DataVersionName from sync_node where NodeID=$2));
`
)
