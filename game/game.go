package game

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/mrsobakin/pixelbattle/auth"
	"github.com/mrsobakin/pixelbattle/internal"
	"github.com/mrsobakin/pixelbattle/internal/mpmc"
	"log"
	"net/http"
	"time"
)

type GameConfig struct {
	CanvasPath string
	Cooldown   time.Duration
	Buffer     uint64
	Width      int
	Height     int
}

type Game struct {
	auth       auth.Authenticator
	canvas     *Canvas
	canvasPath string
	cooldowns  internal.CooldownManager
	mpmc       mpmc.MPMC[Pixel]
	wsUpgrader websocket.Upgrader
}

func NewGame(auth auth.Authenticator, config GameConfig) *Game {
	var canvas *Canvas
	if config.CanvasPath != "" {
		var err error
		canvas, err = ReadCanvasFromFile(config.CanvasPath)

		if err != nil {
			log.Println("Failed to open canvas:", err)
		} else {
			w, h := canvas.Dimensions()
			if w != config.Width || h != config.Height {
				log.Println("Canvas size does not match given dimensions")
			} else {
				log.Println("Loaded canvas from", config.CanvasPath)
			}
		}
	}

	if canvas == nil {
		log.Printf("Creating new white canvas with (%d, %d) dimensions\n", config.Width, config.Height)
		canvas = NewCanvas(config.Width, config.Height)
	}

	return &Game{
		auth:       auth,
		canvas:     canvas,
		canvasPath: config.CanvasPath,
		cooldowns:  *internal.NewCooldownManager(config.Cooldown),
		mpmc:       *mpmc.NewMPMC[Pixel](config.Buffer),
	}
}

func (game *Game) GameLoop() {
	rx := game.mpmc.Subscribe()
	for {
		pxl, err := rx.Receive()

		if err != nil {
			continue
		}

		game.canvas.Paint(*pxl)
	}
}

func (game *Game) CanvasSavingRoutine() {
	for {
		game.SaveCanvas()
		time.Sleep(time.Minute)
	}
}

func (game *Game) SaveCanvas() {
	err := game.canvas.WriteToFile(game.canvasPath)
	if err != nil {
		log.Println("Failed to save canvas:", err)
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

	canvasBytes := game.canvas.ToBytes()

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

		if !game.canvas.IsInBounds(pixel.Pos[0], pixel.Pos[1]) {
			break
		}

		game.mpmc.Send(pixel)
	}
}
