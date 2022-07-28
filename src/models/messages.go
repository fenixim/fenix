package models

// Sends a message to the server.  Clients will recieve this message back, when it broadcasts.
type SendMessage struct {
	T       string `json:"type"`
	Message string `json:"msg"`
}

func (b SendMessage) Type() string {
	return "send_message"
}

// Sends a message to the clients.
type RecvMessage struct {
	T       string `json:"type"`
	Author  string `json:"author"`
	Message string `json:"msg"`
	Time    int64 `json:"time"`
}

func (b RecvMessage) Type() string {
	return "recv_message"
}