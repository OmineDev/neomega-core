package block_nbt

// 烟熏炉
type Smoker struct {
	Furnace
}

// ID ...
func (*Smoker) ID() string {
	return IDSmoker
}
