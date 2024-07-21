package supported_nbt_data

import (
	"math"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/mitchellh/mapstructure"
)

type SignBlockText struct {
	HideGlowOutline   uint8
	IgnoreLighting    uint8
	PersistFormatting uint8
	SignTextColor     int32
	Text              string
	// TextOwner         string
}

func (t SignBlockText) Color() colorful.Color {
	R, G, B := uint8(t.SignTextColor>>16), uint8(t.SignTextColor>>8), uint8(t.SignTextColor)
	return colorful.Color{R: float64(R) / 256.0, G: float64(G) / 256.0, B: float64(B) / 256.0}
}

// data from PhoenixBuilder/fastbuilder/bdump/nbt_assigner/shared_pool.go
var colorToDyeColorName map[colorful.Color]string = map[colorful.Color]string{
	{R: 240 / 256.0, G: 240 / 256.0, B: 240 / 256.0}: "white_dye",      // 白色染料
	{R: 157 / 256.0, G: 151 / 256.0, B: 151 / 256.0}: "light_gray_dye", // 淡灰色染料
	{R: 71 / 256.0, G: 79 / 256.0, B: 82 / 256.0}:    "gray_dye",       // 灰色染料
	{R: 0 / 256.0, G: 0 / 256.0, B: 0 / 256.0}:       "",               // 黑色染料
	{R: 131 / 256.0, G: 84 / 256.0, B: 50 / 256.0}:   "brown_dye",      // 棕色染料
	{R: 176 / 256.0, G: 46 / 256.0, B: 38 / 256.0}:   "red_dye",        // 红色染料
	{R: 249 / 256.0, G: 128 / 256.0, B: 29 / 256.0}:  "orange_dye",     // 橙色染料
	{R: 254 / 256.0, G: 216 / 256.0, B: 61 / 256.0}:  "yellow_dye",     // 黄色染料
	{R: 128 / 256.0, G: 199 / 256.0, B: 31 / 256.0}:  "lime_dye",       // 黄绿色染料
	{R: 94 / 256.0, G: 124 / 256.0, B: 22 / 256.0}:   "green_dye",      // 绿色染料
	{R: 22 / 256.0, G: 156 / 256.0, B: 156 / 256.0}:  "cyan_dye",       // 青色染料
	{R: 58 / 256.0, G: 179 / 256.0, B: 218 / 256.0}:  "light_blue_dye", // 淡蓝色染料
	{R: 60 / 256.0, G: 68 / 256.0, B: 170 / 256.0}:   "blue_dye",       // 蓝色染料
	{R: 137 / 256.0, G: 50 / 256.0, B: 184 / 256.0}:  "purple_dye",     // 紫色染料
	{R: 199 / 256.0, G: 78 / 256.0, B: 189 / 256.0}:  "magenta_dye",    // 品红色染料
	{R: 243 / 256.0, G: 139 / 256.0, B: 170 / 256.0}: "pink_dye",       // 粉红色染料
}

func (t SignBlockText) GetDyeName() string {
	targetC := t.Color()
	bestColor := ""
	distance := math.Inf(1)
	for colorRGB, dyeName := range colorToDyeColorName {
		dis := colorRGB.DistanceLab(targetC)
		if dis == 0 {
			return dyeName
		}
		if dis < distance {
			distance = dis
			bestColor = dyeName
		}
	}
	return bestColor
}

// id HangingSign
type SignBlockSupportedData struct {
	FrontText, BackText SignBlockText
	IsWaxed             uint8
}

func (opt *SignBlockSupportedData) ToNBT() map[string]any {
	out := map[string]any{}
	mapstructure.Decode(opt, &out)
	return out
}

func SimpleTextToSignBlockSupportedData(text string, lighting bool) *SignBlockSupportedData {
	u8lighting := uint8(0)
	if lighting {
		u8lighting = 1
	}
	return &SignBlockSupportedData{
		IsWaxed: 0,
		FrontText: SignBlockText{
			HideGlowOutline:   0,
			Text:              text,
			IgnoreLighting:    u8lighting,
			PersistFormatting: 1,
			SignTextColor:     -16777216,
		},
		BackText: SignBlockText{
			HideGlowOutline:   0,
			Text:              "",
			IgnoreLighting:    0,
			PersistFormatting: 1,
			SignTextColor:     -16777216,
		},
	}
}

func NbtToSignBlockSupportedData(nbt map[string]any) *SignBlockSupportedData {
	if nbt["id"] != "Sign" && nbt["id"] != "HangingSign" {
		return nil
	}
	if _, found := nbt["IsWaxed"]; !found {
		text, ok := nbt["Text"].(string)
		if !ok {
			text = ""
		}
		lighting, _ := nbt["IgnoreLighting"].(uint8)
		return &SignBlockSupportedData{
			IsWaxed: 0,
			FrontText: SignBlockText{
				HideGlowOutline:   0,
				Text:              text,
				IgnoreLighting:    lighting,
				PersistFormatting: 1,
				SignTextColor:     -16777216,
			},
			BackText: SignBlockText{
				HideGlowOutline:   0,
				Text:              text,
				IgnoreLighting:    lighting,
				PersistFormatting: 1,
				SignTextColor:     -16777216,
			},
		}
	}
	opt := &SignBlockSupportedData{}
	if err := mapstructure.Decode(nbt, &opt); err != nil {
		return nil
	}
	return opt
}
