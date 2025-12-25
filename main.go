package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"melogo/internal/config"
	"melogo/internal/handler"
	"melogo/internal/i18n"
	"melogo/internal/routes"
	"melogo/internal/services"
	"melogo/internal/utils"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
)

//go:embed web/templates/*
//go:embed web/templates/partials/*
var templateFiles embed.FS

//go:embed web/assets/*
var assetFiles embed.FS

//go:embed web/locales/*
var localeFiles embed.FS

func main() {
	// 定义命令行参数
	var envFile string
	flag.StringVar(&envFile, "f", "", "Path to .env file")
	flag.Parse()

	// 初始化日志
	logger := utils.NewLogger()

	// 初始化配置
	var cfg *config.Config
	if envFile != "" {
		cfg = config.LoadConfigFromFile(envFile)
	} else {
		cfg = config.LoadConfig()
	}

	// 初始化数据库
	if err := services.InitDatabase(cfg); err != nil {
		logger.Errorf("Failed to initialize database: %v", err)
		os.Exit(1)
	}

	// 初始化JWT密钥
	utils.SetJWTSecret(cfg.Auth.JWTSecret)

	// 初始化i18n
	localesFS, err := fs.Sub(localeFiles, "web/locales")
	if err != nil {
		logger.Errorf("Failed to create locales filesystem: %v", err)
		os.Exit(1)
	}
	if err := i18n.InitWithFS(localesFS); err != nil {
		logger.Errorf("Failed to initialize i18n: %v", err)
		os.Exit(1)
	}

	// 初始化handler
	handler.InitUserHandler(cfg)

	// 初始化播放列表服务
	handler.InitPlaylistHandler(services.NewPlaylistService(services.DB))

	// 初始化收藏服务
	handler.InitFavoriteHandler(services.NewFavoriteService(services.DB))

	// 创建音乐扫描器
	scanner := services.NewMusicScanner(cfg, services.DB)

	// 启动音乐扫描服务
	scanner.Start()

	// 设置Gin模式
	if cfg.Server.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 注册i18n中间件
	r.Use(i18n.Middleware())

	// 设置模板函数
	r.SetFuncMap(template.FuncMap{
		"T": func(c *gin.Context) func(string, ...interface{}) string {
			return i18n.GetT(c)
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	})

	// 设置HTML模板
	// 从嵌入的文件系统加载模板
	tmpl := template.New("")
	tmpl = tmpl.Funcs(template.FuncMap{
		"T": func(c *gin.Context) func(string, ...interface{}) string {
			return i18n.GetT(c)
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	})

	// 遍历嵌入的模板文件
	entries, err := fs.ReadDir(templateFiles, "web/templates")
	if err != nil {
		logger.Errorf("Failed to read templates directory: %v", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
			templatePath := "web/templates/" + entry.Name()
			templateContent, err := templateFiles.ReadFile(templatePath)
			if err != nil {
				logger.Errorf("Failed to read template file %s: %v", templatePath, err)
				os.Exit(1)
			}
			_, err = tmpl.New(entry.Name()).Parse(string(templateContent))
			if err != nil {
				logger.Errorf("Failed to parse template %s: %v", entry.Name(), err)
				os.Exit(1)
			}
		}
	}

	// 加载partials目录中的模板
	partialsEntries, err := fs.ReadDir(templateFiles, "web/templates/partials")
	if err != nil {
		logger.Errorf("Failed to read templates/partials directory: %v", err)
		os.Exit(1)
	}

	for _, entry := range partialsEntries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
			templatePath := "web/templates/partials/" + entry.Name()
			templateContent, err := templateFiles.ReadFile(templatePath)
			if err != nil {
				logger.Errorf("Failed to read partial template file %s: %v", templatePath, err)
				os.Exit(1)
			}
			_, err = tmpl.New(entry.Name()).Parse(string(templateContent))
			if err != nil {
				logger.Errorf("Failed to parse partial template %s: %v", entry.Name(), err)
				os.Exit(1)
			}
		}
	}

	r.SetHTMLTemplate(tmpl)

	// 注册路由
	assetsFS, err := fs.Sub(assetFiles, "web/assets")
	if err != nil {
		logger.Errorf("Failed to create assets filesystem: %v", err)
		os.Exit(1)
	}
	routes.RegisterRoutes(r, assetsFS)

	// 创建信号通道以优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 在goroutine中启动服务器
	go func() {
		logger.Infof("Starting server on %s:%d", cfg.Server.Host, cfg.Server.Port)
		var err error
		if cfg.Server.TLS {
			// TLS模式启动
			logger.Info("Starting server in TLS mode")
			err = r.RunTLS(cfg.Server.Address(), cfg.Server.CertFile, cfg.Server.KeyFile)
		} else {
			// 普通模式启动
			err = r.Run(cfg.Server.Address())
		}
		if err != nil {
			logger.Errorf("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	// 等待中断信号以优雅关闭服务器
	<-quit
	logger.Info("Shutting down server...")

	// 停止音乐扫描服务
	scanner.Stop()

	logger.Info("Server exited")
}
