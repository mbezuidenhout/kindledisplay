package main

import (
	"fmt"
	"image"
	"os"
	"os/signal"
	"regexp"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/mbezuidenhout/kindleland"
	"golang.org/x/exp/slices"
)

type Orientation int

const (
	Landscape Orientation = iota
	Portrait
)

const kindle = false

var (
	AppConfig       Config
	Page            int         = 0
	PageOrientation Orientation = Portrait
)

func main() {

	// Read the config file
	configFile := "config.yml"
	if slices.Contains(os.Args, "-c") {
		pos := slices.Index(os.Args, "-c")
		configFile = os.Args[pos+1]
	}
	config, err := NewConfig(configFile)
	AppConfig = *config
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if strings.Compare(strings.ToLower(AppConfig.Orientation), "landscape") == 0 {
		PageOrientation = Landscape
	}

	ticker := time.NewTicker(time.Millisecond * 250)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool, 1)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	if kindle {
		fiveway, err := kindleland.NewKeyboardListener("/dev/input/event1")
		if err != nil {
			panic(err)
		}
		go func() {
			for kevent := range fiveway {
				if kevent.Type == kindleland.KeyDown {
					if kevent.Key == 105 || kevent.Key == 106 {
						if PageOrientation == Landscape {
							PageOrientation = Portrait
						} else {
							PageOrientation = Landscape
						}
					} else if kevent.Key == 103 {
						Page++
					} else if kevent.Key == 108 {
						Page--
					}
					if Page >= len(AppConfig.Pages) {
						Page = 0
					}
					if Page < 0 {
						Page = len(AppConfig.Pages) - 1
					}
					pageRefresh(Page, PageOrientation)
				}
			}
		}()
	}

	intervalTimer := 0

	minuteOld := -1
	for {
		t := time.Now()
		select {
		case <-ticker.C:
			minuteNow := t.Minute()
			if minuteOld != minuteNow {
				minuteOld = minuteNow
				if minuteNow == 0 {
					debug.FreeOSMemory()
				}
				pageRefresh(Page, PageOrientation)
				break
			}
			if AppConfig.Interval > 0 {
				intervalTimer++
				if intervalTimer > AppConfig.Interval*4 { // Tick happens every 250ms
					intervalTimer = 0
					Page++
					if Page >= len(AppConfig.Pages) {
						Page = 0
					}
					pageRefresh(Page, PageOrientation)
					break
				}
			}
		case <-done:
			fmt.Println("Received done")
			ticker.Stop()
			return
		}
	}
}

func pageRefresh(PageNr int, PageOrientation Orientation) {
	if !kindle {
		fmt.Printf("Page %d has %d blocks\n", PageNr, len(AppConfig.Pages[PageNr].Blocks))
	}

	blockLayout := NewBlockLayout(len(AppConfig.Pages[PageNr].Blocks), PageOrientation)

	urlregex := regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z0-9]{2,4}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`)

	kindlectx := gg.NewContext(blockLayout.Width, blockLayout.Height)
	kindlectx.SetRGB(1, 1, 1)
	kindlectx.Clear()
	for i, element := range AppConfig.Pages[PageNr].Blocks {
		if strings.Compare(strings.ToLower(element), "time") == 0 {
			DrawTime(kindlectx, blockLayout.Blocks[i].X, blockLayout.Blocks[i].Y, blockLayout.Blocks[i].Width, blockLayout.Blocks[i].Height)
		} else if strings.Compare(strings.ToLower(element), "datetime") == 0 {
			DrawDateTime(kindlectx, blockLayout.Blocks[i].X, blockLayout.Blocks[i].Y, blockLayout.Blocks[i].Width, blockLayout.Blocks[i].Height)
		} else if strings.Compare(strings.ToLower(element), "date") == 0 {
			DrawDate(kindlectx, blockLayout.Blocks[i].X, blockLayout.Blocks[i].Y, blockLayout.Blocks[i].Width, blockLayout.Blocks[i].Height)
		} else if urlregex.MatchString(element) {
			DrawFromURL(kindlectx, blockLayout.Blocks[i].X, blockLayout.Blocks[i].Y, blockLayout.Blocks[i].Width, blockLayout.Blocks[i].Height, element)
		} else {
			fmt.Printf("Block %d did not match a known format\n", i)
		}
	}

	if kindle {
		fb, err := kindleland.NewFrameBuffer("/dev/fb0", 600, 800)
		if err != nil {
			panic(err)
		}

		var rotatedImage image.Image

		if PageOrientation == Landscape {
			rotatedImage = imaging.Rotate270(kindlectx.Image())
		} else {
			rotatedImage = kindlectx.Image()
		}

		err = fb.ApplyImage(rotatedImage)
		if err != nil {
			panic(err)
		}

		err = fb.UpdateScreen()
		if err != nil {
			panic(err)
		}
	} else {
		kindlectx.SavePNG("page.png")
	}
}
