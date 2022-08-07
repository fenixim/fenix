package websocket_models

// Base interface for all messages.
type JSONModel interface {
	Type() string
}

// Used to obtain your own client ID
type WhoAmI struct {
	T    string `json:"type"`
	ID   string `json:"id"`
	Nick string `json:"nick"`
}

func (b WhoAmI) Type() string {
	return "whoami"
}

type GenericError struct {
	T string `json:"type"`
	Error string `json:"error"`
	Message string `json:"msg"`
}
func (e GenericError) Type() string {
	return "error"
}