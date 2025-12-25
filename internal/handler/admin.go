package handler

import (
	"database/sql"
	"fmt"
	"melogo/internal/i18n"
	"melogo/internal/middleware"
	"melogo/internal/model"
	"melogo/internal/services"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminSongsPage 管理员歌曲管理页面
func AdminSongsPage(c *gin.Context) {
	i18n.HTML(c, http.StatusOK, "songs.html", gin.H{
		"title": "管理员 - 歌曲管理",
	})
}

// AdminListSongs 管理员获取歌曲列表（支持分页）
func AdminListSongs(c *gin.Context) {
	if _, exists := middleware.GetCurrentUserID(c); !exists {
		errorHandler.HandleUnauthorized(c, "未登录")
		return
	}

	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// 查询歌曲总数
	countQuery := "SELECT COUNT(*) FROM songs WHERE is_deleted = 0"
	var total int
	err = services.DB.QueryRow(countQuery).Scan(&total)
	if err != nil {
		errorHandler.HandleInternalServerError(c, "查询歌曲总数失败", err)
		return
	}

	// 查询分页数据
	query := `
		SELECT id, title, artist, album, duration, file_path, cover_image, lyrics_path, play_count, is_collect, created_at, updated_at
		FROM songs 
		WHERE is_deleted = 0
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := services.DB.Query(query, limit, offset)
	if err != nil {
		errorHandler.HandleInternalServerError(c, "查询歌曲列表失败", err)
		return
	}
	defer rows.Close()

	var songs []model.Song
	for rows.Next() {
		var song model.Song
		var coverImage, lyricsPath sql.NullString
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&song.ID,
			&song.Title,
			&song.Artist,
			&song.Album,
			&song.Duration,
			&song.FilePath,
			&coverImage,
			&lyricsPath,
			&song.PlayCount,
			&song.IsCollect,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			errorHandler.HandleInternalServerError(c, "扫描歌曲数据失败", err)
			return
		}

		// 处理NULL值
		if coverImage.Valid {
			song.CoverImage = &coverImage.String
		}
		if lyricsPath.Valid {
			song.LyricsPath = &lyricsPath.String
		}

		song.CreatedAt = createdAt
		song.UpdatedAt = updatedAt

		songs = append(songs, song)
	}

	totalPages := (total + limit - 1) / limit

	response := gin.H{
		"songs": songs,
		"pagination": gin.H{
			"current_page":   page,
			"total_pages":    totalPages,
			"total_items":    total,
			"items_per_page": limit,
		},
	}

	errorHandler.HandleOK(c, response)
}

// AdminUpdateSong 管理员更新歌曲信息
func AdminUpdateSong(c *gin.Context) {
	if _, exists := middleware.GetCurrentUserID(c); !exists {
		errorHandler.HandleUnauthorized(c, "未登录")
		return
	}

	songIDStr := c.Param("id")
	songID, err := strconv.Atoi(songIDStr)
	if err != nil {
		errorHandler.HandleBadRequest(c, "歌曲ID格式错误", err)
		return
	}

	var req struct {
		Title    string `json:"title" binding:"required"`
		Artist   string `json:"artist" binding:"required"`
		Album    string `json:"album"`
		Duration int    `json:"duration"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorHandler.HandleBadRequest(c, "参数错误: "+err.Error(), err)
		return
	}

	// 检查歌曲是否存在
	song, err := services.GetSongByID(songID)
	if err != nil {
		errorHandler.HandleNotFound(c, "歌曲不存在")
		return
	}

	// 更新歌曲信息
	query := `
		UPDATE songs 
		SET title = ?, artist = ?, album = ?, duration = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err = services.DB.Exec(query, req.Title, req.Artist, req.Album, req.Duration, songID)
	if err != nil {
		errorHandler.HandleInternalServerError(c, "更新歌曲信息失败", err)
		return
	}

	// 使用新的元数据进行歌词和封面的刮削
	if services.GlobalMusicScanner != nil && services.GlobalMusicScanner.LyricsScraper != nil {
		// 刮削歌词和封面
		metadata, err := services.GlobalMusicScanner.LyricsScraper.ScrapeSongMetadata(req.Title, req.Artist, req.Album)
		if err == nil {
			// 获取歌曲的文件路径
			filePath := filepath.Join(services.GlobalMusicScanner.Cfg.Music.Directory, song.FilePath)

			// 保存歌词
			if lyrics, ok := metadata["lyrics"].(string); ok && lyrics != "" {
				lrcPath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".lrc"
				if err := os.WriteFile(lrcPath, []byte(lyrics), 0644); err != nil {
					services.GlobalMusicScanner.Logger.Errorf("Failed to write lyrics to %s: %v", lrcPath, err)
				} else {
					// 更新数据库中的歌词路径
					relLrcPath, _ := filepath.Rel(services.GlobalMusicScanner.Cfg.Music.Directory, lrcPath)
					updateLyricsQuery := `UPDATE songs SET lyrics_path = ? WHERE id = ?`
					services.DB.Exec(updateLyricsQuery, relLrcPath, songID)
					services.GlobalMusicScanner.Logger.Infof("保存歌词到: %s", lrcPath)
				}
			}

			// 保存封面
			if cover, ok := metadata["cover"].(string); ok && cover != "" {
				// 根据API返回的数据类型决定文件扩展名
				ext := ".jpg"
				if strings.HasPrefix(cover, "data:image/png") {
					ext = ".png"
				}
				coverPath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ext
				if err := os.WriteFile(coverPath, []byte(cover), 0644); err != nil {
					services.GlobalMusicScanner.Logger.Errorf("Failed to write cover to %s: %v", coverPath, err)
				} else {
					// 更新数据库中的封面路径
					relCoverPath, _ := filepath.Rel(services.GlobalMusicScanner.Cfg.Music.Directory, coverPath)
					updateCoverQuery := `UPDATE songs SET cover_image = ? WHERE id = ?`
					services.DB.Exec(updateCoverQuery, relCoverPath, songID)
					services.GlobalMusicScanner.Logger.Infof("保存封面到: %s", coverPath)
				}
			}
		}
	}

	errorHandler.HandleOK(c, gin.H{
		"message": "歌曲信息更新成功",
	})
}

// AdminDeleteSongs 管理员删除歌曲（支持批量删除）
func AdminDeleteSongs(c *gin.Context) {
	if _, exists := middleware.GetCurrentUserID(c); !exists {
		errorHandler.HandleUnauthorized(c, "未登录")
		return
	}

	var req struct {
		SongIDs []int `json:"song_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorHandler.HandleBadRequest(c, "参数错误: "+err.Error(), err)
		return
	}

	if len(req.SongIDs) == 0 {
		errorHandler.HandleBadRequest(c, "请选择要删除的歌曲", nil)
		return
	}

	// 构建占位符
	placeholders := strings.Repeat("?,", len(req.SongIDs)-1) + "?"
	query := fmt.Sprintf("UPDATE songs SET is_deleted = 1 WHERE id IN (%s)", placeholders)

	args := make([]interface{}, len(req.SongIDs))
	for i, id := range req.SongIDs {
		args[i] = id
	}

	_, err := services.DB.Exec(query, args...)
	if err != nil {
		errorHandler.HandleInternalServerError(c, "删除歌曲失败", err)
		return
	}

	errorHandler.HandleOK(c, gin.H{
		"message": fmt.Sprintf("成功删除 %d 首歌曲", len(req.SongIDs)),
	})
}

// AdminSearchSongs 管理员搜索歌曲
func AdminSearchSongs(c *gin.Context) {
	if _, exists := middleware.GetCurrentUserID(c); !exists {
		errorHandler.HandleUnauthorized(c, "未登录")
		return
	}

	queryParam := c.Query("q")
	if queryParam == "" {
		errorHandler.HandleBadRequest(c, "搜索关键词不能为空", nil)
		return
	}

	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// 构建搜索查询
	searchQuery := `
		SELECT id, title, artist, album, duration, file_path, cover_image, lyrics_path, play_count, is_collect, created_at, updated_at
		FROM songs 
		WHERE is_deleted = 0 
		AND (title LIKE ? OR artist LIKE ? OR album LIKE ?)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	// 查询总数
	countQuery := `
		SELECT COUNT(*) 
		FROM songs 
		WHERE is_deleted = 0 
		AND (title LIKE ? OR artist LIKE ? OR album LIKE ?)
	`

	var total int
	err = services.DB.QueryRow(countQuery, "%"+queryParam+"%", "%"+queryParam+"%", "%"+queryParam+"%").Scan(&total)
	if err != nil {
		errorHandler.HandleInternalServerError(c, "查询搜索结果总数失败", err)
		return
	}

	// 执行搜索查询
	rows, err := services.DB.Query(searchQuery, "%"+queryParam+"%", "%"+queryParam+"%", "%"+queryParam+"%", limit, offset)
	if err != nil {
		errorHandler.HandleInternalServerError(c, "搜索歌曲失败", err)
		return
	}
	defer rows.Close()

	var songs []model.Song
	for rows.Next() {
		var song model.Song
		var coverImage, lyricsPath sql.NullString
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&song.ID,
			&song.Title,
			&song.Artist,
			&song.Album,
			&song.Duration,
			&song.FilePath,
			&coverImage,
			&lyricsPath,
			&song.PlayCount,
			&song.IsCollect,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			errorHandler.HandleInternalServerError(c, "扫描搜索结果失败", err)
			return
		}

		// 处理NULL值
		if coverImage.Valid {
			song.CoverImage = &coverImage.String
		}
		if lyricsPath.Valid {
			song.LyricsPath = &lyricsPath.String
		}

		song.CreatedAt = createdAt
		song.UpdatedAt = updatedAt

		songs = append(songs, song)
	}

	totalPages := (total + limit - 1) / limit

	response := gin.H{
		"songs": songs,
		"pagination": gin.H{
			"current_page":   page,
			"total_pages":    totalPages,
			"total_items":    total,
			"items_per_page": limit,
		},
	}

	errorHandler.HandleOK(c, response)
}
