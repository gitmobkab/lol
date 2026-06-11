package message_bubble

type MsgKind int

const (
	KindBroadcast MsgKind = iota
	KindSystem
	KindDM
)

type Message struct {
	Sender string
	Body   string
	Ts     string
	Own    bool
	Kind   MsgKind
}
