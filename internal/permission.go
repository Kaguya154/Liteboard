package internal

import (
	"encoding/json"

	"github.com/Kaguya154/dbhelper"
	"github.com/Kaguya154/dbhelper/types"
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
	// Get all projects where user has at least read permission
	cond := dbhelper.Cond().Eq("user_id", userID).Eq("content_type", "project").Build()
	rows, err := db.Query("detail_permission", cond)
	if err != nil {
		return nil, err
	}

	projectIDs := make(map[int64]bool)
	for _, data := range rows.All() {
		action := data["action"].(string)
		if getPermissionLevel(action) >= PermissionRead {
			contentIDsJson := data["content_ids"].(string)
			var contentIDs []int64
			json.Unmarshal([]byte(contentIDsJson), &contentIDs)
			for _, id := range contentIDs {
				projectIDs[id] = true
			}
		}
	}

	var projects []Project
	for id := range projectIDs {
		p, err := GetProject(db, id)
		if err != nil {
			continue // skip if not found
		}
		projects = append(projects, *p)
	}

	return projects, nil
}

func GetContentListsForProject(db types.Conn, projectID int64) ([]ContentList, error) {
	cond := dbhelper.Cond().Eq("project_id", projectID).Build()
	rows, err := db.Query("content_list", cond)
	if err != nil {
		return nil, err
	}

	var lists []ContentList
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
