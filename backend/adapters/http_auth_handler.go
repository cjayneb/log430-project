package adapters

import (
	"brokerx/ports"
	"net/http"

	"github.com/gorilla/sessions"
)


var store = sessions.NewCookieStore([]byte("very-secret-key"))

type AuthHandler struct {
	Service ports.AuthService
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

	handler.initSession(request, writer, user.Email)
	http.Redirect(writer, request, "/", http.StatusFound)
}

func (handler *AuthHandler) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "brokerx-session")
        if session.Values["user_id"] == nil {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func (handler *AuthHandler) initSession(r *http.Request, w http.ResponseWriter, userEmail string) {
	session, _ := store.Get(r, "brokerx-session")
    session.Values["user_id"] = userEmail
    session.Options = &sessions.Options{
        Path:     "/",
        MaxAge:   600,
        HttpOnly: handler.IsProduction,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    }
    err := session.Save(r, w)
    if err != nil {
        http.Error(w, "Failed to save session", http.StatusInternalServerError)
    }
}