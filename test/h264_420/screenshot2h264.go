package main

import (
	"screenshot"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"time"
	// "stringio"
	"reflect"
	"fmt"
	"runtime"
	"codec"
)

func main() {
	threads := runtime.GOMAXPROCS(4)
    threads = runtime.GOMAXPROCS(4)
	fmt.Printf("Server with threads: %v\n", threads)
	
	screenshot.InitConn()

	q := 100
	var nal [][]byte

	img, err := screenshot.CaptureScreen()
	if err != nil {
		panic(err)
	}
	f, err := os.Create("./ss.jpeg")
	if err != nil {
		panic(err)
	}

	fmt.Printf("length: %v\n", len(img.Pix)/4)

	new_img := image.NewYCbCr(img.Rect, image.YCbCrSubsampleRatio420)
	new_img.Y = make([]uint8, len(img.Pix)/4)
	new_img.Cb = make([]uint8, len(img.Pix)/16)
	new_img.Cr = make([]uint8, len(img.Pix)/16)

    // n := 0
	cn := 0
	s := time.Now()
	for y := 0; y < img.Rect.Dy()/2; y += 1 {
		for x := 0; x < img.Rect.Dx()/2; x += 1 {
			x0, y0 := x*2, y*2
			x1, y1 := x*2 + 1, y*2
			x2, y2 := x*2, y*2 + 1
			x3, y3 := x*2 + 1, y*2 + 1

			co0 := img.RGBAAt(x0, y0)
			cy0, cb0, cr0 := color.RGBToYCbCr(co0.R, co0.G, co0.B)
			co1 := img.RGBAAt(x1, y1)
			cy1, cb1, cr1 := color.RGBToYCbCr(co1.R, co1.G, co1.B)
			co2 := img.RGBAAt(x2, y2)
			cy2, cb2, cr2 := color.RGBToYCbCr(co2.R, co2.G, co2.B)
			co3 := img.RGBAAt(x3, y3)
			cy3, cb3, cr3 := color.RGBToYCbCr(co3.R, co3.G, co3.B)

			new_img.Y[x0 + y0 * img.Rect.Dx()] = cy0
			new_img.Y[x1 + y1 * img.Rect.Dx()] = cy1
			new_img.Y[x2 + y2 * img.Rect.Dx()] = cy2
			new_img.Y[x3 + y3 * img.Rect.Dx()] = cy3
			new_img.Cb[cn] = cb0/4 + cb1/4 + cb2/4 + cb3/4
			new_img.Cr[cn] = cr0/4 + cr1/4 + cr2/4 + cr3/4
			cn += 1
		}
	}

	ss := time.Now()
	fmt.Printf("RGB to YCbCr: %v\n", ss.Sub(s))

	// for y := 0; y < img.Rect.Dy(); y += 1 {
	// 	for x := 0; x < img.Rect.Dx(); x += 1 {
	// 		co := img.RGBAAt(x, y)
	// 		y, cb, cr := color.RGBToYCbCr(co.R, co.G, co.B)
	// 		new_img.Y = append(new_img.Y, y)
	// 		new_img.Cb = append(new_img.Cb, cb)
	// 		new_img.Cr = append(new_img.Cr, cr)
	// 	}
	// }

    // err_color := 0
	// for y := 0; y < img.Rect.Dy(); y += 1 {
	// 	for x := 0; x < img.Rect.Dx(); x += 1 {
	// 		// i := img.PixOffset(x, y)
	// 		// fmt.Printf("pix: %v\n", i)
	// 		co := img.RGBAAt(x, y)
	// 		ycbcr := new_img.YCbCrAt(x, y)
	// 		r, g, b := color.YCbCrToRGB(ycbcr.Y, ycbcr.Cb, ycbcr.Cr)
	// 		co2 := color.RGBA{r, g, b, 255}
	// 		if co.R != co2.R || co.G != co2.G || co.B != co2.B {
	// 			fmt.Printf("co: %v, co2: %v\n", co, co2)
	// 			err_color += 1
	// 		}
	// 	}
	// }
	// fmt.Printf("err_color: %v\n", err_color)

	err = jpeg.Encode(f, img, &jpeg.Options{q})
	if err != nil {
		panic(err)
	}

	f.Close()
	
	c, _ := codec.NewH264Encoder(new_img.Rect.Dx(), new_img.Rect.Dy(), 0, 30, 1, 30, 64, image.YCbCrSubsampleRatio420 , "bufsize,0k,0", "pixel_format,yuv420p,0")
    nal = append(nal, c.Header)

    for i := 0; i < 360; i++ {
        p, _ := c.Encode(new_img)
		fmt.Printf("append data: len(%v)\n", len(p.Data))
        if len(p.Data) > 0 {
        	nal = append(nal, p.Data)
        }
    }

    for {
        // flush encoder
        p, err := c.Encode(nil)
        if err != nil {
       		break
        }
        fmt.Printf("p.Data type: %v, len: %v\n", reflect.TypeOf(p.Data), len(p.Data))
        nal = append(nal, p.Data)
    }
    
    f2, err := os.Create("./test.h264")
    if err != nil {
        panic(err)
    }
    
    for _, bytes := range nal {
        n, err := f2.Write(bytes)
        if err != nil {
            panic(err)
        } else {
            fmt.Printf("wrote %d bytes\n", n)
        }
    }
    
    f2.Sync()
	f2.Close()

	dec, err := codec.NewH264Decoder(nal[0])
	for i, n := range nal[1:] {
		img_src, err := dec.Decode(n, image.YCbCrSubsampleRatio420)
		if err == nil {
			fp, _ := os.Create(fmt.Sprintf("./images/dec-%d.jpg", i))
			jpeg.Encode(fp, img_src, nil)
			fp.Close()
			if i == 2 {
				break
			}
		}
	}

	fmt.Printf("rect: %v, dx: %v, dy: %v\n", img.Rect, new_img.Rect.Dx(), new_img.Rect.Dy())
	fmt.Printf("img: Stride: %v\n", img.Stride)
	fmt.Printf("new_img: %v, stringio: %v, %v\n", len(new_img.Y), new_img.YStride, new_img.CStride)
	fmt.Printf("c.W: %v, c.H: %v. c.Header: %v, c.Pixfmt: %v\n", c.W, c.H, c.Header, c.Pixfmt)
    // fmt.Printf("nal: %v\n", nal)
}
