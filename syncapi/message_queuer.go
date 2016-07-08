package syncapi

//MessageQueuing provides services for queuing messages for downstream processes.
type MessageQueuing interface {
	Queue(sessionID string, nodeIDToQueue string) (int, error)
}
