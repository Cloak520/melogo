package services

import (
	"database/sql"
	"fmt"
	"melogo/internal/model"
)

// FavoriteService 收藏服务
type FavoriteService struct {
	db *sql.DB
}

// NewFavoriteService 创建新的收藏服务实例
func NewFavoriteService(database *sql.DB) *FavoriteService {
	return &FavoriteService{
		db: database,
	}
}

// AddFavorite 添加收藏
func (fs *FavoriteService) AddFavorite(userID, songID int) error {
	// 检查收藏是否已存在
	exists, err := fs.IsFavorite(userID, songID)
	if err != nil {
		return fmt.Errorf("检查收藏状态失败: %v", err)
	}

	// 如果已存在，直接返回
	if exists {
		return nil
	}

	// 插入新的收藏记录
	_, err = fs.db.Exec(
		"INSERT INTO favorites (user_id, song_id, created_at) VALUES (?, ?, datetime('now'))",
		userID, songID,
	)
	if err != nil {
		return fmt.Errorf("添加收藏失败: %v", err)
	}

	return nil
}

// RemoveFavorite 取消收藏
func (fs *FavoriteService) RemoveFavorite(userID, songID int) error {
	_, err := fs.db.Exec(
		"DELETE FROM favorites WHERE user_id = ? AND song_id = ?",
		userID, songID,
	)
	if err != nil {
		return fmt.Errorf("取消收藏失败: %v", err)
	}

	return nil
}

// IsFavorite 检查是否已收藏
func (fs *FavoriteService) IsFavorite(userID, songID int) (bool, error) {
	var count int
	err := fs.db.QueryRow(
		"SELECT COUNT(*) FROM favorites WHERE user_id = ? AND song_id = ?",
		userID, songID,
	).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("查询收藏状态失败: %v", err)
	}

	return count > 0, nil
}

// ListFavorites 获取用户收藏列表
func (fs *FavoriteService) ListFavorites(userID int) ([]model.FavoriteWithSong, error) {
	query := `
		SELECT f.id, f.user_id, f.song_id, f.created_at, 
		       s.title, s.artist, s.album, s.duration
		FROM favorites f
		JOIN songs s ON f.song_id = s.id
		WHERE f.user_id = ? AND s.is_deleted = 0
		ORDER BY f.created_at DESC
	`

	rows, err := fs.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("查询收藏列表失败: %v", err)
	}
	defer rows.Close()

	var favorites []model.FavoriteWithSong
	for rows.Next() {
		var fav model.FavoriteWithSong
		err := rows.Scan(
			&fav.ID, &fav.UserID, &fav.SongID, &fav.CreatedAt,
			&fav.Song.Title, &fav.Song.Artist, &fav.Song.Album, &fav.Song.Duration,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描收藏记录失败: %v", err)
		}
		favorites = append(favorites, fav)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历收藏记录失败: %v", err)
	}

	return favorites, nil
}
