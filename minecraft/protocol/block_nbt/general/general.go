package general

import "github.com/OmineDev/neomega-core/minecraft/protocol"

// 描述一个方块实体的通用字段
type Global struct {
	IsMovable      byte   `nbt:"isMovable"`      // TAG_Byte(1) = 1
	CustomName     string `nbt:"CustomName"`     // TAG_String(8) = ""
	UserCustomData string `nbt:"UserCustomData"` // TAG_String(8) = ""
}

func (g *Global) Marshal(r protocol.IO) {
	r.Uint8(&g.IsMovable)
	r.String(&g.CustomName)
	r.String(&g.UserCustomData)
}

func (g *Global) ToNBT() map[string]any {
	return map[string]any{
		"isMovable":      g.IsMovable,
		"CustomName":     g.CustomName,
		"UserCustomData": g.UserCustomData,
	}
}

func (g *Global) FromNBT(x map[string]any) {
	g.IsMovable = x["isMovable"].(byte)
	g.CustomName = x["CustomName"].(string)
	g.UserCustomData = x["UserCustomData"].(string)
}
