package model

import (
	"time"
)

// Favorite represents a user's favorite song
type Favorite struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	SongID    int       `json:"song_id" db:"song_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// FavoriteWithSong represents a favorite with song information
type FavoriteWithSong struct {
	Favorite
	Song SongInfo `json:"song"`
}

// AddFavoriteRequest 添加收藏请求
type AddFavoriteRequest struct {
	SongID int `json:"song_id" binding:"required"`
}
