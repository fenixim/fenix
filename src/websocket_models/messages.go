package websocket_models

import "fenix/src/database"

// Sends a message to the server.  Clients will recieve this message back, when it broadcasts.
type MsgSend struct {
	T       string `json:"type"`
	Message string `json:"msg"`
}

func (b MsgSend) Type() string {
	b.T = "msg_send"
	return b.T
}
func (b MsgSend) SetType() JSONModel {
	b.T = b.Type()
	return b
}

// Sends a message to the clients.
type MsgBroadcast struct {
	T string `json:"type"`

	MessageID string `json:"m_id"`
	Author    Author `json:"author"`
	Message   string `json:"msg"`
	Time      int64  `json:"time"`
}

func (b MsgBroadcast) Type() string {
	b.T = "msg_broadcast"
	return b.T
}
func (b MsgBroadcast) SetType() JSONModel {
	b.T = b.Type()
	return b
}

type Author struct {
	ID       string
	Username string
}

type MsgHistory struct {
	T        string              `json:"type"`
	From     int64               `json:"from,omitempty"`
	To       int64               `json:"to,omitempty"`
	Messages []*database.Message `json:"messages,omitempty"`
}

func (m MsgHistory) Type() string {
	m.T = "msg_history"
	return m.T
}

func (b MsgHistory) SetType() JSONModel {
	b.T = b.Type()
	return b
}
