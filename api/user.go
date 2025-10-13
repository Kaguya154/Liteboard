package api

import (
	"context"
	"liteboard/auth"
	"liteboard/internal"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
)

func RegisterUserRoutes(r *route.RouterGroup) {
	r.GET("/users", GetUsers)
	r.POST("/users", CreateUser)
	r.GET("/users/:id", auth.PermissionCheckMiddleware("user", "read", GetIDFromParam), GetUser)
	r.PUT("/users/:id", auth.PermissionCheckMiddleware("user", "write", GetIDFromParam), UpdateUser)
	r.DELETE("/users/:id", auth.PermissionCheckMiddleware("user", "admin", GetIDFromParam), DeleteUser)
}

// GetUsers @Summary Get all users
// @Description Retrieve list of users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} internal.User
// @Failure 500 {object} internal.ErrorResponse
// @Router /api/users [get]
func GetUsers(ctx context.Context, c *app.RequestContext) {
	users, err := internal.GetUsers(db)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, users)
}

// CreateUser @Summary Create user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body internal.User true "User"
// @Success 201 {object} internal.User
// @Failure 400 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Router /api/users [post]
func CreateUser(ctx context.Context, c *app.RequestContext) {
	var u internal.User
	if err := c.BindJSON(&u); err != nil {
		c.JSON(400, internal.NewErrorResponse(err.Error()))
		return
	}
	u.Groups = []string{"user"}
	id, err := internal.CreateUser(db, &u)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	u.ID = id
	c.JSON(201, u)
}

// GetUser @Summary Get user by ID
// @Description Retrieve a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} internal.User
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 404 {object} internal.ErrorResponse
// @Security Session
// @Router /api/users/{id} [get]
func GetUser(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	u, err := internal.GetUser(db, id)
	if err != nil {
		c.JSON(404, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, u)
}

// UpdateUser @Summary Update user
// @Description Update an existing user
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body internal.User true "User"
// @Success 200 {object} internal.User
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/users/{id} [put]
func UpdateUser(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	var u internal.User
	if err := c.BindJSON(&u); err != nil {
		c.JSON(400, internal.NewErrorResponse(err.Error()))
		return
	}
	err = internal.UpdateUser(db, id, &u)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, u)
}

// DeleteUser @Summary Delete user
// @Description Delete a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} internal.SuccessResponse
// @Failure 400 {object} internal.ErrorResponse
// @Failure 401 {object} internal.ErrorResponse
// @Failure 403 {object} internal.ErrorResponse
// @Failure 500 {object} internal.ErrorResponse
// @Security Session
// @Router /api/users/{id} [delete]
func DeleteUser(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, internal.NewErrorResponse("invalid id"))
		return
	}
	err = internal.DeleteUser(db, id)
	if err != nil {
		c.JSON(500, internal.NewErrorResponse(err.Error()))
		return
	}
	c.JSON(200, internal.NewSuccessResponse("deleted"))
}
