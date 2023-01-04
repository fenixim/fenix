package websocket_models

type YodelCreate struct {
	T string `json:"type"`
}

func (b YodelCreate) Type() string {
	b.T = "yodel_create"
	return b.T
}
func (b YodelCreate) SetType() JSONModel {
	b.T = b.Type()
	return b
}

type Yodel struct {
	T       string `json:"type"`
	YodelID string `json:"y_id"`
}

func (b Yodel) Type() string {
	b.T = "yodel"
	return b.T
}
func (b Yodel) SetType() JSONModel {
	b.T = b.Type()
	return b
}
