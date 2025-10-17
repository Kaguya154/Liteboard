package internal

import (
	"encoding/json"
	"sort"

	"github.com/Kaguya154/dbhelper"
	"github.com/Kaguya154/dbhelper/types"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

const (
	PermissionNone  = 0
	PermissionRead  = 1
	PermissionWrite = 2
	PermissionAdmin = 3
)

func getPermissionLevel(action string) int {
	switch action {
	case "read":
		return PermissionRead
	case "write":
		return PermissionWrite
	case "admin":
		return PermissionAdmin
	default:
		return PermissionNone
	}
}

// GetPermissionLevel is the exported version of getPermissionLevel
func GetPermissionLevel(action string) int {
	return getPermissionLevel(action)
}

func HasPermission(db types.Conn, userID int64, contentType string, contentID int64, requiredAction string) (bool, error) {
	requiredLevel := getPermissionLevel(requiredAction)
	if requiredLevel == PermissionNone {
		return false, nil
	}

	// Query detail permissions for the user and content type
	cond := dbhelper.Cond().Eq("user_id", userID).Eq("content_type", contentType).Build()
	rows, err := db.Query("detail_permission", cond)
	if err != nil {
		return false, err
	}

	maxLevel := PermissionNone
	for _, data := range rows.All() {
		action := data["action"].(string)
		contentIDsJson := data["content_ids"].(string)
		var contentIDs []int64
		json.Unmarshal([]byte(contentIDsJson), &contentIDs)
		for _, id := range contentIDs {
			if id == contentID {
				level := getPermissionLevel(action)
				if level > maxLevel {
					maxLevel = level
				}
			}
		}
	}

	return maxLevel >= requiredLevel, nil
}

func GetProjectsForUser(db types.Conn, userID int64) ([]Project, error) {
	hlog.Debugf("GetProjectsForUser: Starting for userID=%d", userID)

	// Get all projects where user has at least read permission
	cond := dbhelper.Cond().Eq("user_id", userID).Eq("content_type", "project").Build()

	rows, err := db.Query("detail_permission", cond)
	if err != nil {
		hlog.Errorf("GetProjectsForUser: Query failed, error=%v", err)
		return []Project{}, err
	}

	projectIDs := make(map[int64]bool)
	for _, data := range rows.All() {
		action := data["action"].(string)

		if getPermissionLevel(action) >= PermissionRead {
			contentIDsJson := data["content_ids"].(string)

			var contentIDs []int64
			err := json.Unmarshal([]byte(contentIDsJson), &contentIDs)
			if err != nil {
				continue
			}

			for _, id := range contentIDs {
				projectIDs[id] = true
			}
		}
	}

	// Collect IDs into a sorted slice to ensure consistent ordering
	sortedIDs := make([]int64, 0, len(projectIDs))
	for id := range projectIDs {
		sortedIDs = append(sortedIDs, id)
	}
	// Sort by ID in ascending order for consistent display
	sort.Slice(sortedIDs, func(i, j int) bool {
		return sortedIDs[i] < sortedIDs[j]
	})

	projects := make([]Project, 0) // 初始化为空数组而不是 nil
	for _, id := range sortedIDs {
		p, err := GetProject(db, id)
		if err != nil {
			hlog.Errorf("GetProjectsForUser: Failed to get project ID=%d, error=%v", id, err)
			continue // skip if not found
		}
		projects = append(projects, *p)
	}

	hlog.Debugf("GetProjectsForUser: Returning %d projects for userID=%d", len(projects), userID)
	return projects, nil
}

func GetContentListsForProject(db types.Conn, projectID int64) ([]ContentList, error) {
	cond := dbhelper.Cond().Eq("project_id", projectID).Build()
	rows, err := db.Query("content_list", cond)
	if err != nil {
		return []ContentList{}, err
	}

	lists := make([]ContentList, 0)
	for _, data := range rows.All() {
		list := ContentList{
			ID:    data["id"].(int64),
			Type:  data["type"].(string),
			Title: data["title"].(string),
		}
		lists = append(lists, list)
	}

	return lists, nil
}
