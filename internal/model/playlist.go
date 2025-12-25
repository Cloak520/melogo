package model

import (
	"time"
)

// Playlist represents a playlist in the system
type Playlist struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	UserID    int       `json:"user_id" db:"user_id"`
	IsPublic  bool      `json:"is_public" db:"is_public"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PlaylistWithSongs represents a playlist with its songs
type PlaylistWithSongs struct {
	Playlist
	Songs []SongInfo `json:"songs"`
}

// PlaylistSong represents the association between a playlist and a song
type PlaylistSong struct {
	ID         int       `json:"id" db:"id"`
	PlaylistID int       `json:"playlist_id" db:"playlist_id"`
	SongID     int       `json:"song_id" db:"song_id"`
	OrderIndex int       `json:"order_index" db:"order_index"`
	AddedAt    time.Time `json:"added_at" db:"added_at"`
}

// CreatePlaylistRequest 创建播放列表请求
type CreatePlaylistRequest struct {
	Name     string `json:"name" binding:"required"`
	IsPublic bool   `json:"is_public"`
}

// UpdatePlaylistRequest 更新播放列表请求
type UpdatePlaylistRequest struct {
	Name     string `json:"name" binding:"required"`
	IsPublic bool   `json:"is_public"`
}

// AddSongToPlaylistRequest 添加歌曲到播放列表请求
type AddSongToPlaylistRequest struct {
	SongID int `json:"song_id" binding:"required"`
}
