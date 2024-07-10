package block_nbt

// 荧光物品展示框
type GlowFrame struct {
	Frame
}

// ID ...
func (*GlowFrame) ID() string {
	return IDGlowFrame
}
