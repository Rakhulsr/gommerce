package sessions

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

const (
	sessionCookieName = "ecommerce-session"

	userIDSessionKey = "userID"
	cartIDSessionKey = "cartID"
)

type SessionStore interface {
	GetUserID(w http.ResponseWriter, r *http.Request) string
	SetUserID(w http.ResponseWriter, r *http.Request, userID string) error
	ClearUserID(w http.ResponseWriter, r *http.Request) error

	GetCartID(w http.ResponseWriter, r *http.Request) string
	SetCartID(w http.ResponseWriter, r *http.Request, cartID string) error
	ClearCartID(w http.ResponseWriter, r *http.Request) error

	ClearSession(w http.ResponseWriter, r *http.Request) error
	GetSession(w http.ResponseWriter, r *http.Request) (*sessions.Session, error)
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

func (c *CookieSessionStore) GetSession(w http.ResponseWriter, r *http.Request) (*sessions.Session, error) {
	session, err := c.store.Get(r, sessionCookieName)
	if err != nil {

		log.Printf("Error getting session: %v", err)
	}
	return session, nil
}

func (s *CookieSessionStore) GetUserID(w http.ResponseWriter, r *http.Request) string {
	session, err := s.store.Get(r, "auth-session")
	if err != nil {
		log.Printf("SessionStore: Error getting session for UserID: %v", err)
		return ""
	}
	userID, ok := session.Values["user_id"].(string)
	if !ok {
		return ""
	}
	return userID
}

func (s *CookieSessionStore) SetUserID(w http.ResponseWriter, r *http.Request, userID string) error {
	session, err := s.store.Get(r, "auth-session")
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	session.Values["user_id"] = userID
	session.Values["is_logged_in"] = true
	return session.Save(r, w)
}

func (c *CookieSessionStore) ClearUserID(w http.ResponseWriter, r *http.Request) error {
	session, err := c.GetSession(w, r)
	if err != nil || session == nil {
		return err
	}
	delete(session.Values, userIDSessionKey)
	return session.Save(r, w)
}

func (s *CookieSessionStore) GetCartID(w http.ResponseWriter, r *http.Request) string {
	session, err := s.store.Get(r, "auth-session")
	if err != nil {
		log.Printf("SessionStore: Error getting session for CartID: %v", err)
		return ""
	}
	cartID, ok := session.Values["cart_id"].(string)
	if !ok {
		return ""
	}
	return cartID
}

func (s *CookieSessionStore) SetCartID(w http.ResponseWriter, r *http.Request, cartID string) error {
	session, err := s.store.Get(r, "auth-session")
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	session.Values["cart_id"] = cartID
	return session.Save(r, w)
}

func (s *CookieSessionStore) ClearCartID(w http.ResponseWriter, r *http.Request) error {
	session, err := s.store.Get(r, "auth-session")
	if err != nil {
		log.Printf("SessionStore: Error getting session to clear cart ID: %v", err)
		return fmt.Errorf("failed to get session to clear cart ID: %w", err)
	}
	session.Values["cart_id"] = ""
	if err := session.Save(r, w); err != nil {
		log.Printf("SessionStore: Error saving session after clearing cart ID: %v", err)
		return fmt.Errorf("failed to save session after clearing cart ID: %w", err)
	}
	return nil
}

func (s *CookieSessionStore) ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := s.store.Get(r, "auth-session")
	if err != nil {
		log.Printf("SessionStore: Error getting session to clear: %v", err)
		return err
	}
	session.Values["user_id"] = ""
	session.Values["is_logged_in"] = false
	session.Values["cart_id"] = ""
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		log.Printf("SessionStore: Error saving session after clearing: %v", err)
	}

	return nil
}
