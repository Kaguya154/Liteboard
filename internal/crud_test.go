package internal

import (
	"testing"

	"github.com/Kaguya154/dbhelper"
	"github.com/Kaguya154/dbhelper/drivers/sqlite"
	"github.com/Kaguya154/dbhelper/types"
)

func setupDB(t *testing.T) types.Conn {
	db, err := dbhelper.Open(types.DBConfig{
		Driver: sqlite.DriverName,
		DSN:    ":memory:",
	})
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}

	// 创建表
	tables := []string{
		"CREATE TABLE project (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, description TEXT)",
		"CREATE TABLE user (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, email TEXT, openid TEXT, password_hash TEXT)",
		"CREATE TABLE page (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, author_id INTEGER)",
		"CREATE TABLE sidebar (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, description TEXT)",
		"CREATE TABLE sidebar_item (id INTEGER PRIMARY KEY AUTOINCREMENT, parent_id INTEGER, name TEXT, icon TEXT, url TEXT, order_num INTEGER)",
		"CREATE TABLE content_list (id INTEGER PRIMARY KEY AUTOINCREMENT, type TEXT, title TEXT, items TEXT)",
		"CREATE TABLE content_entry (id INTEGER PRIMARY KEY AUTOINCREMENT, type TEXT, title TEXT, content TEXT)",
		"CREATE TABLE detail_permission (id INTEGER PRIMARY KEY AUTOINCREMENT, content_type TEXT, content_ids TEXT, action TEXT)",
		"CREATE TABLE permission (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, description TEXT, content_type TEXT, action TEXT, detail INTEGER)",
		"CREATE TABLE role (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, description TEXT, permissions TEXT)",
	}

	for _, sql := range tables {
		cond := dbhelper.Cond().Raw(sql).Build()
		_, err := db.Exec(cond)
		if err != nil {
			t.Fatalf("建表失败: %v", err)
		}
	}

	return db
}

func TestProjectCRUD(t *testing.T) {
	db := setupDB(t)

	// Create
	p := &Project{Name: "Test Project", Description: "A test project"}
	id, err := CreateProject(db, p)
	if err != nil || id == 0 {
		t.Fatalf("创建项目失败: %v, id=%d", err, id)
	}

	// Read
	got, err := GetProject(db, id)
	if err != nil || got.Name != p.Name || got.Description != p.Description {
		t.Fatalf("获取项目失败: %v, got=%+v", err, got)
	}

	// Update
	updates := &Project{Name: "Updated Project", Description: "Updated description"}
	err = UpdateProject(db, id, updates)
	if err != nil {
		t.Fatalf("更新项目失败: %v", err)
	}

	got, err = GetProject(db, id)
	if err != nil || got.Name != updates.Name || got.Description != updates.Description {
		t.Fatalf("更新后获取项目失败: %v, got=%+v", err, got)
	}

	// Delete
	err = DeleteProject(db, id)
	if err != nil {
		t.Fatalf("删除项目失败: %v", err)
	}

	_, err = GetProject(db, id)
	if err == nil {
		t.Fatalf("删除后仍能获取项目")
	}
}

func TestUserCRUD(t *testing.T) {
	db := setupDB(t)

	// Create
	u := &User{Username: "testuser", Email: "test@example.com"}
	id, err := CreateUser(db, u)
	if err != nil || id == 0 {
		t.Fatalf("创建用户失败: %v, id=%d", err, id)
	}

	// Read
	got, err := GetUser(db, id)
	if err != nil || got.Username != u.Username || got.Email != u.Email {
		t.Fatalf("获取用户失败: %v, got=%+v", err, got)
	}

	// Update
	updates := &User{Username: "updateduser", Email: "updated@example.com"}
	err = UpdateUser(db, id, updates)
	if err != nil {
		t.Fatalf("更新用户失败: %v", err)
	}

	got, err = GetUser(db, id)
	if err != nil || got.Username != updates.Username || got.Email != updates.Email {
		t.Fatalf("更新后获取用户失败: %v, got=%+v", err, got)
	}

	// Delete
	err = DeleteUser(db, id)
	if err != nil {
		t.Fatalf("删除用户失败: %v", err)
	}

	_, err = GetUser(db, id)
	if err == nil {
		t.Fatalf("删除后仍能获取用户")
	}
}

func TestPageCRUD(t *testing.T) {
	db := setupDB(t)

	// Create
	p := &Page{Title: "Test Page", AuthorID: 1}
	id, err := CreatePage(db, p)
	if err != nil || id == 0 {
		t.Fatalf("创建页面失败: %v, id=%d", err, id)
	}

	// Read
	got, err := GetPage(db, id)
	if err != nil || got.Title != p.Title || got.AuthorID != p.AuthorID {
		t.Fatalf("获取页面失败: %v, got=%+v", err, got)
	}

	// Update
	updates := &Page{Title: "Updated Page", AuthorID: 2}
	err = UpdatePage(db, id, updates)
	if err != nil {
		t.Fatalf("更新页面失败: %v", err)
	}

	got, err = GetPage(db, id)
	if err != nil || got.Title != updates.Title || got.AuthorID != updates.AuthorID {
		t.Fatalf("更新后获取页面失败: %v, got=%+v", err, got)
	}

	// Delete
	err = DeletePage(db, id)
	if err != nil {
		t.Fatalf("删除页面失败: %v", err)
	}

	_, err = GetPage(db, id)
	if err == nil {
		t.Fatalf("删除后仍能获取页面")
	}
}

func TestUserInternalCRUD(t *testing.T) {
	db := setupDB(t)

	// Create
	ui := &UserInternal{Username: "testuser", Email: "test@example.com", OpenID: "openid123", PasswordHash: "hash123"}
	id, err := CreateUserInternal(db, ui)
	if err != nil || id == 0 {
		t.Fatalf("创建内部用户失败: %v, id=%d", err, id)
	}

	// Read
	got, err := GetUserInternal(db, id)
	if err != nil || got.Username != ui.Username || got.Email != ui.Email || got.OpenID != ui.OpenID || got.PasswordHash != ui.PasswordHash {
		t.Fatalf("获取内部用户失败: %v, got=%+v", err, got)
	}

	// Update
	updates := &UserInternal{Username: "updateduser", Email: "updated@example.com", OpenID: "updatedopenid", PasswordHash: "updatedhash"}
	err = UpdateUserInternal(db, id, updates)
	if err != nil {
		t.Fatalf("更新内部用户失败: %v", err)
	}

	got, err = GetUserInternal(db, id)
	if err != nil || got.Username != updates.Username || got.Email != updates.Email || got.OpenID != updates.OpenID || got.PasswordHash != updates.PasswordHash {
		t.Fatalf("更新后获取内部用户失败: %v, got=%+v", err, got)
	}

	// Delete
	err = DeleteUserInternal(db, id)
	if err != nil {
		t.Fatalf("删除内部用户失败: %v", err)
	}

	_, err = GetUserInternal(db, id)
	if err == nil {
		t.Fatalf("删除后仍能获取内部用户")
	}
}

func TestSidebarCRUD(t *testing.T) {
	db := setupDB(t)

	// Create
	s := &Sidebar{Name: "Test Sidebar", Items: []SidebarItem{
		{Name: "Item1", Icon: "icon1", URL: "/url1", Order: 1},
		{Name: "Item2", Icon: "icon2", URL: "/url2", Order: 2},
	}}
	id, err := CreateSidebar(db, s)
	if err != nil || id == 0 {
		t.Fatalf("创建侧边栏失败: %v, id=%d", err, id)
	}

	// Read
	got, err := GetSidebar(db, id)
	if err != nil || got.Name != s.Name || len(got.Items) != len(s.Items) {
		t.Fatalf("获取侧边栏失败: %v, got=%+v", err, got)
	}

	// Update
	updates := &Sidebar{Name: "Updated Sidebar", Items: []SidebarItem{
		{Name: "Updated Item1", Icon: "uicon1", URL: "/uurl1", Order: 1},
	}}
	err = UpdateSidebar(db, id, updates)
	if err != nil {
		t.Fatalf("更新侧边栏失败: %v", err)
	}

	got, err = GetSidebar(db, id)
	if err != nil || got.Name != updates.Name || len(got.Items) != len(updates.Items) || got.Items[0].Name != updates.Items[0].Name {
		t.Fatalf("更新后获取侧边栏失败: %v, got=%+v", err, got)
	}

	// Delete
	err = DeleteSidebar(db, id)
	if err != nil {
		t.Fatalf("删除侧边栏失败: %v", err)
	}

	_, err = GetSidebar(db, id)
	if err == nil {
		t.Fatalf("删除后仍能获取侧边栏")
	}
}

func TestSidebarItemCRUD(t *testing.T) {
	db := setupDB(t)

	// Create
	si := &SidebarItem{ParentID: 1, Name: "Test Item", Icon: "icon", URL: "/url", Order: 1}
	id, err := CreateSidebarItem(db, si)
	if err != nil || id == 0 {
		t.Fatalf("创建侧边栏项失败: %v, id=%d", err, id)
	}

	// Read
	got, err := GetSidebarItem(db, id)
	if err != nil || got.Name != si.Name || got.Icon != si.Icon || got.URL != si.URL || got.Order != si.Order {
		t.Fatalf("获取侧边栏项失败: %v, got=%+v", err, got)
	}

	// Update
	updates := &SidebarItem{ParentID: 1, Name: "Updated Item", Icon: "uicon", URL: "/uurl", Order: 2}
	err = UpdateSidebarItem(db, id, updates)
	if err != nil {
		t.Fatalf("更新侧边栏项失败: %v", err)
	}

	got, err = GetSidebarItem(db, id)
	if err != nil || got.Name != updates.Name || got.Icon != updates.Icon || got.URL != updates.URL || got.Order != updates.Order {
		t.Fatalf("更新后获取侧边栏项失败: %v, got=%+v", err, got)
	}

	// Delete
	err = DeleteSidebarItem(db, id)
	if err != nil {
		t.Fatalf("删除侧边栏项失败: %v", err)
	}

	_, err = GetSidebarItem(db, id)
	if err == nil {
		t.Fatalf("删除后仍能获取侧边栏项")
	}
}

func TestContentListCRUD(t *testing.T) {
	db := setupDB(t)

	// Create
	cl := &ContentList{Type: "list", Title: "Test List", Items: []ContentItem{
		{Entry: &ContentEntry{Type: "entry", Title: "E1", Content: "Content1"}},
	}}
	id, err := CreateContentList(db, cl)
	if err != nil || id == 0 {
		t.Fatalf("创建内容列表失败: %v, id=%d", err, id)
	}

	// Read
	got, err := GetContentList(db, id)
	if err != nil || got.Type != cl.Type || got.Title != cl.Title || len(got.Items) != len(cl.Items) {
		t.Fatalf("获取内容列表失败: %v, got=%+v", err, got)
	}

	// Update
	updates := &ContentList{Type: "list", Title: "Updated List", Items: []ContentItem{
		{Entry: &ContentEntry{Type: "entry", Title: "UE1", Content: "UContent1"}},
	}}
	err = UpdateContentList(db, id, updates)
	if err != nil {
		t.Fatalf("更新内容列表失败: %v", err)
	}

	got, err = GetContentList(db, id)
	if err != nil || got.Type != updates.Type || got.Title != updates.Title || len(got.Items) != len(updates.Items) {
		t.Fatalf("更新后获取内容列表失败: %v, got=%+v", err, got)
	}

	// Delete
	err = DeleteContentList(db, id)
	if err != nil {
		t.Fatalf("删除内容列表失败: %v", err)
	}

	_, err = GetContentList(db, id)
	if err == nil {
		t.Fatalf("删除后仍能获取内容列表")
	}
}

func TestContentEntryCRUD(t *testing.T) {
	db := setupDB(t)

	// Create
	ce := &ContentEntry{Type: "entry", Title: "Test Entry", Content: "Test Content"}
	id, err := CreateContentEntry(db, ce)
	if err != nil || id == 0 {
		t.Fatalf("创建内容条目失败: %v, id=%d", err, id)
	}

	// Read
	got, err := GetContentEntry(db, id)
	if err != nil || got.Type != ce.Type || got.Title != ce.Title || got.Content != ce.Content {
		t.Fatalf("获取内容条目失败: %v, got=%+v", err, got)
	}

	// Update
	updates := &ContentEntry{Type: "entry", Title: "Updated Entry", Content: "Updated Content"}
	err = UpdateContentEntry(db, id, updates)
	if err != nil {
		t.Fatalf("更新内容条目失败: %v", err)
	}

	got, err = GetContentEntry(db, id)
	if err != nil || got.Type != updates.Type || got.Title != updates.Title || got.Content != updates.Content {
		t.Fatalf("更新后获取内容条目失败: %v, got=%+v", err, got)
	}

	// Delete
	err = DeleteContentEntry(db, id)
	if err != nil {
		t.Fatalf("删除内容条目失败: %v", err)
	}

	_, err = GetContentEntry(db, id)
	if err == nil {
		t.Fatalf("删除后仍能获取内容条目")
	}
}
