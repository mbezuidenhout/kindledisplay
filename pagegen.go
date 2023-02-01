package main

import (
	"fmt"
	"image"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/mbezuidenhout/kindleland"
	"golang.org/x/exp/shiny/iconvg"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"golang.org/x/exp/slices"
	"golang.org/x/image/draw"
)

type Orientation int

const (
	Landscape Orientation = iota
	Portrait
)

const kindle = true

var (
	AppConfig       Config
	Page            int         = 0
	PageOrientation Orientation = Portrait
	LinkDown        bool        = false
	//lastReconnect   time.Time
	fb *kindleland.FrameBuffer
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

	if kindle {
		fb, err = kindleland.NewFrameBuffer("/dev/fb0", 600, 800)
		if err != nil {
			panic(err)
		}
	}

	if len(AppConfig.Interface) > 0 {
		_, err = net.InterfaceByName(AppConfig.Interface)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Interface '"+AppConfig.Interface+"' not found.")
			panic(err)
		}
	}

	if kindle {
		// Shutdown the kindle framework and powerd
		cmd := exec.Command("/etc/init.d/powerd", "stop")
		cmd.Run()
		cmd = exec.Command("/etc/init.d/framework", "stop")
		cmd.Run()
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
	wifiOff := false
	for {
		t := time.Now()
		select {
		case <-ticker.C:
			minuteNow := t.Minute()
			if minuteOld != minuteNow {
				minuteOld = minuteNow
				pageRefresh(Page, PageOrientation)
				if LinkDown { // && !wifiOff { // && lastReconnect.Add(time.Duration(time.Minute*5)).Before(time.Now()) {
					//lastReconnect = time.Now()
					//exec.Command(fmt.Sprintf("/usr/bin/wpa_cli -i %s disconnect", AppConfig.Interface))
					//cmd := exec.Command("/sbin/udhcpc", "-i", AppConfig.Interface, "reconnect")
					//cmd.Run()
					if !wifiOff {
						cmd := exec.Command("/etc/init.d/wifid", "stop")
						cmd.Run()
						wifiOff = true
					} else {
						cmd := exec.Command("/etc/init.d/wifid", "start")
						cmd.Run()
						cmd = exec.Command("/etc/init.d/wpa_supplicant", "restart")
						cmd.Run()
						wifiOff = false
					}

					//cmd := exec.Command("/etc/init.d/wpa_supplicant", "restart")
					//cmd.Run()
				}
				break
			}
			if AppConfig.Interval > 0 {
				intervalTimer++
				if intervalTimer > AppConfig.Interval*4 { // Tick happens every 250ms
					intervalTimer = 0
					Page++
					if Page >= len(AppConfig.Pages) {
						Page = 0
						if kindle {
							fb.ClearScreen()
						}
					}
					if len(AppConfig.Interface) > 0 {
						// Ignoring the error because it should have been handled in on startup in main()
						defaultInterface, _ := net.InterfaceByName(AppConfig.Interface)
						addrs, err := defaultInterface.Addrs()

						if err != nil || len(addrs) < 1 || (defaultInterface.Flags&net.FlagUp == 0) {
							LinkDown = true
						} else {
							LinkDown = false
						}
					}

					pageRefresh(Page, PageOrientation)
					break
				}
			}
		case <-done:
			fmt.Println("Received done")
			ticker.Stop()

			if kindle {
				// Start the kindle framework and powerd
				cmd := exec.Command("/etc/init.d/powerd", "start")
				cmd.Run()
				cmd = exec.Command("/etc/init.d/framework", "start")
				cmd.Run()
			}

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
		switch {
		case strings.Compare(strings.ToLower(element), "time") == 0:
			DrawTime(kindlectx, blockLayout.Blocks[i].X, blockLayout.Blocks[i].Y, blockLayout.Blocks[i].Width, blockLayout.Blocks[i].Height)
		case strings.Compare(strings.ToLower(element), "datetime") == 0:
			DrawDateTime(kindlectx, blockLayout.Blocks[i].X, blockLayout.Blocks[i].Y, blockLayout.Blocks[i].Width, blockLayout.Blocks[i].Height)
		case strings.Compare(strings.ToLower(element), "date") == 0:
			DrawDate(kindlectx, blockLayout.Blocks[i].X, blockLayout.Blocks[i].Y, blockLayout.Blocks[i].Width, blockLayout.Blocks[i].Height)
		case strings.Compare(strings.ToLower(element), "sun") == 0:
			if AppConfig.Latitude == 0 && AppConfig.Longitude == 0 {
				fmt.Println("latitude and longitude must be set to use block type \"sun\"")
			} else {
				DrawSun(kindlectx, blockLayout.Blocks[i].X, blockLayout.Blocks[i].Y, blockLayout.Blocks[i].Width, blockLayout.Blocks[i].Height)
			}
		case urlregex.MatchString(element):
			DrawFromURL(kindlectx, blockLayout.Blocks[i].X, blockLayout.Blocks[i].Y, blockLayout.Blocks[i].Width, blockLayout.Blocks[i].Height, element)
		default:
			fmt.Printf("Block %d did not match a known format\nBlock types include time, datetime, date, sun and URL to an image", i)
		}
	}

	if LinkDown {
		wifiOff := image.NewAlpha(image.Rect(0, 0, 40, 40))
		var z iconvg.Rasterizer
		z.SetDstImage(wifiOff, wifiOff.Bounds(), draw.Src)
		if err := iconvg.Decode(&z, icons.DeviceSignalWiFiOff, nil); err != nil {
			panic(err)
		}
		invertedWifiOff := imaging.Invert(wifiOff)

		kindlectx.DrawImage(invertedWifiOff, blockLayout.Width-40, 0)
	}

	if kindle {
		var rotatedImage image.Image

		if PageOrientation == Landscape {
			rotatedImage = imaging.Rotate270(kindlectx.Image())
		} else {
			rotatedImage = kindlectx.Image()
		}

		err := fb.ApplyImage(rotatedImage)
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
