package block_nbt

// 投掷器
type Dropper struct {
	Dispenser
}

// ID ...
func (*Dropper) ID() string {
	return IDDropper
}
