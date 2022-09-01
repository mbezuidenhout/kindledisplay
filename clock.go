package main

import (
	"time"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

type DrawArea struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

func getFontSize(ctx *gg.Context, width int) int {
	// Font cannot be much larger than 220px otherwise it runs out of memory
	size := 100
	switch {
	case width > 600:
		size = 250
	case width > 400:
		size = 150
	case width > 300:
		size = 130
	}
	return size
}

func DrawTimeFormat(format string, ctx *gg.Context, x, y, width, height int) (w, h float64) {
	t := time.Now()
	timeString := t.Format(format)
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}

	size := getFontSize(ctx, width)

	if len(format) > 5 {
		size = size / 5 * 2
	}

	face := truetype.NewFace(font, &truetype.Options{Size: float64(size), GlyphCacheEntries: 2})
	defer face.Close()
	ctx.SetRGB(0, 0, 0)
	ctx.SetFontFace(face)
	ctx.DrawStringAnchored(timeString, float64(x+width/2), float64(y+height/2), 0.5, 0.5)
	return ctx.MeasureString(timeString)
}

func DrawTime(ctx *gg.Context, x, y, width, height int) {
	DrawTimeFormat("15:04", ctx, x, y, width, height)
}

func DrawDate(ctx *gg.Context, x, y, width, height int) {
	DrawTimeFormat("Wed Feb 2", ctx, x, y, width, height)
}

func DrawDateTime(ctx *gg.Context, x, y, width, height int) {
	_, h := DrawTimeFormat("15:04", ctx, x, y, width, height)
	ctx.SetRGB(1, 1, 1)
	ctx.Clear()
	DrawTimeFormat("15:04", ctx, x, y-int(h)/2, width, height)
	DrawTimeFormat("Wed Jan 2", ctx, x, y+int(h)/2, width, height)
}
