package auth

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid Request Method", er)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, ok := users[username] // add in memory database for users

	if !ok || !utils.checkPasswordHash(password, user.HashPassword) {
		er := http.StatusUnauthorized
		http.Error(w, "Invalid username or password", er)
		return
	}

	// // generate session token

	// sessionToken := generateSessionToken(32)

	// http.SetCookie(w, &http.Cookie)

	fmt.Fprintln(w, "Login successful!")
}
