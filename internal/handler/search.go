package handler

import (
	"melogo/internal/services"

	"github.com/gin-gonic/gin"
)

// 使用music.go中定义的errorHandler，避免重复声明
// var errorHandler = utils.NewErrorHandler()

// Search searches for songs, artists, or albums
func Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		errorHandler.HandleOK(c, gin.H{
			"songs": []interface{}{},
		})
		return
	}

	songs, err := services.SearchSongs(query)
	if err != nil {
		errorHandler.HandleInternalServerError(c, "Failed to search songs", err)
		return
	}

	errorHandler.HandleOK(c, gin.H{
		"songs": songs,
	})
}

// GetSearchHistory returns the user's search history
func GetSearchHistory(c *gin.Context) {
	// TODO: Implement get search history logic
	history := []string{"song1", "artist1", "album1"}

	errorHandler.HandleOK(c, gin.H{
		"history": history,
	})
}
