package websocket_models

// Base interface for all messages.
type JSONModel interface {
	Type() string
	GetNonce() string
	SetType() JSONModel
}

// Used to obtain your own client ID
type WhoAmI struct {
	T        string `json:"type"`
	ID       string `json:"id"`
	Username string `json:"nick"`
	Nonce    string `json:"n"`
}

func (b WhoAmI) Type() string {
	return "whoami"
}

func (b WhoAmI) SetType() JSONModel {
	b.T = b.Type()
	return b
}

func (n WhoAmI) GetNonce() string {
	return n.Nonce
}

type GenericError struct {
	T       string `json:"type"`
	Nonce   string `json:"n"`
	Error   string `json:"error"`
	Message string `json:"msg"`
}

func (e GenericError) Type() string {
	return "error"
}

func (b GenericError) SetType() JSONModel {
	b.T = b.Type()
	return b
}
func (n GenericError) GetNonce() string {
	return n.Nonce
}
