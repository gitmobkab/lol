package message_bubble

type MsgKind int

const (
	KindBroadcast MsgKind = iota
	KindSystem
	KindDM
	KindFile
)

type Message struct {
	Sender string
	Body   string
	Ts     string
	Own    bool
	Kind   MsgKind
}
