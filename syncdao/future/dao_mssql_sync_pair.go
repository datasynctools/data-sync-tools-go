package main

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"log"
	"time"
)

type SyncPairMsSqlDao struct {
	db *sql.DB
}

func (self SyncPairMsSqlDao) GetPairByNames(requestingNodeName string, toPairWithNodeName string) (SyncPair, error) {
	//TODO: Add SQL injection checks (see http://go-database-sql.org/retrieving.html and http://stackoverflow.com/questions/26345318/how-can-i-prevent-sql-injection-attacks-in-go-while-using-database-sql
	var syncPair SyncPair
	sqlStr := `
SELECT        sync_pair.PairId, sync_pair.PairName, sync_pair.MaxSesDurValue, sync_pair.MaxSesDurUnit, sync_pair.SyncDataTransForm, sync_pair.SyncMsgTransForm, 
                         sync_pair.SyncMsgSecPol, sync_pair.SyncSessionId, sync_pair.SyncSessionState, sync_pair.SyncSessionStart, sync_pair.SyncConflictUri, sync_pair.RecordCreated
FROM            sync_node INNER JOIN
                         sync_pair_nodes ON sync_node.NodeId = sync_pair_nodes.NodeId INNER JOIN
                         sync_pair ON sync_pair_nodes.PairId = sync_pair.PairId where ( 

 ( (sync_pair_nodes.NodeId = (select sync_node.NodeId from sync_node where(sync_node.NodeName = $1))) AND
   (sync_pair_nodes.TargetNodeId = (select sync_node.NodeId from sync_node where(sync_node.NodeName = $2))) ) OR
 ( (sync_pair_nodes.NodeId = (select sync_node.NodeId from sync_node where(sync_node.NodeName = $2))) AND
   (sync_pair_nodes.TargetNodeId = (select sync_node.NodeId from sync_node where(sync_node.NodeName = $1))) )

) limit 3;
`
	rows, err := self.db.Query(sqlStr, requestingNodeName, toPairWithNodeName)
	if err == sql.ErrNoRows {
		//msg := "No records found, expected exactly 2."
		return syncPair, daoNoDataFound
	}
	defer rows.Close()
	var rowCount int = 0
	for rows.Next() {
		rowCount++
		if rowCount == 1 {
			var (
				pairId            string
				pairName          string
				maxSesDurValue    int
				maxSesDurUnit     string
				syncDataTransForm string
				syncMsgTransForm  string
				syncMsgSecPol     string
				syncSessionId     sql.NullString
				syncSessionState  string
				syncSessionStart  pq.NullTime
				syncConflictUri   string
				recordCreated     pq.NullTime
			)
			if err := rows.Scan(&pairId, &pairName, &maxSesDurValue, &maxSesDurUnit, &syncDataTransForm,
				&syncMsgTransForm, &syncMsgSecPol, &syncSessionId, &syncSessionState, &syncSessionStart,
				&syncConflictUri, &recordCreated); err != nil {

				return syncPair, err
			}
			syncPair = SyncPair{
				PairId:            pairId,
				PairName:          pairName,
				MaxSesDurValue:    maxSesDurValue,
				MaxSesDurUnit:     maxSesDurUnit,
				SyncDataTransForm: syncDataTransForm,
				SyncMsgTransForm:  syncMsgTransForm,
				SyncMsgSecPol:     syncMsgSecPol,
				SyncSessionId:     syncSessionId.String,
				SyncSessionState:  syncSessionState,
				SyncSessionStart:  syncSessionStart.Time,
				SyncConflictUri:   syncConflictUri,
				RecordCreated:     recordCreated.Time,
			}
		}
	}
	if rowCount == 0 {
		//msg := "No records found, expected exactly 2."
		return syncPair, daoNoDataFound
	} else if rowCount == 1 {
		msg := "Only 1 record found, expected exactly 2."
		return syncPair, errors.New(msg)
	} else if rowCount > 2 {
		msg := "More than 2 records found, expected exactly 2."
		return syncPair, errors.New(msg)
	}
	//Note: if we're here, the row count was 2, which is correct (we had a pair)
	return syncPair, nil
}

func (self SyncPairMsSqlDao) GetNodePairItem(pairId string, nodeName string) (NodePairItem, error) {
	var nodePairItem NodePairItem
	sqlStr := `
SELECT        sync_node.NodeId, sync_node.Enabled, sync_node.DataMsgConsUri, sync_node.DataMsgProdUri, 
			  sync_node.MgmtMsgConsUri, sync_node.MgmtMsgProdUri, sync_node.SyncDataPersistForm, 
			  sync_node.InMsgBatchSize, sync_node.MaxOutMsgBatchSize, sync_node.InChanDepthSize, 
			  sync_node.MaxOutChanDepthSize, sync_node.RecordCreated
FROM            sync_node INNER JOIN
                         sync_pair_nodes ON sync_node.NodeId = sync_pair_nodes.NodeId INNER JOIN
                         sync_pair ON sync_pair_nodes.PairId = sync_pair.PairId
WHERE        (sync_pair_nodes.PairId = $1 and sync_node.NodeName=$2)
`
	row := self.db.QueryRow(sqlStr, pairId, nodeName)
	var (
		nodeId                    string
		enabled                   bool
		dataMsgConsumerUri        sql.NullString
		dataMsgProducerUri        sql.NullString
		mgmtMsgConsumerUri        sql.NullString
		mgmtMsgProducerUri        sql.NullString
		syncDataPersistanceFormat sql.NullString
		inMsgBatchSize            int
		maxOutMsgBatchSize        int
		inChanDepthSize           int
		maxOutChanDeptSize        int
		//TODO: Leverage record created in output
		recordCreated pq.NullTime
	)
	if err := row.Scan(&nodeId, &enabled, &dataMsgConsumerUri, &dataMsgProducerUri, &mgmtMsgConsumerUri,
		&mgmtMsgProducerUri, &syncDataPersistanceFormat, &inMsgBatchSize, &maxOutMsgBatchSize,
		&inChanDepthSize, &maxOutChanDeptSize, &recordCreated); err != nil {
		if err == sql.ErrNoRows {
			return nodePairItem, daoNoDataFound
		}
		return nodePairItem, err
	}

	nodePairItem = NodePairItem{
		NodeId:                    nodeId,
		NodeName:                  nodeName,
		Enabled:                   enabled,
		DataMsgConsumerUri:        dataMsgConsumerUri.String,
		DataMsgProducerUri:        dataMsgProducerUri.String,
		MgmtMsgConsumerUri:        mgmtMsgConsumerUri.String,
		MgmtMsgProducerUri:        mgmtMsgProducerUri.String,
		SyncDataPersistanceFormat: syncDataPersistanceFormat.String,
		InMsgBatchSize:            inMsgBatchSize,
		MaxOutMsgBatchSize:        maxOutMsgBatchSize,
		InChanDepthSize:           inChanDepthSize,
		MaxOutChanDeptSize:        maxOutChanDeptSize,
		DataVersionName:           "-not implemented-",
	}
	return nodePairItem, nil
}

func (self SyncPairMsSqlDao) GetEntityPairItem(pairId string, nodeName string) ([]EntityPairItem, error) {
	var items []EntityPairItem
	sqlStr := `
SELECT      sync_entity.EntityName, sync_version_entities.ProcOrderAddUpdate, sync_version_entities.ProcessOrderDelete, sync_entity.EntityHandlerUri
FROM            sync_data_version INNER JOIN
                         sync_version_entities ON sync_data_version.VersionId = sync_version_entities.VersionId INNER JOIN
                         sync_entity ON sync_version_entities.EntityId = sync_entity.EntityId CROSS JOIN
                         sync_node
WHERE        (sync_node.NodeName = $1);
`
	rows, err := self.db.Query(sqlStr, nodeName)
	if err == sql.ErrNoRows {
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			entityName            string
			processOrderAddUpdate int
			processOrderDelete    int
			entityHandlerUri      string
		)
		if err := rows.Scan(&entityName, &processOrderAddUpdate, &processOrderDelete, &entityHandlerUri); err != nil {
			if err == sql.ErrNoRows {
				return items, daoNoDataFound
			}
			return items, err
		}
		entityPairItem := EntityPairItem{
			EntityName:            entityName,
			ProcessOrderAddUpdate: processOrderAddUpdate,
			ProcessOrderDelete:    processOrderDelete,
			EntityHandlerUri:      entityHandlerUri,
		}
		items = append(items, entityPairItem)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return items, nil
}

func (self SyncPairMsSqlDao) CreateSyncSession(item CreateSyncSessionRequest) (CreateSyncSessionDaoResult, error) {
	var answer CreateSyncSessionDaoResult
	var actualSessionId string
	sqlStr := `
Update sync_pair SET SyncSessionId=$2, SyncSessionStart=$3, SyncSessionState='Initializing' 
where SyncSessionState='Inactive' and PairId=$1;`
	now := time.Now()
	format := "2006-01-02 15:04:05.000"
	dateString := now.Format(format)
	result, err := self.db.Exec(sqlStr, item.PairId, item.SessionId, now.Format(format))
	if err != nil {
		msg := "Error updating to database, inputData=PairId:'%s', SessionId:'%s', dateString:'%s'. Error:%s"
		log.Printf(msg, item.PairId, item.SessionId, dateString, err.Error())
		//log.Println("Error inserting to database, inputData=", item)
		return answer, err
	}
	var affectedCount int64
	affectedCount, err = result.RowsAffected()
	if err != nil {
		msg := "Error getting results from database, inputData=PairId:'%s', SessionId:'%s', dateString:'%s'. Error:%s"
		log.Printf(msg, item.PairId, item.SessionId, dateString, err.Error())
		//log.Println("Error inserting to database, inputData=", item)
		return answer, err
	}
	if affectedCount == 0 {
		//log.Println("No rows affected")
		//Get the actual id
		sqlStr := "Select SyncSessionId from sync_pair where PairId=$1;"
		row := self.db.QueryRow(sqlStr, item.PairId)
		if err := row.Scan(&actualSessionId); err != nil {
			return answer, err
		}
		//log.Println("Found existing sessionId: " + actualSessionId)
		if item.SessionId == actualSessionId {
			answer = CreateSyncSessionDaoResult{
				Result:          "ThisSessionIdAlreadyActive",
				ActualSessionId: actualSessionId,
			}
		} else {
			answer = CreateSyncSessionDaoResult{
				Result:          "DifferentSessionIdAlreadyActive",
				ActualSessionId: actualSessionId,
			}
		}
		return answer, nil
	} else if affectedCount == 1 {
		answer = CreateSyncSessionDaoResult{
			Result:          "OK",
			ActualSessionId: item.SessionId,
		}
		//log.Println("One row affected")
		return answer, nil
	} else {
		msg := "More than one row was affected by the SQL update. this is very unexpected and suggests a " +
			"mis-configuration or incorrect query"
		err = errors.New(msg)
		log.Printf(msg, item.PairId, item.SessionId, dateString, err.Error())
		return answer, err
	}
}

func (self SyncPairMsSqlDao) CloseSyncSession(item CloseSyncSessionRequest) (CloseSyncSessionDaoResult, error) {
	var answer CloseSyncSessionDaoResult
	var actualSessionId string
	sqlStr := `
Update sync_pair SET SyncSessionId=null, SyncSessionStart=null, SyncSessionState='Inactive' where PairId=$1 and 
SyncSessionId=$2 and SyncSessionState<>'Inactive';`
	result, err := self.db.Exec(sqlStr, item.PairId, item.SessionId)
	if err != nil {
		msg := "Error updating to database, inputData=PairId:'%s', SessionId:'%s'. Error:%s"
		log.Printf(msg, item.PairId, item.SessionId, err.Error())
		//log.Println("Error inserting to database, inputData=", item)
		return answer, err
	}
	var affectedCount int64
	affectedCount, err = result.RowsAffected()
	if err != nil {
		msg := "Error getting results from database, inputData=PairId:'%s', SessionId:'%s'. Error:%s"
		log.Printf(msg, item.PairId, item.SessionId, err.Error())
		//log.Println("Error inserting to database, inputData=", item)
		return answer, err
	}
	if affectedCount == 0 {
		//log.Println("No rows affected")
		//Get the actual id
		sqlStr := "Select SyncSessionId from sync_pair where PairId=$1;"
		row := self.db.QueryRow(sqlStr, item.PairId)
		if err := row.Scan(&actualSessionId); err != nil {
			//actualSessionId was null
			answer = CloseSyncSessionDaoResult{
				Result:          "ThisSessionIdAlreadyInactive",
				ActualSessionId: actualSessionId,
			}
		} else {
			answer = CloseSyncSessionDaoResult{
				Result:          "DifferentSessionIdAlreadyActive",
				ActualSessionId: actualSessionId,
			}
		}
		return answer, nil
	} else if affectedCount == 1 {
		answer = CloseSyncSessionDaoResult{
			Result:          "OK",
			ActualSessionId: item.SessionId,
		}
		//log.Println("One row affected")
		return answer, nil
	} else {
		msg := "More than one row was affected by the SQL update. this is very unexpected and suggests a " +
			"mis-configuration or incorrect query"
		err = errors.New(msg)
		log.Printf(msg, item.PairId, item.SessionId, err.Error())
		return answer, err
	}
}
