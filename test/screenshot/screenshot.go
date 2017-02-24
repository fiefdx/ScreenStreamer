package main

import (
	"fmt"
	"image/jpeg"
	"os"
	"runtime"
	"screenshot"
	"stringio"
	"time"
	// "github.com/lazywei/go-opencv/opencv"
	// "github.com/anthonynsimon/bild/imgio"
	ljpeg "github.com/pixiv/go-libjpeg/jpeg"
)

func main() {
	threads := runtime.GOMAXPROCS(4)
	threads = runtime.GOMAXPROCS(4)
	fmt.Printf("Server with threads: %v\n", threads)

	screenshot.InitConn()

	s := time.Now()
	i := 0
	q := 100
	for i < 100 {
		t := time.Now()
		img, err := screenshot.CaptureWindow(&screenshot.POS{0, 0}, &screenshot.SIZE{0, 0}, &screenshot.RESIZE{0, 0}, false, false)
		// img, err := screenshot.CaptureScreen()
		// c, err := screenshot.CaptureWindowByte(&screenshot.POS{0, 0}, &screenshot.SIZE{0, 0})
		if err != nil {
			panic(err)
		}
		// img := screenshot.CaptureWindowImage(c)
		tt := time.Now()
		sio := stringio.New()
		// fmt.Printf("test\n")
		// err = jpeg.Encode(sio, img, &jpeg.Options{q})
		err = ljpeg.Encode(sio, img, &ljpeg.EncoderOptions{Quality: q})
		// err = imgio.Encode(sio, img, 0)
		// iplimg := opencv.FromImageUnsafe(img)
		// iplimg
		sio.Seek(0, 0)
		out := sio.GetValueString()
		ttt := time.Now()
		fmt.Printf("Use Time: %v, %v\n", tt.Sub(t), ttt.Sub(tt))
		fmt.Printf("Len: %d\n", len(out))
		if err != nil {
			panic(err)
		}
		i += 1
	}
	ss := time.Now()
	fmt.Printf("Total: %.3ffps, %.3fs\n", 100/ss.Sub(s).Seconds(), ss.Sub(s).Seconds())

	img, err := screenshot.CaptureScreen()
	if err != nil {
		panic(err)
	}
	f, err := os.Create("./ss.jpeg")
	if err != nil {
		panic(err)
	}
	err = jpeg.Encode(f, img, &jpeg.Options{q})
	// err = imgio.Encode(f, img, 0)
	if err != nil {
		panic(err)
	}
	f.Close()

	img, err = screenshot.CaptureWindow(&screenshot.POS{0, 0}, &screenshot.SIZE{0, 0}, &screenshot.RESIZE{0, 0}, false, false)
	if err != nil {
		panic(err)
	}
	// c, err := screenshot.CaptureWindowByte(&screenshot.POS{0, 0}, &screenshot.SIZE{0, 0})
	// if err != nil {
	// 	panic(err)
	// }
	// img = screenshot.CaptureWindowImage(c)
	f, err = os.Create("./wss.jpeg")
	if err != nil {
		panic(err)
	}
	err = jpeg.Encode(f, img, &jpeg.Options{q})
	if err != nil {
		panic(err)
	}
	f.Close()

	screenshot.CloseConn()
}
