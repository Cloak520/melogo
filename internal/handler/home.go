package handler

import (
	"melogo/internal/i18n"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Index serves the home page
func Index(c *gin.Context) {
	i18n.HTML(c, http.StatusOK, "index.html", gin.H{
		"title": "MeloGo 音乐播放器",
	})
}

// PlaylistsPage serves the playlists page
func PlaylistsPage(c *gin.Context) {
	i18n.HTML(c, http.StatusOK, "playlists.html", gin.H{
		"title": "我的播放列表 - MeloGo",
	})
}

// PlaylistDetailPage serves a specific playlist detail page
func PlaylistDetailPage(c *gin.Context) {
	id := c.Param("id")
	i18n.HTML(c, http.StatusOK, "playlist_detail.html", gin.H{
		"title":    "播放列表详情 - MeloGo",
		"playlist": id,
	})
}
