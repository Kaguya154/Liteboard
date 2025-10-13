package api

import (
	"context"
	"liteboard/auth"
	"liteboard/internal"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
)

func RegisterPermissionRoutes(r *route.RouterGroup) {
	// DetailPermission routes
	r.GET("/detail_permissions", GetDetailPermissions)
	r.POST("/detail_permissions", CreateDetailPermission)
	r.GET("/detail_permissions/:id", auth.PermissionCheckMiddleware("detail_permission", "read", GetIDFromParam), GetDetailPermission)
	r.PUT("/detail_permissions/:id", auth.PermissionCheckMiddleware("detail_permission", "write", GetIDFromParam), UpdateDetailPermission)
	r.DELETE("/detail_permissions/:id", auth.PermissionCheckMiddleware("detail_permission", "admin", GetIDFromParam), DeleteDetailPermission)

	// Permission routes
	r.GET("/permissions", GetPermissions)
	r.POST("/permissions", CreatePermission)
	r.GET("/permissions/:id", auth.PermissionCheckMiddleware("permission", "read", GetIDFromParam), GetPermission)
	r.PUT("/permissions/:id", auth.PermissionCheckMiddleware("permission", "write", GetIDFromParam), UpdatePermission)
	r.DELETE("/permissions/:id", auth.PermissionCheckMiddleware("permission", "admin", GetIDFromParam), DeletePermission)
}

// GetDetailPermissions @Summary Get all detail permissions
// @Description Retrieve list of detail permissions
// @Tags permissions
// @Accept json
// @Produce json
// @Success 200 {array} internal.DetailPermission
// @Failure 500 {object} internal.ErrorResponse
// @Router /api/detail_permissions [get]
func GetDetailPermissions(ctx context.Context, c *app.RequestContext) {
	dps, err := internal.GetDetailPermissions(db)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, dps)
}

// CreateDetailPermission @Summary Create detail permission
// @Description Create a new detail permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param detailPermission body internal.DetailPermission true "Detail Permission"
// @Success 201 {object} internal.DetailPermission
// @Failure 400 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Router /api/detail_permissions [post]
func CreateDetailPermission(ctx context.Context, c *app.RequestContext) {
	var dp internal.DetailPermission
	if err := c.BindJSON(&dp); err != nil {
		c.JSON(400, internal.NewErrorResponse(err.Error()))
		return
	}
	id, err := internal.CreateDetailPermission(db, &dp)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	dp.ID = id
	c.JSON(201, dp)
}

// GetDetailPermission @Summary Get detail permission by ID
// @Description Retrieve a detail permission by ID
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Detail Permission ID"
// @Success 200 {object} internal.DetailPermission
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 404 {object} internal.ErrorResponse
// @Security Session
// @Router /api/detail_permissions/{id} [get]
func GetDetailPermission(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	dp, err := internal.GetDetailPermission(db, id)
	if err != nil {
		c.JSON(404, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, dp)
}

// UpdateDetailPermission @Summary Update detail permission
// @Description Update an existing detail permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Detail Permission ID"
// @Param detailPermission body internal.DetailPermission true "Detail Permission"
// @Success 200 {object} internal.DetailPermission
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/detail_permissions/{id} [put]
func UpdateDetailPermission(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	var dp internal.DetailPermission
	if err := c.BindJSON(&dp); err != nil {
		c.JSON(400, internal.NewErrorResponse(err.Error()))
		return
	}
	err = internal.UpdateDetailPermission(db, id, &dp)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, dp)
}

// DeleteDetailPermission @Summary Delete detail permission
// @Description Delete a detail permission by ID
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Detail Permission ID"
// @Success 200 {object} internal.SuccessResponse
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/detail_permissions/{id} [delete]
func DeleteDetailPermission(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	err = internal.DeleteDetailPermission(db, id)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, internal.NewSuccessResponse("deleted"))
}

// GetPermissions @Summary Get all permissions
// @Description Retrieve list of permissions
// @Tags permissions
// @Accept json
// @Produce json
// @Success 200 {array} internal.Permission
// @Failure 500 {object} internal.ErrorResponse
// @Router /api/permissions [get]
func GetPermissions(ctx context.Context, c *app.RequestContext) {
	permissions, err := internal.GetPermissions(db)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, permissions)
}

// CreatePermission @Summary Create permission
// @Description Create a new permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param permission body internal.Permission true "Permission"
// @Success 201 {object} internal.Permission
// @Failure 400 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Router /api/permissions [post]
func CreatePermission(ctx context.Context, c *app.RequestContext) {
	var p internal.Permission
	if err := c.BindJSON(&p); err != nil {
		c.JSON(400, internal.NewErrorResponse(err.Error()))
		return
	}
	id, err := internal.CreatePermission(db, &p)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	p.ID = id
	c.JSON(201, p)
}

// GetPermission @Summary Get permission by ID
// @Description Retrieve a permission by ID
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Permission ID"
// @Success 200 {object} internal.Permission
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 404 {object} internal.ErrorResponse
// @Security Session
// @Router /api/permissions/{id} [get]
func GetPermission(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	p, err := internal.GetPermission(db, id)
	if err != nil {
		c.JSON(404, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, p)
}

// UpdatePermission @Summary Update permission
// @Description Update an existing permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Permission ID"
// @Param permission body internal.Permission true "Permission"
// @Success 200 {object} internal.Permission
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/permissions/{id} [put]
func UpdatePermission(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	var p internal.Permission
	if err := c.BindJSON(&p); err != nil {
		c.JSON(400, internal.NewErrorResponse(err.Error()))
		return
	}
	err = internal.UpdatePermission(db, id, &p)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, p)
}

// DeletePermission @Summary Delete permission
// @Description Delete a permission by ID
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Permission ID"
// @Success 200 {object} internal.SuccessResponse
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/permissions/{id} [delete]
func DeletePermission(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	err = internal.DeletePermission(db, id)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, internal.NewSuccessResponse("deleted"))
}
