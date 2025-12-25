package handler

import (
	"melogo/internal/services"
	"melogo/internal/utils"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

var errorHandler = utils.NewErrorHandler()

// ListSongs returns a list of songs
func ListSongs(c *gin.Context) {
	// 获取所有歌曲
	songs, err := services.GetSongs()
	if err != nil {
		errorHandler.HandleInternalServerError(c, "Failed to get songs", err)
		return
	}

	errorHandler.HandleOK(c, gin.H{
		"songs": songs,
	})
}

// GetSong returns details of a specific song
func GetSong(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorHandler.HandleBadRequest(c, "Invalid song ID", err)
		return
	}

	// 获取歌曲详情
	song, err := services.GetSongByID(id)
	if err != nil {
		errorHandler.HandleNotFound(c, "Song not found")
		return
	}

	errorHandler.HandleOK(c, gin.H{
		"song": song,
	})
}

// StreamSong streams a song audio file
func StreamSong(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorHandler.HandleBadRequest(c, "Invalid song ID", err)
		return
	}

	// 获取歌曲详情
	song, err := services.GetSongByID(id)
	if err != nil {
		errorHandler.HandleNotFound(c, "Song not found")
		return
	}

	// 构建完整的文件路径
	filePath := filepath.Join(services.GlobalMusicScanner.Cfg.Music.Directory, song.FilePath)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		errorHandler.HandleNotFound(c, "Audio file not found")
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "audio/mpeg")
	c.Header("Content-Disposition", "inline; filename="+filepath.Base(song.FilePath))
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	// 流式传输文件
	c.File(filePath)
}

// GetLyrics returns lyrics for a specific song
func GetLyrics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorHandler.HandleBadRequest(c, "Invalid song ID", err)
		return
	}

	// 获取歌曲详情
	song, err := services.GetSongByID(id)
	if err != nil {
		errorHandler.HandleNotFound(c, "Song not found")
		return
	}

	// 如果没有歌词文件路径，返回空歌词
	if song.LyricsPath == nil || *song.LyricsPath == "" {
		errorHandler.HandleOK(c, gin.H{
			"song_id": id,
			"lyrics":  "",
		})
		return
	}

	// 构建完整的歌词文件路径
	lyricsPath := filepath.Join(services.GlobalMusicScanner.Cfg.Music.Directory, *song.LyricsPath)

	// 读取歌词文件
	lyricsBytes, err := os.ReadFile(lyricsPath)
	if err != nil {
		errorHandler.HandleOK(c, gin.H{
			"song_id": id,
			"lyrics":  "",
		})
		return
	}

	errorHandler.HandleOK(c, gin.H{
		"song_id": id,
		"lyrics":  string(lyricsBytes),
	})
}

// GetCover serves the cover image for a specific song
func GetCover(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorHandler.HandleBadRequest(c, "Invalid song ID", err)
		return
	}

	// 获取歌曲详情
	song, err := services.GetSongByID(id)
	if err != nil {
		errorHandler.HandleNotFound(c, "Song not found")
		return
	}

	// 如果没有封面图片路径，返回 404
	if song.CoverImage == nil || *song.CoverImage == "" {
		errorHandler.HandleNotFound(c, "Cover image not found")
		return
	}

	// 构建完整的封面文件路径
	coverPath := filepath.Join(services.GlobalMusicScanner.Cfg.Music.Directory, *song.CoverImage)

	// 检查文件是否存在
	if _, err := os.Stat(coverPath); os.IsNotExist(err) {
		errorHandler.HandleNotFound(c, "Cover image file not found")
		return
	}

	// 流式传输文件
	c.File(coverPath)
}
