package adapters

import (
	"brokerx/ports"
	"context"
	"net/http"

	"github.com/gorilla/sessions"
)

type contextKey string

const USER_ID_KEY contextKey = "user_id"

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

    if err := handler.initSession(request, writer, user.ID); err != nil {
		http.Error(writer, "failed to save session: " + err.Error(), http.StatusInternalServerError)
		return
	}
    
	http.Redirect(writer, request, "/", http.StatusFound)
}

func (handler *AuthHandler) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        session, _ := handler.SessionStore.Get(r, "brokerx-session")
        userID, ok := session.Values["user_id"].(string)
        if !ok || userID == "" {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }
        
        context := context.WithValue(r.Context(), USER_ID_KEY, userID)
        r = r.WithContext(context)

        next.ServeHTTP(w, r)
    })
}

func (handler *AuthHandler) initSession(r *http.Request, w http.ResponseWriter, userId string) error {
	session, _ := handler.SessionStore.Get(r, "brokerx-session")
    session.Values["user_id"] = userId
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