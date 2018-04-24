package main

import (
	"strings"
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
	"codec"
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

func RGBAToYCbCr420(img *image.RGBA) *image.YCbCr {
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
	return new_img
}

func ScreenShotToNALU(c *codec.H264Encoder, frameNum int) []codec.H264Out {
	var nal []codec.H264Out
	// finish := 0
	for i := 0; i < frameNum; i++ {
		img, err := screenshot.CaptureScreen()
		if err != nil {
			panic(err)
		}
		new_img := RGBAToYCbCr420(img)
        p, _ := c.Encode(new_img)
		fmt.Printf("append data(%d): len(%v), key: %v\n", i, len(p.Data), p.Key)
        if len(p.Data) > 0 {
        	nal = append(nal, p)
			// finish += 1
        }
    }

    // i := 0
    // for {
    //     // flush encoder
	// 	if finish == frameNum {
	// 		break
	// 	}
    //     p, err := c.Encode(nil)
    //     if err != nil {
	// 		fmt.Printf("flush encoder p.Data(%d) err: %v, key: %v\n", i, err, p.Key)
    //    		continue
    //     }
    //     fmt.Printf("flush encoder p.Data(%d) type: %v, len: %v, key: %v\n", i, reflect.TypeOf(p.Data), len(p.Data), p.Key)
    //     nal = append(nal, p.Data)
	// 	finish += 1
	// 	i += 1
    // }

	return nal
}

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

	new_img := RGBAToYCbCr(img)

	err = jpeg.Encode(f, img, &jpeg.Options{q})
	if err != nil {
		panic(err)
	}

	f.Close()

	f4, err := os.Create("./out.flv")
	if err != nil {
		panic(err)
	}

	flvHeader := new(flv.Header)
	flvHeader.Version = 257
	flvHeader.Body = []byte{70, 76, 86, 1, 1, 0, 0, 0, 9, 0, 0, 0, 0}

	c, _ := codec.NewH264Encoder(new_img.Rect.Dx(), new_img.Rect.Dy(), 0, 30, 1, 30, 64, image.YCbCrSubsampleRatio420 , "bufsize,0k,0", "pixel_format,yuv420p,0")
    nal = append(nal, c.Header)

	w := flv.NewWriter(f4)
	w.WriteHeader(flvHeader)

	avcFrame := new(flv.AVCVideoFrame)
	avcFrame.VideoFrame = new(flv.VideoFrame)
	avcFrame.VideoFrame.CFrame = new(flv.CFrame)
	avcFrame.Stream = 0
	avcFrame.Dts = uint32(0)
	avcFrame.Type = flv.TAG_TYPE_VIDEO
	avcFrame.Flavor = flv.KEYFRAME
	avcFrame.CodecId = flv.VIDEO_CODEC_AVC // flv.VIDEO_CODEC_AVC
	avcFrame.Width = 1920
	avcFrame.Height = 1080
	avcFrame.PacketType = flv.VIDEO_AVC_SEQUENCE_HEADER
	avcFrame.Body = make([]byte, 0)
	avcFrame.Body = append(avcFrame.Body, (uint8(flv.VIDEO_FRAME_TYPE_KEYFRAME) << 4) | uint8(avcFrame.CodecId))
	avcFrame.Body = append(avcFrame.Body, uint8(flv.VIDEO_AVC_SEQUENCE_HEADER))

    SPS := make([]byte, 0)
	PPS := make([]byte, 0)
    headerS := fmt.Sprintf("%X", c.Header)
    i := strings.Index(headerS, "0000000167")
	if i != -1 {
		spsTmp := c.Header[i/2+4:]
		spsTmpS := fmt.Sprintf("%X", spsTmp)
    	j := strings.Index(spsTmpS, "0000000168")
		if j != -1{
			sps := spsTmp[:j/2]
			pps := spsTmp[j/2+4:]
			SPS = append(SPS, sps...)
			PPS = append(PPS, pps...)
		}
	}

	fmt.Printf("SPS: % X\n", SPS)
	fmt.Printf("PPS: % X\n", PPS)

	avcFrame.Body = append(avcFrame.Body, 0)
	avcFrame.Body = append(avcFrame.Body, 0)
	avcFrame.Body = append(avcFrame.Body, 0)

	avcFrame.Body = append(avcFrame.Body, 0x01)
	avcFrame.Body = append(avcFrame.Body, 0x64)
	avcFrame.Body = append(avcFrame.Body, 0x00)
	avcFrame.Body = append(avcFrame.Body, 0x0A)
	avcFrame.Body = append(avcFrame.Body, 0xFF) // reserved | NALULengthSizeMinusOne
	avcFrame.Body = append(avcFrame.Body, 0xE1) // reserved | number of sps NALUs
    lenSPS := len(SPS)
	lenPPS := len(PPS)
	avcFrame.Body = append(avcFrame.Body, uint8(lenSPS >> 8) & 0xFF)
	avcFrame.Body = append(avcFrame.Body, uint8(lenSPS & 0xFF))
	avcFrame.Body = append(avcFrame.Body, SPS...)
	avcFrame.Body = append(avcFrame.Body, 0x01)
	avcFrame.Body = append(avcFrame.Body, uint8(lenPPS >> 8 & 0xFF))
	avcFrame.Body = append(avcFrame.Body, uint8(lenPPS & 0xFF))
	avcFrame.Body = append(avcFrame.Body, PPS...)

	e := w.WriteFrame(avcFrame)
	if e != nil {
		fmt.Printf("write frame err: %v\n", e)
	}

	t := time.Now()
	dts := 66
    for i := 0; i < 3000; i++ {
		nals := ScreenShotToNALU(c, 1)
		for _, data := range nals {
			nal = append(nal, data.Data)

			avcFrame := new(flv.AVCVideoFrame)
			avcFrame.VideoFrame = new(flv.VideoFrame)
			avcFrame.VideoFrame.CFrame = new(flv.CFrame)
			// if i == 0 {
				avcFrame.Stream = 0
				avcFrame.Dts = uint32(i*dts)
				avcFrame.Type = flv.TAG_TYPE_VIDEO
				avcFrame.Flavor = flv.KEYFRAME
				avcFrame.CodecId = flv.VIDEO_CODEC_AVC // flv.VIDEO_CODEC_AVC
				avcFrame.Width = 1920
				avcFrame.Height = 1080
				avcFrame.PacketType = flv.VIDEO_AVC_NALU
				avcFrame.Body = make([]byte, 0)
				if data.Key {
					avcFrame.Body = append(avcFrame.Body, (uint8(flv.VIDEO_FRAME_TYPE_KEYFRAME) << 4) | uint8(avcFrame.CodecId))
				} else {
					avcFrame.Body = append(avcFrame.Body, (uint8(flv.VIDEO_FRAME_TYPE_INTER_FRAME) << 4) | uint8(avcFrame.CodecId))
				}
				avcFrame.Body = append(avcFrame.Body, uint8(flv.VIDEO_AVC_NALU))

				avcFrame.Body = append(avcFrame.Body, 0)
				avcFrame.Body = append(avcFrame.Body, 0)
				avcFrame.Body = append(avcFrame.Body, 0)
				if data.Data[0] == 0 && data.Data[1] == 0 && data.Data[2] == 0 && data.Data[3] == 1 {
					lenBody := len(data.Data[4:])
					avcFrame.Body = append(avcFrame.Body, uint8(lenBody >> 24) & 0xFF)
					avcFrame.Body = append(avcFrame.Body, uint8(lenBody >> 16) & 0xFF)
					avcFrame.Body = append(avcFrame.Body, uint8(lenBody >> 8) & 0xFF)
					avcFrame.Body = append(avcFrame.Body, uint8(lenBody & 0xFF))

					avcFrame.Body = append(avcFrame.Body, data.Data[4:]...)
				} else if data.Data[0] == 0 && data.Data[1] == 0 && data.Data[2] == 1 {
					lenBody := len(data.Data[3:])
					avcFrame.Body = append(avcFrame.Body, uint8(lenBody >> 24) & 0xFF)
					avcFrame.Body = append(avcFrame.Body, uint8(lenBody >> 16) & 0xFF)
					avcFrame.Body = append(avcFrame.Body, uint8(lenBody >> 8) & 0xFF)
					avcFrame.Body = append(avcFrame.Body, uint8(lenBody & 0xFF))

					avcFrame.Body = append(avcFrame.Body, data.Data[3:]...)
				}
			// } else {
			// 	avcFrame.Stream = 0
			// 	avcFrame.Dts = uint32(i*40)
			// 	avcFrame.Type = flv.TAG_TYPE_VIDEO
			// 	avcFrame.Flavor = flv.FRAME
			// 	avcFrame.CodecId = flv.VIDEO_CODEC_AVC // flv.VIDEO_CODEC_AVC
			// 	avcFrame.Width = 1920
			// 	avcFrame.Height = 1080
			// 	avcFrame.PacketType = flv.VIDEO_AVC_NALU
			// 	avcFrame.Body = make([]byte, 0)
			// 	avcFrame.Body = append(avcFrame.Body, (uint8(flv.VIDEO_FRAME_TYPE_INTER_FRAME) << 4) | uint8(avcFrame.CodecId))
			// 	avcFrame.Body = append(avcFrame.Body, uint8(flv.VIDEO_AVC_NALU))

			// 	avcFrame.Body = append(avcFrame.Body, 0)
			// 	avcFrame.Body = append(avcFrame.Body, 0)
			// 	avcFrame.Body = append(avcFrame.Body, 0)
			// 	if data.Data[0] == 0 && data.Data[1] == 0 && data.Data[2] == 0 && data.Data[3] == 1 {
			// 		lenBody := len(data.Data[4:])
			// 		avcFrame.Body = append(avcFrame.Body, uint8(lenBody >> 24) & 0xFF)
			// 		avcFrame.Body = append(avcFrame.Body, uint8(lenBody >> 16) & 0xFF)
			// 		avcFrame.Body = append(avcFrame.Body, uint8(lenBody >> 8) & 0xFF)
			// 		avcFrame.Body = append(avcFrame.Body, uint8(lenBody & 0xFF))

			// 		avcFrame.Body = append(avcFrame.Body, data.Data[4:]...)
			// 	} else if data.Data[0] == 0 && data.Data[1] == 0 && data.Data[2] == 1 {
			// 		lenBody := len(data.Data[3:])
			// 		avcFrame.Body = append(avcFrame.Body, uint8(lenBody >> 24) & 0xFF)
			// 		avcFrame.Body = append(avcFrame.Body, uint8(lenBody >> 16) & 0xFF)
			// 		avcFrame.Body = append(avcFrame.Body, uint8(lenBody >> 8) & 0xFF)
			// 		avcFrame.Body = append(avcFrame.Body, uint8(lenBody & 0xFF))

			// 		avcFrame.Body = append(avcFrame.Body, data.Data[3:]...)
			// 	}
			// }

			e := w.WriteFrame(avcFrame)
			if e != nil {
				fmt.Printf("write frame err: %v\n", e)
				break
			}
		}
	}
	tt := time.Now()
	fmt.Printf("use time: %v\n", tt.Sub(t))
    
    f2, err := os.Create("./test.h264")
    if err != nil {
        panic(err)
    }
    
    for i, bytes := range nal {
		if i < 1 {
			if len(bytes) < 50 {
				fmt.Printf("first frame:\n{[% X]}\n", bytes)
			} else {
				fmt.Printf("first frame:\n{[% X]}\n", bytes)
			}
		}
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

	screenshot.CloseConn()
}
