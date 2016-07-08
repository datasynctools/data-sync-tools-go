package syncdaopq

import (
	"database/sql"
	"data-sync-tools-go/syncdao"
	"data-sync-tools-go/syncutil"

	//"reflect"
	//"strconv"
)

//SyncNodePostgresSQLDao implements the syncdao.SyncNodeDao interface as a postgressql implementation.
type SyncNodePostgresSQLDao struct {
	db *sql.DB
}

//AddNode implements the syncdao.SyncNodeDao.AddNode interface as a postgressql implementation.
func (dao SyncNodePostgresSQLDao) AddNode(item syncdao.SyncNode) error {
	// BUG(doug4j@gmail.com): Add SQL injection checks.
	// (see http://go-database-sql.org/retrieving.html and
	// http://stackoverflow.com/questions/26345318/how-can-i-prevent-sql-injection-attacks-in-go-while-using-database-sql)

	var sql = "INSERT INTO sync_node(nodeId,nodeName,dataVersionName) VALUES($1,$2,$3);"

	_, err := dao.db.Exec(sql, item.NodeID, item.NodeName, item.DataVersionName)
	if err != nil {
		syncutil.Error("Error inserting to database, inputData=", item)
		return err
	}
	return nil
}

//GetOneNodeByNodeName implements the syncdao.SyncNodeDao.GetOneNodeByNodeName interface as a postgressql implementation.
func (dao SyncNodePostgresSQLDao) GetOneNodeByNodeName(nodeName string) (syncdao.SyncNode, error) {
	var (
		nodeID          string
		dataVersionName string
	)
	sqlStr := `
SELECT        sync_node.nodeId, sync_node.dataVersionName
FROM          sync_node
WHERE         (sync_node.nodeName = $1)
`
	err := dao.db.QueryRow(sqlStr, nodeName).Scan(&nodeID, &dataVersionName)
	var answer syncdao.SyncNode
	switch {
	case err == sql.ErrNoRows:
		syncutil.Info("No Data")
		return answer, syncdao.ErrDaoNoDataFound
	case err != nil:
		syncutil.Info("DB Error")
		return answer, err
	default:
		answer = syncdao.SyncNode{NodeID: nodeID, NodeName: nodeName, DataVersionName: dataVersionName}
		return answer, nil
	}

}

//GetOneNodeByNodeID implements the syncdao.SyncNodeDao.GetOneNodeByNodeID interface as a postgressql implementation.
func (dao SyncNodePostgresSQLDao) GetOneNodeByNodeID(nodeID string) (syncdao.SyncNode, error) {
	var (
		nodeName        string
		dataVersionName string
	)
	sqlStr := `
SELECT        sync_node.nodeName, sync_node.dataVersionName
FROM          sync_node
WHERE         (sync_node.nodeId = $1)
`
	err := dao.db.QueryRow(sqlStr, nodeID).Scan(&nodeName, &dataVersionName)
	var answer syncdao.SyncNode
	switch {
	case err == sql.ErrNoRows:
		syncutil.Info("No Data")
		return answer, syncdao.ErrDaoNoDataFound
	case err != nil:
		syncutil.Info("DB Error")
		return answer, err
	default:
		answer = syncdao.SyncNode{NodeID: nodeID, NodeName: nodeName, DataVersionName: dataVersionName}
		return answer, nil
	}
}

//DeleteNodeByNodeID implements the syncdao.SyncNodeDao.DeleteNodeByNodeID interface as a postgressql implementation.
func (dao SyncNodePostgresSQLDao) DeleteNodeByNodeID(nodeID string) error {
	var sql = "DELETE from sync_node WHERE (nodeId = $1);"

	result, err := dao.db.Exec(sql, nodeID)
	if err != nil {
		syncutil.Error(err, ". Error deleting from database key:", nodeID)
		return err
	}
	affectedCount, err := result.RowsAffected()
	if err != nil {
		syncutil.Error(err, ". Error determining the number of rows affected for nodeId:", nodeID)
		return err
	}
	if affectedCount == 0 {
		return syncdao.ErrDaoNoDataFound
	}
	return nil
}
