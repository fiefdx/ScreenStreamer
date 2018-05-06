package screenshot

import (
	"fmt"
	"image"
	"time"

	"log"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

var Conn *xgb.Conn

type POS struct {
	X, Y int
}

type SIZE struct {
	W, H int
}

type RESIZE struct {
	W, H int
}

type CAPTURE struct {
	W, H int
	B    *[]byte
}

func InitConn() {
	var err error
	Conn, err = xgb.NewConn()
	if err != nil {
		panic(err)
	}
}

func CloseConn() {
	Conn.Close()
}

func GetActiveWindow() int64 {
	c := Conn

	screen := xproto.Setup(c).DefaultScreen(c)

	aname := "_NET_ACTIVE_WINDOW"
	activeAtom, err := xproto.InternAtom(c, true, uint16(len(aname)), aname).Reply()
	if err != nil {
		return int64(-1)
	}
	reply, err := xproto.GetProperty(c, false, screen.Root, activeAtom.Atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		return int64(-1)
	}
	windowId := xproto.Window(xgb.Get32(reply.Value))
	return int64(windowId)
}

func ScreenRect() (image.Rectangle, error) {
	c := Conn

	screen := xproto.Setup(c).DefaultScreen(c)
	x := screen.WidthInPixels
	y := screen.HeightInPixels

	return image.Rect(0, 0, int(x), int(y)), nil
}

func CaptureWindowMust(pos *POS, size *SIZE, resize *RESIZE, toSBS bool, cursor bool) *image.RGBA {
	img, err := CaptureWindow(pos, size, resize, toSBS, cursor)
	for err != nil {
		img, err = CaptureWindow(pos, size, resize, toSBS, cursor)
		time.Sleep(10 * time.Millisecond)
	}
	return img
}

func CaptureScreenYCbCrMust(pos *POS, size *SIZE, resize *RESIZE, toSBS, cursor, fullScreen bool, numOfRange, windowId int64) *image.YCbCr {
	img, err := CaptureScreenYCbCr444(pos, size, resize, toSBS, cursor, fullScreen, numOfRange, windowId)
	for err != nil {
		img, err = CaptureScreenYCbCr444(pos, size, resize, toSBS, cursor, fullScreen, numOfRange, windowId)
		time.Sleep(10 * time.Millisecond)
	}
	return img
}

func CaptureScreen() (*image.RGBA, error) {
	r, e := ScreenRect()
	if e != nil {
		return nil, e
	}
	return CaptureRect(r)
}

func CaptureScreenYCbCr444(pos *POS, size *SIZE, resize *RESIZE, toSBS, cursor, fullScreen bool, numOfRange, windowId int64) (*image.YCbCr, error) {
	if fullScreen {
		r, e := ScreenRect() // 20us
		if e != nil {
			return nil, e
		}
		return CaptureRectYCbCr444(r, numOfRange)
	} else {
		return CaptureWindowYCbCr(pos, size, resize, toSBS, cursor, numOfRange, windowId)
	}
}

func CaptureRect(rect image.Rectangle) (*image.RGBA, error) {
	c := Conn

	screen := xproto.Setup(c).DefaultScreen(c)
	x, y := rect.Dx(), rect.Dy()
	xImg, err := xproto.GetImage(c, xproto.ImageFormatZPixmap, xproto.Drawable(screen.Root), int16(rect.Min.X), int16(rect.Min.Y), uint16(x), uint16(y), 0xffffffff).Reply()
	if err != nil {
		return nil, err
	}

	data := xImg.Data
	for i := 0; i < len(data); i += 4 {
		data[i], data[i+2], data[i+3] = data[i+2], data[i], 255
	}

	img := &image.RGBA{data, 4 * x, image.Rect(0, 0, x, y)}
	return img, nil
}

func CaptureRectYCbCr444(rect image.Rectangle, numOfRange int64) (*image.YCbCr, error) {
	c := Conn

	t := time.Now()
	screen := xproto.Setup(c).DefaultScreen(c)
	x, y := rect.Dx(), rect.Dy()
	xImg, err := xproto.GetImage(c, xproto.ImageFormatZPixmap, xproto.Drawable(screen.Root), int16(rect.Min.X), int16(rect.Min.Y), uint16(x), uint16(y), 0xffffffff).Reply()
	if err != nil {
		return nil, err
	}

	data := xImg.Data

	tt := time.Now()
	if ImageCache == nil {
		ImageCache = image.NewYCbCr(image.Rect(0, 0, x, y), image.YCbCrSubsampleRatio444)
	}
	ttt := time.Now()

	// CRGBToYCbCr444(data, ImageCache.Y, ImageCache.Cb, ImageCache.Cr)

	// Data <- data
	// Data <- data
	// Data <- data
	// Y <- ImageCache.Y
	// Cb <- ImageCache.Cb
	// Cr <- ImageCache.Cr
	// <-RY
	// <-RCb
	// <-RCr

	lenData := int64(len(data))
	batchSize := lenData / (4 * numOfRange) * 4
	for i := int64(0); i < numOfRange-1; i++ {
		Range <- []int64{i * batchSize, batchSize, int64(ImageCache.Rect.Dx()), int64(ImageCache.Rect.Dx())}
		Data <- data
		Y <- ImageCache.Y
		Cb <- ImageCache.Cb
		Cr <- ImageCache.Cr
	}
	start := (numOfRange - 1) * batchSize
	Range <- []int64{start, lenData - start, int64(ImageCache.Rect.Dx()), int64(ImageCache.Rect.Dx())}
	Data <- data
	Y <- ImageCache.Y
	Cb <- ImageCache.Cb
	Cr <- ImageCache.Cr
	for i := int64(0); i < numOfRange; i++ {
		<-R
	}
	tttt := time.Now()
	log.Println(fmt.Sprintf("Shot: %v, Create: %v, Convert: %v", tt.Sub(t), ttt.Sub(tt), tttt.Sub(ttt)))
	// Shot: 14.734765ms, Create: 108ns, Convert: 9.515677ms

	return ImageCache, nil
}

func CaptureWindowYCbCr(pos *POS, size *SIZE, resize *RESIZE, toSBS bool, cursor bool, numOfRange, windowId int64) (*image.YCbCr, error) {
	c := Conn
	screen := xproto.Setup(c).DefaultScreen(c)

	aname := "_NET_ACTIVE_WINDOW"
	activeAtom, err := xproto.InternAtom(c, true, uint16(len(aname)), aname).Reply()
	if err != nil {
		return nil, fmt.Errorf("error occurred, when xproto.InternAtom 0 err:%v.\n", err)
	}

	reply, err := xproto.GetProperty(c, false, screen.Root, activeAtom.Atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		return nil, fmt.Errorf("error occurred, when xproto.GetProperty 0 err:%v.\n", err)
	}

	var winId xproto.Window
	if windowId != -1 {
		winId = xproto.Window(windowId)
	} else {
		winId = xproto.Window(xgb.Get32(reply.Value))
	}

	ginfo, err := xproto.GetGeometry(c, xproto.Drawable(winId)).Reply()
	if err != nil {
		return nil, err
	}

	width := int(ginfo.Width) - pos.X
	height := int(ginfo.Height) - pos.Y
	if size.W != 0 && size.H != 0 {
		width = size.W
		height = size.H
	}

	xImg, err := xproto.GetImage(c, xproto.ImageFormatZPixmap, xproto.Drawable(winId), int16(pos.X), int16(pos.Y), uint16(width), uint16(height), 0xffffffff).Reply()
	if err != nil {
		return nil, err
	}

	data := xImg.Data

	if !toSBS {
		if ImageCache == nil || (ImageCache.Rect.Dx() != (width-pos.X) || ImageCache.Rect.Dy() != (height-pos.Y)) {
			ImageCache = image.NewYCbCr(image.Rect(pos.X, pos.Y, width, height), image.YCbCrSubsampleRatio444)
		}
	} else {
		if ImageCache == nil || (ImageCache.Rect.Dx() != (2*width-pos.X) || ImageCache.Rect.Dy() != (height-pos.Y)) {
			ImageCache = image.NewYCbCr(image.Rect(pos.X, pos.Y, 2*width, height), image.YCbCrSubsampleRatio444)
		}
	}

	// CRGBToYCbCr444(data, ImageCache.Y, ImageCache.Cb, ImageCache.Cr)
	// Shot: 14.734765ms, Create: 108ns, Convert: 9.515677ms

	lenData := int64(len(data))
	batchSize := (int64(height) / numOfRange) * int64(width) * 4
	for i := int64(0); i < numOfRange-1; i++ {
		Range <- []int64{i * batchSize, batchSize, int64(width), int64(ImageCache.Rect.Dx())}
		Data <- data
		Y <- ImageCache.Y
		Cb <- ImageCache.Cb
		Cr <- ImageCache.Cr
	}
	start := (numOfRange - 1) * batchSize
	Range <- []int64{start, lenData - start, int64(width), int64(ImageCache.Rect.Dx())}
	Data <- data
	Y <- ImageCache.Y
	Cb <- ImageCache.Cb
	Cr <- ImageCache.Cr
	for i := int64(0); i < numOfRange; i++ {
		<-R
	}

	return ImageCache, nil
}

func CaptureWindow(pos *POS, size *SIZE, resize *RESIZE, toSBS bool, cursor bool) (*image.RGBA, error) {
	c := Conn
	screen := xproto.Setup(c).DefaultScreen(c)

	aname := "_NET_ACTIVE_WINDOW"
	activeAtom, err := xproto.InternAtom(c, true, uint16(len(aname)), aname).Reply()
	if err != nil {
		return nil, fmt.Errorf("error occurred, when xproto.InternAtom 0 err:%v.\n", err)
	}

	reply, err := xproto.GetProperty(c, false, screen.Root, activeAtom.Atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		return nil, fmt.Errorf("error occurred, when xproto.GetProperty 0 err:%v.\n", err)
	}
	windowId := xproto.Window(xgb.Get32(reply.Value))

	ginfo, err := xproto.GetGeometry(c, xproto.Drawable(windowId)).Reply()
	if err != nil {
		return nil, err
	}

	width := int(ginfo.Width) - pos.X
	height := int(ginfo.Height) - pos.Y
	if size.W != 0 && size.H != 0 {
		width = size.W
		height = size.H
	}

	xImg, err := xproto.GetImage(c, xproto.ImageFormatZPixmap, xproto.Drawable(windowId), int16(pos.X), int16(pos.Y), uint16(width), uint16(height), 0xffffffff).Reply()
	if err != nil {
		return nil, err
	}

	data := xImg.Data

	var img *image.RGBA
	if toSBS {
		data2 := make([]byte, len(data)*2)
		ImageToRGBASBSLinux(data, data2, width, height)
		img = &image.RGBA{data2, 8 * width, image.Rect(pos.X, pos.Y, width*2, height)}
	} else {
		ImageToRGBALinux(data)
		img = &image.RGBA{data, 4 * width, image.Rect(pos.X, pos.Y, width, height)}
	}
	return img, nil
}
