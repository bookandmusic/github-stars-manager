package session

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type SessionData struct {
    AccessToken string
    UserName    string
    AvatarURL   string
}

var store = map[string]*SessionData{}
var mu sync.Mutex

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewSessionData() *SessionData {
	return &SessionData{}
}

func generateSessionID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func SetSession(c *gin.Context, data *SessionData) {
	sessionID := generateSessionID()
	mu.Lock()
	store[sessionID] = data
	mu.Unlock()
	
	// 设置cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600 * 24, // 24小时
	})
}

func GetSession(c *gin.Context) (*SessionData, error) {
	// 从cookie中获取session_id
	cookie, err := c.Request.Cookie("session_id")
	if err != nil {
		return nil, err
	}
	
	sessionID := cookie.Value
	
	mu.Lock()
	data, ok := store[sessionID]
	mu.Unlock()
	
	if !ok {
		return nil, nil
	}
	
	return data, nil
}

func ClearSession(c *gin.Context) {
	// 从cookie中获取session_id
	cookie, err := c.Request.Cookie("session_id")
	if err != nil {
		return
	}
	
	sessionID := cookie.Value
	
	mu.Lock()
	delete(store, sessionID)
	mu.Unlock()
	
	// 清除cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

func Set(sessionID string, data *SessionData) {
    mu.Lock()
    defer mu.Unlock()
    store[sessionID] = data
}

func Get(sessionID string) (*SessionData, bool) {
    mu.Lock()
    defer mu.Unlock()
    data, ok := store[sessionID]
    return data, ok
}

func Delete(sessionID string) {
    mu.Lock()
    defer mu.Unlock()
    delete(store, sessionID)
}