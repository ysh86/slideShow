package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"

	"github.com/ysh86/slideShow"
)

// Default window size
const (
	winWidth  = 1920
	winHeight = 1280
)

// UI colors
var backGroundColor = color.Gray{32}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	// feed
	src := []string{
		"https://upload.wikimedia.org/wikipedia/commons/2/2c/Rotating_earth_%28large%29.gif",
		"0",
		"https://upload.wikimedia.org/wikipedia/commons/6/6b/Phalaenopsis_JPEG.jpg",
		"https://upload.wikimedia.org/wikipedia/commons/4/47/PNG_transparency_demonstration_1.png",
		"https://upload.wikimedia.org/wikipedia/commons/b/b2/Vulphere_WebP_OTAGROOVE_demonstration_2.webp",
		"1",
	}

	// src images
	loader, err := slideShow.NewAsyncLoader()
	if err != nil {
		log.Fatal(err)
	}
	loader.SetList(src)
	log.Println("Start loader")

	// renderer
	renderer, err := slideShow.NewNoEffectRenderer()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Start renderer")

	// UI loop
	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(&screen.NewWindowOptions{
			Title:  "Image viewer",
			Width:  winWidth,  // / 2, // TODO: HiDPI
			Height: winHeight, // / 2, // TODO: HiDPI
		})
		if err != nil {
			log.Fatal(err)
		}
		defer w.Release()

		// start loader
		done := make(chan interface{})
		loader.Run(done, w)
		defer func() { close(done); log.Println("Done loader") }()

		// init renderer
		err = renderer.Init(s, image.Pt(winWidth, winHeight), backGroundColor)
		if err != nil {
			log.Fatal(err)
		}
		defer func() { renderer.Release(); log.Println("Done renderer") }()

		var sz size.Event
		for {
			e := w.NextEvent()

			// This print message is to help programmers learn what events this
			// example program generates. A real program shouldn't print such
			// messages; they're not important to end users.
			format := "got %#v\n"
			if _, ok := e.(fmt.Stringer); ok {
				format = "got %v\n"
			}
			log.Printf(format, e)

			switch e := e.(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}

			case key.Event:
				if e.Code == key.CodeEscape {
					return
				}

			case paint.Event:
				select {
				case cur, ok := <-loader.Cur:
					if ok {
						renderer.Render(cur)
					} else {
						// EOF
						renderer.Render(nil)
						log.Println("paint EOF")
					}
				default:
					// nothing to do
				}
				renderer.Swap(w)

			case size.Event:
				sz = e
				log.Printf("size: %#v, PointsPerInch: %#v\n", sz, unit.PointsPerInch)
				// TODO: resize canvas in renderer
				w.Send(paint.Event{})

			case error:
				log.Print(e)
			}
		}
	})

}