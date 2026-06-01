package appicon

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"sync"

	"fyne.io/fyne/v2"
)

var (
	once sync.Once
	res  fyne.Resource
)

// Resource returns a monochrome icon used by the app, window and About dialog.
func Resource() fyne.Resource {
	once.Do(func() {
		const size = 128
		img := image.NewNRGBA(image.Rect(0, 0, size, size))

		transparent := color.NRGBA{0, 0, 0, 0}
		white := color.NRGBA{255, 255, 255, 255}
		black := color.NRGBA{0, 0, 0, 255}

		draw.Draw(img, img.Bounds(), &image.Uniform{transparent}, image.Point{}, draw.Src)

		// Outer rounded-like plate.
		fillRect(img, 8, 8, 112, 112, white)
		fillRect(img, 12, 12, 104, 104, black)

		// 1980s-style computer: monitor + keyboard.
		drawRect(img, 18, 20, 58, 36, white)
		fillRect(img, 23, 25, 48, 26, black)
		drawHLine(img, 27, 70, 40, white)
		fillRect(img, 22, 61, 50, 15, white)
		fillRect(img, 24, 63, 46, 11, black)
		for y := 64; y <= 72; y += 4 {
			drawHLine(img, 26, 68, y, white)
		}

		// Database cylinder.
		drawRect(img, 80, 32, 30, 50, white)
		drawHLine(img, 84, 106, 36, white)
		drawHLine(img, 84, 106, 46, white)
		drawHLine(img, 84, 106, 56, white)
		drawHLine(img, 84, 106, 66, white)
		drawHLine(img, 84, 106, 76, white)

		// Link line between computer and database.
		drawHLine(img, 74, 79, 49, white)

		var b bytes.Buffer
		_ = png.Encode(&b, img)
		res = &fyne.StaticResource{StaticName: "msxdbdown-icon.png", StaticContent: b.Bytes()}
	})

	return res
}

func fillRect(img *image.NRGBA, x, y, w, h int, c color.NRGBA) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			img.SetNRGBA(xx, yy, c)
		}
	}
}

func drawRect(img *image.NRGBA, x, y, w, h int, c color.NRGBA) {
	for xx := x; xx < x+w; xx++ {
		img.SetNRGBA(xx, y, c)
		img.SetNRGBA(xx, y+h-1, c)
	}
	for yy := y; yy < y+h; yy++ {
		img.SetNRGBA(x, yy, c)
		img.SetNRGBA(x+w-1, yy, c)
	}
}

func drawHLine(img *image.NRGBA, x1, x2, y int, c color.NRGBA) {
	for x := x1; x <= x2; x++ {
		img.SetNRGBA(x, y, c)
	}
}
