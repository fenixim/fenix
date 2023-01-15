package websocket_models

type YodelCreate struct {
	T    string `json:"type"`
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

type Yodel struct {
	T       string `json:"type"`
	YodelID string `json:"y_id"`
	Name    string `json:"name"`
	Owner string `json:"o_id"`
}

func (b Yodel) Type() string {
	b.T = "yodel"
	return b.T
}
func (b Yodel) SetType() JSONModel {
	b.T = b.Type()
	return b
}

type YodelGet struct {
	T       string `json:"type"`
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
