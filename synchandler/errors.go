package synchandler

const (
	//MissingParamSessionID denotes a missing parameter 'sessionId'
	MissingParamSessionID = "Missing 'sessionId' parameter from request"
	//MissingParamNodeID denotes a missing parameter 'nodeId'
	MissingParamNodeID = "Missing 'nodeId' parameter from request"
	//MissingParamOrderNum denotes a missing parameter 'nodeId'
	MissingParamOrderNum = "Missing 'orderNum' parameter from request"
	//MissingParamChangeType denotes a missing parameter 'changeType'
	MissingParamChangeType = "Missing 'changeType' parameter from request"
	//MissingParamTransBindId denotes a missing parameter 'changeType'
	MissingParamTransBindId = "Missing 'transactionBindId' parameter from request"
	//BadOrderNumConversion denotes a bad value for parameter 'orderNum'
	BadOrderNumConversion = "Parameter 'orderNum' cannot be converted to a number. Supplied value: %v"
	//BadChangeTypeConversion denotes a bad value for parameter 'changeType'
	BadChangeTypeConversion = "Parameter 'changeType' cannot be converted to a syncapi.ProcessSyncChangeEnumAddOrUpdate. Supplied value: %v"
)
