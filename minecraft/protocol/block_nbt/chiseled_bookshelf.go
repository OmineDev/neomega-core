package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 雕纹书架
type ChiseledBookshelf struct {
	general.Global
}

// ID ...
func (*ChiseledBookshelf) ID() string {
	return IDChiseledBookshelf
}
