//Package synchandler takes routes http traffic and operates against the given syncdao implementation.
package synchandler

import (
	"data-sync-tools-go/syncapi"
	"data-sync-tools-go/syncdao"
	"data-sync-tools-go/syncutil"
	"fmt"
	"net/http"
)

//RequestParmParser abstracts out the parsing of parsing HTTP request args
/*
type RequestParmParser interface {
	ParseParmsToMap(r *http.Request) map[string]string
}

type RequestParmParserType struct {
	VarsHandler func(*http.Request) map[string]string
}

func (parser RequestParmParserType) ParseParmsToMap(r *http.Request) map[string]string {
	return parser.VarsHandler(r)
}
*/

//Handlers defines the type holding all handling logic for http request routes.
type Handlers struct {
	syncapi.Repository
	VarsHandler func(*http.Request) map[string]string
}

//Index processes HTTP requests for a base url to the configured hostname and application
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome!\n")
}

//MARK: SyncSessionMgmt Processing START

//SyncSessionState is the structure for the sync session state.
type SyncSessionState struct {
}

//SyncSessionMgmtMsg contains one way messages for management purposes used for signaling of state changes between nodes
//MessageTypes include the following: 1-Heartbeat, 2-SeedingClosed, 3-Cancel, 4-HaultingError, 5-Start.
//Optional fields: Result.
//Result possible values:
type SyncSessionMgmtMsg struct {
	SessionID   string `json:"sessionId"`
	SenderNode  string `json:"senderNode"`
	MessageType int    `json:"messageType"`
	Result      string `json:"result,omitempty"`
}

//MARK: SyncSessionMgmt Processing END

//dbSetupError provides common database error checks.
func dbSetupError(w http.ResponseWriter) bool {
	if syncdao.DefaultDaos == nil {
		errMsg := "Server Configuration Error: Database not set"
		syncutil.Error(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return true
	}
	return false
}
