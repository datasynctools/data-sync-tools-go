package syncdaopq

import (
	"bytes"
	"database/sql"
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncdao"
	"data-sync-tools-go/syncmsg"
	"data-sync-tools-go/syncutil"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"sort"
	"time"

	"github.com/golang/protobuf/proto"
)

//NewSyncMessagesFetcher creates an instance of the struct SyncMessagesFetcherType
func newMessageProcessor(sessionID string, nodeID string, entitiesByPluralName map[string]syncapi.EntityNameItem, db *sql.DB) (syncapi.MessageProcessing, error) {

	processorType := syncapi.MessageProcessingData{
		SessionID:            sessionID,
		NodeID:               nodeID,
		EntitiesByPluralName: entitiesByPluralName,
	}
	processor := postgresSQLMessageProcessor{
		MessageProcessingData: processorType,
		db: db,
	}
	//syncutil.Info(fetcher)
	return processor, nil
}

//postgresSQLMessageProcessor logically implements MessageProcessor for the PostgressSQL database.
type postgresSQLMessageProcessor struct {
	syncapi.MessageProcessingData
	db *sql.DB
}

//Process transforms request data into persisted data synced to the local model while resolving or reporting
//conflicts according to the sync policy for the session. Logically it works as follows (without error handling):
//
// 1. request(a) -> ready initial changes -> persist all changes as a batch assuming there are no conflicts and
// with database optimizations
//
// 2. check if initial changes to see what was persisted. The one's that were processed are marked as 'AckFastBatch'.
//
// 3. if all were processed, go to step 5.
//
// 4. if all were not processed, process individual messages to the data store trying to resolve the conflicts
// with configured algorithm(s)
//
// 5. Aggregate response(b)
//
// (a) see "IN" for the below state documentation.
//
// (b) see "OUT" for the below state documentation.
//
// State documentation
//
//  IN/OUT LOGICAL ENUM                                          VALUE
//  ------ ------------                                          -----
//  IN     PersistedNeverSentToPeer                              1
//  IN     PersistedFirstTimeSentToPeer                          2
//  IN     PersistedStandardSentToPeer                           3
//  IN     PersistedFastDeleted                                  4
//  OUT    AckFastBatch                                          21
//  OUT    AckRecordLevelConflictResolvedSeparateFieldsChanged   22
//  OUT    AckFieldLevelConflictwithNoAutoResolverAvailable      23
//  OUT    AckFieldLevelConflictResolvedWithAutoResolver         24
//  OUT    AckFieldLevelConflictWithNoAutoResolverResolution     25
//  OUT    AckDeleteAndUpdateConflictWithNoAutoResolverAvailable 26
//  OUT    AckDeleteAndUpdateConflictWithNoAutoResolution        27
//  OUT    AckDeleteAndUpdateConflictWithAutoResolution          28
func (processor postgresSQLMessageProcessor) Process(request *syncmsg.ProtoSyncEntityMessageRequest) *syncmsg.ProtoSyncEntityMessageResponse {
	answer := &syncmsg.ProtoSyncEntityMessageResponse{
		TransactionBindId: request.TransactionBindId,
		Items:             []*syncmsg.ProtoSyncDataMessagesResponse{}, //make([]*syncmsg.ProtoSyncDataMessagesResponse, 0, 1),
	}
	sqlProcessor := &compoundSQLProcessor{
		builders: []sqlBuilder{&changeInitialSQLBuilder{}, &readInitialSQLBuilder{}},
	}
	unprocessedMsgs, err := processor.initialProcessLoop(sqlProcessor, *request, answer, processor.NodeID, *request.TransactionBindId)
	if err != nil {
		//initialProcessLoop fills in the erros on the answer object, so, just return it as is
		return answer
	}
	if len(unprocessedMsgs) == 0 {
		answer.Result = syncmsg.SyncEntityMessageResponseResult_OK.Enum()
		answer.ResultMsg = proto.String("All records are fast batch")
		return answer
	}
	answer.Result = syncmsg.SyncEntityMessageResponseResult_OK.Enum()
	answer.ResultMsg = proto.String("Implement Me: I'm not fast batch")
	syncutil.Debug("Not Fast Batch Items: ", unprocessedMsgs)
	return answer
}

func (processor postgresSQLMessageProcessor) initialProcessLoop(sqlProcessor *compoundSQLProcessor, requestData syncmsg.ProtoSyncEntityMessageRequest, response *syncmsg.ProtoSyncEntityMessageResponse, nodeIDToProcess string, transactionBindID string) (map[string][]readInitialTransactionBindResult, error) {
	unprocessedMsgs := make(map[string][]readInitialTransactionBindResult)
	sqlProcessor.processStart(requestData, nodeIDToProcess, transactionBindID)
	err := processor.initialProcessChangesLoop(sqlProcessor, requestData)
	if err != nil {
		// TODO(doug4j@gmail.com): Add check for 'ERROR:  could not serialize access due to concurrent update' as
		// described in http://www.postgresql.org/docs/current/static/transaction-iso.html
		syncutil.Error(err)
		response.Result = syncmsg.SyncEntityMessageResponseResult_Error.Enum()
		response.ResultMsg = proto.String(err.Error())
		return unprocessedMsgs, err
	}
	sqlProcessor.processEnd(requestData)
	err = processor.initialProcessApplyChanges(sqlProcessor.builders[0].result())
	if err != nil {
		syncutil.Error(err)
		response.Result = syncmsg.SyncEntityMessageResponseResult_Error.Enum()
		response.ResultMsg = proto.String(err.Error())
		return unprocessedMsgs, err
	}
	unprocessedMsgs, err = processor.initialProcessReadChanges(sqlProcessor.builders[1].result(), transactionBindID, response)
	if err != nil {
		syncutil.Error(err)
		response.Result = syncmsg.SyncEntityMessageResponseResult_Error.Enum()
		response.ResultMsg = proto.String(err.Error())
		return unprocessedMsgs, err
	}
	return unprocessedMsgs, nil
}

func (processor postgresSQLMessageProcessor) initialProcessApplyChanges(processInitialChangeSQL string) error {
	//syncutil.Debug("processInitialChangeSQL", processInitialChangeSQL)
	_, err := processor.db.Exec(processInitialChangeSQL)
	if err != nil {
		syncutil.Error(err, ". Error processing initial database changes")
		return err
	}
	return nil
}

type readInitialTransactionBindResult struct {
	entitySingularName, recordID, recordHash, recordData, transactionBindReceiveID string
	isDelete                                                                       bool
}

func (processor postgresSQLMessageProcessor) initialProcessReadChanges(readInitialChangeSQL string, transactionBindID string, response *syncmsg.ProtoSyncEntityMessageResponse) (map[string][]readInitialTransactionBindResult, error) {
	processedFastBatchMsgs := make(map[string]map[string]string)
	unprocessedMsgs := make(map[string][]readInitialTransactionBindResult)
	//syncutil.Debug("readInitialChangeSql", readInitialChangeSQL)
	rows, err := processor.db.Query(readInitialChangeSQL)
	if err != nil {
		syncutil.Error(err, ". Error reading initial database changes")
		return unprocessedMsgs, err
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
			entitySingularName, entityPluralName, recordID, recordHash, recordData string
			//http://marcesher.com/2014/10/13/go-working-effectively-with-database-nulls/
			transactionBindReceiveID sql.NullString
			isDelete                 bool
		)
		if err := rows.Scan(&entitySingularName, &entityPluralName, &recordID, &recordHash, &recordData, &isDelete, &transactionBindReceiveID); err != nil {
			syncutil.Error(err, ". Error reading fields in initial database changes")
			return unprocessedMsgs, err
		}
		// TODO(doug4j@gmail.com): Add post processing of fields into proper objects AND assess which one's are processed and which ones are record conflicted
		if transactionBindReceiveID.Valid && transactionBindID == transactionBindReceiveID.String {
			if _, ok := processedFastBatchMsgs[entityPluralName]; !ok {
				processedFastBatchMsgs[entityPluralName] = make(map[string]string)
			}
			processedFastBatchMsgs[entityPluralName][recordID] = recordHash
		} else {
			if _, ok := unprocessedMsgs[entitySingularName]; !ok {
				unprocessedMsgs[entitySingularName] = []readInitialTransactionBindResult{}
			}
			unprocessedMsgs[entitySingularName] = append(unprocessedMsgs[entitySingularName], readInitialTransactionBindResult{
				entitySingularName:       entitySingularName,
				recordID:                 recordID,
				recordHash:               recordHash,
				recordData:               recordData,
				transactionBindReceiveID: transactionBindReceiveID.String,
				isDelete:                 isDelete,
			})
		}
	}
	for pluralEntityName, fastBatchRecordArray := range processedFastBatchMsgs {
		syncutil.Debug("pluralEntityName='", pluralEntityName, "'")
		for fastBatchRecordID, fastBatchRecordHash := range fastBatchRecordArray {
			response.Items = append(response.Items, &syncmsg.ProtoSyncDataMessagesResponse{
				//EntityPluralName: proto.String(singularAndPluralEntityName.PluralName),
				EntityPluralName: proto.String(pluralEntityName),
				Msgs: []*syncmsg.ProtoSyncDataMessageResponse{
					&syncmsg.ProtoSyncDataMessageResponse{
						RecordId:     proto.String(fastBatchRecordID),
						ResponseHash: proto.String(fastBatchRecordHash),
						SyncState:    syncmsg.AckSyncStateEnum_AckFastBatch.Enum(),
					},
				},
			})
		}
	}
	return unprocessedMsgs, nil
}

func (processor postgresSQLMessageProcessor) initialProcessChangesLoop(sqlProcessor *compoundSQLProcessor, requestData syncmsg.ProtoSyncEntityMessageRequest) error {
	recordIndexLen := len(requestData.Items)
	for recordIndex, item := range requestData.Items {

		err := sqlProcessor.startRecords(item, recordIndex, recordIndexLen, processor)
		if err != nil {
			return err
		}
		msgLen := len(item.Msgs) - 1
		for msgIndex, msg := range item.Msgs {

			err = sqlProcessor.startRecordItem(msg, msgIndex, msgLen)
			if err != nil {
				return err
			}
		}
		//log.Printf("TransactionBindId: 	%v", requestData.TransactionBindId)
	}
	return nil
}

func (processor postgresSQLMessageProcessor) findNodeEntityFields(nodeIDToProcess string, syncEntitySingularName string) (map[string]syncdao.SyncFieldDefinition, error) {

	answer := map[string]syncdao.SyncFieldDefinition{}
	//answer := map[string]syncdao.SyncFieldTypeEnum{}
	sqlStr := `
SELECT        sync_data_field.FieldName, sync_data_field.DataTypeName, sync_data_field.IsPrimaryKey
FROM            sync_node INNER JOIN
                         sync_data_version ON sync_node.DataVersionName = sync_data_version.DataVersionName INNER JOIN
                         sync_data_field ON sync_data_version.DataVersionName = sync_data_field.DataVersionName
WHERE        (sync_node.NodeId = $1 and EntitySingularName = $2);
`
	rows, err := processor.db.Query(sqlStr, nodeIDToProcess, syncEntitySingularName)
	if err != nil {
		syncutil.Error(err)
		return answer, errors.New("Error finding Node Entity fields: " + err.Error())
	}
	closeRowQuietly := func() {
		err := rows.Close()
		if err != nil {
			syncutil.Error("Quietly handling of row close error. Error: " + err.Error())
		}
	}
	defer closeRowQuietly()
	err = processor.mapFoundNodeEntitiesFinal(rows, &answer)
	if err != nil {
		syncutil.Error(err)
		return answer, err
	}
	return answer, nil
}

func (processor postgresSQLMessageProcessor) mapFoundNodeEntitiesFinal(rows *sql.Rows, answer *map[string]syncdao.SyncFieldDefinition) error {
	localAnswer := map[string]syncdao.SyncFieldDefinition{}
	for rows.Next() {
		var (
			fieldName    string
			fieldType    string
			isPrimaryKey bool
		)
		if err := rows.Scan(&fieldName, &fieldType, &isPrimaryKey); err != nil {
			syncutil.Error(err)
			return errors.New("Error finding field information: " + err.Error())
		}
		switch fieldType {
		case "String":
			item := syncdao.SyncFieldDefinition{
				FieldName:    fieldName,
				FieldType:    syncdao.SyncFieldTypeEnumString,
				IsPrimaryKey: isPrimaryKey,
			}
			localAnswer[fieldName] = item
		case "Int":
			item := syncdao.SyncFieldDefinition{
				FieldName:    fieldName,
				FieldType:    syncdao.SyncFieldTypeEnumInt,
				IsPrimaryKey: isPrimaryKey,
			}
			localAnswer[fieldName] = item
		case "Float":
			item := syncdao.SyncFieldDefinition{
				FieldName:    fieldName,
				FieldType:    syncdao.SyncFieldTypeEnumFloat,
				IsPrimaryKey: isPrimaryKey,
			}
			localAnswer[fieldName] = item
		case "Bool":
			item := syncdao.SyncFieldDefinition{
				FieldName:    fieldName,
				FieldType:    syncdao.SyncFieldTypeEnumBool,
				IsPrimaryKey: isPrimaryKey,
			}
			localAnswer[fieldName] = item
		case "Date":
			item := syncdao.SyncFieldDefinition{
				FieldName:    fieldName,
				FieldType:    syncdao.SyncFieldTypeEnumDate,
				IsPrimaryKey: isPrimaryKey,
			}
			localAnswer[fieldName] = item
		case "Binary":
			item := syncdao.SyncFieldDefinition{
				FieldName:    fieldName,
				FieldType:    syncdao.SyncFieldTypeEnumBinary,
				IsPrimaryKey: isPrimaryKey,
			}
			localAnswer[fieldName] = item
		default:
			item := syncdao.SyncFieldDefinition{
				FieldName:    fieldName,
				FieldType:    syncdao.SyncFieldTypeEnumUndefined,
				IsPrimaryKey: isPrimaryKey,
			}
			localAnswer[fieldName] = item
		}
		*answer = localAnswer
	}
	return nil
}

func (processor postgresSQLMessageProcessor) findSingularEntityName(syncEntityPluralName string, dataVersion string) (string, error) {
	var answer string
	sqlStr := `
SELECT        EntitySingularName
FROM            sync_data_entity
WHERE        (DataVersionName = $1 and EntityPluralName = $2);
`
	err := processor.db.QueryRow(sqlStr, dataVersion, syncEntityPluralName).Scan(&answer)
	switch {
	case err == sql.ErrNoRows:
		msg := "Cannot find singular form of entity from plural name '" + syncEntityPluralName + "' and data version '" + dataVersion + "'"
		syncutil.Error(msg)
		return answer, errors.New(msg)
	case err != nil:
		msg := "Cannot find singular form of entity from plural name '" + syncEntityPluralName + "' and data version '" + dataVersion + "'"
		syncutil.Error(err, ".", msg)
		return answer, err
	default:
		return answer, nil
	}
}

func calculateSQLValue(field syncmsg.ProtoField, fieldDefinitions map[string]syncdao.SyncFieldDefinition, syncEntityName string) (string, error) {
	fieldName := *field.FieldName
	fieldDefinition := fieldDefinitions[fieldName]
	if fieldDefinition.FieldType == syncdao.SyncFieldTypeEnumUndefined {
		return "", errors.New("field " + fieldName + " does  not match a known field definition for entity " + syncEntityName)
	}
	valueType := *field.EncodedFieldType
	switch fieldDefinition.FieldType {
	case syncdao.SyncFieldTypeEnumString:
		return fmt.Sprintf("'%v'", syncdao.CalculateValue(valueType, field.FieldValue)), nil
	case syncdao.SyncFieldTypeEnumDate:
		creator := syncmsg.NewCreator()
		rawTimeString := fmt.Sprintf("%v", syncdao.CalculateValue(valueType, field.FieldValue))
		theTime := creator.FormatTimeFromString(rawTimeString)
		if len(creator.Errors) != 0 {
			err := syncmsg.NewCreatorError(creator.Errors)
			syncutil.Error(err.Error())
			return "", err
		}
		timeString := theTime.Format(SQLDateFormat)
		return "'" + timeString + "'", nil
	default:
		return fmt.Sprintf("%v", syncdao.CalculateValue(valueType, field.FieldValue)), nil
	}
}

type sqlBuilder interface {
	processStart(msg changeEntityMessage, nodeIDToProcess string, transactionBindID string)
	processEnd()
	startRecords(item changeDataMessageList)
	startRecordItem(item changeDataMessage, requestData changeEntityMessage, changeDataMessages changeDataMessageList) error
	result() string
}

type compoundSQLProcessor struct {
	builders           []sqlBuilder
	requestData        changeEntityMessage
	changeDataMessages changeDataMessageList
	changeDataMessage  changeDataMessage
}

func (processor *compoundSQLProcessor) processStart(requestData syncmsg.ProtoSyncEntityMessageRequest, nodeIDToProcess string, transactionBindID string) {
	now := time.Now()
	format := "2006-01-02 15:04:05.000"
	processor.requestData = changeEntityMessage{
		isDelete:          *requestData.IsDelete,
		NowDateString:     now.Format(format),
		NodeIDToProcess:   nodeIDToProcess,
		TransactionBindID: transactionBindID,
	}
	for _, builder := range processor.builders {
		builder.processStart(processor.requestData, nodeIDToProcess, transactionBindID)
	}
}

func (processor *compoundSQLProcessor) startRecords(item *syncmsg.ProtoSyncDataMessagesRequest, recordIndex int, recordIndexLen int, msgProcessor postgresSQLMessageProcessor) error {
	// TODO(doug4j@gmail.com): Remove hard coding of data version name
	entityPluralName := *item.EntityPluralName
	entitySingularName, err := msgProcessor.findSingularEntityName(entityPluralName, "Demo Model 1")
	if err != nil {
		syncutil.Error(err)
		return err
	}

	//Get entity field definitions
	var fieldDefinitions map[string]syncdao.SyncFieldDefinition
	fieldDefinitions, err = msgProcessor.findNodeEntityFields(processor.requestData.NodeIDToProcess, entitySingularName)
	if err != nil {
		syncutil.Error(err)
		return err
	}

	entityKeyMap := make(map[string]bool)
	entitySortedKeys := []string{}
	for _, k := range fieldDefinitions {
		if k.IsPrimaryKey {
			name := k.FieldName
			entityKeyMap[name] = true
			entitySortedKeys = append(entitySortedKeys, name)
		}
	}
	sort.Strings(entitySortedKeys)

	processor.changeDataMessages = changeDataMessageList{
		SyncEntitySingularName: entitySingularName,
		SyncEntityPluralName:   entityPluralName,
		fieldDefinitions:       fieldDefinitions,
		entityKeyMap:           entityKeyMap,
		entitySortedKeys:       entitySortedKeys,
		recordIndex:            recordIndex,
		recordIndexLen:         recordIndexLen,
	}

	for _, builder := range processor.builders {
		builder.startRecords(processor.changeDataMessages)
	}
	return nil
}

func (processor *compoundSQLProcessor) startRecordItem(item *syncmsg.ProtoSyncDataMessageRequest, itemIndex int, indexLength int) error {
	//printSyncDebugItem(msgIndex int, msg *syncmsg.ProtoSyncDataMessageRequest, record syncmsg.ProtoRecord)
	var (
		sentSyncStateString = item.SentSyncState.Enum().String()
		lastKnownPeerHash   string
	)
	if item.LastKnownPeerHash != nil {
		lastKnownPeerHash = *item.LastKnownPeerHash
	}
	processor.changeDataMessage = changeDataMessage{
		recordData:    item.RecordData,
		sentSyncState: *item.SentSyncState,
		//recordIndexStr         string                    = fmt.Sprintf("%v", recordIndex+1)
		MsgIndexStr:            fmt.Sprintf("%v", itemIndex+1),
		msgIndex:               itemIndex,
		msgIndexLen:            indexLength,
		RecordID:               *item.RecordId,
		RecordDataHex:          hex.EncodeToString(item.RecordData),
		RecordHash:             *item.RecordHash,
		LastKnownPeerHash:      lastKnownPeerHash,
		RecordBytesSizeStr:     fmt.Sprintf("%v", *item.RecordBytesSize),
		SentSyncStateNumString: fmt.Sprintf("%v", syncmsg.SentSyncStateEnum_value[sentSyncStateString]),
	}
	for _, builder := range processor.builders {
		err := builder.startRecordItem(processor.changeDataMessage, processor.requestData, processor.changeDataMessages)
		if err != nil {
			return err
		}
	}
	return nil
}

func (processor *compoundSQLProcessor) processEnd(requestData syncmsg.ProtoSyncEntityMessageRequest) {
	for _, builder := range processor.builders {
		builder.processEnd()
		//syncutil.Debug(reflect.TypeOf(builder).String() + ".result()=" + builder.result())
	}
}

type changeEntityMessage struct {
	isDelete          bool
	NowDateString     string
	NodeIDToProcess   string
	TransactionBindID string
	//entityFieldDefinitions map[string]map[string]syncdao.SyncFieldDefinition
}

type changeDataMessageList struct {
	SyncEntitySingularName string
	SyncEntityPluralName   string
	entityKeyMap           map[string]bool
	entitySortedKeys       []string
	fieldDefinitions       map[string]syncdao.SyncFieldDefinition
	recordIndex            int
	recordIndexLen         int
}

type changeDataMessage struct {
	recordData             []byte
	sentSyncState          syncmsg.SentSyncStateEnum
	RecordIndexStr         string
	msgIndexLen            int
	msgIndex               int
	MsgIndexStr            string
	LastKnownPeerHash      string
	RecordID               string
	RecordDataHex          string
	RecordHash             string
	RecordBytesSizeStr     string
	SentSyncStateNumString string
}

type changeInitialSQLBuilder struct {
	sql string
}

func (builder *changeInitialSQLBuilder) processStart(msg changeEntityMessage, nodeIDToProcess string, transactionBindID string) {
	builder.sql = "begin;"
}

func (builder *changeInitialSQLBuilder) startRecords(item changeDataMessageList) {
}

func (builder *changeInitialSQLBuilder) startRecordItem(item changeDataMessage, requestData changeEntityMessage, changeDataMessages changeDataMessageList) error {
	switch item.sentSyncState {
	case syncmsg.SentSyncStateEnum_PersistedFirstTimeSentToPeer:
		//syncutil.Debug("Processing first time sent to peer for recordId", item.recordId)

		err := builder.handleSyncFirstTimeSentToPeer(item, requestData, changeDataMessages)
		if err != nil {
			return err
		}
	case syncmsg.SentSyncStateEnum_PersistedStandardSentToPeer:
		//syncutil.Debug("Processing standard to peer for recordId", item.recordId)
		if requestData.isDelete {
			syncutil.NotImplementedMsg("Delete me")
		} else {
			err := builder.handleSyncStandardSentToPeerUpdate(item, requestData, changeDataMessages)
			if err != nil {
				return err
			}
		}
	default:
		errMsg := "Unsupported sentSyncState " + item.sentSyncState.String()
		syncutil.Debug(errMsg)
		return errors.New(errMsg)
	}
	return nil
}

func (builder *changeInitialSQLBuilder) createSyncFirstTimeSentSQL(item changeDataMessage, requestData changeEntityMessage, changeDataMessages changeDataMessageList) (string, error) {
	var templateStr = `
	-- msg {{.Item.MsgIndexStr}} | rec {{.Item.RecordIndexStr}} SyncState
	insert into sync_state (EntitySingularName, RecordId, DataVersionName, RecordHash, RecordData, RecordBytesSize, IsDelete)
	values('{{.ChangeDataMessages.SyncEntitySingularName}}', '{{.Item.RecordID}}', 'Demo Model 1', '{{.Item.RecordHash}}',
	' {{.Item.RecordDataHex}}', {{.Item.RecordBytesSizeStr}}, False);
	-- msg {{.Item.MsgIndexStr}} | rec {{.Item.RecordIndexStr}} SyncPeerState
	insert into sync_peer_state (NodeId, EntitySingularName, RecordId, TransactionBindReceiveId, SentLastKnownHash,
		PeerLastKnownHash, SentSyncState, RecordBytesSize, LastUpdated, RecordCreated)
		values ('{{.RequestData.NodeIDToProcess}}', '{{.ChangeDataMessages.SyncEntitySingularName}}', '{{.Item.RecordID}}',
			'{{.RequestData.TransactionBindID}}', '{{.Item.RecordHash}}', '{{.Item.RecordHash}}', {{.Item.SentSyncStateNumString}},
			{{.Item.RecordBytesSizeStr}}, '{{.RequestData.NowDateString}}', '{{.RequestData.NowDateString}}');`
	sqlTemplate := template.Must(template.New("syncFirstTimeSentSQL").Parse(templateStr))
	var substitutedSQLBytes bytes.Buffer
	var data = struct {
		Item               changeDataMessage
		RequestData        changeEntityMessage
		ChangeDataMessages changeDataMessageList
	}{item, requestData, changeDataMessages}
	err := sqlTemplate.Execute(&substitutedSQLBytes, data)
	if err != nil {
		return "", err

	}
	substitutedSQL := substitutedSQLBytes.String()
	//syncutil.Debug(substitutedSQL)
	return substitutedSQL, nil
}

func (builder *changeInitialSQLBuilder) handleSyncFirstTimeSentToPeer(item changeDataMessage, requestData changeEntityMessage, changeDataMessages changeDataMessageList) error {
	sql, err := builder.createSyncFirstTimeSentSQL(item, requestData, changeDataMessages)
	if err != nil {
		syncutil.Error(err.Error())
		return err
	}
	builder.sql = builder.sql + sql

	//Process Custom Table Start
	record := &syncmsg.ProtoRecord{}
	err = proto.Unmarshal(item.recordData, record)
	if err != nil {
		return err
	}
	//Process Custom Table Name
	for fieldIndex, field := range record.Fields {
		if fieldIndex == 0 {
			builder.sql = builder.sql + `
-- msg ` + item.MsgIndexStr + ` | rec ` + item.RecordIndexStr + ` CustomTable
insert into ` + changeDataMessages.SyncEntityPluralName + ` (` + *field.FieldName
		} else {
			builder.sql = builder.sql + `, ` + *field.FieldName
		}
	}
	//Process Custom Table Value
	for fieldIndex, field := range record.Fields {
		rawValue, err := calculateSQLValue(*field, changeDataMessages.fieldDefinitions, changeDataMessages.SyncEntitySingularName)
		if err != nil {
			return err
		}
		valueAsString := fmt.Sprintf("%v", rawValue)
		if fieldIndex == 0 {
			builder.sql = builder.sql + `)
	values (` + valueAsString
		} else {
			builder.sql = builder.sql + `, ` + valueAsString
		}
	}
	builder.sql = builder.sql + `);`
	return nil
}

func (builder *changeInitialSQLBuilder) handleSyncStandardSentToPeerUpdate(item changeDataMessage, requestData changeEntityMessage, changeDataMessages changeDataMessageList) error {

	var newSQLStr = ""

	//Process Custom Table Start
	record := &syncmsg.ProtoRecord{}
	err := proto.Unmarshal(item.recordData, record)
	if err != nil {
		syncutil.Debug(err)
		return err
	}

	newSQLStr, err = builder.processCustomTableUpdateBase(newSQLStr, item, changeDataMessages, record)
	if err != nil {
		syncutil.Debug(err)
		return err
	}

	newSQLStr, err = builder.processCustomTableUpdateWhere(newSQLStr, item, changeDataMessages, requestData, record)
	if err != nil {
		syncutil.Debug(err)
		return err
	}
	builder.sql = builder.sql + newSQLStr
	return nil
}

func (builder *changeInitialSQLBuilder) processCustomTableUpdateBase(newSQLStr string, item changeDataMessage, changeDataMessages changeDataMessageList, record *syncmsg.ProtoRecord) (string, error) {
	//Process Custom Table Name
	fieldLen := len(record.Fields) - 1
	for fieldIndex, field := range record.Fields {
		if fieldIndex == 0 {
			newSQLStr = newSQLStr + `
-- msg ` + item.MsgIndexStr + ` | rec ` + item.RecordIndexStr + ` CustomTable
update ` + changeDataMessages.SyncEntityPluralName + ` set ` + *field.FieldName + `=`

		} else {
			newSQLStr = newSQLStr + *field.FieldName + `=`
		}
		rawValue, err := calculateSQLValue(*field, changeDataMessages.fieldDefinitions, changeDataMessages.SyncEntitySingularName)
		if err != nil {
			return newSQLStr, err
		}
		valueAsString := fmt.Sprintf("%v", rawValue)
		if fieldIndex == fieldLen { //Is last field
			newSQLStr = newSQLStr + valueAsString
		} else {
			newSQLStr = newSQLStr + valueAsString + `, `
		}
	}
	return newSQLStr, nil
}

func (builder *changeInitialSQLBuilder) processCustomTableUpdateWhere(newSQLStr string, item changeDataMessage, changeDataMessages changeDataMessageList, requestData changeEntityMessage, record *syncmsg.ProtoRecord) (string, error) {
	//The extra where clause '((select count(... where(make it so we don't
	//update the custom table unless the item previous hash matches

	nameValues, err := builder.createKeysAndSQLValues(changeDataMessages, record)
	if err != nil {
		syncutil.Error(err)
		return "", err
	}

	for index, key := range changeDataMessages.entitySortedKeys {
		if index == 0 {
			newSQLStr = newSQLStr + ` where (` + key + `=` + nameValues[key]
		} else {
			newSQLStr = newSQLStr + ` AND ` + key + `=` + nameValues[key]
		}
	}
	sqlLockViaWhereOnRecordItem := `((select count(RecordId) as syncRecordCount from sync_state where (EntitySingularName='` + changeDataMessages.SyncEntitySingularName + `' AND RecordId='` + item.RecordID + `' AND RecordHash='` + item.LastKnownPeerHash + `'))) = 1)`

	newSQLStr = newSQLStr + ` AND ((select count(RecordId) as syncRecordCount from sync_state where (EntitySingularName='` + changeDataMessages.SyncEntitySingularName + `' AND RecordId='` + item.RecordID + `' and RecordHash='` + item.LastKnownPeerHash + `'))) = 1);`
	// TODO(doug4j@gmail.com): Is there a reason for updating PeerLastKnownHash? Is it possible there is a PeerLastKnownSendHash and PeerLastKnownReceiveHash instead of just PeerLastKnownHash? And with it, do we update a PeerLastKnownReceiveHash here?
	newSQLStr = newSQLStr + `
-- msg ` + item.MsgIndexStr + ` | rec ` + item.RecordIndexStr + ` SyncPeerState
update sync_peer_state set TransactionBindReceiveId='` + requestData.TransactionBindID + `', PeerLastKnownHash='` + item.RecordHash + `', LastUpdated='` + requestData.NowDateString + `' where (NodeId='` + requestData.NodeIDToProcess + `' AND EntitySingularName='` + changeDataMessages.SyncEntitySingularName + `' AND RecordId='` + item.RecordID + `' AND ` + sqlLockViaWhereOnRecordItem + `;` + `
-- msg ` + item.MsgIndexStr + ` | rec ` + item.RecordIndexStr + ` SyncState
update sync_state set RecordHash='` + item.RecordHash + `', RecordData='` + item.RecordDataHex + `', RecordBytesSize=` + item.RecordBytesSizeStr + `, IsDelete=false where (EntitySingularName='` + changeDataMessages.SyncEntitySingularName + `' AND RecordId='` + item.RecordID + `' AND RecordHash='` + item.LastKnownPeerHash +
		`');`

	return newSQLStr, nil
}

func (builder *changeInitialSQLBuilder) createKeysAndSQLValues(changeDataMessages changeDataMessageList, record *syncmsg.ProtoRecord) (map[string]string, error) {
	answer := make(map[string]string)
	var err error
	for _, field := range record.Fields {
		fieldName := *field.FieldName
		if _, found := changeDataMessages.entityKeyMap[fieldName]; found {
			answer[fieldName], err = calculateSQLValue(*field, changeDataMessages.fieldDefinitions, changeDataMessages.SyncEntitySingularName)
			if err != nil {
				syncutil.Error(err)
				return answer, err
			}
		}
	}
	return answer, nil
}

func (builder *changeInitialSQLBuilder) processEnd() {
	builder.sql = builder.sql + `
commit;`
	//syncutil.Debug(self.sql)
}

func (builder *changeInitialSQLBuilder) result() string {
	answer := builder.sql
	//debug("answer=" + answer)
	return answer
}

type readInitialSQLBuilder struct {
	sql string
}

func (builder *readInitialSQLBuilder) processStart(msg changeEntityMessage, nodeIDToProcess string, transactionBindID string) {
	builder.sql = ""
}

func (builder *readInitialSQLBuilder) startRecords(item changeDataMessageList) {

	if item.recordIndex == 0 {
		builder.sql = builder.sql + `
select sync_state.EntitySingularName, sync_data_entity.EntityPluralName, sync_state.RecordId, sync_state.RecordHash, sync_state.RecordData, sync_state.IsDelete, sync_peer_state.TransactionBindReceiveId from sync_state INNER JOIN sync_peer_state ON sync_state.EntitySingularName = sync_peer_state.EntitySingularName AND sync_state.RecordId = sync_peer_state.RecordId INNER JOIN sync_data_entity ON sync_state.EntitySingularName = sync_data_entity.EntitySingularName where (
	(sync_state.EntitySingularName='` + item.SyncEntitySingularName + `' AND`
	} else if item.recordIndex == item.recordIndexLen {
		builder.sql = builder.sql + `
	(sync_state.EntitySingluarName='` + item.SyncEntitySingularName + `'`
	} else {
		builder.sql = builder.sql + `
	(sync_state.EntitySingularName='` + item.SyncEntitySingularName + `' AND`
	}
	//debug("self.sql=" + self.sql)
}

func (builder *readInitialSQLBuilder) processEnd() {
	builder.sql = builder.sql + `
);`
}

func (builder *readInitialSQLBuilder) startRecordItem(item changeDataMessage, requestData changeEntityMessage, changeDataMessages changeDataMessageList) error {
	if (item.msgIndex == 0) && (item.msgIndex == item.msgIndexLen) {
		builder.sql = builder.sql + ` sync_state.RecordId='` + item.RecordID + `')`
	} else if item.msgIndex == 0 {
		builder.sql = builder.sql + ` sync_state.RecordId='` + item.RecordID + `' OR`
	} else if item.msgIndex == item.msgIndexLen {
		builder.sql = builder.sql + ` sync_state.RecordId='` + item.RecordID + `')`
	} else {
		builder.sql = builder.sql + ` sync_state.RecordId='` + item.RecordID + `' OR`
	}
	//syncutil.Debug("self.sql", self.sql)
	return nil
}

func (builder *readInitialSQLBuilder) result() string {
	answer := builder.sql
	//debug("answer=" + answer)
	return answer
}
