package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"liteboard/auth"
	"liteboard/internal"
	"strconv"
	"time"

	"github.com/Kaguya154/dbhelper"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/hertz-contrib/sessions"
)

func RegisterShareRoutes(r *route.RouterGroup) {
	// Project permission management
	r.POST("/projects/:id/permissions", auth.PermissionCheckMiddleware("project", "admin", GetIDFromParam), AddProjectPermission)
	r.GET("/projects/:id/permissions", auth.PermissionCheckMiddleware("project", "read", GetIDFromParam), GetProjectPermissions)
	r.DELETE("/projects/:id/permissions/:userId", auth.PermissionCheckMiddleware("project", "admin", GetIDFromParam), RemoveProjectPermission)

	// Share token management
	r.POST("/projects/:id/share", auth.PermissionCheckMiddleware("project", "admin", GetIDFromParam), GenerateShareToken)
	r.GET("/projects/:id/shares", auth.PermissionCheckMiddleware("project", "admin", GetIDFromParam), GetProjectShareTokens)
	r.DELETE("/shares/:id", auth.LoginRequired(), DeleteShareToken)

	// Public route for joining via share token
	r.POST("/share/:token/join", JoinViaShareToken)
}

// AddProjectPermission @Summary Add permission to project
// @Description Add a user permission to a project
// @Tags share
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param permission body internal.ProjectPermission true "Permission"
// @Success 200 {object} internal.SuccessResponse
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/projects/{id}/permissions [post]
func AddProjectPermission(ctx context.Context, c *app.RequestContext) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid project id"))
		return
	}

	var req struct {
		UserID          int64  `json:"user_id"`
		PermissionLevel string `json:"permission_level"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, internal.NewErrorResponse(err.Error()))
		return
	}

	// Validate permission level
	if req.PermissionLevel != "read" && req.PermissionLevel != "write" && req.PermissionLevel != "admin" {
		c.JSON(400, internal.NewErrorResponse("invalid permission level"))
		return
	}

	// Check if permission already exists
	cond := dbhelper.Cond().Eq("user_id", req.UserID).Eq("content_type", "project").Build()
	rows, err := db.Query("detail_permission", cond)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}

	// Update or create permission
	for _, data := range rows.All() {
		contentIDsJson := data["content_ids"].(string)
		var contentIDs []int64
		json.Unmarshal([]byte(contentIDsJson), &contentIDs)

		// Check if this project is in the list
		for i, id := range contentIDs {
			if id == projectID {
				// Update existing permission if needed
				if data["action"].(string) != req.PermissionLevel {
					// Remove from this list
					contentIDs = append(contentIDs[:i], contentIDs[i+1:]...)
					newContentIDsJson, _ := json.Marshal(contentIDs)

					updateCond := dbhelper.Cond().Eq("id", data["id"].(int64)).Build()
					upd := dbhelper.Cond().Eq("content_ids", string(newContentIDsJson)).Build()
					db.Update("detail_permission", updateCond, upd)
				}
				break
			}
		}
	}

	// Create or update the permission entry for the new level
	cond = dbhelper.Cond().Eq("user_id", req.UserID).Eq("content_type", "project").Eq("action", req.PermissionLevel).Build()
	rows, err = db.Query("detail_permission", cond)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}

	if rows.Count() == 0 {
		// Create new permission entry
		dp := internal.DetailPermission{
			UserID:      req.UserID,
			ContentType: "project",
			ContentIDs:  []int64{projectID},
			Action:      req.PermissionLevel,
		}
		_, err = internal.CreateDetailPermission(db, &dp)
		if err != nil {
			c.JSON(500, internal.NewErrorResponse(err.Error()))
			return
		}
	} else {
		// Add to existing permission entry
		data := rows.All()[0]
		contentIDsJson := data["content_ids"].(string)
		var contentIDs []int64
		json.Unmarshal([]byte(contentIDsJson), &contentIDs)

		// Check if already in list
		alreadyExists := false
		for _, id := range contentIDs {
			if id == projectID {
				alreadyExists = true
				break
			}
		}

		if !alreadyExists {
			contentIDs = append(contentIDs, projectID)
			newContentIDsJson, _ := json.Marshal(contentIDs)

			updateCond := dbhelper.Cond().Eq("id", data["id"].(int64)).Build()
			upd := dbhelper.Cond().Eq("content_ids", string(newContentIDsJson)).Build()
			_, err = db.Update("detail_permission", updateCond, upd)
			if err != nil {
				c.JSON(500, internal.NewErrorResponse(err.Error()))
				return
			}
		}
	}

	c.JSON(200, internal.NewSuccessResponse("permission added"))
}

// GetProjectPermissions @Summary Get project permissions
// @Description Get all user permissions for a project
// @Tags share
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {array} internal.ProjectPermission
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/projects/{id}/permissions [get]
func GetProjectPermissions(ctx context.Context, c *app.RequestContext) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid project id"))
		return
	}

	// Get all permissions for this project
	cond := dbhelper.Cond().Eq("content_type", "project").Build()
	rows, err := db.Query("detail_permission", cond)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}

	permMap := make(map[int64]string) // userID -> permission level

	for _, data := range rows.All() {
		contentIDsJson := data["content_ids"].(string)
		var contentIDs []int64
		json.Unmarshal([]byte(contentIDsJson), &contentIDs)

		for _, id := range contentIDs {
			if id == projectID {
				userID := data["user_id"].(int64)
				action := data["action"].(string)

				// Keep the highest permission level
				if existing, ok := permMap[userID]; ok {
					if internal.GetPermissionLevel(action) > internal.GetPermissionLevel(existing) {
						permMap[userID] = action
					}
				} else {
					permMap[userID] = action
				}
			}
		}
	}

	// Get user details
	permissions := make([]internal.ProjectPermission, 0)
	for userID, permLevel := range permMap {
		userInternal, err := auth.GetUserInternalByID(userID)
		if err != nil {
			continue
		}
		permissions = append(permissions, internal.ProjectPermission{
			UserID:          userID,
			Username:        userInternal.Username,
			Email:           userInternal.Email,
			PermissionLevel: permLevel,
		})
	}

	c.JSON(200, permissions)
}

// RemoveProjectPermission @Summary Remove permission from project
// @Description Remove a user's permission from a project
// @Tags share
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param userId path int true "User ID"
// @Success 200 {object} internal.SuccessResponse
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/projects/{id}/permissions/{userId} [delete]
func RemoveProjectPermission(ctx context.Context, c *app.RequestContext) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid project id"))
		return
	}

	userIDStr := c.Param("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid user id"))
		return
	}

	// Find and remove permissions
	cond := dbhelper.Cond().Eq("user_id", userID).Eq("content_type", "project").Build()
	rows, err := db.Query("detail_permission", cond)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}

	for _, data := range rows.All() {
		contentIDsJson := data["content_ids"].(string)
		var contentIDs []int64
		json.Unmarshal([]byte(contentIDsJson), &contentIDs)

		// Remove projectID from list
		newContentIDs := make([]int64, 0)
		for _, id := range contentIDs {
			if id != projectID {
				newContentIDs = append(newContentIDs, id)
			}
		}

		if len(newContentIDs) == 0 {
			// Delete the entire permission entry
			deleteCond := dbhelper.Cond().Eq("id", data["id"].(int64)).Build()
			db.Delete("detail_permission", deleteCond)
		} else {
			// Update with new list
			newContentIDsJson, _ := json.Marshal(newContentIDs)
			updateCond := dbhelper.Cond().Eq("id", data["id"].(int64)).Build()
			upd := dbhelper.Cond().Eq("content_ids", string(newContentIDsJson)).Build()
			db.Update("detail_permission", updateCond, upd)
		}
	}

	c.JSON(200, internal.NewSuccessResponse("permission removed"))
}

// GenerateShareToken @Summary Generate share token
// @Description Generate a share token for a project
// @Tags share
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param request body object{permission_level=string,expires_in_hours=int} true "Share request"
// @Success 200 {object} internal.ShareToken
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/projects/{id}/share [post]
func GenerateShareToken(ctx context.Context, c *app.RequestContext) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid project id"))
		return
	}

	var req struct {
		PermissionLevel string `json:"permission_level"`
		ExpiresInHours  int    `json:"expires_in_hours"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, internal.NewErrorResponse(err.Error()))
		return
	}

	// Validate permission level
	if req.PermissionLevel != "read" && req.PermissionLevel != "write" {
		c.JSON(400, internal.NewErrorResponse("invalid permission level"))
		return
	}

	// Default to 24 hours if not specified
	if req.ExpiresInHours == 0 {
		req.ExpiresInHours = 24
	}

	// Generate random token
	tokenBytes := make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse("failed to generate token"))
		return
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	now := time.Now().Unix()
	expiresAt := now + int64(req.ExpiresInHours*3600)

	st := &internal.ShareToken{
		Token:           token,
		ProjectID:       projectID,
		PermissionLevel: req.PermissionLevel,
		CreatedAt:       now,
		ExpiresAt:       expiresAt,
	}

	id, err := internal.CreateShareToken(db, st)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	st.ID = id

	c.JSON(200, st)
}

// GetProjectShareTokens @Summary Get project share tokens
// @Description Get all active share tokens for a project
// @Tags share
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {array} internal.ShareToken
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/projects/{id}/shares [get]
func GetProjectShareTokens(ctx context.Context, c *app.RequestContext) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid project id"))
		return
	}

	tokens, err := internal.GetShareTokensByProjectID(db, projectID)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}

	// Filter out expired tokens
	now := time.Now().Unix()
	activeTokens := make([]internal.ShareToken, 0)
	for _, token := range tokens {
		if token.ExpiresAt > now {
			activeTokens = append(activeTokens, token)
		}
	}

	c.JSON(200, activeTokens)
}

// DeleteShareToken @Summary Delete share token
// @Description Delete a share token
// @Tags share
// @Accept json
// @Produce json
// @Param id path int true "Share Token ID"
// @Success 200 {object} internal.SuccessResponse
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/shares/{id} [delete]
func DeleteShareToken(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}

	err = internal.DeleteShareToken(db, id)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, internal.NewSuccessResponse("share token deleted"))
}

// JoinViaShareToken @Summary Join project via share token
// @Description Join a project using a share token
// @Tags share
// @Accept json
// @Produce json
// @Param token path string true "Share Token"
// @Success 200 {object} internal.SuccessResponse
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 404 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/share/{token}/join [post]
func JoinViaShareToken(ctx context.Context, c *app.RequestContext) {
	token := c.Param("token")

	// Check if user is logged in
	sess := sessions.Default(c)
	userVal := sess.Get("user")
	if userVal == nil {
		c.JSON(401, internal.NewErrorResponse("not logged in"))
		return
	}

	user, ok := userVal.(*auth.User)
	if !ok {
		c.JSON(500, internal.NewErrorResponse("invalid user session"))
		return
	}

	// Get share token
	st, err := internal.GetShareToken(db, token)
	if err != nil {
		c.JSON(404, internal.NewErrorResponse("share token not found"))
		return
	}

	// Check if token is expired
	now := time.Now().Unix()
	if st.ExpiresAt < now {
		c.JSON(400, internal.NewErrorResponse("share token expired"))
		return
	}

	// Add permission to user
	req := struct {
		UserID          int64  `json:"user_id"`
		PermissionLevel string `json:"permission_level"`
	}{
		UserID:          user.ID,
		PermissionLevel: st.PermissionLevel,
	}

	// Check if permission already exists
	cond := dbhelper.Cond().Eq("user_id", req.UserID).Eq("content_type", "project").Eq("action", req.PermissionLevel).Build()
	rows, err := db.Query("detail_permission", cond)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}

	if rows.Count() == 0 {
		// Create new permission entry
		dp := internal.DetailPermission{
			UserID:      req.UserID,
			ContentType: "project",
			ContentIDs:  []int64{st.ProjectID},
			Action:      req.PermissionLevel,
		}
		_, err = internal.CreateDetailPermission(db, &dp)
		if err != nil {
			c.JSON(500, internal.NewErrorResponse(err.Error()))
			return
		}
	} else {
		// Add to existing permission entry
		data := rows.All()[0]
		contentIDsJson := data["content_ids"].(string)
		var contentIDs []int64
		json.Unmarshal([]byte(contentIDsJson), &contentIDs)

		// Check if already in list
		alreadyExists := false
		for _, id := range contentIDs {
			if id == st.ProjectID {
				alreadyExists = true
				break
			}
		}

		if !alreadyExists {
			contentIDs = append(contentIDs, st.ProjectID)
			newContentIDsJson, _ := json.Marshal(contentIDs)

			updateCond := dbhelper.Cond().Eq("id", data["id"].(int64)).Build()
			upd := dbhelper.Cond().Eq("content_ids", string(newContentIDsJson)).Build()
			_, err = db.Update("detail_permission", updateCond, upd)
			if err != nil {
				c.JSON(500, internal.NewErrorResponse(err.Error()))
				return
			}
		}
	}

	hlog.Infof("User %d joined project %d via share token with %s permission", user.ID, st.ProjectID, st.PermissionLevel)
	c.JSON(200, internal.NewSuccessResponseWithData("joined project", map[string]interface{}{
		"project_id": st.ProjectID,
	}))
}
