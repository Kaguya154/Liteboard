package api

import (
	"context"
	"liteboard/auth"
	"liteboard/internal"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/hertz-contrib/sessions"
)

// GetIDFromParam 从请求参数中获取ID
func GetIDFromParam(c *app.RequestContext) (int64, error) {
	idStr := c.Param("id")
	return strconv.ParseInt(idStr, 10, 64)
}

func RegisterProjectRoutes(r *route.RouterGroup) {
	r.GET("/projects", GetProjects)
	r.POST("/projects", CreateProject)
	r.GET("/projects/:id", auth.PermissionCheckMiddleware("project", "read", GetIDFromParam), GetProject)
	r.PUT("/projects/:id", auth.PermissionCheckMiddleware("project", "write", GetIDFromParam), UpdateProject)
	r.DELETE("/projects/:id", auth.PermissionCheckMiddleware("project", "admin", GetIDFromParam), DeleteProject)
}

// GetProjects @Summary Get all projects
// @Description Retrieve list of projects
// @Tags projects
// @Accept json
// @Produce json
// @Success 200 {array} internal.Project
// @Router /api/projects [get]
func GetProjects(ctx context.Context, c *app.RequestContext) {
	sess := sessions.Default(c)
	userID := sess.Get("user")
	if userID == nil {
		c.JSON(401, map[string]string{"error": "not logged in"})
		return
	}
	projects, err := internal.GetProjectsForUser(db, userID.(int64))
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(200, projects)
}

// CreateProject @Summary Create project
// @Description Create a new project
// @Tags projects
// @Accept json
// @Produce json
// @Param project body internal.Project true "Project"
// @Success 201 {object} internal.Project
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/projects [post]
func CreateProject(ctx context.Context, c *app.RequestContext) {
	sess := sessions.Default(c)
	userID := sess.Get("user")
	if userID == nil {
		c.JSON(401, map[string]string{"error": "not logged in"})
		return
	}
	var p internal.Project
	if err := c.BindJSON(&p); err != nil {
		c.JSON(400, map[string]string{"error": err.Error()})
		return
	}
	p.CreatorID = userID.(int64)
	id, err := internal.CreateProject(db, &p)
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}
	p.ID = id
	// Give admin permission to creator
	dp := internal.DetailPermission{
		UserID:      p.CreatorID,
		ContentType: "project",
		ContentIDs:  []int64{p.ID},
		Action:      "admin",
	}
	_, err = internal.CreateDetailPermission(db, &dp)
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}
	// Also give read permission
	dpRead := internal.DetailPermission{
		UserID:      p.CreatorID,
		ContentType: "project",
		ContentIDs:  []int64{p.ID},
		Action:      "read",
	}
	_, err = internal.CreateDetailPermission(db, &dpRead)
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(201, p)
}

// GetProject @Summary Get project by ID
// @Description Retrieve a project by ID
// @Tags projects
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} internal.Project
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/projects/{id} [get]
func GetProject(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid id"})
		return
	}
	p, err := internal.GetProject(db, id)
	if err != nil {
		c.JSON(404, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(200, p)
}

// UpdateProject @Summary Update project
// @Description Update an existing project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param project body internal.Project true "Project"
// @Success 200 {object} internal.Project
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/projects/{id} [put]
func UpdateProject(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid id"})
		return
	}
	var p internal.Project
	if err := c.BindJSON(&p); err != nil {
		c.JSON(400, map[string]string{"error": err.Error()})
		return
	}
	err = internal.UpdateProject(db, id, &p)
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(200, p)
}

// DeleteProject @Summary Delete project
// @Description Delete a project by ID
// @Tags projects
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/projects/{id} [delete]
func DeleteProject(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid id"})
		return
	}
	err = internal.DeleteProject(db, id)
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(200, map[string]string{"message": "deleted"})
}
