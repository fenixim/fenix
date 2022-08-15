package websocket_models

type Invite struct {
	InviteID   string
	Channel    Channel
	ValidUntil int64
	UseLimit  int64
}

type InviteCreate struct {
	T          string `json:"type"`
	ChannelID  string `json:"channelID"`
	ValidUntil int64  `json:"validUntil"`
	UseLimit  int64  `json:"timesUsed"`
}

func (c InviteCreate) Type() string {
	return "inv_create"
}

func (c InviteCreate) SetType() JSONModel {
	c.T = c.Type()
	return c
}

type InviteGet struct {
	T          string `json:"type"`
	InviteID   string `json:"inviteID"`
	ChannelID  string `json:"channelID,omitempty"`
	ValidUntil int64  `json:"validUntil,omitempty"`
	UseLimit  int64  `json:"timesUsed,omitempty"`
}

func (c InviteGet) Type() string {
	return "inv_get"
}

func (c InviteGet) SetType() JSONModel {
	c.T = c.Type()
	return c
}

type InviteDelete struct {
	T          string `json:"type"`
	InviteID   string `json:"inviteID"`
}

func (c InviteDelete) Type() string {
	return "inv_delete"
}

func (c InviteDelete) SetType() JSONModel {
	c.T = c.Type()
	return c
}

