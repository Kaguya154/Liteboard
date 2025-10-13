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

func RegisterContentRoutes(r *route.RouterGroup) {
	r.GET("/content_lists", GetContentLists)
	r.POST("/content_lists", CreateContentList)
	r.GET("/content_lists/:id", auth.PermissionCheckMiddleware("content_list", "read", GetIDFromParam), GetContentList)
	r.PUT("/content_lists/:id", auth.PermissionCheckMiddleware("content_list", "write", GetIDFromParam), UpdateContentList)
	r.DELETE("/content_lists/:id", auth.PermissionCheckMiddleware("content_list", "admin", GetIDFromParam), DeleteContentList)

	r.GET("/content_entries", GetContentEntries)
	r.POST("/content_entries", CreateContentEntry)
	r.GET("/content_entries/:id", auth.PermissionCheckMiddleware("content_entry", "read", GetIDFromParam), GetContentEntry)
	r.PUT("/content_entries/:id", auth.PermissionCheckMiddleware("content_entry", "write", GetIDFromParam), UpdateContentEntry)
	r.DELETE("/content_entries/:id", auth.PermissionCheckMiddleware("content_entry", "admin", GetIDFromParam), DeleteContentEntry)
}

// GetContentLists @Summary Get all content lists
// @Description Retrieve list of content lists for a project
// @Tags content
// @Accept json
// @Produce json
// @Param projectid query int true "Project ID"
// @Success 200 {array} internal.ContentList
// @Failure 400 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Router /api/content_lists [get]
func GetContentLists(ctx context.Context, c *app.RequestContext) {
	projectIDStr := c.Query("projectid")
	if projectIDStr == "" {
		c.JSON(400, internal.NewErrorResponse("projectid is required"))
		return
	}
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid projectid"))
		return
	}
	lists, err := internal.GetContentListsByProject(db, projectID)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, lists)
}

// CreateContentList @Summary Create content list
// @Description Create a new content list
// @Tags content
// @Accept json
// @Produce json
// @Param contentList body internal.ContentList true "Content List"
// @Success 201 {object} internal.ContentList
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Router /api/content_lists [post]
func CreateContentList(ctx context.Context, c *app.RequestContext) {
	hlog.Debug("CreateContentList: Starting request")
	sess := sessions.Default(c)
	userVal := sess.Get("user")

	if userVal == nil {
		hlog.Debug("CreateContentList: user not in session")
		c.JSON(401, internal.NewErrorResponse("not logged in"))
		return
	}
	user, ok := userVal.(*auth.User)
	if !ok {
		hlog.Errorf("CreateContentList: invalid user session type, got %T", userVal)
		c.JSON(500, internal.NewErrorResponse("invalid user session"))
		return
	}

	var cl internal.ContentList
	if err := c.BindJSON(&cl); err != nil {
		hlog.Errorf("CreateContentList: BindJSON failed, error=%v", err)
		c.JSON(400, internal.NewErrorResponse(err.Error()))
		return
	}

	cl.CreatorID = user.ID
	id, err := internal.CreateContentList(db, &cl)
	if err != nil {
		hlog.Errorf("CreateContentList: CreateContentList failed, error=%v", err)
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	cl.ID = id

	// Give admin permission to creator
	dp := internal.DetailPermission{
		UserID:      cl.CreatorID,
		ContentType: "content_list",
		ContentIDs:  []int64{cl.ID},
		Action:      "admin",
	}
	_, err = internal.CreateDetailPermission(db, &dp)
	if err != nil {
		hlog.Errorf("CreateContentList: CreateDetailPermission(admin) failed, error=%v", err)
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}

	// Also give read permission
	dpRead := internal.DetailPermission{
		UserID:      cl.CreatorID,
		ContentType: "content_list",
		ContentIDs:  []int64{cl.ID},
		Action:      "read",
	}
	_, err = internal.CreateDetailPermission(db, &dpRead)
	if err != nil {
		hlog.Errorf("CreateContentList: CreateDetailPermission(read) failed, error=%v", err)
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}

	hlog.Debugf("CreateContentList: successfully created content list ID=%d", cl.ID)
	c.JSON(201, cl)
}

// GetContentList @Summary Get content list by ID
// @Description Retrieve a content list by ID
// @Tags content
// @Accept json
// @Produce json
// @Param id path int true "Content List ID"
// @Success 200 {object} internal.ContentList
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 404 {object} internal.ErrorResponse
// @Security Session
// @Router /api/content_lists/{id} [get]
func GetContentList(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	cl, err := internal.GetContentList(db, id)
	if err != nil {
		c.JSON(404, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, cl)
}

// UpdateContentList @Summary Update content list
// @Description Update an existing content list
// @Tags content
// @Accept json
// @Produce json
// @Param id path int true "Content List ID"
// @Param contentList body internal.ContentList true "Content List"
// @Success 200 {object} internal.ContentList
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/content_lists/{id} [put]
func UpdateContentList(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	var cl internal.ContentList
	if err := c.BindJSON(&cl); err != nil {
		c.JSON(400, internal.NewErrorResponse(err.Error()))
		return
	}
	err = internal.UpdateContentList(db, id, &cl)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, cl)
}

// DeleteContentList @Summary Delete content list
// @Description Delete a content list by ID
// @Tags content
// @Accept json
// @Produce json
// @Param id path int true "Content List ID"
// @Success 200 {object} internal.SuccessResponse
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/content_lists/{id} [delete]
func DeleteContentList(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	err = internal.DeleteContentList(db, id)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, internal.NewSuccessResponse("deleted"))
}

// GetContentEntries @Summary Get all content entries
// @Description Retrieve list of content entries. Each entry includes creator_id and project_id.
// @Tags content
// @Accept json
// @Produce json
// @Success 200 {array} internal.ContentEntry
// @Failure 500 {object} internal.ErrorResponse
// @Router /api/content_entries [get]
func GetContentEntries(ctx context.Context, c *app.RequestContext) {
	entries, err := internal.GetContentEntries(db)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, entries)
}

// CreateContentEntry @Summary Create content entry
// @Description Create a new content entry. The creator_id will be automatically set to the current user.
// @Tags content
// @Accept json
// @Produce json
// @Param contentEntry body internal.ContentEntry true "Content Entry (creator_id will be set automatically)"
// @Success 201 {object} internal.ContentEntry
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Router /api/content_entries [post]
func CreateContentEntry(ctx context.Context, c *app.RequestContext) {
	hlog.Debug("CreateContentEntry: Starting request")
	sess := sessions.Default(c)
	userVal := sess.Get("user")

	if userVal == nil {
		hlog.Debug("CreateContentEntry: user not in session")
		c.JSON(401, internal.NewErrorResponse("not logged in"))
		return
	}
	user, ok := userVal.(*auth.User)
	if !ok {
		hlog.Errorf("CreateContentEntry: invalid user session type, got %T", userVal)
		c.JSON(500, internal.NewErrorResponse("invalid user session"))
		return
	}

	var ce internal.ContentEntry
	if err := c.BindJSON(&ce); err != nil {
		hlog.Errorf("CreateContentEntry: BindJSON failed, error=%v", err)
		c.JSON(400, internal.NewErrorResponse(err.Error()))
		return
	}

	ce.CreatorID = user.ID
	id, err := internal.CreateContentEntry(db, &ce)
	if err != nil {
		hlog.Errorf("CreateContentEntry: CreateContentEntry failed, error=%v", err)
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	ce.ID = id

	// Give admin permission to creator
	dp := internal.DetailPermission{
		UserID:      ce.CreatorID,
		ContentType: "content_entry",
		ContentIDs:  []int64{ce.ID},
		Action:      "admin",
	}
	_, err = internal.CreateDetailPermission(db, &dp)
	if err != nil {
		hlog.Errorf("CreateContentEntry: CreateDetailPermission(admin) failed, error=%v", err)
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}

	// Also give read permission
	dpRead := internal.DetailPermission{
		UserID:      ce.CreatorID,
		ContentType: "content_entry",
		ContentIDs:  []int64{ce.ID},
		Action:      "read",
	}
	_, err = internal.CreateDetailPermission(db, &dpRead)
	if err != nil {
		hlog.Errorf("CreateContentEntry: CreateDetailPermission(read) failed, error=%v", err)
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}

	hlog.Debugf("CreateContentEntry: successfully created content entry ID=%d", ce.ID)
	c.JSON(201, ce)
}

// GetContentEntry @Summary Get content entry by ID
// @Description Retrieve a content entry by ID. Returns the entry with creator_id and project_id.
// @Tags content
// @Accept json
// @Produce json
// @Param id path int true "Content Entry ID"
// @Success 200 {object} internal.ContentEntry
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 404 {object} internal.ErrorResponse
// @Security Session
// @Router /api/content_entries/{id} [get]
func GetContentEntry(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	ce, err := internal.GetContentEntry(db, id)
	if err != nil {
		c.JSON(404, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, ce)
}

// UpdateContentEntry @Summary Update content entry
// @Description Update an existing content entry
// @Tags content
// @Accept json
// @Produce json
// @Param id path int true "Content Entry ID"
// @Param contentEntry body internal.ContentEntry true "Content Entry"
// @Success 200 {object} internal.ContentEntry
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/content_entries/{id} [put]
func UpdateContentEntry(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	var ce internal.ContentEntry
	if err := c.BindJSON(&ce); err != nil {
		c.JSON(400, internal.NewErrorResponse(err.Error()))
		return
	}
	err = internal.UpdateContentEntry(db, id, &ce)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, ce)
}

// DeleteContentEntry @Summary Delete content entry
// @Description Delete a content entry by ID
// @Tags content
// @Accept json
// @Produce json
// @Param id path int true "Content Entry ID"
// @Success 200 {object} internal.SuccessResponse
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/content_entries/{id} [delete]
func DeleteContentEntry(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	err = internal.DeleteContentEntry(db, id)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, internal.NewSuccessResponse("deleted"))
}
