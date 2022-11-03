package main

import (
	"image"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"golang.org/x/exp/shiny/iconvg"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"golang.org/x/image/draw"
)

func DrawSun(ctx *gg.Context, x, y, width, height int) {
	// now := time.Now()
	// times := suncalc.GetTimes(now, float64(AppConfig.Latitude), float64(AppConfig.Longitude))
	// sunrise := times[suncalc.Sunrise]
	// If the block is higher than it is high put sun up at the top and sun down below
	sun := image.NewAlpha(image.Rect(0, 0, width, height))
	var z iconvg.Rasterizer
	z.SetDstImage(sun, sun.Bounds(), draw.Src)
	if err := iconvg.Decode(&z, icons.ImageWBSunny, nil); err != nil {
		panic(err)
	}
	invertedSun := imaging.Invert(sun)
	ctx.DrawImage(invertedSun, x, y)
}
