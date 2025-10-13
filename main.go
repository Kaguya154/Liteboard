package main

import (
	"context"
	"encoding/gob"
	"flag"
	"liteboard/api"
	"liteboard/auth"

	"github.com/Kaguya154/dbhelper"
	"github.com/Kaguya154/dbhelper/types"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/sessions"
	"github.com/hertz-contrib/sessions/cookie"
	"github.com/hertz-contrib/swagger"
	swaggerFiles "github.com/swaggo/files"

	_ "liteboard/docs"

	_ "liteboard/internal"
)

func main() {
	// Register the type for gob encoding to allow storing auth.User in sessions
	gob.Register(&auth.User{})

	port := flag.String("p", "8080", "监听端口")
	address := flag.String("a", "0.0.0.0", "监听地址")
	help := flag.Bool("h", false, "显示帮助")
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	hlog.Debug("Starting liteboard application")

	// 初始化数据库
	hlog.Debug("Opening database connection")
	conn, err := dbhelper.Open(types.DBConfig{Driver: "sqlite3", DSN: "liteboard.db"})
	if err != nil {
		hlog.Fatal("Failed to open database:", err)
	}
	hlog.Debug("Database connection established")

	// 创建表
	hlog.Debug("Creating database tables")
	tables := []string{
		"CREATE TABLE IF NOT EXISTS project (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, description TEXT, creator_id INTEGER)",
		"CREATE TABLE IF NOT EXISTS user (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, email TEXT, openid TEXT, password_hash TEXT, groups TEXT)",
		"CREATE TABLE IF NOT EXISTS page (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, author_id INTEGER)",
		"CREATE TABLE IF NOT EXISTS sidebar (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, description TEXT)",
		"CREATE TABLE IF NOT EXISTS sidebar_item (id INTEGER PRIMARY KEY AUTOINCREMENT, parent_id INTEGER, name TEXT, icon TEXT, url TEXT, order_num INTEGER)",
		"CREATE TABLE IF NOT EXISTS content_list (id INTEGER PRIMARY KEY AUTOINCREMENT, type TEXT, title TEXT, items TEXT, creator_id INTEGER, project_id INTEGER)",
		"CREATE TABLE IF NOT EXISTS content_entry (id INTEGER PRIMARY KEY AUTOINCREMENT, type TEXT, title TEXT, content TEXT, creator_id INTEGER, project_id INTEGER)",
		"CREATE TABLE IF NOT EXISTS detail_permission (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, content_type TEXT, content_ids TEXT, action TEXT)",
		"CREATE TABLE IF NOT EXISTS permission (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, description TEXT, content_type TEXT, action TEXT, detail INTEGER)",
		"CREATE TABLE IF NOT EXISTS role (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, description TEXT, permissions TEXT)",
	}

	for _, sql := range tables {
		cond := dbhelper.Cond().Raw(sql).Build()
		_, err := conn.Exec(cond)
		if err != nil {
			hlog.Fatal("Failed to create table:", err)
		}
	}
	hlog.Debug("Database tables created successfully")

	api.SetDB(conn)
	auth.SetDB(conn)
	hlog.Debug("Database connections set for api and auth packages")

	h := server.Default(server.WithHostPorts(*address + ":" + *port))
	store := cookie.NewStore([]byte("secret"))
	h.Use(sessions.New("user", store))

	r := h.Group("/")

	// Serve static files
	r.GET("/css/styles.css", func(ctx context.Context, c *app.RequestContext) {
		c.File("./frontend/css/styles.css")
	})
	r.GET("/js/api.js", func(ctx context.Context, c *app.RequestContext) {
		c.File("./frontend/js/api.js")
	})
	r.GET("/js/auth.js", func(ctx context.Context, c *app.RequestContext) {
		c.File("./frontend/js/auth.js")
	})
	r.GET("/js/dashboard.js", func(ctx context.Context, c *app.RequestContext) {
		c.File("./frontend/js/dashboard.js")
	})
	r.GET("/js/board.js", func(ctx context.Context, c *app.RequestContext) {
		c.File("./frontend/js/board.js")
	})

	// Serve HTML pages
	r.GET("/", func(ctx context.Context, c *app.RequestContext) {
		//如果已登录，跳转到/dashboard
		sess := sessions.Default(c)
		user := sess.Get("user")
		if user != nil {
			c.Redirect(302, []byte("/dashboard"))
			return
		}
		//否则显示首页
		c.File("./frontend/index.html")
	})
	r.GET("/dashboard", auth.LoginRequired(), func(ctx context.Context, c *app.RequestContext) {
		c.File("./frontend/dashboard_new.html")
	})
	r.GET("/board.html", auth.LoginRequired(), func(ctx context.Context, c *app.RequestContext) {
		c.File("./frontend/board.html")
	})
	// 注册认证路由 (公开)
	api.RegisterAuthRoutes(r)

	// 受保护路由组，需登录且具备组权限
	apiRoute := r.Group("/api")

	apiRoute.Use(auth.LoginRequired(), auth.PermissionMiddleware("user", "admin"))

	api.RegisterContentRoutes(apiRoute)
	api.RegisterPermissionRoutes(apiRoute)
	api.RegisterProjectRoutes(apiRoute)
	api.RegisterUserRoutes(apiRoute)

	// 404 handler
	h.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		c.String(404, "404 Not Found: "+string(c.Request.Path()))
	})

	url := swagger.URL("http://localhost:8080/swagger/doc.json") // The url pointing to API definition
	h.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler, url))

	h.Spin()
}
