package api

import (
	"context"
	"liteboard/auth"
	"liteboard/internal"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
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
// @Router /api/content_lists [get]
func GetContentLists(ctx context.Context, c *app.RequestContext) {
	projectIDStr := c.Query("projectid")
	if projectIDStr == "" {
		c.JSON(400, map[string]string{"error": "projectid is required"})
		return
	}
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid projectid"})
		return
	}
	lists, err := internal.GetContentListsByProject(db, projectID)
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
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
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/content_lists [post]
func CreateContentList(ctx context.Context, c *app.RequestContext) {
	var cl internal.ContentList
	if err := c.BindJSON(&cl); err != nil {
		c.JSON(400, map[string]string{"error": err.Error()})
		return
	}
	id, err := internal.CreateContentList(db, &cl)
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}
	cl.ID = id
	c.JSON(201, cl)
}

// GetContentList @Summary Get content list by ID
// @Description Retrieve a content list by ID
// @Tags content
// @Accept json
// @Produce json
// @Param id path int true "Content List ID"
// @Success 200 {object} internal.ContentList
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/content_lists/{id} [get]
func GetContentList(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid id"})
		return
	}
	cl, err := internal.GetContentList(db, id)
	if err != nil {
		c.JSON(404, map[string]string{"error": err.Error()})
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
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/content_lists/{id} [put]
func UpdateContentList(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid id"})
		return
	}
	var cl internal.ContentList
	if err := c.BindJSON(&cl); err != nil {
		c.JSON(400, map[string]string{"error": err.Error()})
		return
	}
	err = internal.UpdateContentList(db, id, &cl)
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
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
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/content_lists/{id} [delete]
func DeleteContentList(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid id"})
		return
	}
	err = internal.DeleteContentList(db, id)
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(200, map[string]string{"message": "deleted"})
}

// GetContentEntries @Summary Get all content entries
// @Description Retrieve list of content entries
// @Tags content
// @Accept json
// @Produce json
// @Success 200 {array} internal.ContentEntry
// @Router /api/content_entries [get]
func GetContentEntries(ctx context.Context, c *app.RequestContext) {
	entries, err := internal.GetContentEntries(db)
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(200, entries)
}

// CreateContentEntry @Summary Create content entry
// @Description Create a new content entry
// @Tags content
// @Accept json
// @Produce json
// @Param contentEntry body internal.ContentEntry true "Content Entry"
// @Success 201 {object} internal.ContentEntry
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/content_entries [post]
func CreateContentEntry(ctx context.Context, c *app.RequestContext) {
	var ce internal.ContentEntry
	if err := c.BindJSON(&ce); err != nil {
		c.JSON(400, map[string]string{"error": err.Error()})
		return
	}
	id, err := internal.CreateContentEntry(db, &ce)
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}
	ce.ID = id
	c.JSON(201, ce)
}

// GetContentEntry @Summary Get content entry by ID
// @Description Retrieve a content entry by ID
// @Tags content
// @Accept json
// @Produce json
// @Param id path int true "Content Entry ID"
// @Success 200 {object} internal.ContentEntry
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/content_entries/{id} [get]
func GetContentEntry(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid id"})
		return
	}
	ce, err := internal.GetContentEntry(db, id)
	if err != nil {
		c.JSON(404, map[string]string{"error": err.Error()})
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
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/content_entries/{id} [put]
func UpdateContentEntry(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid id"})
		return
	}
	var ce internal.ContentEntry
	if err := c.BindJSON(&ce); err != nil {
		c.JSON(400, map[string]string{"error": err.Error()})
		return
	}
	err = internal.UpdateContentEntry(db, id, &ce)
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
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
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/content_entries/{id} [delete]
func DeleteContentEntry(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, map[string]string{"error": "invalid id"})
		return
	}
	err = internal.DeleteContentEntry(db, id)
	if err != nil {
		c.JSON(500, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(200, map[string]string{"message": "deleted"})
}
