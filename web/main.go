package main

import (
	"log"
	"net/http"

	"github.com/captncraig/mosaicChallenge/imgur"
)

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/oauth", imgur.HandleCallback)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, imgur.ImgurLoginUrl(), 302) })
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/", rootOnly(selectHandler(loggedOutHome, loggedInHome)))
	log.Println("Listening on port 7777.")
	http.ListenAndServe(":7777", nil)
}

func loggedOutHome(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "loggedOut",
		struct{ Token *imgur.ImgurAccessToken }{nil})
}

func loggedInHome(w http.ResponseWriter, r *http.Request, token *imgur.ImgurAccessToken) {
	renderTemplate(w, "loggedIn", struct{ Token *imgur.ImgurAccessToken }{token})
}

func logout(w http.ResponseWriter, r *http.Request) {
	imgur.ClearImgurCookie(w)
	http.Redirect(w, r, "/", 302)
}

// Special handler type that accepts an access token prepopulated by the authentication middleware
type credentialHandler func(w http.ResponseWriter, r *http.Request, token *imgur.ImgurAccessToken)

func imgurHandler(handler credentialHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := imgur.TokenForRequest(w, r)
		handler(w, r, tok)
	}
}

// Select handler based on whether or not the user is logged in or not.
func selectHandler(loggedOut http.HandlerFunc, loggedIn credentialHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := imgur.TokenForRequest(w, r)
		if tok != nil {
			loggedIn(w, r, tok)
		} else {
			loggedOut(w, r)
		}
	}
}

// Turns "/" into a strict match.
func rootOnly(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		handler(w, r)
	}
}
