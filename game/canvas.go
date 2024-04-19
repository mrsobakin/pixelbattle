package game

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"sync"
)

type Canvas struct {
	lock  sync.RWMutex
	image *image.RGBA
}

func NewCanvas(width int, height int) *Canvas {
	image := image.NewRGBA(image.Rect(0, 0, width, height))

	for x := range width {
		for y := range height {
			image.Set(x, y, color.White)
		}
	}

	return &Canvas{
		lock:  sync.RWMutex{},
		image: image,
	}
}

func (canvas *Canvas) Paint(pixel Pixel) {
	canvas.lock.Lock()
	defer canvas.lock.Unlock()

	canvas.image.Set(pixel.Pos[0], pixel.Pos[1], color.NRGBA{
		R: pixel.Color[0],
		G: pixel.Color[1],
		B: pixel.Color[2],
	})
}

func (canvas *Canvas) ToBytes() []byte {
	canvas.lock.RLock()
	defer canvas.lock.RUnlock()

	w := bytes.NewBuffer(nil)
	png.Encode(w, canvas.image)
	return w.Bytes()
}

func (canvas *Canvas) WriteToFile(path string) error {
	canvas.lock.RLock()
	defer canvas.lock.RUnlock()

	f, err := os.Create(path)

	if err != nil {
		return err
	}

	return png.Encode(f, canvas.image)
}

func (canvas *Canvas) Dimensions() (int, int) {
	point := canvas.image.Rect.Max
	return point.X, point.Y
}

func (canvas *Canvas) IsInBounds(x int, y int) bool {
	w, h := canvas.Dimensions()
	return x >= 0 && y >= 0 && x < w && y < h
}

func ReadCanvasFromFile(path string) (*Canvas, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	img_i, err := png.Decode(f)

	if err != nil {
		return nil, err
	}

	img, ok := img_i.(*image.RGBA)

	if !ok {
		return nil, fmt.Errorf("image is not in RGBA format")
	}

	return &Canvas{
		image: img,
	}, nil
}
