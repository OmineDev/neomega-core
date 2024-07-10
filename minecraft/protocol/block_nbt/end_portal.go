package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 末地折跃门
type EndPortal struct {
	general.Global
}

// ID ...
func (*EndPortal) ID() string {
	return IDEndPortal
}
