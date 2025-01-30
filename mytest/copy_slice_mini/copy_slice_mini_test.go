package copyslicemini

import "testing"

// SourceItem 是源结构体
type SourceItem struct {
	ID	int
	Name	string
	Value	float64
}

// DestItem 是目标结构体，与 SourceItem 相似但不完全相同
type DestItem struct {
	ID	int64	// 类型不同
	Name	string
	Amount	float64	// 字段名不同
}

// SourceContainer 包含 SourceItem 切片的结构体
type SourceContainer struct {
	Items []SourceItem
}

// DestContainer 包含 DestItem 切片的结构体
type DestContainer struct {
	Items []DestItem
}

func TestCopyContainer(t *testing.T) {
	src := &SourceContainer{
		Items: []SourceItem{
			{ID: 1, Name: "Item 1", Value: 10.5},
			{ID: 2, Name: "Item 2", Value: 20.7},
		},
	}

	dst := &DestContainer{}
	CopyContainer(dst, src)

	if len(dst.Items) != len(src.Items) {
		t.Errorf("Items length mismatch, got %d, want %d", len(dst.Items), len(src.Items))
	}

	for i, item := range dst.Items {
		if item.ID != int64(src.Items[i].ID) {
			t.Errorf("Item[%d] ID mismatch, got %d, want %d", i, item.ID, src.Items[i].ID)
		}
		if item.Name != src.Items[i].Name {
			t.Errorf("Item[%d] Name mismatch, got %s, want %s", i, item.Name, src.Items[i].Name)
		}
		// if item.Amount != src.Items[i].Value {
		// 	t.Errorf("Item[%d] Amount/Value mismatch, got %f, want %f", i, item.Amount, src.Items[i].Value)
		// }
	}
}

// :quickcopy
func CopyContainer(dst *DestContainer, src *SourceContainer) {

	dst.Items = copySliceDestItemFromSliceSourceItem(src.Items)
}
func copyDestItemFromSourceItem(dst *DestItem, src *SourceItem) {
	dst.ID = int64(src.ID)

	dst.Name = src.Name
}

func copySliceDestItemFromSliceSourceItem(
	src []SourceItem) []DestItem {
	if src == nil {
		return nil
	}
	dst := make([]DestItem, len(src),
	)
	for i := range src {
		copyDestItemFromSourceItem(&dst[i], &src[i])
	}
	return dst
}
