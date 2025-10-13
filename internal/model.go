package internal

import (
	"encoding/json"
	"fmt"
)

type Project struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type User struct {
	ID       int64    `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Groups   []string `json:"groups"`
}

type UserInternal struct {
	ID           int64
	Username     string
	Email        string
	OpenID       string
	PasswordHash string
	Groups       []string
}

type Sidebar struct {
	ID          int64         `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Items       []SidebarItem `json:"items"`
}

type SidebarItem struct {
	ID       int64  `json:"id"`
	ParentID int64  `json:"parent_id"`
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	URL      string `json:"url"`
	Order    int    `json:"order"`
}

type Page struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	AuthorID int64  `json:"author_id"`
}

// ContentItem is a union type that can hold either a ContentEntry or a ContentList.
// It enables nesting ContentList within another ContentList.
// Only one of Entry or List should be non-nil.
type ContentItem struct {
	Entry *ContentEntry
	List  *ContentList
}

func (ci ContentItem) MarshalJSON() ([]byte, error) {
	switch {
	case ci.Entry != nil && ci.List == nil:
		return json.Marshal(ci.Entry)
	case ci.List != nil && ci.Entry == nil:
		return json.Marshal(ci.List)
	case ci.Entry == nil && ci.List == nil:
		return []byte("null"), nil
	default:
		return nil, fmt.Errorf("ContentItem must have exactly one of Entry or List set")
	}
}

func (ci *ContentItem) UnmarshalJSON(data []byte) error {
	// Probe to decide whether it's a list (has items) or an entry (has content)
	var probe struct {
		Items   json.RawMessage `json:"items"`
		Content json.RawMessage `json:"content"`
		Type    string          `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}

	// Heuristics: if it has an "items" field or type == "list", treat as ContentList
	if len(probe.Items) != 0 || probe.Type == "list" {
		var list ContentList
		if err := json.Unmarshal(data, &list); err != nil {
			return err
		}
		ci.List = &list
		ci.Entry = nil
		return nil
	}

	// Otherwise treat it as ContentEntry
	var entry ContentEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return err
	}
	ci.Entry = &entry
	ci.List = nil
	return nil
}

type ContentList struct {
	ID    int64         `json:"id"`
	Type  string        `json:"type"`
	Title string        `json:"title"`
	Items []ContentItem `json:"items"`
}

type ContentEntry struct {
	ID      int64  `json:"id"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type DetailPermission struct {
	ID          int64   `json:"id"`
	ContentType string  `json:"content_type"`
	ContentIDs  []int64 `json:"content_ids"`
	Action      string  `json:"action"`
}

type Permission struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ContentType string `json:"content_type"`
	Action      string `json:"action"`
	Detail      int64  `json:"detail"`
}

type Role struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Permissions []int64 `json:"permissions"`
}
