package synchandler

import (
	"data-sync-tools-go/syncdao"
	"data-sync-tools-go/syncutil"
	"data-sync-tools-go/testhelper"
	"net/http"

	"github.com/gorilla/mux"
)

//IntegrationTestReset re-establishes the test data for a particular integration test
func IntegrationTestReset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var testName string
	var hasValue bool
	testName, hasValue = vars["testName"]
	if !hasValue {
		syncutil.Error("testName not found")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	syncutil.Info("Resetting test", testName)
	if dbSetupError(w) {
		return
	}
	dbFactory := syncdao.DefaultDaos
	sqlDbable, ok := dbFactory.(syncdao.SQLDbable)
	if !ok {
		syncutil.Error("dao factory does not contain sql SQLDbable interface")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	db := sqlDbable.SQLDb()
	testhelper.SetupData(db, testName)
	return
}
