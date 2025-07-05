package sessions

import (
	"net/http"

	"github.com/Rakhulsr/go-ecommerce/app/configs"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

const (
	SessionCartKey   = "cart_session"
	CartSessionIDKey = "cart_id"
)

var (
	secret = configs.LoadEnv()
	store  = sessions.NewCookieStore([]byte(secret.SESSION_KEY))
)

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   false,
	}
}

func GetCartID(w http.ResponseWriter, r *http.Request) (string, error) {
	session, err := store.Get(r, SessionCartKey)
	if err != nil {
		return "", err
	}

	if cartID, ok := session.Values[CartSessionIDKey].(string); ok && cartID != "" {
		return cartID, nil
	}

	newCartID := uuid.New().String()
	session.Values[CartSessionIDKey] = newCartID
	err = session.Save(r, w)
	if err != nil {
		return "", err
	}

	return newCartID, nil
}
