package supported_item

import "strings"

type ItemTypeDescription string

const (
	SimpleNonBlockItem     = ItemTypeDescription("simple non-block item")
	SimpleBlockItem        = ItemTypeDescription("simple block item")
	NeedHotBarNonBlockItem = ItemTypeDescription("need hotbar non-block item")
	NeedHotBarBlockItem    = ItemTypeDescription("need hotbar block item")

	NeedAuxBlockNonBlockItem  = ItemTypeDescription("need aux block non-block item")
	NeedAuxBlockBlockItem     = ItemTypeDescription("need aux block block item")
	ComplexBlockItemContainer = ItemTypeDescription("complex block item: container")
	ComplexBlockItemUnknown   = ItemTypeDescription("complex block item: unknown")

	ComplexBlockContainer = "container"
	ComplexBlockUnknown   = "unknown"

	KnownItemWrittenBook  = "written_book"
	KnownItemWritableBook = "writable_book"
)

// IsBlock: if item can be put as a block, it could have RelatedBlockStateString
func (d ItemTypeDescription) IsBlock() bool { return strings.Contains(string(d), " block item") }

// IsSimple: if can be fully get by replace/give, etc., return true
func (d ItemTypeDescription) IsSimple() bool { return strings.HasPrefix(string(d), "simple") }

// NeedHotbar: item that requires putted in hot bar when generating, usually enchant
func (d ItemTypeDescription) NeedHotbar() bool {
	return strings.Contains(string(d), "need hotbar") || d.NeedAuxBlock()
}

// KnownItem: item that is known to have specific operations in generating, e.g. book
func (d ItemTypeDescription) KnownItem() string {
	if d.IsBlock() {
		return ""
	} else if strings.ContainsAny(string(d), ":") {
		return strings.TrimSpace(strings.Split(string(d), ":")[1])
	}
	return ""
}

// NeedAuxItem: item need aux block when generating, e.g. rename
func (d ItemTypeDescription) NeedAuxBlock() bool {
	return strings.Contains(string(d), "need aux block") || d.IsComplexBlock()
}

// IsComplexBlock: when item put as a block, the block contains certain info
func (d ItemTypeDescription) IsComplexBlock() bool {
	return strings.HasPrefix(string(d), "complex block item")
}

func (d ItemTypeDescription) ComplexBlockSubType() string {
	ss := strings.Split(string(d), ":")
	if len(ss) == 0 {
		return "unknown"
	} else {
		subType := strings.TrimSpace(ss[1])
		if subType == ComplexBlockContainer {
			return ComplexBlockContainer
		}
		return ComplexBlockUnknown
	}
}
