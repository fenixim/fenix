package websocket_models

type YodelCreate struct {
	T    string `json:"type"`
	Nonce string `json:"n"`
	Name string `json:"name"`
}

func (b YodelCreate) Type() string {
	b.T = "yodel_create"
	return b.T
}
func (b YodelCreate) SetType() JSONModel {
	b.T = b.Type()
	return b
}
func (n YodelCreate) GetNonce() string {
	return n.Nonce
}

type Yodel struct {
	T       string `json:"type"`
	YodelID string `json:"y_id"`
	Name    string `json:"name"`
	Owner   string `json:"o_id"`
	Nonce string `json:"n"`
}

func (b Yodel) Type() string {
	b.T = "yodel"
	return b.T
}
func (b Yodel) SetType() JSONModel {
	b.T = b.Type()
	return b
}
func (n Yodel) GetNonce() string {
	return n.Nonce
}

type YodelGet struct {
	T       string `json:"type"`
	Nonce string `json:"n"`
	YodelID string `json:"y_id"`
}

func (b YodelGet) Type() string {
	b.T = "yodel_get"
	return b.T
}
func (b YodelGet) SetType() JSONModel {
	b.T = b.Type()
	return b
}
func (n YodelGet) GetNonce() string {
	return n.Nonce
}