package internal

import (
	"encoding/json"
	"errors"

	"github.com/Kaguya154/dbhelper"
	"github.com/Kaguya154/dbhelper/types"
)

// Project CRUD

func CreateProject(db types.Conn, p *Project) (int64, error) {
	cond := dbhelper.Cond().Eq("name", p.Name).Eq("description", p.Description).Eq("creator_id", p.CreatorID).Build()
	return db.Insert("project", cond)
}

func GetProject(db types.Conn, id int64) (*Project, error) {
	cond := dbhelper.Cond().Eq("id", id).Build()
	rows, err := db.Query("project", cond)
	if err != nil {
		return nil, err
	}
	if rows.Count() == 0 {
		return nil, errors.New("project not found")
	}
	data := rows.All()[0]
	p := &Project{
		ID:          data["id"].(int64),
		Name:        data["name"].(string),
		Description: data["description"].(string),
		CreatorID:   data["creator_id"].(int64),
	}
	return p, nil
}

func UpdateProject(db types.Conn, id int64, updates *Project) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	upd := dbhelper.Cond().Eq("name", updates.Name).Eq("description", updates.Description).Eq("creator_id", updates.CreatorID).Build()
	_, err := db.Update("project", cond, upd)
	return err
}

func DeleteProject(db types.Conn, id int64) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	_, err := db.Delete("project", cond)
	return err
}

// User CRUD

func CreateUser(db types.Conn, u *User) (int64, error) {
	groupsJson, _ := json.Marshal(u.Groups)
	cond := dbhelper.Cond().Eq("username", u.Username).Eq("email", u.Email).Eq("groups", string(groupsJson)).Build()
	return db.Insert("user", cond)
}

func GetUser(db types.Conn, id int64) (*User, error) {
	cond := dbhelper.Cond().Eq("id", id).Build()
	rows, err := db.Query("user", cond)
	if err != nil {
		return nil, err
	}
	if rows.Count() == 0 {
		return nil, errors.New("user not found")
	}
	data := rows.All()[0]
	u := &User{
		ID:       data["id"].(int64),
		Username: data["username"].(string),
		Email:    data["email"].(string),
	}
	if groupsData, ok := data["groups"].(string); ok {
		json.Unmarshal([]byte(groupsData), &u.Groups)
	}
	return u, nil
}

func UpdateUser(db types.Conn, id int64, updates *User) error {
	groupsJson, _ := json.Marshal(updates.Groups)
	cond := dbhelper.Cond().Eq("id", id).Build()
	upd := dbhelper.Cond().Eq("username", updates.Username).Eq("email", updates.Email).Eq("groups", string(groupsJson)).Build()
	_, err := db.Update("user", cond, upd)
	return err
}

func DeleteUser(db types.Conn, id int64) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	_, err := db.Delete("user", cond)
	return err
}

// Page CRUD

func CreatePage(db types.Conn, p *Page) (int64, error) {
	cond := dbhelper.Cond().Eq("title", p.Title).Eq("author_id", p.AuthorID).Build()
	return db.Insert("page", cond)
}

func GetPage(db types.Conn, id int64) (*Page, error) {
	cond := dbhelper.Cond().Eq("id", id).Build()
	rows, err := db.Query("page", cond)
	if err != nil {
		return nil, err
	}
	if rows.Count() == 0 {
		return nil, errors.New("page not found")
	}
	data := rows.All()[0]
	p := &Page{
		ID:       data["id"].(int64),
		Title:    data["title"].(string),
		AuthorID: data["author_id"].(int64),
	}
	return p, nil
}

func UpdatePage(db types.Conn, id int64, updates *Page) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	upd := dbhelper.Cond().Eq("title", updates.Title).Eq("author_id", updates.AuthorID).Build()
	_, err := db.Update("page", cond, upd)
	return err
}

func DeletePage(db types.Conn, id int64) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	_, err := db.Delete("page", cond)
	return err
}

// UserInternal CRUD (shares table with User)

func CreateUserInternal(db types.Conn, u *UserInternal) (int64, error) {
	groupsJSON, _ := json.Marshal(u.Groups)
	cond := dbhelper.Cond().Eq("username", u.Username).Eq("email", u.Email).Eq("openid", u.OpenID).Eq("password_hash", u.PasswordHash).Eq("groups", string(groupsJSON)).Build()
	return db.Insert("user", cond)
}

func GetUserInternal(db types.Conn, id int64) (*UserInternal, error) {
	cond := dbhelper.Cond().Eq("id", id).Build()
	rows, err := db.Query("user", cond)
	if err != nil {
		return nil, err
	}
	if rows.Count() == 0 {
		return nil, errors.New("user not found")
	}
	data := rows.All()[0]
	u := &UserInternal{
		ID:           data["id"].(int64),
		Username:     data["username"].(string),
		Email:        data["email"].(string),
		OpenID:       data["openid"].(string),
		PasswordHash: data["password_hash"].(string),
	}
	return u, nil
}

func UpdateUserInternal(db types.Conn, id int64, updates *UserInternal) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	upd := dbhelper.Cond().Eq("username", updates.Username).Eq("email", updates.Email).Eq("openid", updates.OpenID).Eq("password_hash", updates.PasswordHash).Build()
	_, err := db.Update("user", cond, upd)
	return err
}

func DeleteUserInternal(db types.Conn, id int64) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	_, err := db.Delete("user", cond)
	return err
}

// Sidebar CRUD

func CreateSidebar(db types.Conn, s *Sidebar) (int64, error) {
	cond := dbhelper.Cond().Eq("name", s.Name).Eq("description", s.Description).Build()
	id, err := db.Insert("sidebar", cond)
	if err != nil {
		return 0, err
	}
	// Insert items
	for _, item := range s.Items {
		item.ParentID = id
		_, err := CreateSidebarItem(db, &item)
		if err != nil {
			return 0, err
		}
	}
	return id, nil
}

func GetSidebar(db types.Conn, id int64) (*Sidebar, error) {
	cond := dbhelper.Cond().Eq("id", id).Build()
	rows, err := db.Query("sidebar", cond)
	if err != nil {
		return nil, err
	}
	if rows.Count() == 0 {
		return nil, errors.New("sidebar not found")
	}
	data := rows.All()[0]
	s := &Sidebar{
		ID:          data["id"].(int64),
		Name:        data["name"].(string),
		Description: data["description"].(string),
	}
	// Get items
	itemCond := dbhelper.Cond().Eq("parent_id", id).Build()
	itemRows, err := db.Query("sidebar_item", itemCond)
	if err != nil {
		return nil, err
	}
	for _, itemData := range itemRows.All() {
		item := SidebarItem{
			ID:       itemData["id"].(int64),
			ParentID: itemData["parent_id"].(int64),
			Name:     itemData["name"].(string),
			Icon:     itemData["icon"].(string),
			URL:      itemData["url"].(string),
			Order:    int(itemData["order_num"].(int64)),
		}
		s.Items = append(s.Items, item)
	}
	return s, nil
}

func UpdateSidebar(db types.Conn, id int64, updates *Sidebar) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	upd := dbhelper.Cond().Eq("name", updates.Name).Eq("description", updates.Description).Build()
	_, err := db.Update("sidebar", cond, upd)
	if err != nil {
		return err
	}
	// Delete old items and insert new ones
	itemCond := dbhelper.Cond().Eq("parent_id", id).Build()
	_, err = db.Delete("sidebar_item", itemCond)
	if err != nil {
		return err
	}
	for _, item := range updates.Items {
		item.ParentID = id
		_, err := CreateSidebarItem(db, &item)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteSidebar(db types.Conn, id int64) error {
	// Delete items first
	itemCond := dbhelper.Cond().Eq("parent_id", id).Build()
	_, err := db.Delete("sidebar_item", itemCond)
	if err != nil {
		return err
	}
	cond := dbhelper.Cond().Eq("id", id).Build()
	_, err = db.Delete("sidebar", cond)
	return err
}

// SidebarItem CRUD

func CreateSidebarItem(db types.Conn, si *SidebarItem) (int64, error) {
	cond := dbhelper.Cond().Eq("parent_id", si.ParentID).Eq("name", si.Name).Eq("icon", si.Icon).Eq("url", si.URL).Eq("order_num", si.Order).Build()
	return db.Insert("sidebar_item", cond)
}

func GetSidebarItem(db types.Conn, id int64) (*SidebarItem, error) {
	cond := dbhelper.Cond().Eq("id", id).Build()
	rows, err := db.Query("sidebar_item", cond)
	if err != nil {
		return nil, err
	}
	if rows.Count() == 0 {
		return nil, errors.New("sidebar item not found")
	}
	data := rows.All()[0]
	si := &SidebarItem{
		ID:       data["id"].(int64),
		ParentID: data["parent_id"].(int64),
		Name:     data["name"].(string),
		Icon:     data["icon"].(string),
		URL:      data["url"].(string),
		Order:    int(data["order_num"].(int64)),
	}
	return si, nil
}

func UpdateSidebarItem(db types.Conn, id int64, updates *SidebarItem) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	upd := dbhelper.Cond().Eq("parent_id", updates.ParentID).Eq("name", updates.Name).Eq("icon", updates.Icon).Eq("url", updates.URL).Eq("order_num", updates.Order).Build()
	_, err := db.Update("sidebar_item", cond, upd)
	return err
}

func DeleteSidebarItem(db types.Conn, id int64) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	_, err := db.Delete("sidebar_item", cond)
	return err
}

// ContentList CRUD

func CreateContentList(db types.Conn, cl *ContentList) (int64, error) {
	itemsJson, _ := json.Marshal(cl.Items)
	cond := dbhelper.Cond().Eq("type", cl.Type).Eq("title", cl.Title).Eq("items", string(itemsJson)).Eq("creator_id", cl.CreatorID).Eq("project_id", cl.ProjectID).Build()
	return db.Insert("content_list", cond)
}

func GetContentList(db types.Conn, id int64) (*ContentList, error) {
	cond := dbhelper.Cond().Eq("id", id).Build()
	rows, err := db.Query("content_list", cond)
	if err != nil {
		return nil, err
	}
	if rows.Count() == 0 {
		return nil, errors.New("content list not found")
	}
	data := rows.All()[0]
	cl := &ContentList{
		ID:        data["id"].(int64),
		Type:      data["type"].(string),
		Title:     data["title"].(string),
		CreatorID: data["creator_id"].(int64),
		ProjectID: data["project_id"].(int64),
	}
	itemsJson := data["items"].(string)
	json.Unmarshal([]byte(itemsJson), &cl.Items)
	return cl, nil
}

func UpdateContentList(db types.Conn, id int64, updates *ContentList) error {
	itemsJson, _ := json.Marshal(updates.Items)
	cond := dbhelper.Cond().Eq("id", id).Build()
	upd := dbhelper.Cond().Eq("type", updates.Type).Eq("title", updates.Title).Eq("items", string(itemsJson)).Eq("creator_id", updates.CreatorID).Eq("project_id", updates.ProjectID).Build()
	_, err := db.Update("content_list", cond, upd)
	return err
}

func DeleteContentList(db types.Conn, id int64) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	_, err := db.Delete("content_list", cond)
	return err
}

// ContentEntry CRUD

func CreateContentEntry(db types.Conn, ce *ContentEntry) (int64, error) {
	cond := dbhelper.Cond().Eq("type", ce.Type).Eq("title", ce.Title).Eq("content", ce.Content).Eq("creator_id", ce.CreatorID).Eq("project_id", ce.ProjectID).Build()
	return db.Insert("content_entry", cond)
}

func GetContentEntry(db types.Conn, id int64) (*ContentEntry, error) {
	cond := dbhelper.Cond().Eq("id", id).Build()
	rows, err := db.Query("content_entry", cond)
	if err != nil {
		return nil, err
	}
	if rows.Count() == 0 {
		return nil, errors.New("content entry not found")
	}
	data := rows.All()[0]
	ce := &ContentEntry{
		ID:        data["id"].(int64),
		Type:      data["type"].(string),
		Title:     data["title"].(string),
		Content:   data["content"].(string),
		CreatorID: data["creator_id"].(int64),
		ProjectID: data["project_id"].(int64),
	}
	return ce, nil
}

func UpdateContentEntry(db types.Conn, id int64, updates *ContentEntry) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	upd := dbhelper.Cond().Eq("type", updates.Type).Eq("title", updates.Title).Eq("content", updates.Content).Eq("creator_id", updates.CreatorID).Eq("project_id", updates.ProjectID).Build()
	_, err := db.Update("content_entry", cond, upd)
	return err
}

func DeleteContentEntry(db types.Conn, id int64) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	_, err := db.Delete("content_entry", cond)
	return err
}

// DetailPermission CRUD

func CreateDetailPermission(db types.Conn, dp *DetailPermission) (int64, error) {
	contentIDsJson, _ := json.Marshal(dp.ContentIDs)
	cond := dbhelper.Cond().Eq("user_id", dp.UserID).Eq("content_type", dp.ContentType).Eq("content_ids", string(contentIDsJson)).Eq("action", dp.Action).Build()
	return db.Insert("detail_permission", cond)
}

func GetDetailPermission(db types.Conn, id int64) (*DetailPermission, error) {
	cond := dbhelper.Cond().Eq("id", id).Build()
	rows, err := db.Query("detail_permission", cond)
	if err != nil {
		return nil, err
	}
	if rows.Count() == 0 {
		return nil, errors.New("detail permission not found")
	}
	data := rows.All()[0]
	dp := &DetailPermission{
		ID:          data["id"].(int64),
		UserID:      data["user_id"].(int64),
		ContentType: data["content_type"].(string),
		Action:      data["action"].(string),
	}
	contentIDsJson := data["content_ids"].(string)
	json.Unmarshal([]byte(contentIDsJson), &dp.ContentIDs)
	return dp, nil
}

func UpdateDetailPermission(db types.Conn, id int64, updates *DetailPermission) error {
	contentIDsJson, _ := json.Marshal(updates.ContentIDs)
	cond := dbhelper.Cond().Eq("id", id).Build()
	upd := dbhelper.Cond().Eq("user_id", updates.UserID).Eq("content_type", updates.ContentType).Eq("content_ids", string(contentIDsJson)).Eq("action", updates.Action).Build()
	_, err := db.Update("detail_permission", cond, upd)
	return err
}

func DeleteDetailPermission(db types.Conn, id int64) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	_, err := db.Delete("detail_permission", cond)
	return err
}

// Permission CRUD

func CreatePermission(db types.Conn, p *Permission) (int64, error) {
	cond := dbhelper.Cond().Eq("name", p.Name).Eq("description", p.Description).Eq("content_type", p.ContentType).Eq("action", p.Action).Eq("detail", p.Detail).Build()
	return db.Insert("permission", cond)
}

func GetPermission(db types.Conn, id int64) (*Permission, error) {
	cond := dbhelper.Cond().Eq("id", id).Build()
	rows, err := db.Query("permission", cond)
	if err != nil {
		return nil, err
	}
	if rows.Count() == 0 {
		return nil, errors.New("permission not found")
	}
	data := rows.All()[0]
	p := &Permission{
		ID:          data["id"].(int64),
		Name:        data["name"].(string),
		Description: data["description"].(string),
		ContentType: data["content_type"].(string),
		Action:      data["action"].(string),
		Detail:      data["detail"].(int64),
	}
	return p, nil
}

func UpdatePermission(db types.Conn, id int64, updates *Permission) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	upd := dbhelper.Cond().Eq("name", updates.Name).Eq("description", updates.Description).Eq("content_type", updates.ContentType).Eq("action", updates.Action).Eq("detail", updates.Detail).Build()
	_, err := db.Update("permission", cond, upd)
	return err
}

func DeletePermission(db types.Conn, id int64) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	_, err := db.Delete("permission", cond)
	return err
}

// Role CRUD

func CreateRole(db types.Conn, r *Role) (int64, error) {
	permissionsJson, _ := json.Marshal(r.Permissions)
	cond := dbhelper.Cond().Eq("name", r.Name).Eq("description", r.Description).Eq("permissions", string(permissionsJson)).Build()
	return db.Insert("role", cond)
}

func GetRole(db types.Conn, id int64) (*Role, error) {
	cond := dbhelper.Cond().Eq("id", id).Build()
	rows, err := db.Query("role", cond)
	if err != nil {
		return nil, err
	}
	if rows.Count() == 0 {
		return nil, errors.New("role not found")
	}
	data := rows.All()[0]
	r := &Role{
		ID:          data["id"].(int64),
		Name:        data["name"].(string),
		Description: data["description"].(string),
	}
	permissionsJson := data["permissions"].(string)
	json.Unmarshal([]byte(permissionsJson), &r.Permissions)
	return r, nil
}

func UpdateRole(db types.Conn, id int64, updates *Role) error {
	permissionsJson, _ := json.Marshal(updates.Permissions)
	cond := dbhelper.Cond().Eq("id", id).Build()
	upd := dbhelper.Cond().Eq("name", updates.Name).Eq("description", updates.Description).Eq("permissions", string(permissionsJson)).Build()
	_, err := db.Update("role", cond, upd)
	return err
}

func DeleteRole(db types.Conn, id int64) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	_, err := db.Delete("role", cond)
	return err
}

// Get all projects
func GetProjects(db types.Conn) ([]Project, error) {
	rows, err := db.Query("project", nil)
	if err != nil {
		return []Project{}, err
	}
	projects := make([]Project, 0)
	for _, data := range rows.All() {
		p := Project{
			ID:          data["id"].(int64),
			Name:        data["name"].(string),
			Description: data["description"].(string),
			CreatorID:   data["creator_id"].(int64),
		}
		projects = append(projects, p)
	}
	return projects, nil
}

// Get all users
func GetUsers(db types.Conn) ([]User, error) {
	rows, err := db.Query("user", nil)
	if err != nil {
		return []User{}, err
	}
	users := make([]User, 0)
	for _, data := range rows.All() {
		u := User{
			ID:       data["id"].(int64),
			Username: data["username"].(string),
			Email:    data["email"].(string),
		}
		if groupsData, ok := data["groups"].(string); ok {
			json.Unmarshal([]byte(groupsData), &u.Groups)
		}
		users = append(users, u)
	}
	return users, nil
}

// Get all content lists for a project
func GetContentListsByProject(db types.Conn, projectID int64) ([]ContentList, error) {
	cond := dbhelper.Cond().Eq("project_id", projectID).Build()
	rows, err := db.Query("content_list", cond)
	if err != nil {
		return []ContentList{}, err
	}
	lists := make([]ContentList, 0)
	for _, data := range rows.All() {
		cl := ContentList{
			ID:        data["id"].(int64),
			Type:      data["type"].(string),
			Title:     data["title"].(string),
			CreatorID: data["creator_id"].(int64),
			ProjectID: data["project_id"].(int64),
		}
		if itemsJson, ok := data["items"].(string); ok && itemsJson != "" {
			json.Unmarshal([]byte(itemsJson), &cl.Items)
		} else {
			cl.Items = make([]int64, 0)
		}
		lists = append(lists, cl)
	}
	return lists, nil
}

// Get all detail permissions
func GetDetailPermissions(db types.Conn) ([]DetailPermission, error) {
	rows, err := db.Query("detail_permission", nil)
	if err != nil {
		return []DetailPermission{}, err
	}
	dps := make([]DetailPermission, 0)
	for _, data := range rows.All() {
		dp := DetailPermission{
			ID:          data["id"].(int64),
			UserID:      data["user_id"].(int64),
			ContentType: data["content_type"].(string),
			Action:      data["action"].(string),
		}
		if contentIDsJson, ok := data["content_ids"].(string); ok {
			json.Unmarshal([]byte(contentIDsJson), &dp.ContentIDs)
		}
		dps = append(dps, dp)
	}
	return dps, nil
}

// Get all permissions
func GetPermissions(db types.Conn) ([]Permission, error) {
	rows, err := db.Query("permission", nil)
	if err != nil {
		return []Permission{}, err
	}
	permissions := make([]Permission, 0)
	for _, data := range rows.All() {
		p := Permission{
			ID:          data["id"].(int64),
			Name:        data["name"].(string),
			Description: data["description"].(string),
			ContentType: data["content_type"].(string),
			Action:      data["action"].(string),
			Detail:      data["detail"].(int64),
		}
		permissions = append(permissions, p)
	}
	return permissions, nil
}

// Get all content entries
func GetContentEntries(db types.Conn) ([]ContentEntry, error) {
	rows, err := db.Query("content_entry", nil)
	if err != nil {
		return []ContentEntry{}, err
	}
	entries := make([]ContentEntry, 0)
	for _, data := range rows.All() {
		ce := ContentEntry{
			ID:        data["id"].(int64),
			Type:      data["type"].(string),
			Title:     data["title"].(string),
			Content:   data["content"].(string),
			CreatorID: data["creator_id"].(int64),
			ProjectID: data["project_id"].(int64),
		}
		entries = append(entries, ce)
	}
	return entries, nil
}

// ShareToken CRUD

func CreateShareToken(db types.Conn, st *ShareToken) (int64, error) {
	cond := dbhelper.Cond().
		Eq("token", st.Token).
		Eq("project_id", st.ProjectID).
		Eq("permission_level", st.PermissionLevel).
		Eq("created_at", st.CreatedAt).
		Eq("expires_at", st.ExpiresAt).
		Build()
	return db.Insert("share_token", cond)
}

func GetShareToken(db types.Conn, token string) (*ShareToken, error) {
	cond := dbhelper.Cond().Eq("token", token).Build()
	rows, err := db.Query("share_token", cond)
	if err != nil {
		return nil, err
	}
	if rows.Count() == 0 {
		return nil, errors.New("share token not found")
	}
	data := rows.All()[0]
	st := &ShareToken{
		ID:              data["id"].(int64),
		Token:           data["token"].(string),
		ProjectID:       data["project_id"].(int64),
		PermissionLevel: data["permission_level"].(string),
		CreatedAt:       data["created_at"].(int64),
		ExpiresAt:       data["expires_at"].(int64),
	}
	return st, nil
}

func DeleteShareToken(db types.Conn, id int64) error {
	cond := dbhelper.Cond().Eq("id", id).Build()
	_, err := db.Delete("share_token", cond)
	return err
}

func GetShareTokensByProjectID(db types.Conn, projectID int64) ([]ShareToken, error) {
	cond := dbhelper.Cond().Eq("project_id", projectID).Build()
	rows, err := db.Query("share_token", cond)
	if err != nil {
		return []ShareToken{}, err
	}
	tokens := make([]ShareToken, 0)
	for _, data := range rows.All() {
		st := ShareToken{
			ID:              data["id"].(int64),
			Token:           data["token"].(string),
			ProjectID:       data["project_id"].(int64),
			PermissionLevel: data["permission_level"].(string),
			CreatedAt:       data["created_at"].(int64),
			ExpiresAt:       data["expires_at"].(int64),
		}
		tokens = append(tokens, st)
	}
	return tokens, nil
}
