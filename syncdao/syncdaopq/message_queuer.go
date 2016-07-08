package syncdaopq

import (
	"database/sql"
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncutil"
)

func newMessageQueuer(db *sql.DB) (syncapi.MessageQueuing, error) {

	queuer := postgresSQLMessageQueuer{
		db: db,
	}
	//syncutil.Info(fetcher)
	return queuer, nil
}

//postgresSQLMessageQueuer logically implements MessageQueuing for PostgressSQL database.
type postgresSQLMessageQueuer struct {
	db *sql.DB
}

//Queue implements the syncapi.MessageQueuing interface as a postgressql implementation.
func (queuer postgresSQLMessageQueuer) Queue(sessionID string, nodeIDToQueue string) (int, error) {
	// TODO(doug4j@gmail.com): Add SQL injection checks (see http://go-database-sql.org/retrieving.html and http://stackoverflow.com/questions/26345318/how-can-i-prevent-sql-injection-attacks-in-go-while-using-database-sql
	//Injection Checks are particularly important due to the need for a dynamically set fixed field in the SQL for the
	// Node Id.

	var sqlStr string
	//For queuing previously unqueued records
	/*
	   sqlStr = `
	   insert into sync_peer_state (NodeId, EntitySingularName, RecordId, SentLastKnownHash, ChangedByClient)
	   select '` + nodeIDToQueue
	   sqlStr = sqlStr + `' NodeId, EntitySingularName, RecordId, RecordHash, '1' ChangedByClient from sync_state where
	   					RecordId not in (
	   SELECT        sync_peer_state.RecordId
	   FROM            sync_peer_state INNER JOIN
	   											 sync_state ON sync_peer_state.EntitySingularName = sync_state.EntitySingularName AND
	   											 sync_peer_state.RecordId = sync_state.RecordId
	   WHERE        (sync_peer_state.NodeId = $1)
	   );
	   `
	*/
	sqlStr = `
insert into sync_peer_state (NodeId, EntitySingularName, RecordId, SentLastKnownHash, ChangedByClient, RecordBytesSize)
select '` + nodeIDToQueue
	sqlStr = sqlStr + `', EntitySingularName, RecordId, RecordHash, '1', RecordBytesSize from sync_state where
						RecordId not in (
SELECT        sync_peer_state.RecordId
FROM            sync_peer_state INNER JOIN
                         sync_state ON sync_peer_state.EntitySingularName = sync_state.EntitySingularName AND
                         sync_peer_state.RecordId = sync_state.RecordId
WHERE        (sync_peer_state.NodeId = $1)
);
`
	//syncutil.Debug("sqlStr:", sqlStr)
	_, err := queuer.db.Exec(sqlStr, nodeIDToQueue)
	if err != nil {
		syncutil.Error(err, ". Error inserting with nodeIdToQueue:", nodeIDToQueue, "sql:", sqlStr)
		return 0, err
	}
	sqlStr = `
--For Queuing, how to mark changes to previously queued records
update sync_peer_state set ChangedByClient = '1' where
	EntitySingularName in (
		select EntitySingularName from sync_peer_state where
			ChangedByClient = '0'
		and
				NodeId = $1
		and
			RecordId in (
				SELECT RecordId from sync_peer_state 	where SentLastKnownHash not in	(
				SELECT  SentLastKnownHash
				FROM            sync_peer_state INNER JOIN
								 sync_state ON sync_peer_state.EntitySingularName = sync_state.EntitySingularName AND
								 sync_peer_state.RecordId = sync_state.RecordId AND
								 sync_peer_state.SentLastKnownHash = sync_state.RecordHash)
			)
		and
			EntitySingularName in (
				SELECT EntitySingularName from sync_peer_state 	where SentLastKnownHash not in	(
				SELECT  SentLastKnownHash
				FROM            sync_peer_state INNER JOIN
								 sync_state ON sync_peer_state.EntitySingularName = sync_state.EntitySingularName AND
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
								 sync_state ON sync_peer_state.EntitySingularName = sync_state.EntitySingularName AND
								 sync_peer_state.RecordId = sync_state.RecordId AND
								 sync_peer_state.SentLastKnownHash = sync_state.RecordHash)
			)
		and
			EntitySingularName in (
				SELECT EntitySingularName from sync_peer_state 	where SentLastKnownHash not in	(
				SELECT  SentLastKnownHash
				FROM            sync_peer_state INNER JOIN
								 sync_state ON sync_peer_state.EntitySingularName = sync_state.EntitySingularName AND
								 sync_peer_state.RecordId = sync_state.RecordId AND
								 sync_peer_state.SentLastKnownHash = sync_state.RecordHash)
			)
	)
and
	NodeId = $1;`
	//syncutil.Debug("sqlStr:", sqlStr)
	_, err = queuer.db.Exec(sqlStr, nodeIDToQueue)
	if err != nil {
		syncutil.Error(err, ". Error inserting with nodeIdToQueue:", nodeIDToQueue)
		return 0, err
	}
	sqlStr = `
	update sync_peer_state set SessionBindId=$1 where NodeId=$2 and ChangedByClient = '1';`
	_, err = queuer.db.Exec(sqlStr, sessionID, nodeIDToQueue)
	if err != nil {
		syncutil.Error(err, ". Error updating with nodeIdToQueue:", nodeIDToQueue)
		return 0, err
	}

	sqlStr = "select count(*) from sync_peer_state where NodeId=$1 and ChangedByClient='1';"
	rows, err := queuer.db.Query(sqlStr, nodeIDToQueue)
	if err != nil {
		syncutil.Error(err, ". Error querying with nodeIdToQueue:", nodeIDToQueue, "sessionId:", sessionID)
		return 0, err
	}
	var answer int //This is 0 by default, of course
	closeRowQuietly := func() {
		err := rows.Close()
		if err != nil {
			syncutil.Error("Quietly handling of row close error. Error: " + err.Error())
		}
	}
	defer closeRowQuietly()
	for rows.Next() {
		err = rows.Scan(&answer)
		if err != nil {
			syncutil.Error("Error scanning for change count. Error:", err.Error())
			return answer, err
		}
	}
	return answer, nil
}
