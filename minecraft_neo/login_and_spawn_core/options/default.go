package options

import (
	"crypto/ecdsa"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"log"
	rand2 "math/rand"
	"os"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/login"
	"github.com/OmineDev/neomega-core/minecraft_neo/login_and_spawn_core/skin"
	"github.com/google/uuid"
)

func NewDefaultOptions(
	address string,
	authResp map[string]any,
	PrivateKey *ecdsa.PrivateKey,

) *Options {
	var err error
	opt := &Options{
		Salt:       make([]byte, 16),
		PrivateKey: PrivateKey,
		ErrorLog:   log.New(os.Stderr, "", log.LstdFlags),
	}
	_, _ = rand.Read(opt.Salt)
	opt.ClientData = defaultClientData(address, authResp)
	chainData, _ := authResp["chainInfo"].(string)
	opt.Request = login.Encode(chainData, opt.ClientData, PrivateKey)
	opt.IdentityData, _, _, err = login.Parse(opt.Request)
	if err != nil {
		panic(err)
	}
	opt.ClientData.ThirdPartyName = opt.IdentityData.DisplayName
	if opt.IdentityData.DisplayName == "" {
		panic("invalid identity data: display name")
	}
	if opt.IdentityData.Identity == "" {
		panic("invalid identity data: identity in uuid")
	}
	return opt
}

// defaultClientData edits the ClientData passed to have defaults set to all fields that were left unchanged.
func defaultClientData(address string, authResp map[string]any) login.ClientData {
	bot_level, _ := authResp["growth_level"].(float64)
	skin_info, _ := authResp["skin_info"].(map[string]any)
	skin_iid, _ := skin_info["entity_id"].(string)
	skin_url, _ := skin_info["res_url"].(string)
	skin_is_slim, _ := skin_info["is_slim"].(bool)
	growthLevel := int(bot_level)

	skin, skinErr := skin.ProcessInfoToSkin(skin_url, skin_is_slim)

	d := login.ClientData{}
	d.PremiumSkin = (skinErr == nil)
	d.ServerAddress = address
	d.DeviceOS = protocol.DeviceAndroid
	d.GameVersion = protocol.CurrentVersion
	d.GrowthLevel = growthLevel
	d.ClientRandomID = rand2.Int63()
	d.DeviceID = uuid.New().String()
	d.LanguageCode = "zh_CN"
	d.AnimatedImageData = make([]login.SkinAnimation, 0)
	d.ArmSize = skin.GetSlimStatus()
	d.PersonaPieces = make([]login.PersonaPiece, 0)
	d.PieceTintColours = make([]login.PersonaPieceTintColour, 0)
	d.SelfSignedID = uuid.New().String()
	d.SkinID = skin.SkinUUID
	d.SkinImageHeight = skin.SkinHight
	d.SkinImageWidth = skin.SkinWidth
	d.SkinData = base64.StdEncoding.EncodeToString(skin.SkinPixels)
	d.SkinResourcePatch = base64.StdEncoding.EncodeToString(skin.SkinResourcePatch)
	d.SkinGeometry = base64.StdEncoding.EncodeToString(skin.SkinGeometry)
	d.SkinGeometryVersion = base64.StdEncoding.EncodeToString([]byte("0.0.0"))
	d.SkinItemID = skin_iid
	return d
}
