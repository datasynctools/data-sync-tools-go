package syncdaopq

import (
	"database/sql"
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncmsg"
	"data-sync-tools-go/syncutil"
	"encoding/hex"

	"github.com/golang/protobuf/proto"
	"github.com/twinj/uuid"
)

//NewSyncMessagesFetcher creates an instance of the struct SyncMessagesFetcherType
func newMessageFetcher(SessionID string, NodeID string, db *sql.DB) (syncapi.MessageFetching, error) {

	// BUG(doug4j@gmail.com): MaxGroupSize and MaxMsgs Should come from configuration and not be hard coded

	fetcherType := syncapi.MessageFetchingData{
		SessionID:         SessionID,
		NodeID:            NodeID,
		MaxGroupBytesSize: 50, //Hard Coded for Now
		MaxMsgs:           1,  //Hard Coded for Now
	}
	fetcher := postgresSQLMessageFetcher{
		MessageFetchingData: fetcherType,
		db:                  db,
	}
	//syncutil.Info(fetcher)
	return fetcher, nil
}

//postgresSQLMessageFetcher glogically implements SyncMessagesFetcherStruct for the PostgressSQL database.
type postgresSQLMessageFetcher struct {
	syncapi.MessageFetchingData
	db *sql.DB
}

//Fetch retrieves a group of sync messages for downstream processes returning a value with no request items
//in it should it be (Copied from SyncMessagesFetcher.Fetch() interface)
func (fetcher postgresSQLMessageFetcher) Fetch(entities []syncapi.EntityNameItem, changeType syncapi.ProcessSyncChangeEnum) (*syncmsg.ProtoRequestSyncEntityMessageResponse, error) {
	lastState := newFetchedState()
	bindID := uuid.Formatter(uuid.NewV4(), uuid.FormatCanonical)
	isDelete := *changeType.Enum() == syncapi.ProcessSyncChangeEnumDelete

	answer := &syncmsg.ProtoRequestSyncEntityMessageResponse{}

	request := &syncmsg.ProtoSyncEntityMessageRequest{
		IsDelete:          proto.Bool(isDelete),
		TransactionBindId: proto.String(bindID),
		Items:             make([]*syncmsg.ProtoSyncDataMessagesRequest, 0),
	}
	for _, entity := range entities {

		msgsResponse, lastState, err := fetcher.processEntity(&lastState, entity, changeType)
		if err != nil {
			syncutil.Error(err)
			answer.Result = syncmsg.SyncRequestEntityMessageResponseResult_ErrorCreatingMsgs.Enum()
			answer.ResultMsg = proto.String(err.Error())
			return answer, err
		}
		if len(msgsResponse.Msgs) > 0 {
			request.Items = append(request.Items, msgsResponse)
		}

		if lastState.readAnotherEntity == false {
			var entityMapByPluralName = fetcher.createEntityMapByPluralName(entities)
			err = fetcher.markItemsWithBindID(bindID, request, entityMapByPluralName)
			if err != nil {
				syncutil.Error(err)
				answer.Result = syncmsg.SyncRequestEntityMessageResponseResult_ErrorCreatingMsgs.Enum()
				answer.ResultMsg = proto.String(err.Error())
				return answer, err
			}
			//syncutil.Debug("Fetch completed. bytes:", lastState.totalBytesProcessed, "count:", lastState.lastProcessedCount)

			answer.Request = request
			if len(request.Items) > 0 {
				answer.Result = syncmsg.SyncRequestEntityMessageResponseResult_HasMsgs.Enum()
			} else {
				answer.Result = syncmsg.SyncRequestEntityMessageResponseResult_NoMsgs.Enum()
			}
			answer.ResultMsg = proto.String("")

			return answer, nil
		}

	}
	var entityMapByPluralName = fetcher.createEntityMapByPluralName(entities)
	err := fetcher.markItemsWithBindID(bindID, request, entityMapByPluralName)
	if err != nil {
		syncutil.Error(err)
		answer.Result = syncmsg.SyncRequestEntityMessageResponseResult_ErrorCreatingMsgs.Enum()
		answer.ResultMsg = proto.String(err.Error())
		return answer, err
	}

	answer.Request = request
	if len(request.Items) > 0 {
		answer.Result = syncmsg.SyncRequestEntityMessageResponseResult_HasMsgs.Enum()
	} else {
		answer.Result = syncmsg.SyncRequestEntityMessageResponseResult_NoMsgs.Enum()
	}
	answer.ResultMsg = proto.String("")

	return answer, nil
}

func (fetcher postgresSQLMessageFetcher) createEntityMapByPluralName(items []syncapi.EntityNameItem) map[string]syncapi.EntityNameItem {
	var entityMapByPluralName = map[string]syncapi.EntityNameItem{}
	for _, item := range items {
		entityMapByPluralName[item.PluralName] = item
	}
	return entityMapByPluralName
}

func (fetcher postgresSQLMessageFetcher) markItemsWithBindID(bindID string, request *syncmsg.ProtoSyncEntityMessageRequest, entityMapByPluralName map[string]syncapi.EntityNameItem) error {
	if len(request.Items) == 0 {
		return nil
	}

	var sql = "update sync_peer_state set transactionBindSendId='" + bindID + "' where nodeId='" + fetcher.NodeID + "' AND"
	for entityCount, item := range request.Items {
		singularEntityName := entityMapByPluralName[*item.EntityPluralName].SingularName
		if entityCount == 0 {
			sql = sql + " (entitySingularName='" + singularEntityName + "' AND ("
		} else {
			sql = sql + " OR (entitySingularName='" + singularEntityName + "' AND ("
		}
		for recordCount, msg := range item.Msgs {
			if recordCount == 0 {
				sql = sql + "recordId='" + *msg.RecordId + "'"
			} else {
				sql = sql + " OR recordId='" + *msg.RecordId + "'"
			}
		}
		sql = sql + "))"
	}
	sql = sql + ";"
	//syncutil.Debug("\n\n", sql, "\n\n")
	_, err := fetcher.db.Exec(sql)
	if err != nil {
		syncutil.Error(err)
		return err
	}

	return nil
}

func (fetcher postgresSQLMessageFetcher) processEntity(lastState *fetchedState, entity syncapi.EntityNameItem, changeType syncapi.ProcessSyncChangeEnum) (*syncmsg.ProtoSyncDataMessagesRequest, *fetchedState, error) {

	//syncutil.Debug("Processing Entity {", entity, "}")
	//var localState = lastState
	var firstTimeReadingEntity = true
	var msgsRequest = &syncmsg.ProtoSyncDataMessagesRequest{
		EntityPluralName: proto.String(entity.PluralName),
	}
	for true {
		queueID := uuid.Formatter(uuid.NewV4(), uuid.FormatCanonical)
		err := fetcher.reserveFetchItems(entity.SingularName, changeType, queueID)
		if err != nil {
			syncutil.Error(err)
			return msgsRequest, lastState, err
		}
		err = fetcher.fetchReserveFetchItems(changeType, queueID, msgsRequest, lastState)
		if err != nil {
			syncutil.Error(err)
			return msgsRequest, lastState, err
		}

		if lastState.readMoreFromEntity == false && firstTimeReadingEntity == true && lastState.lastProcessedCount == 0 {
			//There were not any records on the first read of this entity. Tell the client that we never need to try to
			//fetch items from this entity for the remainder of this session
			lastState.completedSingularEntities = append(lastState.completedSingularEntities, entity.SingularName)
		}
		//syncutil.Info(localState.readMoreFromEntity)
		//time.Sleep(1 * time.Second)
		if lastState.readMoreFromEntity == false {
			return msgsRequest, lastState, nil
		}
	}
	return msgsRequest, lastState, nil
}

type fetchedState struct {
	completedSingularEntities []string //updated by 'processEntity'
	readMoreFromEntity        bool     //updated by 'fetchReserveFetchItems'
	readAnotherEntity         bool     //updated by 'fetchReserveFetchItems'
	lastProcessedCount        int      //updated by 'fetchReserveFetchItems'
	totalBytesProcessed       uint32   //updated by 'fetchReserveFetchItems'
	//entityResults      map[string]*syncmsg.ProtoSyncDataMessagesResponse
}

func newFetchedState() fetchedState {
	return fetchedState{
		completedSingularEntities: make([]string, 0),
		readMoreFromEntity:        true,
		readAnotherEntity:         true,
		lastProcessedCount:        0,
		totalBytesProcessed:       0,
		//entityResults:      make(map[string]*syncmsg.ProtoSyncDataMessagesResponse),
	}
}

func (fetcher postgresSQLMessageFetcher) fetchReserveFetchItems(changeType syncapi.ProcessSyncChangeEnum, queueID string, request *syncmsg.ProtoSyncDataMessagesRequest, previousState *fetchedState) error {
	//Reset the reading state for ineligibility check downstream
	previousState.readMoreFromEntity = true
	rows, err := fetcher.db.Query(sqlFetchReserved, queueID, fetcher.NodeID)
	if err != nil {
		syncutil.Error(err)
		return err
	}

	closeRowQuietly := func() {
		err := rows.Close()
		if err != nil {
			syncutil.Error("Quietly handling of row close error. Error: " + err.Error())
		}
	}
	defer closeRowQuietly()

	var (
		recordID, recordHash string
		sentSyncState        int32
		lastKnownPeerHash    sql.NullString
		recordBytesSize      uint32
		recordBytesAsHex     string
		//recordBytes          []byte
		//recordCreated, lastUpdated time.Time
	)

	var rowsInThisMethod = 0

	for rows.Next() {
		err = rows.Scan(&recordID, &recordHash, &lastKnownPeerHash, &sentSyncState, &recordBytesSize, &recordBytesAsHex)
		if err != nil {
			syncutil.Error("Error scanning for change count. Error:", err.Error())
			return err
		}
		recordBytes, err := hex.DecodeString(recordBytesAsHex)
		if err != nil {
			msg := "Error decoding hex bytes from database. This should not happen as long as data is written in a uniform manner. Error:"
			syncutil.Error(msg, err.Error())
			return err
		}
		err = fetcher.addDataMessagesResponse(queueID, recordID, recordHash, lastKnownPeerHash, sentSyncState, recordBytesSize, recordBytes, request)
		if err != nil {
			syncutil.Error(err)
			return err
		}
		previousState.lastProcessedCount = previousState.lastProcessedCount + 1
		previousState.totalBytesProcessed = previousState.totalBytesProcessed + recordBytesSize
		//syncutil.Info("Processing recordID "+recordID+". lastProcessedCount:", previousState.lastProcessedCount, ", totalBytesProcessed:", previousState.totalBytesProcessed)
		rowsInThisMethod++
		//time.Sleep(1 * time.Second)
	}

	if previousState.totalBytesProcessed >= uint32(fetcher.MaxGroupBytesSize) {
		previousState.readMoreFromEntity = false
		previousState.readAnotherEntity = false
		//syncutil.Info("Total bytes greater than or equal to max group bytes size", fetcher.MaxGroupBytesSize, ". processed", previousState.totalBytesProcessed)
	} else if rowsInThisMethod == 0 {
		previousState.readMoreFromEntity = false
		//previousState.readMoreFromEntity = false
		//syncutil.Debug("No more values from entity {", entity, "}")
	}
	//syncutil.Debug("rowsInThisMethod:", rowsInThisMethod)

	return nil
}

func (fetcher postgresSQLMessageFetcher) addDataMessagesResponse(bindID string, recordID string, recordHash string, lastKnownPeerHash sql.NullString, sentSyncState int32, recordBytesSize uint32, recordBytes []byte, request *syncmsg.ProtoSyncDataMessagesRequest) error {

	sentSyncStateEnum, err := syncmsg.CreateSentSyncState(sentSyncState)
	if err != nil {
		syncutil.Error(err)
		return err
	}
	//If this msg has never been sent to the client, mark it as the first time sent to the client
	if sentSyncStateEnum == syncmsg.SentSyncStateEnum_PersistedNeverSentToPeer {
		sentSyncStateEnum = syncmsg.SentSyncStateEnum_PersistedFirstTimeSentToPeer
	}
	if lastKnownPeerHash.Valid {
		request.Msgs = append(request.Msgs, &syncmsg.ProtoSyncDataMessageRequest{
			RecordId:          proto.String(recordID),
			RecordHash:        proto.String(recordHash),
			LastKnownPeerHash: proto.String(lastKnownPeerHash.String),
			SentSyncState:     sentSyncStateEnum.Enum(),
			RecordBytesSize:   proto.Uint32(recordBytesSize),
			RecordData:        recordBytes,
		})
	} else {
		//Don't include optional value LastKnownPeerHash
		request.Msgs = append(request.Msgs, &syncmsg.ProtoSyncDataMessageRequest{
			RecordId:        proto.String(recordID),
			RecordHash:      proto.String(recordHash),
			SentSyncState:   sentSyncStateEnum.Enum(),
			RecordBytesSize: proto.Uint32(recordBytesSize),
			RecordData:      recordBytes,
		})
	}

	return nil
}

func (fetcher postgresSQLMessageFetcher) reserveFetchItems(entitySingularName string, changeType syncapi.ProcessSyncChangeEnum, bindID string) error {
	isDelete := *changeType.Enum() == syncapi.ProcessSyncChangeEnumDelete
	_, err := fetcher.db.Exec(sqlReserveForFetching, bindID, entitySingularName, fetcher.SessionID, fetcher.NodeID, isDelete, fetcher.MaxMsgs)
	if err != nil {
		syncutil.Error(err)
		return err
	}

	return nil
}

const (
	sqlReserveForFetching = `
update sync_peer_state set queueBindSendId=$1 where entitySingularName=$2 AND nodeId=$4 AND recordId in (select recordId from sync_peer_state where entitySingularName=$2 AND sessionBindId=$3 AND queueBindSendId is null AND nodeId=$4 AND isDelete=$5 AND changedByClient=true limit $6);
`

	//The sort order doesn't matter logically, but for testing purposes we can get interminant results, so we're
	//being strict on the sort order so that results are determinant. We might want to remove this from the
	//production code. Though, the performance hit is believed to be negligible.
	sqlFetchReserved = `
select a.recordId, b.recordHash, a.peerLastKnownHash, a.sentSyncState, b.recordBytesSize, b.recordData from sync_peer_state a INNER JOIN
                         sync_state b ON a.EntitySingularName = b.EntitySingularName AND
                         a.RecordId = b.RecordId where (a.recordId in (select RecordId from sync_peer_state where queueBindSendId=$1)) AND a.nodeId=$2 order by a.entitySingularName, a.recordId;
`
)
