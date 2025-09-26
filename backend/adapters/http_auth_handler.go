package adapters

import (
	"brokerx/ports"
	"net/http"

	"github.com/gorilla/sessions"
)

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

    if err := handler.initSession(request, writer, user.Email); err != nil {
		http.Error(writer, "failed to save session: " + err.Error(), http.StatusInternalServerError)
		return
	}
    
	http.Redirect(writer, request, "/", http.StatusFound)
}

func (handler *AuthHandler) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        session, _ := handler.SessionStore.Get(r, "brokerx-session")
        if session.Values["user_id"] == nil {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func (handler *AuthHandler) initSession(r *http.Request, w http.ResponseWriter, userEmail string) error {
	session, _ := handler.SessionStore.Get(r, "brokerx-session")
    session.Values["user_id"] = userEmail
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