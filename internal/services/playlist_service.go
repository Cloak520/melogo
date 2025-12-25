package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// PlaylistService 播放列表服务
type PlaylistService struct {
	db *sql.DB
}

// Playlist 播放列表结构
type Playlist struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	UserID    int       `json:"user_id"`
	IsPublic  bool      `json:"is_public"`
	SongCount int       `json:"song_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PlaylistSong 播放列表歌曲结构
type PlaylistSong struct {
	ID         int       `json:"id"`
	PlaylistID int       `json:"playlist_id"`
	SongID     int       `json:"song_id"`
	OrderIndex int       `json:"order_index"`
	AddedAt    time.Time `json:"added_at"`
	// 歌曲详细信息
	Title    string `json:"title,omitempty"`
	Artist   string `json:"artist,omitempty"`
	Duration int    `json:"duration,omitempty"`
}

var playlistService *PlaylistService

// NewPlaylistService 创建播放列表服务实例
func NewPlaylistService(db *sql.DB) *PlaylistService {
	service := &PlaylistService{db: db}
	playlistService = service
	return service
}

// GetPlaylistService 获取全局播放列表服务实例
func GetPlaylistService() *PlaylistService {
	return playlistService
}

// GetUserPlaylists 获取用户的所有播放列表
func (ps *PlaylistService) GetUserPlaylists(userID int) ([]*Playlist, error) {
	query := `
		SELECT p.id, p.name, p.user_id, p.is_public, p.created_at, p.updated_at,
		       COUNT(ps.song_id) as song_count
		FROM playlists p
		LEFT JOIN playlist_songs ps ON p.id = ps.playlist_id
		WHERE p.user_id = ?
		GROUP BY p.id
		ORDER BY p.updated_at DESC
	`

	rows, err := ps.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("查询播放列表失败: %v", err)
	}
	defer rows.Close()

	playlists := []*Playlist{}
	for rows.Next() {
		var p Playlist
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.UserID,
			&p.IsPublic,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.SongCount,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描播放列表数据失败: %v", err)
		}
		playlists = append(playlists, &p)
	}

	return playlists, nil
}

// CreatePlaylist 创建新播放列表
func (ps *PlaylistService) CreatePlaylist(userID int, name string, isPublic bool) (*Playlist, error) {
	if name == "" {
		return nil, errors.New("播放列表名称不能为空")
	}

	isPublicInt := 0
	if isPublic {
		isPublicInt = 1
	}

	query := `
		INSERT INTO playlists (name, user_id, is_public, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := ps.db.Exec(query, name, userID, isPublicInt)
	if err != nil {
		return nil, fmt.Errorf("创建播放列表失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取播放列表ID失败: %v", err)
	}

	playlist := &Playlist{
		ID:        int(id),
		Name:      name,
		UserID:    userID,
		IsPublic:  isPublic,
		SongCount: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return playlist, nil
}

// GetPlaylistByID 根据ID获取播放列表
func (ps *PlaylistService) GetPlaylistByID(playlistID int) (*Playlist, error) {
	query := `
		SELECT p.id, p.name, p.user_id, p.is_public, p.created_at, p.updated_at,
		       COUNT(ps.song_id) as song_count
		FROM playlists p
		LEFT JOIN playlist_songs ps ON p.id = ps.playlist_id
		WHERE p.id = ?
		GROUP BY p.id
	`

	var p Playlist
	err := ps.db.QueryRow(query, playlistID).Scan(
		&p.ID,
		&p.Name,
		&p.UserID,
		&p.IsPublic,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.SongCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("播放列表不存在")
		}
		return nil, fmt.Errorf("查询播放列表失败: %v", err)
	}

	return &p, nil
}

// UpdatePlaylist 更新播放列表
func (ps *PlaylistService) UpdatePlaylist(playlistID, userID int, name string, isPublic bool) error {
	// 检查权限
	playlist, err := ps.GetPlaylistByID(playlistID)
	if err != nil {
		return err
	}
	if playlist.UserID != userID {
		return errors.New("无权限修改此播放列表")
	}

	isPublicInt := 0
	if isPublic {
		isPublicInt = 1
	}

	query := `
		UPDATE playlists 
		SET name = ?, is_public = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err = ps.db.Exec(query, name, isPublicInt, playlistID)
	if err != nil {
		return fmt.Errorf("更新播放列表失败: %v", err)
	}

	return nil
}

// DeletePlaylist 删除播放列表
func (ps *PlaylistService) DeletePlaylist(playlistID, userID int) error {
	// 检查权限
	playlist, err := ps.GetPlaylistByID(playlistID)
	if err != nil {
		return err
	}
	if playlist.UserID != userID {
		return errors.New("无权限删除此播放列表")
	}

	query := `DELETE FROM playlists WHERE id = ?`
	_, err = ps.db.Exec(query, playlistID)
	if err != nil {
		return fmt.Errorf("删除播放列表失败: %v", err)
	}

	return nil
}

// AddSongToPlaylist 添加歌曲到播放列表
func (ps *PlaylistService) AddSongToPlaylist(playlistID, songID int) error {
	// 检查歌曲是否已存在
	checkQuery := `SELECT COUNT(*) FROM playlist_songs WHERE playlist_id = ? AND song_id = ?`
	var count int
	err := ps.db.QueryRow(checkQuery, playlistID, songID).Scan(&count)
	if err != nil {
		return fmt.Errorf("检查歌曲是否存在失败: %v", err)
	}
	if count > 0 {
		return errors.New("歌曲已在播放列表中")
	}

	// 获取当前最大序号
	maxOrderQuery := `SELECT COALESCE(MAX(order_index), 0) FROM playlist_songs WHERE playlist_id = ?`
	var maxOrder int
	err = ps.db.QueryRow(maxOrderQuery, playlistID).Scan(&maxOrder)
	if err != nil {
		return fmt.Errorf("获取最大序号失败: %v", err)
	}

	// 插入歌曲
	insertQuery := `
		INSERT INTO playlist_songs (playlist_id, song_id, order_index, added_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err = ps.db.Exec(insertQuery, playlistID, songID, maxOrder+1)
	if err != nil {
		return fmt.Errorf("添加歌曲失败: %v", err)
	}

	return nil
}

// RemoveSongFromPlaylist 从播放列表移除歌曲
func (ps *PlaylistService) RemoveSongFromPlaylist(playlistID, songID int) error {
	query := `DELETE FROM playlist_songs WHERE playlist_id = ? AND song_id = ?`
	result, err := ps.db.Exec(query, playlistID, songID)
	if err != nil {
		return fmt.Errorf("移除歌曲失败: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("歌曲不在播放列表中")
	}

	return nil
}

// GetPlaylistSongs 获取播放列表中的所有歌曲
func (ps *PlaylistService) GetPlaylistSongs(playlistID int) ([]*PlaylistSong, error) {
	query := `
		SELECT ps.id, ps.playlist_id, ps.song_id, ps.order_index, ps.added_at,
		       s.title, s.artist, s.duration
		FROM playlist_songs ps
		INNER JOIN songs s ON ps.song_id = s.id
		WHERE ps.playlist_id = ? AND s.is_deleted = 0
		ORDER BY ps.order_index ASC
	`

	rows, err := ps.db.Query(query, playlistID)
	if err != nil {
		return nil, fmt.Errorf("查询播放列表歌曲失败: %v", err)
	}
	defer rows.Close()

	songs := []*PlaylistSong{}
	for rows.Next() {
		var ps PlaylistSong
		err := rows.Scan(
			&ps.ID,
			&ps.PlaylistID,
			&ps.SongID,
			&ps.OrderIndex,
			&ps.AddedAt,
			&ps.Title,
			&ps.Artist,
			&ps.Duration,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描歌曲数据失败: %v", err)
		}
		songs = append(songs, &ps)
	}

	return songs, nil
}
