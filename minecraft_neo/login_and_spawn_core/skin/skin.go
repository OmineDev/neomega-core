package skin

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"strings"

	_ "embed"

	"github.com/OmineDev/neomega-core/utils/download_wrapper"
	"github.com/google/uuid"
)

//go:embed default_skin_resource_patch_slim.json
var defaultSkinResourcePatchSlim []byte

//go:embed default_skin_geometry.json
var defaultSkinGeometry []byte

// 描述皮肤信息
type Skin struct {
	// 皮肤的UUID
	SkinUUID string
	// 储存皮肤数据的二进制负载。
	// 对于普通皮肤，这是一个二进制形式的 PNG；
	// 对于高级皮肤(如 4D 皮肤)，
	// 这是一个压缩包形式的二进制表示
	FullSkinData []byte
	// 皮肤的一维密集像素矩阵
	SkinPixels []byte
	// 皮肤的骨架信息
	SkinGeometry []byte
	// SkinResourcePatch 是一个 JSON 编码对象，
	// 其中包含一些指向皮肤所具有的几何形状的字段。
	// 它包含的 JSON 对象指定动画的几何形状，
	// 以及播放器的默认皮肤的组合方式
	SkinResourcePatch []byte
	// 皮肤的宽度
	SkinWidth int
	// 皮肤的高度
	SkinHight int
	// 是否为纤细皮肤
	IsSlim bool
}

func (s *Skin) GetSlimStatus() string{
	if s.IsSlim {
		return "slim"
	}
	return "wide"
}

func makeDefaultSkin(isSlim bool) *Skin {
	defaultSkinResourcePatchCopy := make([]byte, len(defaultSkinResourcePatchSlim))
	copy(defaultSkinResourcePatchCopy, defaultSkinResourcePatchSlim)
	defaultSkinGeometryCopy := make([]byte, len(defaultSkinGeometry))
	copy(defaultSkinGeometryCopy, defaultSkinGeometry)
	if !isSlim {
		defaultSkinResourcePatchCopy = bytes.ReplaceAll(
			defaultSkinResourcePatchCopy,
			[]byte("geometry.humanoid.customSlim"),
			[]byte("geometry.humanoid.custom"),
		)
	}
	return &Skin{
		SkinUUID:          uuid.NewString(),
		SkinResourcePatch: defaultSkinResourcePatchCopy,
		SkinGeometry:      defaultSkinGeometryCopy,
		SkinWidth:         64,
		SkinHight:         32,
		SkinPixels: bytes.Repeat(
			[]byte{0, 0, 0, 255},
			32*64,
		),
		IsSlim: isSlim,
	}
}

// 从 url 指定的网址下载文件，
// 并处理为有效的皮肤数据，
// 然后保存在 skin 中
func ProcessInfoToSkin(url string, isSlim bool) (skin *Skin, err error) {
	// 初始化默认皮肤信息
	var skinImageData []byte
	skin = makeDefaultSkin(isSlim)
	// 从远程服务器下载皮肤文件
	res, err := download_wrapper.DownloadMicroContent(url)
	if err != nil {
		return skin, fmt.Errorf("ProcessURLToSkin: %v", err)
	}
	// 获取皮肤数据
	{
		// 如果这是一个普通的皮肤，
		// 那么 res 就是该皮肤的 PNG 二进制形式，
		// 并且该皮肤使用的骨架格式为默认格式
		skin.FullSkinData, skin.SkinGeometry = res, defaultSkinGeometry
		skinImageData = res
		// 如果这是一个高级的皮肤(比如 4D 皮肤)，
		// 那么 res 是一个压缩包，
		// 我们需要处理这个压缩包以得到皮肤文件
		if len(res) >= 4 && bytes.Equal(res[0:4], []byte("PK\x03\x04")) {
			skinImageData, err = convertZIPToSkin(skin)
			if err != nil {
				return skin, fmt.Errorf("ProcessURLToSkin: %v", err)
			}
		}
	}
	// 将皮肤 PNG 二进制形式解码为图片
	img, err := convertToPNG(skinImageData)
	if err != nil {
		return skin, fmt.Errorf("ProcessURLToSkin: %v", err)
	}
	// 设置皮肤像素、高度、宽度等数据
	skin.SkinPixels = img.(*image.NRGBA).Pix
	skin.SkinWidth, skin.SkinHight = img.Bounds().Dx(), img.Bounds().Dy()
	// 返回值
	return
}

type SkinManifest struct {
	Header struct {
		UUID string `json:"uuid"`
	} `json:"header"`
}

// 从 zipData 指代的 ZIP 二进制数据负载提取皮肤数据，
// 并把处理好的皮肤数据保存在 skin 中，
// 同时返回皮肤图片(PNG)的二进制表示
func convertZIPToSkin(skin *Skin) (skinImageData []byte, err error) {
	// 创建 ZIP 读取器
	reader, err := zip.NewReader(
		bytes.NewReader(skin.FullSkinData),
		int64(len(skin.FullSkinData)),
	)
	if err != nil {
		return nil, fmt.Errorf("convertZIPToSkin: %v", err)
	}
	// 设置皮肤默认资源路径
	skin.SkinResourcePatch = defaultSkinResourcePatchSlim
	// 查找皮肤内容
	for _, file := range reader.File {
		// 皮肤数据
		if strings.HasSuffix(file.Name, ".png") && !strings.HasSuffix(file.Name, "_bloom.png") {
			r, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("convertZIPToSkin: %v", err)
			}
			defer r.Close()
			skinImageData, err = io.ReadAll(r)
			if err != nil {
				return nil, fmt.Errorf("convertZIPToSkin: %v", err)
			}
		}
		// 皮肤骨架信息
		if strings.HasSuffix(file.Name, "geometry.json") {
			r, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("convertZIPToSkin: %v", err)
			}
			defer r.Close()
			geometryData, err := io.ReadAll(r)
			if err != nil {
				return nil, fmt.Errorf("convertZIPToSkin: %v", err)
			}
			processGeometry(skin, geometryData)
		}
		// manifest.json OR pack_manifest.json
		if strings.HasSuffix(file.Name, "manifest.json") {
			r, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("convertZIPToSkin: %v", err)
			}
			defer r.Close()
			manifestData, err := io.ReadAll(r)
			if err != nil {
				return nil, fmt.Errorf("convertZIPToSkin: %v", err)
			}
			var manifest SkinManifest
			if err = json.Unmarshal(manifestData, &manifest); err != nil {
				return nil, fmt.Errorf("convertZIPToSkin: %v", err)
			}
			skin.SkinUUID = manifest.Header.UUID
		}
	}
	// 返回值
	return
}
