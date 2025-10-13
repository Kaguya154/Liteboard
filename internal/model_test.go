package internal

import (
	"encoding/json"
	"testing"
)

func TestNestedContentListMarshalUnmarshal(t *testing.T) {
	// Build a nested structure: list -> [entry, list -> [entry]]
	top := ContentList{
		ID:    1,
		Type:  "list",
		Title: "Top List",
		Items: []ContentItem{
			{Entry: &ContentEntry{ID: 10, Type: "entry", Title: "E1", Content: "Hello"}},
			{List: &ContentList{
				ID:    2,
				Type:  "list",
				Title: "Nested List",
				Items: []ContentItem{
					{Entry: &ContentEntry{ID: 11, Type: "entry", Title: "E2", Content: "World"}},
				},
			}},
		},
	}

	// Marshal
	b, err := json.Marshal(top)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Unmarshal back
	var got ContentList
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Validate shape
	if got.Title != "Top List" || len(got.Items) != 2 {
		t.Fatalf("unexpected top content: %+v", got)
	}

	if got.Items[0].Entry == nil || got.Items[0].List != nil {
		t.Fatalf("first item should be entry: %+v", got.Items[0])
	}
	if got.Items[0].Entry.Title != "E1" {
		t.Fatalf("first entry title mismatch: %s", got.Items[0].Entry.Title)
	}

	if got.Items[1].List == nil || got.Items[1].Entry != nil {
		t.Fatalf("second item should be list: %+v", got.Items[1])
	}

	nested := got.Items[1].List
	if nested.Title != "Nested List" || len(nested.Items) != 1 {
		t.Fatalf("unexpected nested content: %+v", nested)
	}
	if nested.Items[0].Entry == nil || nested.Items[0].Entry.Title != "E2" {
		t.Fatalf("unexpected nested entry: %+v", nested.Items[0])
	}
}

func TestContentItemUnmarshalEntry(t *testing.T) {
	data := []byte(`{"id":42,"type":"entry","title":"Single","content":"body"}`)
	var ci ContentItem
	if err := json.Unmarshal(data, &ci); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if ci.Entry == nil || ci.List != nil {
		t.Fatalf("expected entry, got: %+v", ci)
	}
	if ci.Entry.ID != 42 || ci.Entry.Title != "Single" {
		t.Fatalf("entry content mismatch: %+v", ci.Entry)
	}
}

func TestContentItemUnmarshalList(t *testing.T) {
	data := []byte(`{"id":1,"type":"list","title":"L","items":[{"id":2,"type":"entry","title":"E","content":"c"}]}`)
	var ci ContentItem
	if err := json.Unmarshal(data, &ci); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if ci.List == nil || ci.Entry != nil {
		t.Fatalf("expected list, got: %+v", ci)
	}
	if ci.List.Title != "L" || len(ci.List.Items) != 1 || ci.List.Items[0].Entry == nil {
		t.Fatalf("list content mismatch: %+v", ci.List)
	}
}

func TestContentItemMarshalErrorWhenBothSet(t *testing.T) {
	ci := ContentItem{
		Entry: &ContentEntry{ID: 1, Type: "entry", Title: "E", Content: "c"},
		List:  &ContentList{ID: 2, Type: "list", Title: "L"},
	}
	if _, err := json.Marshal(ci); err == nil {
		t.Fatalf("expected error when both Entry and List are set")
	}
}
