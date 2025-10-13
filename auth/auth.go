package auth

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"liteboard/internal"

	"github.com/Kaguya154/dbhelper"
	"github.com/Kaguya154/dbhelper/types"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/sessions"
)

var db types.Conn

func SetDB(conn types.Conn) {
	db = conn
}

// 登录保护中间件
func LoginRequired() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		sess := sessions.Default(c)
		if sess.Get("user") == nil {
			c.Redirect(302, []byte("/auth/login"))
			c.Abort()
			return
		}
		hlog.Debug("User is logged in")
		c.Next(ctx)
	}
}

// 权限检查中间件
func PermissionMiddleware(requiredPermissions ...string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		sess := sessions.Default(c)
		userVal := sess.Get("user")

		if userVal == nil {
			c.String(403, "未登录")
			c.Abort()
			return
		}

		user, ok := userVal.(*User)
		if !ok {
			c.String(500, "无效的用户会话")
			c.Abort()
			return
		}

		// 从数据库查询用户的 groups
		userInternal, err := GetUserInternalByID(user.ID)
		if err != nil {
			hlog.Debug(err)
			c.String(500, "无法获取用户权限")
			c.Abort()
			return
		}

		hasPermission := false
		for _, group := range userInternal.Groups {
			for _, req := range requiredPermissions {
				if group == req {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			c.String(403, "没有权限")
			c.Abort()
			return
		}

		hlog.Debug("Permission check passed for user:", user.Username)
		c.Next(ctx)
	}
}

func GetUserFromSession(c *app.RequestContext) *User {
	sess := sessions.Default(c)
	user, ok := sess.Get("user").(*User)
	if !ok {
		return nil
	}
	return user
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

func NewUser(userinfo map[string]interface{}) *User {
	idStr := ""
	if v, ok := userinfo["sub"].(string); ok {
		idStr = v
	} else if v, ok := userinfo["id"].(string); ok {
		idStr = v
	}
	id, _ := strconv.ParseInt(idStr, 10, 64)
	username := "未知用户"
	if v, ok := userinfo["name"].(string); ok {
		username = v
	}
	email := ""
	if v, ok := userinfo["email"].(string); ok {
		email = v
	}
	groups := []string{}
	if v, ok := userinfo["groups"].([]interface{}); ok {
		for _, g := range v {
			if gs, ok := g.(string); ok {
				groups = append(groups, gs)
			}
		}
	} else if v, ok := userinfo["groups"].([]string); ok {
		groups = v
	}
	return &User{ID: id, Username: username, Email: email, Groups: groups}
}

func NewUserFromInternal(userInternal *UserInternal) *User {
	return &User{
		ID:       userInternal.ID,
		Username: userInternal.Username,
		Email:    userInternal.Email,
		Groups:   userInternal.Groups,
	}
}

func GetUserInternalByOpenID(openid string) (*UserInternal, error) {
	cond := dbhelper.Cond().Eq("openid", openid).Build()
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

func GetUserInternalByID(id int64) (*UserInternal, error) {
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
		Groups:       []string{},
	}
	if data["groups"] != nil && data["groups"].(string) != "" {
		err = json.Unmarshal([]byte(data["groups"].(string)), &u.Groups)
		if err != nil {
			return nil, err
		}
	}
	return u, nil
}

func CreateUserInternal(ui *UserInternal) (int64, error) {
	cond := dbhelper.Cond().Eq("username", ui.Username).Eq("email", ui.Email).Eq("openid", ui.OpenID).Eq("password_hash", ui.PasswordHash).Build()
	return db.Insert("user", cond)
}

// 通用的权限检查中间件
// contentType: 内容类型，如 "project", "content_list"
// action: 操作类型，如 "read", "write", "admin"
// getID: 从请求中获取内容ID的函数，如果为nil则不检查ID（适用于列表操作）
func PermissionCheckMiddleware(contentType string, action string, getID func(*app.RequestContext) (int64, error)) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		sess := sessions.Default(c)
		userVal := sess.Get("user")
		if userVal == nil {
			c.JSON(401, map[string]string{"error": "not logged in"})
			c.Abort()
			return
		}

		user, ok := userVal.(*User)
		if !ok {
			c.JSON(500, map[string]string{"error": "invalid user session"})
			c.Abort()
			return
		}

		if getID != nil {
			id, err := getID(c)
			if err != nil {
				c.JSON(400, map[string]string{"error": "invalid id"})
				c.Abort()
				return
			}
			hasPerm, err := internal.HasPermission(db, user.ID, contentType, id, action)
			if err != nil {
				c.JSON(500, map[string]string{"error": err.Error()})
				c.Abort()
				return
			}
			if !hasPerm {
				c.JSON(403, map[string]string{"error": "forbidden"})
				c.Abort()
				return
			}
		}
		// 如果getID为nil，跳过ID检查（适用于列表操作或创建操作）
		c.Next(ctx)
	}
}
