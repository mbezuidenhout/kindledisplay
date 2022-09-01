package main

import (
	"fmt"
	"image"
	"net/http"
	"time"

	"github.com/fogleman/gg"
	"golang.org/x/image/draw"
)

type Image struct {
	Image     image.Image
	ValidTill time.Time
}

var cache map[string]Image = make(map[string]Image)

func UrlImageFromCache(url string) image.Image {
	if k, ok := cache[url]; ok {
		if time.Now().Before(k.ValidTill) {
			return k.Image
		}
		go func() {
			res, err := http.Get(url)
			if err != nil || res.StatusCode != 200 {
				fmt.Printf("Could not get image from %s\n", url)
				return
			}
			defer res.Body.Close()
			src, _, err := image.Decode(res.Body)
			if err != nil {
				fmt.Printf("Could not decode %s as an image\n", url)
				return
			}
			if _, ok := cache[url]; !ok {
				cache[url] = Image{Image: src, ValidTill: time.Now().Add(time.Duration(time.Second * time.Duration(AppConfig.CacheTimeout)))}
			}
			val := cache[url]
			val.ValidTill = time.Now().Add(time.Duration(time.Second * time.Duration(AppConfig.CacheTimeout)))
			val.Image = src

		}()
		return k.Image
	}

	res, err := http.Get(url)
	if err != nil || res.StatusCode != 200 {
		fmt.Printf("Could not get image from %s\n", url)
		return image.NewGray(image.Rect(0, 0, 0, 0))
	}
	defer res.Body.Close()
	src, _, err := image.Decode(res.Body)
	if err != nil {
		fmt.Printf("Could not decode %s as an image\n", url)
		return image.NewGray(image.Rect(0, 0, 0, 0))
	}
	if _, ok := cache[url]; !ok {
		cache[url] = Image{Image: src, ValidTill: time.Now().Add(time.Duration(time.Second * time.Duration(AppConfig.CacheTimeout)))}
	}
	val := cache[url]
	val.ValidTill = time.Now().Add(time.Duration(time.Second * time.Duration(AppConfig.CacheTimeout)))
	val.Image = src

	return val.Image
}

func DrawFromURL(ctx *gg.Context, x, y, width, height int, url string) {
	src := UrlImageFromCache(url)

	dst := image.NewGray(image.Rect(0, 0, width, height))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	ctx.DrawImage(dst, x, y)
}