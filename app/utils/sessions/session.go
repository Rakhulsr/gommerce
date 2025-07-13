package sessions

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

const (
	sessionStoreKey = "ecommerce-session-key"

	sessionCookieName = "ecommerce-session"

	userIDSessionKey = "userID"
	cartIDSessionKey = "cartID"
)

type SessionStore interface {
	GetUserID(r *http.Request) string
	SetUserID(w http.ResponseWriter, r *http.Request, userID string) error
	ClearUserID(w http.ResponseWriter, r *http.Request) error

	GetCartID(r *http.Request) string
	SetCartID(w http.ResponseWriter, r *http.Request, cartID string) error
	ClearCartID(w http.ResponseWriter, r *http.Request) error

	ClearSession(w http.ResponseWriter, r *http.Request) error
}

type CookieSessionStore struct {
	store *sessions.CookieStore
}

func NewCookieSessionStore(keyPairs ...[]byte) *CookieSessionStore {
	store := sessions.NewCookieStore(keyPairs...)

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(30 * 24 * time.Hour / time.Second),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	return &CookieSessionStore{store: store}
}

func (c *CookieSessionStore) getSession(w http.ResponseWriter, r *http.Request) (*sessions.Session, error) {
	session, err := c.store.Get(r, sessionCookieName)
	if err != nil {

		log.Printf("Error getting session: %v", err)
	}
	return session, nil
}

func (c *CookieSessionStore) GetUserID(r *http.Request) string {
	session, err := c.getSession(nil, r)
	if err != nil || session == nil {
		return ""
	}
	userID, ok := session.Values[userIDSessionKey].(string)
	if !ok {
		return ""
	}
	return userID
}

func (c *CookieSessionStore) SetUserID(w http.ResponseWriter, r *http.Request, userID string) error {
	session, err := c.getSession(w, r)
	if err != nil || session == nil {
		return err
	}
	session.Values[userIDSessionKey] = userID
	return session.Save(r, w)
}

func (c *CookieSessionStore) ClearUserID(w http.ResponseWriter, r *http.Request) error {
	session, err := c.getSession(w, r)
	if err != nil || session == nil {
		return err
	}
	delete(session.Values, userIDSessionKey)
	return session.Save(r, w)
}

func (c *CookieSessionStore) GetCartID(r *http.Request) string {
	session, err := c.getSession(nil, r)
	if err != nil || session == nil {
		return ""
	}
	cartID, ok := session.Values[cartIDSessionKey].(string)
	if !ok {
		return ""
	}
	return cartID
}

func (c *CookieSessionStore) SetCartID(w http.ResponseWriter, r *http.Request, cartID string) error {
	session, err := c.getSession(w, r)
	if err != nil || session == nil {
		return err
	}
	session.Values[cartIDSessionKey] = cartID
	return session.Save(r, w)
}

func (c *CookieSessionStore) ClearCartID(w http.ResponseWriter, r *http.Request) error {
	session, err := c.getSession(w, r)
	if err != nil || session == nil {
		return err
	}
	delete(session.Values, cartIDSessionKey)
	return session.Save(r, w)
}

func (c *CookieSessionStore) ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := c.getSession(w, r)
	if err != nil || session == nil {
		return err
	}
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1
	return session.Save(r, w)
}
