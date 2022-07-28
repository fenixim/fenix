package models

// Base interface for all messages.
type JSONModel interface {
	Type() string
}

// Sent when the websocket message is incorrectly formatted, ie JSON error, missing type, etc.
type BadFormat struct {
	T       string `json:"type"`
	Message string `json:"msg"`
}

func (b BadFormat) Type() string {
	return "err_bad_format"
}

// Used to obtain your own client ID
type WhoAmI struct {
	T       string `json:"type"`
	ID      string `json:"id"`
	Nick    string `json:"nick"`

}
func (b WhoAmI) Type() string {
	return "whoami"
}
