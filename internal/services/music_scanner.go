package services

import (
	"context"
	"database/sql"
	"fmt"
	"melogo/internal/config"
	"melogo/internal/model"
	"melogo/internal/utils"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"go.senan.xyz/taglib"
)

var (
	// GlobalMusicScanner 全局音乐扫描器实例
	GlobalMusicScanner *MusicScanner
	Collecting         bool = false
)

// MusicScanner 音乐扫描器
type MusicScanner struct {
	Cfg           *config.Config
	Db            *sql.DB
	Logger        *utils.Logger
	LyricsScraper *utils.LyricsScraper
	cancel        context.CancelFunc
}

// NewMusicScanner 创建新的音乐扫描器
func NewMusicScanner(cfg *config.Config, db *sql.DB) *MusicScanner {
	logger := utils.NewLogger()
	scanner := &MusicScanner{
		Cfg:           cfg,
		Db:            db,
		Logger:        logger,
		LyricsScraper: utils.NewLyricsScraper(cfg.Music.LyricsAPIURL, logger.GetStandardLogger()),
	}
	GlobalMusicScanner = scanner
	return scanner
}

// Start 启动定时扫描任务
func (ms *MusicScanner) Start() {
	// 创建上下文用于控制定时任务
	ctx, cancel := context.WithCancel(context.Background())
	ms.cancel = cancel

	// 立即执行一次扫描
	// go ms.scanMusicDirectory()

	// 启动定时任务，根据配置的时间间隔扫描
	scanInterval := time.Duration(ms.Cfg.Music.ScanInterval) * time.Minute
	go func() {
		ticker := time.NewTicker(scanInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ms.scanMusicDirectory()
			case <-ctx.Done():
				ms.Logger.Info("Music scanner stopped")
				return
			}
		}
	}()

	ms.Logger.Infof("Music scanner started, scanning every %d minutes", ms.Cfg.Music.ScanInterval)
}

// Stop 停止定时扫描任务
func (ms *MusicScanner) Stop() {
	if ms.cancel != nil {
		ms.cancel()
	}
}

// scanMusicDirectory 扫描音乐目录
func (ms *MusicScanner) scanMusicDirectory() {
	if Collecting {
		ms.Logger.Info("Collector is already running")
		return
	}
	ms.Logger.Info("Starting music directory scan...")

	// 检查音乐目录是否存在
	if _, err := os.Stat(ms.Cfg.Music.Directory); os.IsNotExist(err) {
		ms.Logger.Warningf("Music directory does not exist: %s", ms.Cfg.Music.Directory)
		return
	}

	// 支持的音频格式
	supportedFormats := make(map[string]bool)
	for _, format := range ms.Cfg.Music.AllowedFormats {
		supportedFormats[format] = true
	}

	// 收集meta信息
	var metas []*songMetadata

	// 遍历音乐目录
	err := filepath.Walk(ms.Cfg.Music.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			ms.Logger.Errorf("Error accessing path %s: %v", path, err)
			return nil
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(path))
		if !supportedFormats[ext] {
			return nil
		}

		// 处理音频文件
		if meta, err := ms.processAudioFile(path, info); err != nil {
			ms.Logger.Errorf("Error processing file %s: %v", path, err)
		} else if meta != nil {
			metas = append(metas, meta)
		}

		return nil
	})

	if err != nil {
		ms.Logger.Errorf("Error walking music directory: %v", err)
	}

	ms.Logger.Info("Music directory scan completed")

	// 刮削歌曲缺失的歌词或者封面
	ms.scrapeMissingMetadata(metas)
}

// songMetadata 内部结构，用于在处理过程中传递歌曲元数据
type songMetadata struct {
	Title        string
	Artist       string
	Album        string
	Duration     int
	LyricsPath   string
	CoverPath    string
	FilePath     string // 绝对路径
	RelativePath string // 相对数据库音乐目录的路径
}

// processAudioFile 处理音频文件
func (ms *MusicScanner) processAudioFile(filePath string, fileInfo os.FileInfo) (*songMetadata, error) {
	// 获取相对路径
	relPath, err := filepath.Rel(ms.Cfg.Music.Directory, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %v", err)
	}

	// 查询历史数据到song信息中，同时检查记录是否存在和is_collect状态
	var isCollect, isDeleted int
	err = ms.Db.QueryRow("SELECT is_collect, is_deleted FROM songs WHERE file_path = ?", relPath).Scan(&isCollect, &isDeleted)
	if err != nil && err != sql.ErrNoRows {
		// 如果查询出错但不是因为记录不存在，记录错误但继续处理
		ms.Logger.Warningf("Error querying is_collect for %s: %v", relPath, err)
	} else if err == nil && isDeleted == 1 {
		ms.Logger.Debugf("Song already deleted, skipping: %s", relPath)
		return nil, nil
	} else if err == nil && isCollect == 1 {
		// 如果记录存在且is_collect为1，跳过后续处理
		ms.Logger.Debugf("Song already processed (is_collect=1), skipping: %s", relPath)
		return nil, nil
	}

	// 确定记录是否存在（用于决定是插入还是更新）
	exists := err != sql.ErrNoRows

	// 1. 解析及提取元数据（包括处理歌词和封面文件）
	meta, err := ms.resolveSongMetadata(filePath, relPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve metadata for %s: %v", filePath, err)
	}

	// 2. 保存到数据库 (Insert 或 Update)
	return meta, ms.saveSongToDB(meta, exists)
}

// resolveSongMetadata 解析音频文件，提取元数据，处理歌词和封面保存
func (ms *MusicScanner) resolveSongMetadata(filePath string, relPath string) (*songMetadata, error) {

	meta := &songMetadata{
		FilePath:     filePath,
		RelativePath: relPath,
		Title:        "Unknown Title",
		Artist:       "Unknown Artist",
		Album:        "Unknown Album",
	}

	// 提取元数据
	tags, duration, coverMimeType, err := ms.extractAudioMetadata(filePath)
	if err != nil {
		ms.Logger.Errorf("Error extracting metadata from %s: %v", filePath, err)
		// 出错也继续，使用默认值
	}
	meta.Duration = duration

	// 处理标签信息 (Title, Artist, Album)
	if tags != nil {
		if val, ok := tags[taglib.Title]; ok && len(val) > 0 && val[0] != "" {
			meta.Title = val[0]
		}
		if val, ok := tags[taglib.Artist]; ok && len(val) > 0 && val[0] != "" {
			meta.Artist = val[0]
		}
		if val, ok := tags[taglib.Album]; ok && len(val) > 0 && val[0] != "" {
			meta.Album = val[0]
		}

		// 处理歌词 (Extract & Save)
		if lyrics, ok := tags[taglib.Lyrics]; ok && len(lyrics) > 0 && lyrics[0] != "" {
			lrcPath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".lrc"
			if _, err := os.Stat(lrcPath); os.IsNotExist(err) {
				if err := os.WriteFile(lrcPath, []byte(lyrics[0]), 0644); err != nil {
					ms.Logger.Errorf("Failed to write lyrics to %s: %v", lrcPath, err)
				} else {
					ms.Logger.Infof("Extracted lyrics to %s", lrcPath)
				}
			}
		}
	}

	// 再次检查文件名兜底 (如果在Tags里没找到标题)
	if meta.Title == "Unknown Title" {
		filename := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		if filename == "" {
			ms.Logger.Warningf("Invalid filename: %s", filePath)
			return nil, fmt.Errorf("invalid filename: %s", filePath)
		}
		parts := strings.Split(filename, " - ")
		if len(parts) >= 2 {
			meta.Artist = parts[0]
			meta.Title = strings.Join(parts[1:], " - ")
		} else {
			meta.Title = filename
		}
	}

	// 确定最终的LyricsPath (检查文件是否存在)
	lrcAbsPath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".lrc"

	// Check if exists using absolute path
	if _, err := os.Stat(lrcAbsPath); os.IsNotExist(err) {
		meta.LyricsPath = ""
	} else {
		// Store relative path in metadata
		relLrcPath, err := filepath.Rel(ms.Cfg.Music.Directory, lrcAbsPath)
		if err != nil {
			ms.Logger.Warningf("Failed to get relative lyrics path: %v", err)
			meta.LyricsPath = ""
		} else {
			meta.LyricsPath = relLrcPath
		}
	}

	// 处理封面图片 (Extract & Save)
	if coverMimeType != "" {
		ext := ".jpg"
		if coverMimeType == "image/png" {
			ext = ".png"
		}
		coverAbsPath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ext
		relCoverPath, _ := filepath.Rel(ms.Cfg.Music.Directory, coverAbsPath)

		// 检查是否存在
		if _, err := os.Stat(coverAbsPath); os.IsNotExist(err) {
			// Save
			imageData, err := taglib.ReadImage(filePath)
			if err == nil && len(imageData) > 0 {
				if err := os.WriteFile(coverAbsPath, imageData, 0644); err != nil {
					ms.Logger.Errorf("Failed to write cover image to %s: %v", coverAbsPath, err)
				} else {
					ms.Logger.Infof("Extracted cover image to %s", coverAbsPath)
					meta.CoverPath = relCoverPath
				}
			}
		} else {
			// Already exists
			meta.CoverPath = relCoverPath
		}
	}

	return meta, nil
}

// saveSongToDB 保存歌曲信息到数据库
func (ms *MusicScanner) saveSongToDB(meta *songMetadata, exists bool) error {
	// 判断是否歌词和封面都存在，如果都存在，则设置is_collect为1,否则更新为0
	var isCollect int
	if meta.LyricsPath != "" && meta.CoverPath != "" {
		isCollect = 1
	} else {
		isCollect = 0
	}

	if exists {
		// Update
		query := `
			UPDATE songs 
			SET title = ?, artist = ?, album = ?, duration = ?, lyrics_path = ?, play_count = play_count, is_collect = ?, updated_at = CURRENT_TIMESTAMP
			WHERE file_path = ?
		`
		args := []interface{}{meta.Title, meta.Artist, meta.Album, meta.Duration, meta.LyricsPath, isCollect, meta.RelativePath}

		if meta.CoverPath != "" {
			query = `
				UPDATE songs 
				SET title = ?, artist = ?, album = ?, duration = ?, lyrics_path = ?, cover_image = ?, play_count = play_count, is_collect = ?, updated_at = CURRENT_TIMESTAMP
				WHERE file_path = ?
			`
			args = []interface{}{meta.Title, meta.Artist, meta.Album, meta.Duration, meta.LyricsPath, meta.CoverPath, isCollect, meta.RelativePath}
		}

		_, err := ms.Db.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("failed to update song: %v", err)
		}
		ms.Logger.Infof("Updated song info: %s - %s", meta.Artist, meta.Title)
	} else {
		// Insert
		query := `
			INSERT INTO songs (title, artist, album, duration, file_path, lyrics_path, cover_image, is_collect, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		_, err := ms.Db.Exec(query, meta.Title, meta.Artist, meta.Album, meta.Duration, meta.RelativePath, meta.LyricsPath, meta.CoverPath, isCollect)

		if err != nil {
			// 唯一性约束检查
			if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
				ms.Logger.Infof("Song already exists in database (constraint): %s", meta.RelativePath)
				return nil
			}
			return fmt.Errorf("failed to insert song: %v", err)
		}
		ms.Logger.Infof("Added new song: %s - %s", meta.Artist, meta.Title)
	}

	return nil
}

// GetSongs 获取所有歌曲
func (ms *MusicScanner) GetSongs() ([]model.SongInfo, error) {
	query := `
		SELECT id, title, artist, album, duration, cover_image, is_deleted, updated_at
		FROM songs
		WHERE is_deleted = 0
		ORDER BY created_at DESC
	`
	rows, err := ms.Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []model.SongInfo
	for rows.Next() {
		var song model.SongInfo
		err := rows.Scan(&song.ID, &song.Title, &song.Artist, &song.Album, &song.Duration, &song.CoverImage, &song.IsDeleted, &song.UpdatedAt)
		if err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}

	return songs, rows.Err()
}

// GetSongByID 根据ID获取歌曲详情
func (ms *MusicScanner) GetSongByID(id int) (*model.Song, error) {
	ms.Logger.Infof("Getting song by ID: %d", id)
	query := `
		SELECT id, title, artist, album, duration, file_path, cover_image, lyrics_path, play_count, is_deleted, created_at, updated_at
		FROM songs
		WHERE id = ?
	`
	row := ms.Db.QueryRow(query, id)

	var song model.Song
	err := row.Scan(
		&song.ID, &song.Title, &song.Artist, &song.Album,
		&song.Duration, &song.FilePath, &song.CoverImage, &song.LyricsPath,
		&song.PlayCount, &song.IsDeleted, &song.CreatedAt, &song.UpdatedAt,
	)
	if err != nil {
		ms.Logger.Errorf("Error getting song by ID %d: %v", id, err)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("song not found")
		}
		return nil, err
	}

	ms.Logger.Debugf("Found song: %+v", song)
	return &song, nil
}

// GetSongs 获取所有歌曲的便捷函数
func GetSongs() ([]model.SongInfo, error) {
	if GlobalMusicScanner == nil {
		return nil, fmt.Errorf("music scanner not initialized")
	}
	return GlobalMusicScanner.GetSongs()
}

// GetSongByID 根据ID获取歌曲详情的便捷函数
func GetSongByID(id int) (*model.Song, error) {
	if GlobalMusicScanner == nil {
		return nil, fmt.Errorf("music scanner not initialized")
	}
	return GlobalMusicScanner.GetSongByID(id)
}

// extractAudioMetadata 从音频文件中提取元数据和时长
func (ms *MusicScanner) extractAudioMetadata(filePath string) (map[string][]string, int, string, error) {
	// 使用go-taglib库读取元数据
	tags, err := taglib.ReadTags(filePath)
	if err != nil {
		ms.Logger.Warningf("Could not read tags from %s: %v", filePath, err)
		// 即使无法读取标签，也尝试读取属性或估算时长
	}

	// 尝试读取音频属性获取时长
	props, propErr := taglib.ReadProperties(filePath)
	var coverMimeType string
	if propErr == nil {
		if props.Length > 0 {
			// 检查是否有封面图片
			if len(props.Images) > 0 {
				coverMimeType = props.Images[0].MIMEType
			}
			return tags, int(props.Length.Seconds()), coverMimeType, nil
		}
	}

	ms.Logger.Warningf("Could not read properties from %s or duration is 0 (err: %v). Falling back to estimation.", filePath, propErr)

	// 如果无法读取属性或时长为0，使用估算方法获取时长
	duration, estErr := ms.estimateDurationFromSize(filePath)
	if estErr != nil {
		// 如果连估算都失败了，且之前读tags也失败了，那就真的失败了
		if err != nil {
			return nil, 0, "", fmt.Errorf("failed to read metadata and estimate duration: %v, %v", err, estErr)
		}
		// 如果tags读取成功但时长失败，返回tags和0时长
		return tags, 0, "", nil
	}

	return tags, int(duration), "", nil
}

// estimateBitrateForFormat 估算特定格式的比特率
func (ms *MusicScanner) estimateBitrateForFormat(filePath string) float64 {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".mp3":
		return 128.0 // 假设MP3平均比特率为128 kbps
	case ".flac":
		return 1000.0 // FLAC是无损压缩，比特率较高
	case ".wav":
		return 1411.0 // 假设WAV为CD质量 (44.1kHz, 16bit, 立体声)
	case ".m4a", ".aac":
		return 128.0 // 假设AAC/M4A平均比特率为128 kbps
	case ".ogg":
		return 128.0 // 假设OGG平均比特率为128 kbps
	default:
		return 128.0 // 默认比特率
	}
}

// estimateDurationFromSize 基于文件大小估算音频时长
func (ms *MusicScanner) estimateDurationFromSize(filePath string) (float64, error) {
	// 获取文件信息
	stat, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	fileSize := stat.Size()

	// 基于文件大小估算时长
	approxBitrate := ms.estimateBitrateForFormat(filePath) // kbps
	// size (bytes) * 8 (bits/byte) / (bitrate (kbps) * 1000 (bits/kbit))
	approxDuration := float64(fileSize) * 8.0 / (approxBitrate * 1000.0)

	return approxDuration, nil
}

// SearchSongs 搜索歌曲
func (ms *MusicScanner) SearchSongs(query string) ([]model.SongInfo, error) {
	searchQuery := "%" + query + "%"
	sqlQuery := `
		SELECT id, title, artist, album, duration, cover_image, is_deleted, updated_at
		FROM songs
		WHERE (title LIKE ? OR artist LIKE ? OR album LIKE ?) AND is_deleted = 0
		ORDER BY created_at DESC
	`
	rows, err := ms.Db.Query(sqlQuery, searchQuery, searchQuery, searchQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []model.SongInfo
	for rows.Next() {
		var song model.SongInfo
		err := rows.Scan(&song.ID, &song.Title, &song.Artist, &song.Album, &song.Duration, &song.CoverImage, &song.IsDeleted, &song.UpdatedAt)
		if err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}

	return songs, rows.Err()
}

// scrapeMissingMetadata 刮削缺失的歌词或封面
func (ms *MusicScanner) scrapeMissingMetadata(metas []*songMetadata) {
	if len(metas) == 0 {
		return
	}
	ms.Logger.Info("开始刮削缺失的歌词和封面")
	Collecting = true

	for _, meta := range metas {
		// 检查是否缺少歌词或封面
		missingLyrics := meta.LyricsPath == ""
		missingCover := meta.CoverPath == ""

		if missingLyrics || missingCover {
			ms.Logger.Debugf("发现缺少元数据的歌曲: %s - %s", meta.Title, meta.Artist)

			// 使用歌词刮削工具刮削缺失的数据
			if missingLyrics {
				if lyrics, err := ms.LyricsScraper.ScrapeLyrics(meta.Title, meta.Artist, meta.Album); err == nil && lyrics != "" {
					// 保存歌词到文件
					lrcPath := strings.TrimSuffix(meta.FilePath, filepath.Ext(meta.FilePath)) + ".lrc"
					if err := os.WriteFile(lrcPath, []byte(lyrics), 0644); err != nil {
						ms.Logger.Errorf("Failed to write lyrics to %s: %v", lrcPath, err)
					} else {
						ms.Logger.Infof("保存歌词到: %s", lrcPath)
						// 更新元数据中的歌词路径
						relLrcPath, _ := filepath.Rel(ms.Cfg.Music.Directory, lrcPath)
						meta.LyricsPath = relLrcPath
					}
				}
			}

			if missingCover {
				if coverData, err := ms.LyricsScraper.ScrapeCover(meta.Title, meta.Artist, meta.Album); err == nil && coverData != "" {
					// 保存封面到文件
					// 根据API响应的格式，coverData可能是URL或base64编码的图像数据
					// 这里我们假设它是可以直接保存的图像数据
					coverExt := ".jpg" // 默认扩展名，可以根据实际API返回的数据类型调整
					coverPath := strings.TrimSuffix(meta.FilePath, filepath.Ext(meta.FilePath)) + coverExt
					if err := os.WriteFile(coverPath, []byte(coverData), 0644); err != nil {
						ms.Logger.Errorf("Failed to write cover to %s: %v", coverPath, err)
					} else {
						ms.Logger.Infof("保存封面到: %s", coverPath)
						// 更新元数据中的封面路径
						relCoverPath, _ := filepath.Rel(ms.Cfg.Music.Directory, coverPath)
						meta.CoverPath = relCoverPath
					}
				}
			}

			// 更新数据库中的歌曲信息，反映元数据的更新状态
			if meta.LyricsPath != "" || meta.CoverPath != "" {
				var isCollect int
				if meta.LyricsPath != "" && meta.CoverPath != "" {
					isCollect = 1
				} else {
					isCollect = 0
				}

				query := "UPDATE songs SET lyrics_path = ?, cover_image = ?, is_collect = ? WHERE file_path = ?"
				_, err := ms.Db.Exec(query, meta.LyricsPath, meta.CoverPath, isCollect, meta.RelativePath)
				if err != nil {
					ms.Logger.Errorf("更新歌曲元数据失败: %v", err)
				} else {
					ms.Logger.Infof("更新歌曲元数据: %s - %s", meta.Artist, meta.Title)
				}
			}
		}
	}

	Collecting = false
	ms.Logger.Info("歌词和封面刮削完成")
}

// SearchSongs 搜索歌曲的便捷函数
func SearchSongs(query string) ([]model.SongInfo, error) {
	if GlobalMusicScanner == nil {
		return nil, fmt.Errorf("music scanner not initialized")
	}
	return GlobalMusicScanner.SearchSongs(query)
}
