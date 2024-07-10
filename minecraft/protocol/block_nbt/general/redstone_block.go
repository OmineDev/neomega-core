package general

import "github.com/OmineDev/neomega-core/minecraft/protocol"

// 描述一个 红石类 方块实体的通用字段
type RedstoneBlock struct {
	Powered byte `nbt:"powered"` // TAG_Byte(1) = 0
}

func (rb *RedstoneBlock) Marshal(r protocol.IO) {
	r.Uint8(&rb.Powered)
}

func (rb *RedstoneBlock) ToNBT() map[string]any {
	return map[string]any{
		"powered": rb.Powered,
	}
}

func (rb *RedstoneBlock) FromNBT(x map[string]any) {
	rb.Powered = x["powered"].(byte)
}
