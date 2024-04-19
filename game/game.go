package game

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/mrsobakin/pixelbattle/auth"
	"github.com/mrsobakin/pixelbattle/internal"
	"github.com/mrsobakin/pixelbattle/internal/mpmc"
	"log"
	"net/http"
	"sync"
	"time"
)

type GameConfig struct {
	Cooldown time.Duration
	Buffer   int
	Width    int
	Height   int
}

type Game struct {
	auth       auth.Authenticator
	canvas     Canvas
	canvasLock sync.RWMutex
	cooldowns  internal.CooldownManager
	mpmc       mpmc.MPMC[Pixel]
	wsUpgrader websocket.Upgrader
}

func NewGame(auth auth.Authenticator, config GameConfig) *Game {
	return &Game{
		auth:      auth,
		canvas:    *NewCanvas(config.Width, config.Height),
		cooldowns: *internal.NewCooldownManager(config.Cooldown),
		mpmc:      *mpmc.NewMPMC[Pixel](config.Buffer),
	}
}

func (game *Game) GameLoop() {
	rx := game.mpmc.Subscribe()
	for {
		pxl, err := rx.Receive()
		fmt.Println(pxl)

		if err != nil {
			continue
		}

		game.canvasLock.Lock()
		game.canvas.Paint(*pxl)
		game.canvasLock.Unlock()
	}
}

func pipeChanToWs(rx mpmc.Consumer[Pixel], conn *websocket.Conn) {
	defer conn.Close()

	for {
		msg, err := rx.Receive()

		if err != nil {
			conn.WriteMessage(websocket.CloseMessage, nil)
			return
		}

		err = conn.WriteJSON(msg)

		if err != nil {
			return
		}
	}
}

func (game *Game) HandleConnection(w http.ResponseWriter, r *http.Request) {
	uid := game.auth.Authenticate(r)

	if uid == nil {
		log.Println(r.RemoteAddr, "tried to connect, but authentication failed")
		w.WriteHeader(403)
		return
	}
	log.Println(r.RemoteAddr, "connected")

	conn, err := game.wsUpgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	defer conn.Close()

	game.canvasLock.RLock()
	canvasBytes := game.canvas.ToBytes()
	game.canvasLock.RUnlock()

	rx := game.mpmc.Subscribe()

	err = conn.WriteMessage(websocket.BinaryMessage, canvasBytes)
	if err != nil {
		return
	}

	go pipeChanToWs(*rx, conn)

	for {
		mt, msg, err := conn.ReadMessage()

		if err != nil || mt == websocket.CloseMessage {
			break
		}

		if mt != websocket.TextMessage {
			break
		}

		// We can test this even before we decode json
		if can, _ := game.cooldowns.Attempt(*uid); !can {
			continue
		}

		var pixel Pixel
		if err := json.Unmarshal(msg, &pixel); err != nil {
			break
		}

		game.mpmc.Send(pixel)
	}
}
