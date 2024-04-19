package game

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

type Canvas struct {
	image *image.NRGBA
}

func NewCanvas(width int, height int) *Canvas {
	return &Canvas{
		image.NewNRGBA(image.Rect(0, 0, width, height)),
	}
}

func (canvas *Canvas) Paint(pixel Pixel) {
	canvas.image.Set(pixel.Pos[0], pixel.Pos[1], color.NRGBA{
		R: pixel.Color[0],
		G: pixel.Color[1],
		B: pixel.Color[2],
	})
}

func (canvas *Canvas) ToBytes() []byte {
	w := bytes.NewBuffer(nil)
	png.Encode(w, canvas.image)
	return w.Bytes()
}
