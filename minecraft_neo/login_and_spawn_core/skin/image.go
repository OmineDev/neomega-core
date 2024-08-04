package skin

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"strings"

	_ "embed"

	"github.com/OmineDev/neomega-core/utils/download_wrapper"
)

//go:embed default_skin_resource_patch.json
var defaultSkinResourcePatch []byte

//go:embed default_skin_geometry.json
var defaultSkinGeometry []byte

func makeDefaultSkin() *Skin {
	defaultSkinResourcePatchCopy := make([]byte, len(defaultSkinResourcePatch))
	copy(defaultSkinResourcePatchCopy, defaultSkinResourcePatch)
	defaultSkinGeometryCopy := make([]byte, len(defaultSkinGeometry))
	copy(defaultSkinGeometryCopy, defaultSkinGeometry)
	return &Skin{
		SkinResourcePatch: defaultSkinResourcePatchCopy,
		SkinGeometry:      defaultSkinGeometryCopy,
		SkinWidth:         64,
		SkinHight:         32,
		SkinPixels: bytes.Repeat(
			[]byte{0, 0, 0, 255},
			32*64,
		),
	}
}

/*
从 url 指定的网址下载文件，
并处理为有效的皮肤数据。
*/
func ProcessURLToSkin(url string) (skin *Skin, err error) {
	// create default skin
	skin = makeDefaultSkin()
	// download skin file from remote server
	res, err := download_wrapper.DownloadMicroContent(url)
	if err != nil {
		return skin, fmt.Errorf("ProcessURLToSkin: %v", err)
	}
	// get skin data
	skin.SkinImageData, skin.SkinGeometry = res, defaultSkinGeometry
	if len(res) >= 4 && bytes.Equal(res[0:4], []byte("PK\x03\x04")) {
		if err = convertZIPToSkin(skin, res); err != nil {
			return skin, fmt.Errorf("ProcessURLToSkin: %v", err)
		}
	}
	// decode to image
	img, err := convertToPNG(skin.SkinImageData)
	if err != nil {
		return skin, fmt.Errorf("ProcessURLToSkin: %v", err)
	}
	// encode to pixels and return
	skin.SkinPixels = img.(*image.NRGBA).Pix
	skin.SkinWidth, skin.SkinHight = img.Bounds().Dx(), img.Bounds().Dy()
	return
}

// 从 zipData 指代的 ZIP 二进制数据负载提取皮肤数据。
// skinImageData 代表皮肤的 PNG 二进制形式，
// skinGeometry 代表皮肤的骨架信息。
func convertZIPToSkin(skin *Skin, zipData []byte) (err error) {
	// create reader
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("ConvertZIPToSkin: %v", err)
	}
	// set default resource patch
	skin.SkinResourcePatch = defaultSkinResourcePatch
	// find skin contents
	for _, file := range reader.File {
		// skin data
		if strings.HasSuffix(file.Name, ".png") && !strings.HasSuffix(file.Name, "_bloom.png") {
			r, err := file.Open()
			if err != nil {
				return fmt.Errorf("ConvertZIPToSkin: %v", err)
			}
			defer r.Close()
			skin.SkinImageData, err = io.ReadAll(r)
			if err != nil {
				return fmt.Errorf("ConvertZIPToSkin: %v", err)
			}
		}
		// skin geometry
		if strings.HasSuffix(file.Name, "geometry.json") {
			r, err := file.Open()
			if err != nil {
				return fmt.Errorf("ConvertZIPToSkin: %v", err)
			}
			defer r.Close()
			geometryData, err := io.ReadAll(r)
			if err != nil {
				return fmt.Errorf("ConvertZIPToSkin: %v", err)
			}
			processGeometry(skin, geometryData)
		}
	}
	return
}

// 将 imageData 解析为 PNG 图片
func convertToPNG(imageData []byte) (image.Image, error) {
	buffer := bytes.NewBuffer(imageData)
	img, err := png.Decode(buffer)
	if err != nil {
		return nil, fmt.Errorf("ConvertToPNG: %v", err)
	}
	return img, nil
}
