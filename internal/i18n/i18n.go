package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	bundle *i18n.Bundle
)

// Init initializes the i18n bundle
func Init(localesPath string) error {
	bundle = i18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Walk locales directory and load all json files
	err := filepath.Walk(localesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			_, err = bundle.LoadMessageFile(path)
			if err != nil {
				return fmt.Errorf("failed to load message file %s: %v", path, err)
			}
		}
		return nil
	})

	return err
}

// InitWithFS initializes the i18n bundle from an embedded filesystem
func InitWithFS(localesFS fs.FS) error {
	bundle = i18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Walk locales directory and load all json files from embedded filesystem
	entries, err := fs.ReadDir(localesFS, ".")
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			path := filepath.Join(".", entry.Name())
			data, err := fs.ReadFile(localesFS, path)
			if err != nil {
				return fmt.Errorf("failed to read message file %s: %v", path, err)
			}
			_, err = bundle.ParseMessageFileBytes(data, path)
			if err != nil {
				return fmt.Errorf("failed to parse message file %s: %v", path, err)
			}
		}
	}

	return nil
}

// GetLocalizer returns a localizer for the given language tags
func GetLocalizer(langs ...string) *i18n.Localizer {
	return i18n.NewLocalizer(bundle, langs...)
}

// T translates a message by ID
func T(localizer *i18n.Localizer, messageID string, data map[string]interface{}) string {
	translation, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil {
		return messageID
	}
	return translation
}

// Middleware is a gin middleware for language detection
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check query parameter ?lang=
		lang := c.Query("lang")

		// 2. Check cookie
		if lang == "" {
			var err error
			lang, err = c.Cookie("lang")
			if err != nil {
				lang = ""
			}
		}

		// 3. Check Accept-Language header
		if lang == "" {
			lang = c.GetHeader("Accept-Language")
		}

		// Save language to cookie if it comes from query
		if c.Query("lang") != "" {
			c.SetCookie("lang", lang, 3600*24*30, "/", "", false, false)
		}

		localizer := GetLocalizer(lang)
		c.Set("localizer", localizer)
		c.Next()
	}
}

// GetT returns a function that can translate strings in templates
func GetT(c *gin.Context) func(string, ...interface{}) string {
	localizerObj, exists := c.Get("localizer")
	if !exists {
		// Fallback if middleware not used
		return func(id string, args ...interface{}) string { return id }
	}
	localizer := localizerObj.(*i18n.Localizer)

	return func(id string, args ...interface{}) string {
		var data map[string]interface{}
		if len(args) > 0 {
			if d, ok := args[0].(map[string]interface{}); ok {
				data = d
			}
		}
		return T(localizer, id, data)
	}
}

// HTML is a wrapper for c.HTML that injects the T function
func HTML(c *gin.Context, code int, name string, obj gin.H) {
	if obj == nil {
		obj = gin.H{}
	}
	obj["T"] = GetT(c)
	obj["CurrentLang"] = GetCurrentLang(c)
	c.HTML(code, name, obj)
}

// GetCurrentLang returns the current language detected by the middleware
func GetCurrentLang(c *gin.Context) string {
	// 1. Check query parameter ?lang=
	lang := c.Query("lang")

	// 2. Check cookie
	if lang == "" {
		var err error
		lang, err = c.Cookie("lang")
		if err != nil {
			lang = ""
		}
	}

	// 3. Check Accept-Language header
	if lang == "" {
		lang = c.GetHeader("Accept-Language")
	}

	if lang == "" {
		return "zh" // Default
	}

	// Simplify to "en" or "zh"
	if len(lang) >= 2 {
		lang = lang[:2]
	}
	return lang
}
