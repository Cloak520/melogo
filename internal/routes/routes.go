package routes

import (
	"io/fs"
	"melogo/internal/handler"
	"melogo/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all application routes
func RegisterRoutes(r *gin.Engine, assetsFS fs.FS) {
	// Create a http.FileServer for the embedded assets
	fileServer := http.FileServer(http.FS(assetsFS))

	// Route to serve embedded assets
	r.GET("/assets/*filepath", func(c *gin.Context) {
		// Get the requested file path
		filePath := c.Param("filepath")
		// Remove leading slash if present
		if len(filePath) > 0 && filePath[0] == '/' {
			filePath = filePath[1:]
		}
		// Serve the file using the http.FileServer
		c.Request.URL.Path = "/" + filePath
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// Serve favicon
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Request.URL.Path = "/favicon.ico"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Public routes - no authentication required
		api.POST("/register", handler.Register)
		api.POST("/login", handler.Login)

		// Protected routes - authentication required
		authenticated := api.Group("")
		authenticated.Use(middleware.AuthMiddleware())
		{
			// User routes
			authenticated.POST("/logout", handler.Logout)
			authenticated.GET("/user/profile", handler.GetUserProfile)
			authenticated.PUT("/user/profile", handler.UpdateUserProfile)
			authenticated.GET("/user/:id/avatar", handler.GetUserAvatar)

			// Song routes
			authenticated.GET("/songs", handler.ListSongs)
			authenticated.GET("/songs/:id", handler.GetSong)
			authenticated.GET("/songs/:id/stream", handler.StreamSong)
			authenticated.GET("/songs/:id/lyrics", handler.GetLyrics)
			authenticated.GET("/songs/:id/cover", handler.GetCover)

			// Playlist routes // 播放列表相关路由
			authenticated.GET("/playlists", handler.ListPlaylists)
			authenticated.POST("/playlists", handler.CreatePlaylist)
			authenticated.GET("/playlists/:id/detail", handler.GetPlaylistDetail)
			authenticated.PUT("/playlists/:id", handler.UpdatePlaylist)
			authenticated.DELETE("/playlists/:id", handler.DeletePlaylist)
			authenticated.POST("/playlists/:id/songs", handler.AddSongToPlaylist)
			authenticated.DELETE("/playlists/:id/songs/:song_id", handler.RemoveSongFromPlaylist)

			// Favorite routes
			authenticated.GET("/favorites", handler.ListFavorites)
			authenticated.POST("/favorites", handler.AddFavorite)
			authenticated.DELETE("/favorites/:song_id", handler.RemoveFavorite)

			// Search routes
			authenticated.GET("/search", handler.Search)
			authenticated.GET("/search/history", handler.GetSearchHistory)
		}

		// Admin routes - admin authentication required
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.AdminMiddleware())
		{
			// Admin song management routes
			admin.GET("/songs", handler.AdminListSongs)
			admin.PUT("/songs/:id", handler.AdminUpdateSong)
			admin.DELETE("/songs", handler.AdminDeleteSongs)
			admin.GET("/songs/search", handler.AdminSearchSongs)
		}
	}

	// Web routes - public pages
	r.GET("/", handler.Index)
	r.GET("/login", handler.LoginPage)
	r.GET("/register", handler.RegisterPage)

	// Web routes - protected pages
	webAuth := r.Group("")
	webAuth.Use(middleware.AuthMiddleware())
	{
		webAuth.GET("/profile", handler.ProfilePage)
		webAuth.GET("/playlists", handler.PlaylistsPage)
		webAuth.GET("/playlist/:id", handler.PlaylistDetailPage)

		// Admin web routes - admin authentication required
		adminWebAuth := r.Group("")
		adminWebAuth.Use(middleware.AuthMiddleware())
		adminWebAuth.Use(middleware.AdminMiddleware())
		{
			adminWebAuth.GET("/admin/songs", handler.AdminSongsPage)
		}
	}
}
