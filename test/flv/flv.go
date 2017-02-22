package main

import (
	"screenshot"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"time"
	// "stringio"
	// "reflect"
	"fmt"
	"runtime"
	// "codec"
	"flv.go/flv"
)

func RGBAToYCbCr(img *image.RGBA) *image.YCbCr {
    new_img := image.NewYCbCr(img.Rect, image.YCbCrSubsampleRatio444)
	new_img.Y = make([]uint8, len(img.Pix)/4)
	new_img.Cb = make([]uint8, len(img.Pix)/4)
	new_img.Cr = make([]uint8, len(img.Pix)/4)

	n := 0
	s := time.Now()
	for i := 0; i < len(img.Pix); i += 4 {
		y, cb, cr := color.RGBToYCbCr(img.Pix[i], img.Pix[i + 1], img.Pix[i + 2])
		new_img.Y[n] = y
		new_img.Cb[n] = cb
		new_img.Cr[n] = cr
		n += 1
		// new_img.Y = append(new_img.Y, y)
		// new_img.Cb = append(new_img.Cb, cb)
		// new_img.Cr = append(new_img.Cr, cr)
	}
	ss := time.Now()
	fmt.Printf("RGB to YCbCr: %v\n", ss.Sub(s))
	return new_img
}

func main() {
	threads := runtime.GOMAXPROCS(4)
    threads = runtime.GOMAXPROCS(4)
    fmt.Printf("Server with threads: %v\n", threads)

	screenshot.InitConn()

	q := 100
	// var nal [][]byte

	img, err := screenshot.CaptureScreen()
	if err != nil {
		panic(err)
	}
	f, err := os.Create("./ss.jpeg")
	if err != nil {
		panic(err)
	}

	fmt.Printf("length: %v\n", len(img.Pix)/4)

	// new_img := RGBAToYCbCr(img)

	err = jpeg.Encode(f, img, &jpeg.Options{q})
	if err != nil {
		panic(err)
	}

	f.Close()

	// c, _ := codec.NewH264Encoder(new_img.Rect.Dx(), new_img.Rect.Dy(), image.YCbCrSubsampleRatio444, "bufsize,0k,0", "pixel_format,yuv444p,0")
    // nal = append(nal, c.Header)

	flvHeader := new(flv.Header)
	flvHeader.Version = 257
	flvHeader.Body = []byte{70, 76, 86, 1, 1, 0, 0, 0, 9, 0, 0, 0, 0}

	// f3, err := os.Open("./test_good.flv")
	f3, err := os.Open("./test.flv")
	if err != nil {
		panic(err)
	}

	r := flv.NewReader(f3)
	header, err := r.ReadHeader()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		panic(err)
	}
	fmt.Printf("static header: %v\n", flvHeader)
	fmt.Printf("header: %v\n", header)

	// f4, err := os.Create("./out.flv")
	// if err != nil {
	// 	panic(err)
	// }

	// w := flv.NewWriter(f4)
	// w.WriteHeader(flvHeader)

	n := 0
	for {
		frame, err := r.ReadFrame()
		if err != nil {
			fmt.Printf("read frame err: %v\n", err)
			break
		}
		if frame == nil {
			break
		}
		// avcFrame := new(flv.AVCVideoFrame)
		if frame.GetType() == flv.TAG_TYPE_VIDEO {
			switch frame.(type) {
				case flv.AVCVideoFrame:
					avcFrame := frame.(flv.AVCVideoFrame)
					fmt.Printf("AVCFrame:\nStream:%v, Dts:%v, Type:%v, Flavor:%v, Position:%v, PrevTagSize:%v, CodecId:%v, Width:%v, Height:%v, PacketType:%v\n",
					avcFrame.Stream, avcFrame.Dts, avcFrame.Type, avcFrame.Flavor, avcFrame.Position, 
					avcFrame.PrevTagSize, avcFrame.CodecId, avcFrame.Width, avcFrame.Height, avcFrame.PacketType,
					)
					fmt.Printf("AVCFrame:\nlen(Body):%v\n", len(avcFrame.Body))
				case flv.VideoFrame:
					avcFrame := frame.(flv.VideoFrame)
					fmt.Printf("AVCFrame:\nStream:%v, Dts:%v, Type:%v, Flavor:%v, Position:%v, PrevTagSize:%v, CodecId:%v, Width:%v, Height:%v, PacketType:%v\n",
					avcFrame.Stream, avcFrame.Dts, avcFrame.Type, avcFrame.Flavor, avcFrame.Position, 
					avcFrame.PrevTagSize, avcFrame.CodecId, avcFrame.Width, avcFrame.Height, 0,
					)
					fmt.Printf("AVCFrame:\nlen(Body):%v\n", len(avcFrame.Body))
			}
			
		}
		
		// e := w.WriteFrame(frame)
		// if e != nil {
		// 	fmt.Printf("write frame err: %v\n", e)
		// 	break
		// }
		// fmt.Printf("frame.TagType: %v, %v\n", frame.GetType(), reflect.TypeOf(frame))
		n += 1
	}
	fmt.Printf("frames: %v\n", n)


    // for i := 0; i < 3000; i++ {
	// 	img, err := screenshot.CaptureScreen()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	new_img := RGBAToYCbCr(img)
    //     p, _ := c.Encode(new_img)
	// 	fmt.Printf("append data(%d): len(%v)\n", i, len(p.Data))
    //     if len(p.Data) > 0 {
    //     	nal = append(nal, p.Data)
    //     }
    // }

    // i := 0
    // for {
    //     // flush encoder
    //     p, err := c.Encode(nil)
    //     if err != nil {
	// 		fmt.Printf("p.Data(%d) err: %v\n", i, err)
    //    		break
    //     }
    //     fmt.Printf("p.Data(%d) type: %v, len: %v\n", i, reflect.TypeOf(p.Data), len(p.Data))
    //     nal = append(nal, p.Data)
	// 	i += 1
    // }
    
    // f2, err := os.Create("./test.h264")
    // if err != nil {
    //     panic(err)
    // }
    
    // for _, bytes := range nal {
    //     n, err := f2.Write(bytes)
    //     if err != nil {
    //         panic(err)
    //     } else {
    //         fmt.Printf("wrote %d bytes\n", n)
    //     }
    // }
    
    // f2.Sync()
	// f2.Close()

	// dec, err := codec.NewH264Decoder(nal[0])
	// for i, n := range nal[1:] {
	// 	img_src, err := dec.Decode(n, image.YCbCrSubsampleRatio444)
	// 	if err == nil {
	// 		fp, _ := os.Create(fmt.Sprintf("./images/dec-%d.jpg", i))
	// 		jpeg.Encode(fp, img_src, nil)
	// 		fp.Close()
	// 		if i == 2 {
	// 			break
	// 		}
	// 	}
	// }

	// fmt.Printf("rect: %v, dx: %v, dy: %v\n", img.Rect, new_img.Rect.Dx(), new_img.Rect.Dy())
	// fmt.Printf("img: Stride: %v\n", img.Stride)
	// fmt.Printf("new_img: %v, stringio: %v, %v\n", len(new_img.Y), new_img.YStride, new_img.CStride)
	// fmt.Printf("c.W: %v, c.H: %v. c.Header: %v, c.Pixfmt: %v\n", c.W, c.H, c.Header, c.Pixfmt)

	screenshot.CloseConn()
}
