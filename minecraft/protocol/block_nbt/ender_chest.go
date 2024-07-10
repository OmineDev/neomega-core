package block_nbt

// 末影箱
type EnderChest struct {
	Chest
}

// ID ...
func (*EnderChest) ID() string {
	return IDEnderChest
}
