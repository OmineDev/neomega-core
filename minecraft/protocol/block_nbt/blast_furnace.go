package block_nbt

// 高炉
type BlastFurnace struct {
	Furnace
}

// ID ...
func (*BlastFurnace) ID() string {
	return IDBlastFurnace
}
