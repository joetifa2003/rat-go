package models

type MessageType int

const (
	MessagePing = iota + 1
	MessageScreenShot
	MessageMsg
	MessageExec
)

type Message struct {
	Type               MessageType
	PingResponse       string
	ScreenShotResponse []byte
	MsgRequest         string
	ExecRequest        string
	ExecResponse       string
}
