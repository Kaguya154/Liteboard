package api

import (
	"context"
	"liteboard/auth"
	"liteboard/internal"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
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
// @Failure 401 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Router /api/projects [get]
func GetProjects(ctx context.Context, c *app.RequestContext) {
	hlog.Debug("GetProjects: Starting request")
	sess := sessions.Default(c)
	userVal := sess.Get("user")
	hlog.Debugf("GetProjects: userVal type=%T, value=%v", userVal, userVal)

	if userVal == nil {
		hlog.Debug("GetProjects: user not in session")
		c.JSON(401, internal.NewErrorResponse("not logged in"))
		return
	}
	user, ok := userVal.(*auth.User)
	if !ok {
		hlog.Errorf("GetProjects: invalid user session type, got %T", userVal)
		c.JSON(500, internal.NewErrorResponse("invalid user session"))
		return
	}
	hlog.Debugf("GetProjects: user authenticated, ID=%d, Username=%s", user.ID, user.Username)

	projects, err := internal.GetProjectsForUser(db, user.ID)
	if err != nil {
		hlog.Errorf("GetProjects: GetProjectsForUser failed, userID=%d, error=%v", user.ID, err)
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	hlog.Debugf("GetProjects: successfully retrieved %d projects for user %d", len(projects), user.ID)
	c.JSON(200, projects)
}

// CreateProject @Summary Create project
// @Description Create a new project
// @Tags projects
// @Accept json
// @Produce json
// @Param project body internal.Project true "Project"
// @Success 201 {object} internal.Project
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Router /api/projects [post]
func CreateProject(ctx context.Context, c *app.RequestContext) {
	hlog.Debug("CreateProject: Starting request")
	sess := sessions.Default(c)
	userVal := sess.Get("user")

	if userVal == nil {
		hlog.Debug("CreateProject: user not in session")
		c.JSON(401, internal.NewErrorResponse("not logged in"))
		return
	}
	user, ok := userVal.(*auth.User)
	if !ok {
		hlog.Errorf("CreateProject: invalid user session type, got %T", userVal)
		c.JSON(500, internal.NewErrorResponse("invalid user session"))
		return
	}
	hlog.Debugf("CreateProject: user authenticated, ID=%d, Username=%s", user.ID, user.Username)

	var p internal.Project
	if err := c.BindJSON(&p); err != nil {
		hlog.Errorf("CreateProject: BindJSON failed, error=%v", err)
		c.JSON(400, internal.NewErrorResponse(err.Error()))
		return
	}
	hlog.Debugf("CreateProject: project data bound, Name=%s", p.Name)

	p.CreatorID = user.ID
	id, err := internal.CreateProject(db, &p)
	if err != nil {
		hlog.Errorf("CreateProject: CreateProject failed, error=%v", err)
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	p.ID = id
	hlog.Debugf("CreateProject: project created with ID=%d", id)

	// Give admin permission to creator
	dp := internal.DetailPermission{
		UserID:      p.CreatorID,
		ContentType: "project",
		ContentIDs:  []int64{p.ID},
		Action:      "admin",
	}
	_, err = internal.CreateDetailPermission(db, &dp)
	if err != nil {
		hlog.Errorf("CreateProject: CreateDetailPermission(admin) failed, error=%v", err)
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	hlog.Debug("CreateProject: admin permission created")

	// Also give read permission
	dpRead := internal.DetailPermission{
		UserID:      p.CreatorID,
		ContentType: "project",
		ContentIDs:  []int64{p.ID},
		Action:      "read",
	}
	_, err = internal.CreateDetailPermission(db, &dpRead)
	if err != nil {
		hlog.Errorf("CreateProject: CreateDetailPermission(read) failed, error=%v", err)
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	hlog.Debug("CreateProject: read permission created")
	hlog.Debugf("CreateProject: successfully created project ID=%d", p.ID)
	c.JSON(201, p)
}

// GetProject @Summary Get project by ID
// @Description Retrieve a project by ID
// @Tags projects
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} internal.Project
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 404 {object} internal.ErrorResponse
// @Security Session
// @Router /api/projects/{id} [get]
func GetProject(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	p, err := internal.GetProject(db, id)
	if err != nil {
		c.JSON(404, internal.NewErrorResponse(err.Error()))
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
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/projects/{id} [put]
func UpdateProject(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	var p internal.Project
	if err := c.BindJSON(&p); err != nil {
		c.JSON(400, internal.NewErrorResponse(err.Error()))
		return
	}
	err = internal.UpdateProject(db, id, &p)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
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
// @Success 200 {object} internal.SuccessResponse
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/projects/{id} [delete]
func DeleteProject(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	err = internal.DeleteProject(db, id)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, internal.NewSuccessResponse("deleted"))
}
