package block_set

import "github.com/OmineDev/neomega-core/neomega/blocks/convertor"

func (bs *BlockSet) CreateEmptyConvertor() *convertor.ToNEMCConvertor {
	c := convertor.NewToNEMCConverter(bs.unknownRuntimeID, bs.airRuntimeID)
	for _, b := range bs.blocks {
		c.LoadTargetBlock(b)
	}
	return c
}
