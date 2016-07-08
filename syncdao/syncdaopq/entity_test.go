package syncdaopq

import (
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncutil"
	"data-sync-tools-go/testhelper"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/twinj/uuid"
)

func TestSyncFetcher_FindEntitiesForFetchOK(t *testing.T) {
	testName := syncutil.GetCallingName()
	testhelper.StartTest(testName)
	db, err := createAndVerifyDBConn(testDbUser, testDbPassword, testDbHost, testDbName, testDbPort)
	if err != nil {
		t.Error("Failed to connect to database: " + err.Error())
		return
	}
	closeQuietly := func() {
		err := db.Close()
		if err != nil {
			syncutil.Error("Quietly handling of db close error. Error: " + err.Error())
		}
	}
	defer closeQuietly()

	testhelper.RemoveSyncTablesIfNeeded(db)
	testhelper.CreateSyncTables(db)
	testhelper.AddProfile4SampleSyncData(db)

	sessionUUID := uuid.Formatter(uuid.NewV4(), uuid.CleanHyphen)
	nodeID := "*node-hub"
	entityFetcher, err := newEntityFetcher(sessionUUID, nodeID, db)
	if err != nil {
		syncutil.Error(err.Error())
		t.Error(err.Error())
	}
	orderNum := 1
	changeType := syncapi.ProcessSyncChangeEnumAddOrUpdate

	expectedEntities := []syncapi.EntityNameItem{
		syncapi.EntityNameItem{
			SingularName: "A",
			PluralName:   "A",
		},
		syncapi.EntityNameItem{
			SingularName: "B",
			PluralName:   "B",
		},
		syncapi.EntityNameItem{
			SingularName: "C",
			PluralName:   "C",
		},
	}

	actualEntities, err := entityFetcher.FindEntitiesForFetch(orderNum, sessionUUID, nodeID, changeType)
	if err != nil {
		syncutil.Error(err.Error())
		t.Error(err.Error())
	}
	assert.Equal(t, expectedEntities, actualEntities, "entities should be equal")

	testhelper.EndTest(testName)
}
