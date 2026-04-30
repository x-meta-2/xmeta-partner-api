package middlewares

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"time"
	"xmeta-partner/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

var cleanupOnce sync.Once

var recentLogs = struct {
	sync.RWMutex
	data map[string]time.Time
}{data: make(map[string]time.Time)}

const deduplicateInterval = 30 * time.Second

// ActivityLog - log all API requests
func ActivityLog(db *gorm.DB) gin.HandlerFunc {
	cleanupOnce.Do(func() {
		go func() {
			cleanupOldLogs(db)
			ticker := time.NewTicker(24 * time.Hour)
			for range ticker.C {
				cleanupOldLogs(db)
			}
		}()
	})

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		start := time.Now()

		captureRequestBody(c)

		rw := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = rw

		c.Next()

		// Try partner auth first, then admin auth
		partnerID := ""
		adminID := ""
		if partner := PartnerGetAuth(c); partner != nil {
			partnerID = partner.ID
		} else if admin := AdminGetAuth(c); admin != nil {
			adminID = admin.ID
		}

		if partnerID == "" && adminID == "" {
			return
		}

		duration := time.Since(start).Milliseconds()
		action := detectAction(c.Request.Method, path)

		if action == "LIST" {
			cacheKey := partnerID + adminID + ":" + path
			recentLogs.RLock()
			lastTime, exists := recentLogs.data[cacheKey]
			recentLogs.RUnlock()
			if exists && time.Since(lastTime) < deduplicateInterval {
				return
			}
			recentLogs.Lock()
			recentLogs.data[cacheKey] = time.Now()
			recentLogs.Unlock()
		}

		ip := c.ClientIP()
		userAgent := c.Request.UserAgent()

		go func() {
			log := database.PartnerActivityLog{
				PartnerID:  partnerID,
				AdminID:    adminID,
				Action:     action,
				Method:     c.Request.Method,
				Path:       path,
				StatusCode: c.Writer.Status(),
				IP:         ip,
				UserAgent:  userAgent,
				Duration:   duration,
			}
			db.Create(&log)
		}()
	}
}

func captureRequestBody(c *gin.Context) {
	if c.Request.Method != "POST" && c.Request.Method != "PUT" && c.Request.Method != "PATCH" {
		return
	}
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
}

func detectAction(method string, path string) string {
	if method == "POST" {
		if strings.HasSuffix(path, "/list") {
			return "LIST"
		}
		if strings.Contains(path, "/dashboard/") {
			return "LIST"
		}
	}
	switch method {
	case "GET":
		return "VIEW"
	case "POST":
		return "CREATE"
	case "PUT":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return method
	}
}

func cleanupOldLogs(db *gorm.DB) {
	cutoff := time.Now().AddDate(0, 0, -30)
	db.Where("created_at < ?", cutoff).Delete(&database.PartnerActivityLog{})
}
