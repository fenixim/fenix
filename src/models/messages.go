package websocket_models

// Sends a message to the server.  Clients will recieve this message back, when it broadcasts.
type SendMessage struct {
	T       string `json:"type"`
	Message string `json:"msg"`
}

func (b SendMessage) Type() string {
	return "msg_send"
}

// Sends a message to the clients.
type BroadcastMessage struct {
	T         string `json:"type"`
	MessageID string `json:"m_id"`
	Author    Author `json:"author"`
	Message   string `json:"msg"`
	Time      int64  `json:"time"`
}

func (b BroadcastMessage) Type() string {
	return "msg_broadcast"
}

type Author struct {
	ID string
	Nick string
}