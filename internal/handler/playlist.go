package handler

import (
	"melogo/internal/middleware"
	"melogo/internal/model"
	"melogo/internal/services"
	"melogo/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

var playlistService *services.PlaylistService

// 使用music.go中定义的errorHandler，避免重复声明
// var errorHandler = utils.NewErrorHandler()

// InitPlaylistHandler 初始化播放列表处理器
func InitPlaylistHandler(service *services.PlaylistService) {
	playlistService = service
	utils.NewLogger().Info("Playlist handler initialized")
}

// ListPlaylists 获取播放列表列表
func ListPlaylists(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		errorHandler.HandleUnauthorized(c, "未登录")
		return
	}

	playlists, err := playlistService.GetUserPlaylists(userID)
	if err != nil {
		errorHandler.HandleInternalServerError(c, "获取播放列表失败", err)
		return
	}

	errorHandler.HandleOK(c, gin.H{
		"playlists": playlists,
	})
}

// CreatePlaylist 创建播放列表
func CreatePlaylist(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		errorHandler.HandleUnauthorized(c, "未登录")
		return
	}

	var req model.CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorHandler.HandleBadRequest(c, "参数错误: "+err.Error(), err)
		return
	}

	playlist, err := playlistService.CreatePlaylist(userID, req.Name, req.IsPublic)
	if err != nil {
		errorHandler.HandleInternalServerError(c, "创建播放列表失败", err)
		return
	}

	errorHandler.HandleOK(c, gin.H{
		"message":  "创建成功",
		"playlist": playlist,
	})
}

// UpdatePlaylist 更新播放列表
func UpdatePlaylist(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		errorHandler.HandleUnauthorized(c, "未登录")
		return
	}

	playlistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errorHandler.HandleBadRequest(c, "无效的播放列表ID", err)
		return
	}

	var req model.UpdatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorHandler.HandleBadRequest(c, "参数错误: "+err.Error(), err)
		return
	}

	err = playlistService.UpdatePlaylist(playlistID, userID, req.Name, req.IsPublic)
	if err != nil {
		errorHandler.HandleForbidden(c, err.Error())
		return
	}

	errorHandler.HandleOK(c, model.SuccessResponse{Message: "更新成功"})
}

// DeletePlaylist 删除播放列表
func DeletePlaylist(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		errorHandler.HandleUnauthorized(c, "未登录")
		return
	}

	playlistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errorHandler.HandleBadRequest(c, "无效的播放列表ID", err)
		return
	}

	err = playlistService.DeletePlaylist(playlistID, userID)
	if err != nil {
		errorHandler.HandleForbidden(c, err.Error())
		return
	}

	errorHandler.HandleOK(c, model.SuccessResponse{Message: "删除成功"})
}

// AddSongToPlaylist 添加歌曲到播放列表
func AddSongToPlaylist(c *gin.Context) {
	playlistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errorHandler.HandleBadRequest(c, "无效的播放列表ID", err)
		return
	}

	var req model.AddSongToPlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorHandler.HandleBadRequest(c, "参数错误: "+err.Error(), err)
		return
	}

	err = playlistService.AddSongToPlaylist(playlistID, req.SongID)
	if err != nil {
		errorHandler.HandleBadRequest(c, err.Error(), err)
		return
	}

	errorHandler.HandleOK(c, model.SuccessResponse{Message: "添加成功"})
}

// RemoveSongFromPlaylist 从播放列表移除歌曲
func RemoveSongFromPlaylist(c *gin.Context) {
	playlistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errorHandler.HandleBadRequest(c, "无效的播放列表ID", err)
		return
	}

	songID, err := strconv.Atoi(c.Param("song_id"))
	if err != nil {
		errorHandler.HandleBadRequest(c, "无效的歌曲ID", err)
		return
	}

	err = playlistService.RemoveSongFromPlaylist(playlistID, songID)
	if err != nil {
		errorHandler.HandleBadRequest(c, err.Error(), err)
		return
	}

	errorHandler.HandleOK(c, model.SuccessResponse{Message: "移除成功"})
}

// GetPlaylistDetail 获取播放列表详情（包含歌曲）
func GetPlaylistDetail(c *gin.Context) {
	playlistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errorHandler.HandleBadRequest(c, "无效的播放列表ID", err)
		return
	}

	playlist, err := playlistService.GetPlaylistByID(playlistID)
	if err != nil {
		errorHandler.HandleNotFound(c, err.Error())
		return
	}

	songs, err := playlistService.GetPlaylistSongs(playlistID)
	if err != nil {
		errorHandler.HandleInternalServerError(c, "获取播放列表歌曲失败", err)
		return
	}

	errorHandler.HandleOK(c, gin.H{
		"playlist": playlist,
		"songs":    songs,
	})
}
