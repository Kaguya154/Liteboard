package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/Kaguya154/dbhelper"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/sessions"
	"github.com/joho/godotenv"
)

var (
	githubClientID     string
	githubClientSecret string
	githubRedirectURI  string
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		hlog.Error("Error loading .env file")
	}
	githubClientID = os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	githubRedirectURI = os.Getenv("GITHUB_REDIRECT_URI")
	if githubRedirectURI == "" {
		githubRedirectURI = "http://localhost:8080/auth/github/callback"
	}
	hlog.Info("githubClientID:", githubClientID)
}

func GitHubLoginHandler(ctx context.Context, c *app.RequestContext) {
	authURL := "https://github.com/login/oauth/authorize?client_id=" + githubClientID +
		"&redirect_uri=" + url.QueryEscape(githubRedirectURI) + "&scope=user:email"
	c.Redirect(302, []byte(authURL))
}

func GitHubCallbackHandler(ctx context.Context, c *app.RequestContext) {
	code := string(c.Query("code"))
	if code == "" {
		c.String(400, "No code provided")
		return
	}

	if githubClientID == "" || githubClientSecret == "" {
		c.String(500, "GitHub OAuth not configured")
		return
	}

	// Exchange code for access token
	data := url.Values{}
	data.Set("client_id", githubClientID)
	data.Set("client_secret", githubClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", githubRedirectURI)

	resp, err := http.PostForm("https://github.com/login/oauth/access_token", data)
	if err != nil {
		hlog.Error("Failed to post form:", err)
		c.String(500, "Failed to exchange code for token")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		c.String(500, "Failed to exchange code, status: "+strconv.Itoa(resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.String(500, "Failed to read token response")
		return
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		c.String(500, "Failed to parse token response")
		return
	}

	if errorMsg := values.Get("error"); errorMsg != "" {
		c.String(500, "OAuth error: "+values.Get("error_description"))
		return
	}

	accessToken := values.Get("access_token")
	if accessToken == "" {
		c.String(500, "No access token received")
		return
	}

	// Get user info from GitHub
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		c.String(500, "Failed to create user request")
		return
	}
	req.Header.Set("Authorization", "token "+accessToken)

	client := &http.Client{}
	resp2, err := client.Do(req)
	if err != nil {
		c.String(500, "Failed to get user info")
		return
	}
	defer resp2.Body.Close()

	userBody, err := io.ReadAll(resp2.Body)
	if err != nil {
		c.String(500, "Failed to read user response")
		return
	}

	var githubUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	err = json.Unmarshal(userBody, &githubUser)
	if err != nil {
		c.String(500, "Failed to parse user info")
		return
	}

	// Create or get user
	openid := strconv.Itoa(githubUser.ID)

	groups := []string{"user"} // 默认给所有用户 "user" 组权限
	// Configure permissions based on user
	if openid == "92249309" {
		groups = append(groups, "admin") // 特定用户额外添加 "admin" 组
	}
	userInternal, err := CreateUserInternalIfNotExist(openid, githubUser.Login, githubUser.Email, githubUser.AvatarURL, groups)
	if err != nil {
		hlog.Errorf("Failed to create or get user: %v", err)
		c.String(500, "Failed to create or get user")
		return
	}

	user := NewUserFromInternal(userInternal)
	hlog.Debugf("User logged in: ID=%d, Username=%s, Groups=%v", user.ID, user.Username, user.Groups)
	session := sessions.Default(c)
	session.Set("user", user)
	err = session.Save()
	if err != nil {
		hlog.Errorf("Failed to save session: %v", err)
		c.String(500, "Failed to save session")
		return
	}

	c.Redirect(302, []byte("/dashboard")) // Redirect to home
}

func CreateUserInternalIfNotExist(openid, username, email, avatarURL string, groups []string) (*UserInternal, error) {
	hlog.Debugf("CreateUserInternalIfNotExist: openid=%s, username=%s, groups=%v", openid, username, groups)

	u, err := GetUserInternalByOpenID(openid)
	if err == nil {
		// 用户已存在，更新组权限和头像
		hlog.Debugf("User exists with ID=%d, updating groups from %v to %v", u.ID, u.Groups, groups)
		u.Groups = groups
		u.AvatarURL = avatarURL

		// 更新数据库中的 groups 和 avatar_url 字段
		groupsJson, err := json.Marshal(groups)
		if err != nil {
			hlog.Errorf("Failed to marshal groups: %v", err)
			return nil, err
		}

		cond := dbhelper.Cond().Eq("id", u.ID).Build()
		upd := dbhelper.Cond().Eq("groups", string(groupsJson)).Eq("avatar_url", avatarURL).Build()
		_, err = db.Update("user", cond, upd)
		if err != nil {
			hlog.Errorf("Failed to update user groups: %v", err)
			return nil, err
		}
		hlog.Debugf("Successfully updated user groups and avatar for ID=%d", u.ID)
		return u, nil
	}

	hlog.Debugf("User not found, creating new user")
	groupsJson, err := json.Marshal(groups)
	if err != nil {
		return nil, err
	}

	ui := &UserInternal{
		Username:     username,
		Email:        email,
		OpenID:       openid,
		PasswordHash: "",
		Groups:       groups,
		AvatarURL:    avatarURL,
	}

	cond := dbhelper.Cond().
		Eq("username", username).
		Eq("email", email).
		Eq("openid", openid).
		Eq("password_hash", "").
		Eq("groups", string(groupsJson)).
		Eq("avatar_url", avatarURL).
		Build()

	id, err := db.Insert("user", cond)
	if err != nil {
		hlog.Errorf("Failed to create user: %v", err)
		return nil, err
	}
	ui.ID = id
	hlog.Debugf("Successfully created user with ID=%d", id)
	return ui, nil
}
