package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

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
		ID    int    `json:"id"`
		Login string `json:"login"`
		Email string `json:"email"`
	}
	err = json.Unmarshal(userBody, &githubUser)
	if err != nil {
		c.String(500, "Failed to parse user info")
		return
	}

	// Create or get user
	openid := strconv.Itoa(githubUser.ID)

	groups := []string{}
	// Configure permissions based on user
	if openid == "92249309" {
		groups = []string{"admin"}
	}
	userInternal, err := CreateUserInternalIfNotExist(openid, githubUser.Login, githubUser.Email, groups)
	if err != nil {
		c.String(500, "Failed to create or get user")
		return
	}

	user := NewUserFromInternal(userInternal)
	session := sessions.Default(c)
	session.Set("user", user)
	session.Save()

	c.Redirect(302, []byte("/dashboard")) // Redirect to home
}

func CreateUserInternalIfNotExist(openid, username, email string, groups []string) (*UserInternal, error) {
	u, err := GetUserInternalByOpenID(openid)
	if err == nil {
		return u, nil
	}

	ui := &UserInternal{
		Username:     username,
		Email:        email,
		OpenID:       openid,
		PasswordHash: "",
		Groups:       groups,
	}
	id, err := CreateUserInternal(ui)
	if err != nil {
		return nil, err
	}
	ui.ID = id
	return ui, nil
}
