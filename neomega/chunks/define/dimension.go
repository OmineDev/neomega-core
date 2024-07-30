package define

import (
	"fmt"
)

type Dimension int32

const (
	DimensionOverworld = Dimension(0)
	DimensionNether    = Dimension(1)
	DimensionTheEnd    = Dimension(2)
)

func (d Dimension) String() string {
	switch d {
	case DimensionOverworld:
		return "overworld"
	case DimensionNether:
		return "nether"
	case DimensionTheEnd:
		return "the_end"
	default:
		return fmt.Sprintf("dm%v", int32(d))
	}
}

func (d Dimension) RangeUpperInclude() Range {
	switch d {
	case DimensionOverworld:
		return OverWorldRangeUpperInclude
	case DimensionNether:
		return NetherRangeUpperInclude
	case DimensionTheEnd:
		return TheEndRangeUpperInclude
	default:
		return OverWorldRangeUpperInclude
	}
}

func (d Dimension) RangeUpperExclude() Range {
	switch d {
	case DimensionOverworld:
		return OverWorldRangeUpperExclude
	case DimensionNether:
		return NetherRangeUpperExclude
	case DimensionTheEnd:
		return TheEndRangeUpperExclude
	default:
		return OverWorldRangeUpperExclude
	}
}

var OverWorldRangeUpperInclude = Range{-64, 319}
var NetherRangeUpperInclude = Range{0, 127}
var TheEndRangeUpperInclude = Range{0, 255}

var OverWorldRangeUpperExclude = Range{-64, 320}
var NetherRangeUpperExclude = Range{0, 128}
var TheEndRangeUpperExclude = Range{0, 254}
