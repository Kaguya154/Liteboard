package api

import (
	"context"
	"liteboard/auth"
	"liteboard/internal"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/hertz-contrib/sessions"
)

// Login @Summary Login page
// @Description Redirect to GitHub OAuth login
// @Tags auth
// @Accept json
// @Produce json
// @Success 302 {string} string "Redirect to GitHub"
// @Router /auth/login [get]
func Login(ctx context.Context, c *app.RequestContext) {
	// 显示登录页面，或重定向到 GitHub
	c.Redirect(302, []byte("/auth/github/login"))
}

// Logout @Summary Logout
// @Description Logout user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} internal.SuccessResponse "Logged out"
// @Failure 500 {object} internal.ErrorResponse
// @Router /auth/logout [post]
func Logout(ctx context.Context, c *app.RequestContext) {
	sess := sessions.Default(c)
	sess.Delete("user")
	if err := sess.Save(); err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, internal.NewSuccessResponse("logged out"))
}

// RegisterAuthRoutes 注册认证路由
func RegisterAuthRoutes(h *route.RouterGroup) {
	authGroup := h.Group("/auth")

	// @Summary GitHub OAuth login
	// @Description Initiate GitHub OAuth login
	// @Tags auth
	// @Accept json
	// @Produce json
	// @Success 302 {string} string "Redirect to GitHub"
	// @Router /auth/github/login [get]
	authGroup.GET("/github/login", auth.GitHubLoginHandler)

	// @Summary GitHub OAuth callback
	// @Description Handle GitHub OAuth callback
	// @Tags auth
	// @Accept json
	// @Produce json
	// @Success 302 {string} string "Redirect to home"
	// @Router /auth/github/callback [get]
	authGroup.GET("/github/callback", auth.GitHubCallbackHandler)

	authGroup.GET("/login", Login)
	authGroup.GET("/logout", Logout)
	authGroup.POST("/logout", Logout) // Keep POST for compatibility
}
