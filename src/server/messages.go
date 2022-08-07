package server


// Sends a message to the server.  Clients will recieve this message back, when it broadcasts.
type SendMessage struct {
	T    string `json:"type"`
	Message string `json:"msg"`
}

func (b SendMessage) Type() string {
	b.T = "msg_send"
	return b.T
}
func (b SendMessage) SetType() JSONModel {
	b.T = b.Type()
	return b
}

// Sends a message to the clients.
type BroadcastMessage struct {
	T    string `json:"type"`

	MessageID string `json:"m_id"`
	Author    Author `json:"author"`
	Message   string `json:"msg"`
	Time      int64  `json:"time"`
}

func (b BroadcastMessage) Type() string {
	b.T = "msg_broadcast"
	return b.T
}
func (b BroadcastMessage) SetType() JSONModel {
	b.T = b.Type()
	return b
}

type Author struct {
	ID string
	Nick string
}

type MessageHistory struct {
	T    string `json:"type"`
	From int64 `json:"from,omitempty"`
	To int64 `json:"to,omitempty"`
	Messages []Message `json:"messages,omitempty"`
}

func (m MessageHistory) Type() string {
	m.T = "msg_history"
	return m.T
}

func (b MessageHistory) SetType() JSONModel {
	b.T = b.Type()
	return b
}