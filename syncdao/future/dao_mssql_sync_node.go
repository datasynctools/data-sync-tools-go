package main

import (
	"database/sql"
	"log"
)

type SyncNodeMsSqlDao struct {
	db *sql.DB
}

func (self SyncNodeMsSqlDao) Add(item SyncNode) error {
	//TODO: Add SQL injection checks (see http://go-database-sql.org/retrieving.html and http://stackoverflow.com/questions/26345318/how-can-i-prevent-sql-injection-attacks-in-go-while-using-database-sql

	_, err := self.db.Exec("INSERT INTO sync_node(nodeId,nodeName) VALUES($1,$2);", item.MsgId, item.Node)
	if err != nil {
		log.Println("Error inserting to database, inputData=", item)
		return err
	}
	return nil
}

func (self SyncNodeMsSqlDao) GetOne(itemId string) (SyncNode, error) {
	//TODO: Implement me
	return SyncNode{MsgId: "", Node: ""}, nil
}

func (self SyncNodeMsSqlDao) QueueChanges(sessionId string, nodeIdToQueue string) (int, error) {
	//TODO: Add SQL injection checks (see http://go-database-sql.org/retrieving.html and http://stackoverflow.com/questions/26345318/how-can-i-prevent-sql-injection-attacks-in-go-while-using-database-sql
	//Injection Checks are particularly important due to the need for a dynamically set fixed field in the SQL for the
	// Node Id.

	var sqlStr string
	//For queuing previously unqueued records
	sqlStr = `
insert into sync_peer_state (NodeId, EntityId, RecordId, SentLastKnownHash, ChangedByClient) 
select '` + nodeIdToQueue
	sqlStr = sqlStr + `' NodeId, EntityId, RecordId, RecordHash, '1' ChangedByClient from sync_state where 
						RecordId not in (
SELECT        sync_peer_state.RecordId
FROM            sync_peer_state INNER JOIN
                         sync_state ON sync_peer_state.EntityId = sync_state.EntityId AND 
                         sync_peer_state.RecordId = sync_state.RecordId
WHERE        (sync_peer_state.NodeId = $1)
);`
	_, err := self.db.Exec(sqlStr, nodeIdToQueue)
	if err != nil {
		msg := "Error inserting, nodeIdToQueue='%s', Error:%s"
		log.Printf(msg, nodeIdToQueue, err.Error())
		return 0, err
	}
	sqlStr = `
--For Queuing, how to mark changes to previously queued records
update sync_peer_state set ChangedByClient = '1' where 
	EntityId in (
		select EntityId from sync_peer_state where 
			ChangedByClient = '0' 
		and
				NodeId = $1
		and
			RecordId in (
				SELECT RecordId from sync_peer_state 	where SentLastKnownHash not in	(
				SELECT  SentLastKnownHash       
				FROM            sync_peer_state INNER JOIN
								 sync_state ON sync_peer_state.EntityId = sync_state.EntityId AND 
								 sync_peer_state.RecordId = sync_state.RecordId AND 
								 sync_peer_state.SentLastKnownHash = sync_state.RecordHash)
			) 
		and
			EntityId in (
				SELECT EntityId from sync_peer_state 	where SentLastKnownHash not in	(
				SELECT  SentLastKnownHash       
				FROM            sync_peer_state INNER JOIN
								 sync_state ON sync_peer_state.EntityId = sync_state.EntityId AND 
								 sync_peer_state.RecordId = sync_state.RecordId AND 
								 sync_peer_state.SentLastKnownHash = sync_state.RecordHash)
			)
	)
and
	RecordId in (
		select RecordId from sync_peer_state where 
			ChangedByClient = '0' 
		and
				NodeId = $1
		and
			RecordId in (
				SELECT RecordId from sync_peer_state 	where SentLastKnownHash not in	(
				SELECT  SentLastKnownHash       
				FROM            sync_peer_state INNER JOIN
								 sync_state ON sync_peer_state.EntityId = sync_state.EntityId AND 
								 sync_peer_state.RecordId = sync_state.RecordId AND 
								 sync_peer_state.SentLastKnownHash = sync_state.RecordHash)
			) 
		and
			EntityId in (
				SELECT EntityId from sync_peer_state 	where SentLastKnownHash not in	(
				SELECT  SentLastKnownHash       
				FROM            sync_peer_state INNER JOIN
								 sync_state ON sync_peer_state.EntityId = sync_state.EntityId AND 
								 sync_peer_state.RecordId = sync_state.RecordId AND 
								 sync_peer_state.SentLastKnownHash = sync_state.RecordHash)
			)
	)
and
	NodeId = $1;`
	_, err = self.db.Exec(sqlStr, nodeIdToQueue)
	if err != nil {
		msg := "Error updating, nodeIdToQueue='%s', Error:%s"
		log.Printf(msg, nodeIdToQueue, err.Error())
		return 0, err
	}

	/*
		var affectedCount int64
		affectedCount, err = result.RowsAffected()
		if err != nil {
			msg := "Error getting results from database, inputData=PairId:'%s', SessionId:'%s'. Error:%s"
			log.Printf(msg, item.PairId, item.SessionId, err.Error())
			//log.Println("Error inserting to database, inputData=", item)
			return answer, err
		}
	*/
	sqlStr = `
	update sync_peer_state set SessionBindId=$1 where NodeId=$2 and ChangedByClient = '1';`
	_, err = self.db.Exec(sqlStr, sessionId, nodeIdToQueue)
	if err != nil {
		msg := "Error updating, nodeIdToQueue='%s', Error:%s"
		log.Printf(msg, sessionId, nodeIdToQueue, err.Error())
		return 0, err
	}

	sqlStr = "select count(*) from sync_peer_state where NodeId=$1 and ChangedByClient='1';"
	rows, err := self.db.Query(sqlStr, nodeIdToQueue)
	if err != nil {
		msg := "Error querying, nodeIdToQueue='%s', Error:%s"
		log.Printf(msg, sessionId, nodeIdToQueue, err.Error())
		return 0, err
	}
	var answer int = 0
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&answer)
	}
	return answer, nil
}
