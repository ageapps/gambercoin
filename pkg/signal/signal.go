package signal

type Signal string

var (
	Stop  Signal = "stop"
	Reset Signal = "reset"
	Sync  Signal = "sync"
)
