package adapters

import (
	"brokerx/ports"
	"context"
	"net/http"

	"github.com/gorilla/sessions"
)

type contextKey string

const USER_ID_KEY contextKey = "user_id"
const USER_EMAIL_KEY contextKey = "email"

type AuthHandler struct {
	Service ports.AuthService
    SessionStore sessions.Store
    IsProduction bool
}

func (handler *AuthHandler) Login(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil || request.FormValue("email") == "" || request.FormValue("password") == "" {
		http.Error(writer, "badly formed user", http.StatusBadRequest)
		return
	}

	user, e := handler.Service.Authenticate(request.FormValue("email"), request.FormValue("password"))
	if e != nil {
		http.Error(writer, "unauthorized: " + e.Error(), http.StatusUnauthorized)
		return
	}

    if err := handler.initSession(request, writer, user.ID, user.Email); err != nil {
		http.Error(writer, "failed to save session: " + err.Error(), http.StatusInternalServerError)
		return
	}
    
	http.Redirect(writer, request, "/", http.StatusFound)
}

func (handler *AuthHandler) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        session, _ := handler.SessionStore.Get(r, "brokerx-session")
        userID, idOk := session.Values["user_id"].(string)
        userEmail, ok := session.Values["email"].(string)
        if !idOk || !ok || userID == "" {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }
        
        ctx := context.WithValue(r.Context(), USER_ID_KEY, userID)
        ctx = context.WithValue(ctx, USER_EMAIL_KEY, userEmail)
        r = r.WithContext(ctx)

        next.ServeHTTP(w, r)
    })
}

func (handler *AuthHandler) initSession(r *http.Request, w http.ResponseWriter, userId string, userEmail string) error {
	session, _ := handler.SessionStore.Get(r, "brokerx-session")
    session.Values["user_id"] = userId
    session.Values["email"] = userEmail
    session.Options = &sessions.Options{
        Path:     "/",
        MaxAge:   600,
        HttpOnly: true,
        Secure:   handler.IsProduction,
        SameSite: http.SameSiteLaxMode,
    }

    if err := session.Save(r, w); err != nil {
        return err
    }
    
    return nil
}