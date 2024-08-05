package skin

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
)

// 将 imageData 解析为 PNG 图片
func convertToPNG(imageData []byte) (image.Image, error) {
	buffer := bytes.NewBuffer(imageData)
	img, err := png.Decode(buffer)
	if err != nil {
		return nil, fmt.Errorf("convertToPNG: %v", err)
	}
	return img, nil
}
