package websocket_models

type Channel struct {
	ChannelID string
	Name      string
	Owner     User
}

type ChannelCreate struct {
	T    string `json:"type"`
	Name string `json:"name"`
}

func (c ChannelCreate) Type() string {
	return "chnl_create"
}

func (c ChannelCreate) SetType() JSONModel {
	c.T = c.Type()
	return c
}

type ChannelDelete struct {
	T         string `json:"type"`
	ChannelID string `json:"channelID"`
}

func (c ChannelDelete) Type() string {
	return "chnl_delete"
}

func (c ChannelDelete) SetType() JSONModel {
	c.T = c.Type()
	return c
}

type ChannelGet struct {
	T         string `json:"type"`
	ChannelID string `json:"channelID"`
	Name      string `json:"name,omitempty"`
	Owner     User   `json:"owner,omitempty"`
}

func (c ChannelGet) Type() string {
	return "chnl_get"
}

func (c ChannelGet) SetType() JSONModel {
	c.T = c.Type()
	return c
}

type ChannelLeave struct {
	T         string `json:"type"`
	ChannelID string `json:"channelID"`
}

func (c ChannelLeave) Type() string {
	return "chnl_leave"
}

func (c ChannelLeave) SetType() JSONModel {
	c.T = c.Type()
	return c
}

type ChannelJoin struct {
	T        string `json:"type"`
	InviteID string `json:"inviteID"`
}

func (c ChannelJoin) Type() string {
	return "chnl_join"
}

func (c ChannelJoin) SetType() JSONModel {
	c.T = c.Type()
	return c
}

type ChannelMembers struct {
	T         string  `json:"type"`
	ChannelID string  `json:"channelID"`
	Members   []*User `json:"members,omitempty"`
}

func (c ChannelMembers) Type() string {
	return "chnl_members"
}

func (c ChannelMembers) SetType() JSONModel {
	c.T = c.Type()
	return c
}

type ChannelSubscribe struct {
	T         string `json:"type"`
	ChannelID string `json:"channelID"`
}

func (c ChannelSubscribe) Type() string {
	return "chnl_subscribe"
}

func (c ChannelSubscribe) SetType() JSONModel {
	c.T = c.Type()
	return c
}

type ChannelUnsubscribe struct {
	T         string `json:"type"`
	ChannelID string `json:"channelID"`
}

func (c ChannelUnsubscribe) Type() string {
	return "chnl_unsubscribe"
}

func (c ChannelUnsubscribe) SetType() JSONModel {
	c.T = c.Type()
	return c
}

type ChannelSubscriptions struct {
	T             string    `json:"type"`
	Subscriptions []Channel `json:"subscriptions,omitempty"`
}

func (c ChannelSubscriptions) Type() string {
	return "chnl_subscriptions"
}

func (c ChannelSubscriptions) SetType() JSONModel {
	c.T = c.Type()
	return c
}
