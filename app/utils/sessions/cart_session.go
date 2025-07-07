// app/utils/sessions/sessions.go

package sessions

import (
	"log"
	"net/http"

	"github.com/Rakhulsr/go-ecommerce/app/configs"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

const (
	SessionCartKey   = "cart_session_gommerce_last_chance" // <--- UBAH INI!
	CartSessionIDKey = "cart_id"
)

var (
	secret = configs.LoadEnv()
	Store  *sessions.CookieStore
)

func init() {
	// log.Printf("Loaded SESSION_KEY: %s (Length: %d)", secret.SESSION_KEY, len(secret.SESSION_KEY))

	if len(secret.SESSION_KEY) < 32 {
		log.Fatalf("SESSION_KEY is too short or empty. Please set a strong, random key of at least 32 characters in your .env file. Current length: %d", len(secret.SESSION_KEY))
	}

	Store = sessions.NewCookieStore([]byte(secret.SESSION_KEY))

	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	log.Println("Gorilla sessions CookieStore initialized successfully.")
}

func GetCartID(w http.ResponseWriter, r *http.Request) (string, error) {
	session, err := Store.Get(r, SessionCartKey)
	if err != nil {
		log.Printf("GetCartID: Error getting session for %s: %v", r.URL.Path, err)
	}

	sessionID := session.ID
	if sessionID == "" {
		if cartID, ok := session.Values[CartSessionIDKey].(string); ok && cartID != "" {
			sessionID = "FALLBACK_CART_ID:" + cartID
		} else {
			sessionID = "EMPTY_SESSION_ID_AND_CART_ID"
		}
	}

	if cartID, ok := session.Values[CartSessionIDKey].(string); ok && cartID != "" {
		log.Printf("GetCartID: Found existing cart ID: %s for %s. Session ID: %s", cartID, r.URL.Path, sessionID)
		return cartID, nil
	}

	newCartID := uuid.New().String()
	session.Values[CartSessionIDKey] = newCartID

	err = session.Save(r, w)
	if err != nil {
		log.Printf("GetCartID: Error saving new cart ID to session: %v", err)
		return "", err
	}
	// log.Printf("GetCartID: New cart ID created (%s) and saved. Session ID: %s", newCartID, sessionID)

	return newCartID, nil
}
