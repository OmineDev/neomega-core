package tiled_buffer

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt"
	"github.com/pterm/pterm"
)

// 测试 __tag NBT <-> NBT Go Struct <-> NBT Map 是否正常工作
func TestFull(t *testing.T) {
	for _, element := range NewPool() {
		var blockNBTMap map[string]any
		// prepare
		{
			block, err := Decode(element.ID, bytes.NewBuffer(element.Buffer))
			if err != nil {
				t.Errorf("TestFull: %v", err)
			}
			blockNBTMap = block.ToNBT()
			// read
			secondBlockMap, err := WriteAndRead(element.ID, block)
			if err != nil {
				t.Errorf("TestFull: %v", err)
			}
			if !reflect.DeepEqual(blockNBTMap, secondBlockMap) {
				t.Errorf("TestFull: Marshal and unmarshal is unequivalence; element.ID = %#v", element.ID)
			}
			// write
		}
		// __tag NBT <-> NBT Go Struct
		{
			new := block_nbt.NewPool()[element.ID]
			new.FromNBT(blockNBTMap)
			if newBlockNBTMap := new.ToNBT(); !reflect.DeepEqual(blockNBTMap, newBlockNBTMap) {
				t.Errorf("TestFull: FromNBT and ToNBT is unequivalence; element.ID = %#v", element.ID)
				pterm.Warning.Printf("%#v\n", blockNBTMap)
				pterm.Warning.Printf("%#v\n", newBlockNBTMap)
			}
		}
		// NBT Go Struct <-> NBT Map
		pterm.Success.Printf("%v\n", blockNBTMap)
		// print success
	}
}
