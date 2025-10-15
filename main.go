package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"flag"
	"fmt"
	"liteboard/api"
	"liteboard/auth"
	"os"

	_ "liteboard/docs"

	"github.com/Kaguya154/dbhelper"
	"github.com/Kaguya154/dbhelper/types"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/sessions"
	"github.com/hertz-contrib/sessions/cookie"
	"github.com/hertz-contrib/swagger"
	swaggerFiles "github.com/swaggo/files"
	"golang.org/x/net/http2"

	_ "liteboard/internal"
)

func main() {
	// Register the type for gob encoding to allow storing auth.User in sessions
	gob.Register(&auth.User{})

	port := flag.String("p", "8080", "Listen port/监听端口")
	address := flag.String("a", "0.0.0.0", "Listen address/监听地址")
	help := flag.Bool("h", false, "Show helps/显示帮助")
	swaggerFlag := flag.Bool("swagger", false, "Enable swagger docs/启用 Swagger 文档")
	sessionSecret := flag.String("s", "secret", "Session secret")
	enableTLS := flag.Bool("tls", false, "Enable TLS/启用 TLS")
	crtPath := flag.String("crt", "server.crt", "TLS certificate path/TLS 证书路径")
	keyPath := flag.String("key", "server.key", "TLS key path/TLS 密钥路径")
	caPath := flag.String("ca", "ca.crt", "TLS CA certificate path/TLS CA 证书路径")

	flag.Parse()
	if *help {
		flag.Usage()
		return
	}
	if *sessionSecret == "" {
		hlog.Error("The session secret is strongly recommended to be set for security reasons.")
	}

	initDB()

	hlog.Debug("Starting liteboard application")

	var h *server.Hertz
	listenAddr := *address + ":" + *port
	var serverAddr string

	if *enableTLS {
		hlog.Info("TLS is enabled. Make sure to configure TLS certificates in production.")
		hlog.Info("Remember to set crtPath, keyPath, and caPath correctly.")

		h = server.Default(withTLS(*crtPath, *keyPath, *caPath), server.WithHostPorts(listenAddr))
		hlog.Info("Listening on https://" + listenAddr)
		serverAddr = "https://" + listenAddr
	} else {
		h = server.Default(server.WithHostPorts(listenAddr))
		hlog.Info("Listening on http://" + listenAddr)
		serverAddr = "http://" + listenAddr
	}

	store := cookie.NewStore([]byte(*sessionSecret))
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
	r.GET("/share", func(ctx context.Context, c *app.RequestContext) {
		c.File("./frontend/share.html")
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
	api.RegisterShareRoutes(apiRoute)

	// User profile endpoint (requires login only, no permission check)
	apiRoute.GET("/user/profile", api.GetUserProfile)

	// 404 handler
	h.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		c.String(404, "404 Not Found: "+string(c.Request.Path()))
	})
	// Swagger docs
	if *swaggerFlag {
		hlog.Info("Swagger docs enabled at " + serverAddr + "/swagger/index.html")

		url := swagger.URL(serverAddr + "/swagger/doc.json") // The url pointing to API definition
		h.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler, url))
	}

	h.Spin()
}

func initDB() {
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
		"CREATE TABLE IF NOT EXISTS user (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, email TEXT, openid TEXT, password_hash TEXT, groups TEXT, avatar_url TEXT)",
		"CREATE TABLE IF NOT EXISTS page (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, author_id INTEGER)",
		"CREATE TABLE IF NOT EXISTS sidebar (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, description TEXT)",
		"CREATE TABLE IF NOT EXISTS sidebar_item (id INTEGER PRIMARY KEY AUTOINCREMENT, parent_id INTEGER, name TEXT, icon TEXT, url TEXT, order_num INTEGER)",
		"CREATE TABLE IF NOT EXISTS content_list (id INTEGER PRIMARY KEY AUTOINCREMENT, type TEXT, title TEXT, items TEXT, creator_id INTEGER, project_id INTEGER)",
		"CREATE TABLE IF NOT EXISTS content_entry (id INTEGER PRIMARY KEY AUTOINCREMENT, type TEXT, title TEXT, content TEXT, creator_id INTEGER, project_id INTEGER)",
		"CREATE TABLE IF NOT EXISTS detail_permission (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, content_type TEXT, content_ids TEXT, action TEXT)",
		"CREATE TABLE IF NOT EXISTS permission (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, description TEXT, content_type TEXT, action TEXT, detail INTEGER)",
		"CREATE TABLE IF NOT EXISTS role (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, description TEXT, permissions TEXT)",
		"CREATE TABLE IF NOT EXISTS share_token (id INTEGER PRIMARY KEY AUTOINCREMENT, token TEXT UNIQUE, project_id INTEGER, permission_level TEXT, created_at INTEGER, expires_at INTEGER)",
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
}

func withTLS(crtPath, keyPath, caPath string) config.Option {

	// load server certificate
	cert, err := tls.LoadX509KeyPair(crtPath, keyPath)
	if err != nil {
		fmt.Println(err.Error())
	}
	// load root certificate
	certBytes, err := os.ReadFile(caPath)
	if err != nil {
		fmt.Println(err.Error())
	}
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		panic("Failed to parse root certificate.")
	}
	// set server tls.Config
	cfg := &tls.Config{
		// add certificate
		Certificates: []tls.Certificate{cert},
		MaxVersion:   tls.VersionTLS13,
		// enable client authentication
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  caCertPool,
		// cipher suites supported
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		// set application protocol http2
		NextProtos: []string{http2.NextProtoTLS},
	}
	return server.WithTLS(cfg)
}
