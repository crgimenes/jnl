package session

import (
	"crypto/rand"
	"net/http"
	"time"

	"github.com/crgimenes/jnl/sillyname"
)

type Control struct {
	cookieName     string
	SessionDataMap map[string]SessionData
}

type SessionData struct {
	ExpireAt      time.Time
	CurrentScreen int
	Nick          string
}

func New(cookieName string) *Control {
	return &Control{
		cookieName:     cookieName,
		SessionDataMap: make(map[string]SessionData),
	}
}

func (c *Control) Get(r *http.Request) (string, *SessionData, bool) {
	cookies := r.Cookies()
	if len(cookies) == 0 {
		return "", nil, false
	}

	cookie, err := r.Cookie(c.cookieName)
	if err != nil {
		return "", nil, false
	}

	s, ok := c.SessionDataMap[cookie.Value]
	if !ok {
		return "", nil, false
	}

	if s.ExpireAt.Before(time.Now()) {
		delete(c.SessionDataMap, cookie.Value)
		return "", nil, false
	}

	if s.Nick == "" {
		s.Nick = sillyname.Generate()
	}

	return cookie.Value, &s, true
}

func (c *Control) Delete(w http.ResponseWriter, id string) {
	delete(c.SessionDataMap, id)
	// Path has to match the cookie Save set, or the browser keeps the one at
	// "/" and the session survives the delete.
	cookie := http.Cookie{ // #nosec G124 -- deletion cookie: the value is empty
		Path:   "/",
		Name:   c.cookieName,
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, &cookie)
}

func (c *Control) Save(w http.ResponseWriter, r *http.Request, id string, sessionData *SessionData) {
	expireAt := time.Now().Add(3 * time.Hour)

	// A Secure cookie is silently discarded by browsers on a plain-http
	// origin, leaving every following request unauthenticated. Mark it Secure
	// only when the request actually arrived over TLS, directly or via a
	// reverse proxy.
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	cookie := &http.Cookie{ // #nosec G124 -- Secure follows the transport; plain http on a private network is a supported deployment
		Path:     "/",
		Name:     c.cookieName,
		Value:    id,
		Expires:  expireAt,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	}

	if sessionData == nil {
		sessionData = &SessionData{}
	}

	sessionData.ExpireAt = expireAt
	c.SessionDataMap[id] = *sessionData

	http.SetCookie(w, cookie)
}

func (c *Control) Create() (string, *SessionData) {
	sessionData := &SessionData{
		CurrentScreen: 0,
		Nick:          sillyname.Generate(),
		ExpireAt:      time.Now().Add(3 * time.Hour),
	}

	return RandomID(), sessionData
}

func (c *Control) RemoveExpired() {
	for k, v := range c.SessionDataMap {
		if v.ExpireAt.Before(time.Now()) {
			delete(c.SessionDataMap, k)
		}
	}
}

func RandomID() string {
	const (
		length  = 16
		charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	)
	lenCharset := byte(len(charset))
	b := make([]byte, length)
	_, _ = rand.Read(b)
	for i := range length {
		b[i] = charset[b[i]%lenCharset]
	}
	return string(b)
}

func (c *Control) List() map[string]SessionData {
	return c.SessionDataMap
}
