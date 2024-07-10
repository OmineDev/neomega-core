package block_nbt

// 木桶
type Barrel struct {
	Chest
}

// ID ...
func (*Barrel) ID() string {
	return IDBarrel
}
