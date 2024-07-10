package block_nbt

// 悬挂式告示牌
type HangingSign struct {
	Sign
}

// ID ...
func (*HangingSign) ID() string {
	return IDHangingSign
}
