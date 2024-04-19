package main

import (
	"github.com/mrsobakin/pixelbattle/auth"
	"github.com/mrsobakin/pixelbattle/game"
	"net/http"
	"net/url"
	"time"
)

const AUTH_ENDPOINT = "http://localhost:8239/authorize"

func main() {
	url, _ := url.Parse(AUTH_ENDPOINT)

	auth := auth.NewRemoteAuthenticator(*url)
	game := game.NewGame(auth, game.GameConfig{
		Width:    500,
		Height:   250,
		Buffer:   10000000,
		Cooldown: 10 * time.Second,
	})

	go game.GameLoop()

	http.HandleFunc("/", game.HandleConnection)
	http.ListenAndServe(":8080", nil)
}
