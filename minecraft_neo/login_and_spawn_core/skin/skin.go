package skin

/*
skinImageData 指代皮肤的 PNG 二进制形式，
skinData 指代皮肤的一维的密集像素矩阵，
skinGeometryData 指代皮肤的骨架信息，
skinWidth 和 skinHight 则分别指代皮肤的
宽度和高度。
*/
type Skin struct {
	SkinImageData     []byte
	SkinPixels        []byte
	SkinGeometry      []byte
	SkinResourcePatch []byte
	SkinWidth         int
	SkinHight         int
}
