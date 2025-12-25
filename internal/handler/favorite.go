package handler

import (
	"melogo/internal/middleware"
	"melogo/internal/model"
	"melogo/internal/services"
	"melogo/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

var favoriteService *services.FavoriteService

// 使用user.go中定义的errorHandler，避免重复声明
// var errorHandler = utils.NewErrorHandler()

// InitFavoriteHandler 初始化收藏处理器
func InitFavoriteHandler(service *services.FavoriteService) {
	favoriteService = service
	utils.NewLogger().Info("Favorite handler initialized")
}

// ListFavorites returns a list of favorite songs
func ListFavorites(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		errorHandler.HandleUnauthorized(c, "未登录")
		return
	}

	favorites, err := favoriteService.ListFavorites(userID)
	if err != nil {
		errorHandler.HandleInternalServerError(c, "获取收藏列表失败", err)
		return
	}

	// 转换为前端需要的格式
	result := make([]map[string]interface{}, len(favorites))
	for i, fav := range favorites {
		result[i] = map[string]interface{}{
			"id":       fav.Song.ID,
			"title":    fav.Song.Title,
			"artist":   fav.Song.Artist,
			"album":    fav.Song.Album,
			"duration": fav.Song.Duration,
		}
	}

	errorHandler.HandleOK(c, gin.H{
		"favorites": result,
	})
}

// AddFavorite adds a song to favorites
func AddFavorite(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		errorHandler.HandleUnauthorized(c, "未登录")
		return
	}

	var req model.AddFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorHandler.HandleBadRequest(c, "参数错误: "+err.Error(), err)
		return
	}

	err := favoriteService.AddFavorite(userID, req.SongID)
	if err != nil {
		errorHandler.HandleInternalServerError(c, "添加收藏失败", err)
		return
	}

	errorHandler.HandleOK(c, gin.H{
		"message": "收藏成功",
	})
}

// RemoveFavorite removes a song from favorites
func RemoveFavorite(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		errorHandler.HandleUnauthorized(c, "未登录")
		return
	}

	songID, err := strconv.Atoi(c.Param("song_id"))
	if err != nil {
		errorHandler.HandleBadRequest(c, "无效的歌曲ID", err)
		return
	}

	err = favoriteService.RemoveFavorite(userID, songID)
	if err != nil {
		errorHandler.HandleInternalServerError(c, "取消收藏失败", err)
		return
	}

	errorHandler.HandleOK(c, gin.H{
		"message": "取消收藏成功",
	})
}
