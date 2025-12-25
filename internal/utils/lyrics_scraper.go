package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// LyricsScraper 用于处理歌词刮削相关的功能
type LyricsScraper struct {
	APIURL string
	Client *http.Client
	Logger *log.Logger
}

// NewLyricsScraper 创建一个新的歌词刮削器实例
func NewLyricsScraper(apiURL string, logger *log.Logger) *LyricsScraper {
	return &LyricsScraper{
		APIURL: apiURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		Logger: logger,
	}
}

// ScrapeLyrics 根据歌曲信息刮削歌词
func (ls *LyricsScraper) ScrapeLyrics(title, artist, album string) (string, error) {
	if ls.Logger != nil {
		ls.Logger.Printf("正在刮削歌词: %s - %s - %s", title, artist, album)
	}

	// 构建API请求URL
	apiURL := fmt.Sprintf("%s/lyrics?%s", ls.APIURL, buildLyricsQuery(title, artist, album))

	// 发送HTTP请求
	resp, err := ls.Client.Get(apiURL)
	if err != nil {
		if ls.Logger != nil {
			ls.Logger.Printf("请求歌词API失败: %v", err)
		}
		return "", err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		if ls.Logger != nil {
			ls.Logger.Printf("歌词API返回错误状态码: %d", resp.StatusCode)
		}
		return "", fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if ls.Logger != nil {
			ls.Logger.Printf("读取歌词API响应失败: %v", err)
		}
		return "", err
	}

	// 检查返回的歌词内容是否为空
	lyrics := string(body)
	if lyrics == "" {
		if ls.Logger != nil {
			ls.Logger.Printf("未找到歌词: %s - %s - %s", title, artist, album)
		}
		return "", fmt.Errorf("未找到歌词")
	}

	if ls.Logger != nil {
		ls.Logger.Printf("成功获取歌词: %s - %s - %s", title, artist, album)
	}

	return lyrics, nil
}

// buildLyricsQuery 构建歌词查询参数
func buildLyricsQuery(title, artist, album string) string {
	// title为必填，artist和album为选填
	v := url.Values{}
	v.Set("title", title)

	if artist != "" {
		v.Set("artist", artist)
	}

	if album != "" {
		v.Set("album", album)
	}

	return v.Encode()
}

// buildCoverQuery 构建封面查询参数
func buildCoverQuery(title, artist, album string) string {
	// title为必填，artist和album为选填
	v := url.Values{}
	v.Set("title", title)

	if artist != "" {
		v.Set("artist", artist)
	}

	if album != "" {
		v.Set("album", album)
	}

	return v.Encode()
}

// ScrapeCover 根据歌曲信息刮削封面
func (ls *LyricsScraper) ScrapeCover(title, artist, album string) (string, error) {
	if ls.Logger != nil {
		ls.Logger.Printf("正在刮削封面: %s - %s - %s", title, artist, album)
	}

	// 构建API请求URL
	apiURL := fmt.Sprintf("%s/cover?%s", ls.APIURL, buildCoverQuery(title, artist, album))

	// 发送HTTP请求
	resp, err := ls.Client.Get(apiURL)
	if err != nil {
		if ls.Logger != nil {
			ls.Logger.Printf("请求封面API失败: %v", err)
		}
		return "", err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		if ls.Logger != nil {
			ls.Logger.Printf("封面API返回错误状态码: %d", resp.StatusCode)
		}
		return "", fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if ls.Logger != nil {
			ls.Logger.Printf("读取封面API响应失败: %v", err)
		}
		return "", err
	}

	// 检查返回的封面内容是否为空
	coverData := string(body)
	if coverData == "" {
		if ls.Logger != nil {
			ls.Logger.Printf("未找到封面: %s - %s - %s", title, artist, album)
		}
		return "", fmt.Errorf("未找到封面")
	}

	if ls.Logger != nil {
		ls.Logger.Printf("成功获取封面: %s - %s - %s", title, artist, album)
	}

	return coverData, nil
}

// ScrapeSongMetadata 刮削完整的歌曲元数据
func (ls *LyricsScraper) ScrapeSongMetadata(title, artist, album string) (map[string]interface{}, error) {
	if ls.Logger != nil {
		ls.Logger.Printf("正在刮削歌曲元数据: %s - %s - %s", title, artist, album)
	}

	// 同时获取歌词和封面
	results := make(map[string]interface{})

	// 获取歌词
	lyrics, err := ls.ScrapeLyrics(title, artist, album)
	if err == nil {
		results["lyrics"] = lyrics
	} else {
		ls.Logger.Printf("获取歌词失败: %v", err)
		results["lyrics"] = ""
	}

	// 获取封面
	cover, err := ls.ScrapeCover(title, artist, album)
	if err == nil {
		results["cover"] = cover
	} else {
		ls.Logger.Printf("获取封面失败: %v", err)
		results["cover"] = ""
	}

	if ls.Logger != nil {
		ls.Logger.Printf("完成歌曲元数据刮削: %s - %s - %s", title, artist, album)
	}

	return results, nil
}
