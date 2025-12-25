package model

import (
	"time"
)

// Song represents a song in the system
type Song struct {
	ID         int       `json:"id" db:"id"`
	Title      string    `json:"title" db:"title"`
	Artist     string    `json:"artist" db:"artist"`
	Album      string    `json:"album" db:"album"`
	Duration   int       `json:"duration" db:"duration"` // Duration in seconds
	FilePath   string    `json:"file_path" db:"file_path"`
	CoverImage *string   `json:"cover_image,omitempty" db:"cover_image"`
	LyricsPath *string   `json:"lyrics_path,omitempty" db:"lyrics_path"`
	PlayCount  int       `json:"play_count" db:"play_count"`
	IsCollect  int       `json:"is_collect" db:"is_collect"`
	IsDeleted  int       `json:"is_deleted" db:"is_deleted"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// SongInfo represents basic song information for listing
type SongInfo struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Artist     string    `json:"artist"`
	Album      string    `json:"album"`
	Duration   int       `json:"duration"`
	CoverImage *string   `json:"cover_image,omitempty"`
	IsDeleted  int       `json:"is_deleted"`
	UpdatedAt  time.Time `json:"updated_at"`
}
