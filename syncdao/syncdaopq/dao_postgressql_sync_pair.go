package syncdaopq

import (
	"database/sql"
	"data-sync-tools-go/syncdao"
	"data-sync-tools-go/syncutil"
	"errors"
	"log"
	"time"

	"github.com/lib/pq"
)

//SyncPairPostgresSQLDao implements the syncdao.SyncPairDao interface via an postgressql database.
type SyncPairPostgresSQLDao struct {
	db *sql.DB
}

//GetPairByNames gets the SyncPair by sync node names via a postgressql database.
func (dao SyncPairPostgresSQLDao) GetPairByNames(requestingNodeName string, toPairWithNodeName string) (syncdao.SyncPair, error) {
	// BUG(doug4j@gmail.com): Add SQL injection checks (see http://go-database-sql.org/retrieving.html and http://stackoverflow.com/questions/26345318/how-can-i-prevent-sql-injection-attacks-in-go-while-using-database-sql
	var syncPair syncdao.SyncPair
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
	rows, err := dao.db.Query(sqlStr, requestingNodeName, toPairWithNodeName)
	if err == sql.ErrNoRows {
		//msg := "No records found, expected exactly 2."
		return syncPair, syncdao.ErrDaoNoDataFound
	}
	closeRowQuietly := func() {
		err := rows.Close()
		if err != nil {
			syncutil.Error("Quietly handling of row close error. Error: " + err.Error())
		}
	}
	defer closeRowQuietly()
	var rowCount int //by default this is of course 0
	for rows.Next() {
		rowCount++
		if rowCount == 1 {
			var (
				pairID            string
				pairName          string
				maxSesDurValue    int
				maxSesDurUnit     string
				syncDataTransForm string
				syncMsgTransForm  string
				syncMsgSecPol     string
				syncSessionID     sql.NullString
				syncSessionState  string
				syncSessionStart  pq.NullTime
				syncConflictURI   string
				recordCreated     pq.NullTime
			)
			if err := rows.Scan(&pairID, &pairName, &maxSesDurValue, &maxSesDurUnit, &syncDataTransForm,
				&syncMsgTransForm, &syncMsgSecPol, &syncSessionID, &syncSessionState, &syncSessionStart,
				&syncConflictURI, &recordCreated); err != nil {

				return syncPair, err
			}
			syncPair = syncdao.SyncPair{
				PairID:            pairID,
				PairName:          pairName,
				MaxSesDurValue:    maxSesDurValue,
				MaxSesDurUnit:     maxSesDurUnit,
				SyncDataTransForm: syncDataTransForm,
				SyncMsgTransForm:  syncMsgTransForm,
				SyncMsgSecPol:     syncMsgSecPol,
				SyncSessionID:     syncSessionID.String,
				SyncSessionState:  syncSessionState,
				SyncSessionStart:  syncSessionStart.Time,
				SyncConflictURI:   syncConflictURI,
				RecordCreated:     recordCreated.Time,
			}
		}
	}
	if rowCount == 0 {
		//msg := "No records found, expected exactly 2."
		return syncPair, syncdao.ErrDaoNoDataFound
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

//GetNodePairItem gets the NodePairItem by pairId and node name via a postgressql database.
func (dao SyncPairPostgresSQLDao) GetNodePairItem(pairID string, nodeName string) (syncdao.NodePairItem, error) {
	var nodePairItem syncdao.NodePairItem
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
	row := dao.db.QueryRow(sqlStr, pairID, nodeName)
	var (
		nodeID                    string
		enabled                   bool
		dataMsgConsumerURI        sql.NullString
		dataMsgProducerURI        sql.NullString
		mgmtMsgConsumerURI        sql.NullString
		mgmtMsgProducerURI        sql.NullString
		syncDataPersistanceFormat sql.NullString
		inMsgBatchSize            int
		maxOutMsgBatchSize        int
		inChanDepthSize           int
		maxOutChanDeptSize        int
		// BUG(doug4j@gmail.com): Leverage record created in output
		recordCreated pq.NullTime
	)
	if err := row.Scan(&nodeID, &enabled, &dataMsgConsumerURI, &dataMsgProducerURI, &mgmtMsgConsumerURI,
		&mgmtMsgProducerURI, &syncDataPersistanceFormat, &inMsgBatchSize, &maxOutMsgBatchSize,
		&inChanDepthSize, &maxOutChanDeptSize, &recordCreated); err != nil {
		if err == sql.ErrNoRows {
			return nodePairItem, syncdao.ErrDaoNoDataFound
		}
		return nodePairItem, err
	}

	nodePairItem = syncdao.NodePairItem{
		NodeID:                    nodeID,
		NodeName:                  nodeName,
		Enabled:                   enabled,
		DataMsgConsumerURI:        dataMsgConsumerURI.String,
		DataMsgProducerURI:        dataMsgProducerURI.String,
		MgmtMsgConsumerURI:        mgmtMsgConsumerURI.String,
		MgmtMsgProducerURI:        mgmtMsgProducerURI.String,
		SyncDataPersistanceFormat: syncDataPersistanceFormat.String,
		InMsgBatchSize:            inMsgBatchSize,
		MaxOutMsgBatchSize:        maxOutMsgBatchSize,
		InChanDepthSize:           inChanDepthSize,
		MaxOutChanDeptSize:        maxOutChanDeptSize,
		DataVersionName:           "-not implemented-",
	}
	return nodePairItem, nil
}

//GetEntityPairItem retrieves the EntityPairId by pairId and node name via an postgressql database.
func (dao SyncPairPostgresSQLDao) GetEntityPairItem(pairID string, nodeName string) ([]syncdao.EntityPairItem, error) {
	var items []syncdao.EntityPairItem
	sqlStr := `
SELECT        sync_data_entity.EntitySingularName, sync_data_entity.EntityPluralName, sync_data_entity.ProcOrderAddUpdate, sync_data_entity.ProcOrderDelete, sync_data_entity.EntityHandlerUri
FROM            sync_node INNER JOIN
                         sync_data_version ON sync_node.DataVersionName = sync_data_version.DataVersionName INNER JOIN
                         sync_data_entity ON sync_data_version.DataVersionName = sync_data_entity.DataVersionName
WHERE			(sync_node.NodeName = $1);
`
	rows, err := dao.db.Query(sqlStr, nodeName)
	if err == sql.ErrNoRows {
		return items, err
	} else if err != nil {
		msg := "Error getting entity pairinputData=PairId:'%s', NodeName:'%s'. Error:%s"
		log.Printf(msg, pairID, nodeName, err.Error())
		//log.Println("Error inserting to database, inputData=", item)
		return items, err
	}
	closeRowQuietly := func() {
		err := rows.Close()
		if err != nil {
			syncutil.Error("Quietly handling of row close error. Error: " + err.Error())
		}
	}
	defer closeRowQuietly()
	for rows.Next() {
		var (
			entitySingularName, entityPluralName, entityHandlerURI string
			processOrderAddUpdate, processOrderDelete              int
		)
		if err := rows.Scan(&entitySingularName, &entityPluralName, &processOrderAddUpdate, &processOrderDelete, &entityHandlerURI); err != nil {
			if err == sql.ErrNoRows {
				return items, syncdao.ErrDaoNoDataFound
			}
			return items, err
		}
		entityPairItem := syncdao.EntityPairItem{
			EntitySingularName:    entitySingularName,
			EntityPluralName:      entityPluralName,
			ProcessOrderAddUpdate: processOrderAddUpdate,
			ProcessOrderDelete:    processOrderDelete,
			EntityHandlerURI:      entityHandlerURI,
		}
		items = append(items, entityPairItem)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
		return items, err
	}
	return items, nil
}

//NullTime represents an empty time value.
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

//CreateSyncSession creates a session between sync pairs via a postgressql database.
func (dao SyncPairPostgresSQLDao) CreateSyncSession(item syncdao.CreateSyncSessionRequest) (syncdao.CreateSyncSessionDaoResult, error) {
	var answer syncdao.CreateSyncSessionDaoResult
	var actualSessionID string
	sqlStr := `
Update sync_pair SET SyncSessionId=$2, SyncSessionStart=$3, SyncSessionState='Initializing'
where SyncSessionState='Inactive' and PairId=$1;`
	now := time.Now()
	format := "2006-01-02 15:04:05.000"
	dateString := now.Format(format)
	result, err := dao.db.Exec(sqlStr, item.PairID, item.SessionID, now.Format(format))
	if err != nil {
		msg := "Error updating to database, inputData=PairId:'%s', SessionId:'%s', dateString:'%s'. Error:%s"
		log.Printf(msg, item.PairID, item.SessionID, dateString, err.Error())
		//log.Println("Error inserting to database, inputData=", item)
		return answer, err
	}
	var affectedCount int64
	affectedCount, err = result.RowsAffected()
	if err != nil {
		msg := "Error getting results from database, inputData=PairId:'%s', SessionId:'%s', dateString:'%s'. Error:%s"
		log.Printf(msg, item.PairID, item.SessionID, dateString, err.Error())
		//log.Println("Error inserting to database, inputData=", item)
		return answer, err
	}
	if affectedCount == 0 {
		//log.Println("No rows affected")
		//Get the actual id
		sqlStr := "Select SyncSessionId from sync_pair where PairId=$1;"
		row := dao.db.QueryRow(sqlStr, item.PairID)
		if err := row.Scan(&actualSessionID); err != nil {
			return answer, err
		}
		//log.Println("Found existing sessionId: " + actualSessionId)
		if item.SessionID == actualSessionID {
			answer = syncdao.CreateSyncSessionDaoResult{
				Result:          "ThisSessionIdAlreadyActive",
				ActualSessionID: actualSessionID,
			}
		} else {
			answer = syncdao.CreateSyncSessionDaoResult{
				Result:          "DifferentSessionIdAlreadyActive",
				ActualSessionID: actualSessionID,
			}
		}
		return answer, nil
	} else if affectedCount == 1 {
		answer = syncdao.CreateSyncSessionDaoResult{
			Result:          "OK",
			ActualSessionID: item.SessionID,
		}
		//log.Println("One row affected")
		return answer, nil
	} else {
		msg := "More than one row was affected by the SQL update. this is very unexpected and suggests a " +
			"mis-configuration or incorrect query"
		err = errors.New(msg)
		log.Printf(msg, item.PairID, item.SessionID, dateString, err.Error())
		return answer, err
	}
}

//CloseSyncSession closes a sync session via an postgressql database.
func (dao SyncPairPostgresSQLDao) CloseSyncSession(item syncdao.CloseSyncSessionRequest) (syncdao.CloseSyncSessionDaoResult, error) {
	var answer syncdao.CloseSyncSessionDaoResult
	var actualSessionID string
	sqlStr := `
Update sync_pair SET SyncSessionId=null, SyncSessionStart=null, SyncSessionState='Inactive' where PairId=$1 and
SyncSessionId=$2 and SyncSessionState<>'Inactive';`
	result, err := dao.db.Exec(sqlStr, item.PairID, item.SessionID)
	if err != nil {
		msg := "Error updating to database, inputData=PairId:'%s', SessionId:'%s'. Error:%s"
		log.Printf(msg, item.PairID, item.SessionID, err.Error())
		//log.Println("Error inserting to database, inputData=", item)
		return answer, err
	}
	var affectedCount int64
	affectedCount, err = result.RowsAffected()
	if err != nil {
		msg := "Error getting results from database, inputData=PairId:'%s', SessionId:'%s'. Error:%s"
		log.Printf(msg, item.PairID, item.SessionID, err.Error())
		//log.Println("Error inserting to database, inputData=", item)
		return answer, err
	}
	if affectedCount == 0 {
		//log.Println("No rows affected")
		//Get the actual id
		sqlStr := "Select SyncSessionId from sync_pair where PairId=$1;"
		row := dao.db.QueryRow(sqlStr, item.PairID)
		if err := row.Scan(&actualSessionID); err != nil {
			//actualSessionId was null
			answer = syncdao.CloseSyncSessionDaoResult{
				Result:          "ThisSessionIdAlreadyInactive",
				ActualSessionID: actualSessionID,
			}
		} else {
			answer = syncdao.CloseSyncSessionDaoResult{
				Result:          "DifferentSessionIdAlreadyActive",
				ActualSessionID: actualSessionID,
			}
		}
		return answer, nil
	} else if affectedCount == 1 {
		answer = syncdao.CloseSyncSessionDaoResult{
			Result:          "OK",
			ActualSessionID: item.SessionID,
		}
		//log.Println("One row affected")
		return answer, nil
	} else {
		msg := "More than one row was affected by the SQL update. this is very unexpected and suggests a " +
			"mis-configuration or incorrect query"
		err = errors.New(msg)
		log.Printf(msg, item.PairID, item.SessionID, err.Error())
		return answer, err
	}
}

//UpdateSyncSessionState updates the sync session state via a postgressql database.
func (dao SyncPairPostgresSQLDao) UpdateSyncSessionState(item syncdao.UpdateSyncSessionStateRequest) (syncdao.UpdateSyncSessionStateResult, error) {
	var (
		answer          syncdao.UpdateSyncSessionStateResult
		actualSessionID string
		resultingState  string
	)
	sqlStr := `
Update sync_pair SET SyncSessionState=$3 where PairId=$1 and SyncSessionId=$2 and SyncSessionState<>'Inactive';
`
	result, err := dao.db.Exec(sqlStr, item.PairID, item.SessionID, item.State)
	if err != nil {
		msg := "Error getting results from database, inputData=SessionId:'%s', State:'%s'. Error:%s"
		log.Printf(msg, item.SessionID, item.State, err.Error())
		//log.Println("Error inserting to database, inputData=", item)
		return answer, err
	}
	var affectedCount int64
	affectedCount, err = result.RowsAffected()
	if affectedCount == 0 {
		//log.Println("No rows affected")
		//Get the actual id
		sqlStr := "Select SyncSessionId, SyncSessionState from sync_pair where PairId=$1;"
		row := dao.db.QueryRow(sqlStr, item.PairID)
		if err := row.Scan(&actualSessionID, &resultingState); err != nil {
			//actualSessionId was null
			answer = syncdao.UpdateSyncSessionStateResult{
				Result:             "CouldNoFindActiveSessionToUpdate",
				ResultMsg:          "",
				RequestedSessionID: item.SessionID,
				ActualSessionID:    "",
				RequestedState:     item.State,
				ResultingState:     "",
			}
		} else {
			answer = syncdao.UpdateSyncSessionStateResult{
				Result:             "CouldNoFindActiveSessionToUpdate",
				ResultMsg:          "",
				RequestedSessionID: item.SessionID,
				ActualSessionID:    actualSessionID,
				RequestedState:     item.State,
				ResultingState:     resultingState,
			}
		}
		return answer, nil
	} else if affectedCount == 1 {
		answer = syncdao.UpdateSyncSessionStateResult{
			Result:             "OK",
			ResultMsg:          "",
			RequestedSessionID: item.SessionID,
			ActualSessionID:    item.SessionID,
			RequestedState:     item.State,
			ResultingState:     item.State,
		}
		//log.Println("One row affected")
		return answer, nil
	} else {
		msg := "More than one row was affected by the SQL update. this is very unexpected and suggests a " +
			"mis-configuration or incorrect query"
		err = errors.New(msg)
		//log.Printf(msg, item.PairId, item.SessionId, err.Error())
		return answer, err
	}
}

//QueryPairState queries the current pair state via a postgressql database.
func (dao SyncPairPostgresSQLDao) QueryPairState(item syncdao.QueryPairStateRequest) (syncdao.QueryPairStateDaoResult, error) {
	var (
		sessionID    sql.NullString
		sessionState string
		sessionStart pq.NullTime
	)
	sqlStr := `
SELECT        sync_pair.SyncSessionState, sync_pair.SyncSessionStart, sync_pair.SyncSessionId
FROM          sync_pair
WHERE         (sync_pair.PairId = $1)
`
	err := dao.db.QueryRow(sqlStr, item.PairID).Scan(&sessionState, &sessionStart, &sessionID)
	var answer syncdao.QueryPairStateDaoResult
	var lastUpdated time.Time
	switch {
	case err == sql.ErrNoRows:
		log.Println("No Data")
		return answer, syncdao.ErrDaoNoDataFound
	case err != nil:
		log.Println("DB Error")
		return answer, err
	default:
		answer = syncdao.QueryPairStateDaoResult{
			State:        sessionState,
			SessionID:    sessionID.String,
			SessionStart: sessionStart.Time,
			LastUpdated:  lastUpdated,
		}
	}

	return answer, nil
}
