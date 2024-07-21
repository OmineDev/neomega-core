package supported_item

import "fmt"

// to avoid using interface or "any" we enum all known item supported here
type SpecificKnownNonBlockItemData struct {
	// is a book
	Pages []string `json:"pages,omitempty"`
	// author if is a written book
	BookAuthor string `json:"book_author,omitempty"`
	// book name
	BookName string `json:"book_name,omitempty"`
	// unknown (describe as a nbt)
	Unknown map[string]any `json:"known_nbt,omitempty"`
}

func (d *SpecificKnownNonBlockItemData) String() string {
	if d == nil {
		return ""
	}
	out := fmt.Sprintf("书名: %v 作者: %v 页数: %v", d.BookName, d.BookAuthor, len(d.Pages))
	return out
}
