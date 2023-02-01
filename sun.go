package main

import (
	"fmt"
	"image"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/sixdouglas/suncalc"
	"golang.org/x/exp/shiny/iconvg"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"golang.org/x/image/draw"
)

func DrawSun(ctx *gg.Context, x, y, width, height int) {
	now := time.Now()
	times := suncalc.GetTimes(now, float64(AppConfig.Latitude), float64(AppConfig.Longitude))
	sunrise := times[suncalc.Sunrise]
	fmt.Println(sunrise)
	// If the block is higher than it is wide put "sun rise" at the top and "sun down" below
	sun := image.NewAlpha(image.Rect(0, 0, width, height))
	var z iconvg.Rasterizer
	z.SetDstImage(sun, sun.Bounds(), draw.Src)
	if err := iconvg.Decode(&z, icons.ImageWBSunny, nil); err != nil {
		panic(err)
	}
	invertedSun := imaging.Invert(sun)
	ctx.DrawImage(invertedSun, x, y)
}

func DrawMoonPhase(ctx *gg.Context, x, y, width, height int) {
	now := time.Now()
	phase := suncalc.GetMoonIllumination(now)
	fmt.Println(phase) // 0 = New Moon, 0.25 = First Quarter, 0.5 = Full Moon, 0.75 = Last Quarter
}

func DrawMoon(ctx *gg.Context, x, y, width, height int) {
	now := time.Now()
	moonrise := suncalc.GetMoonTimes(now, float64(AppConfig.Latitude), float64(AppConfig.Longitude), false)
	if moonrise.AlwaysUp {
		fmt.Println("Always above horizon for this day")
	}
	if moonrise.AlwaysDown {
		fmt.Println("Always below horizon for this day")
	}

}
