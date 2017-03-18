package screenshot

import (
	"fmt"
	"image"
	"log"
	"reflect"
	"time"
	"unsafe"
	"w32"
)

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
}

func CloseConn() {
}

func GetActiveWindow() int64 {
	hWND := w32.GetForegroundWindow()
	if hWND == 0 {
		return int64(-1)
	}
	return int64(hWND)
}

func ScreenRect() (image.Rectangle, error) {
	hDC := w32.GetDC(0)
	if hDC == 0 {
		return image.Rectangle{}, fmt.Errorf("Could not Get primary display err:%d\n", w32.GetLastError())
	}
	defer w32.ReleaseDC(0, hDC)
	x := w32.GetDeviceCaps(hDC, w32.HORZRES)
	y := w32.GetDeviceCaps(hDC, w32.VERTRES)
	return image.Rect(0, 0, x, y), nil
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
	hDC := w32.GetDC(0)
	if hDC == 0 {
		return nil, fmt.Errorf("Could not Get primary display err:%d.\n", w32.GetLastError())
	}
	defer w32.ReleaseDC(0, hDC)

	m_hDC := w32.CreateCompatibleDC(hDC)
	if m_hDC == 0 {
		return nil, fmt.Errorf("Could not Create Compatible DC err:%d.\n", w32.GetLastError())
	}
	defer w32.DeleteDC(m_hDC)

	x, y := rect.Dx(), rect.Dy()

	bt := w32.BITMAPINFO{}
	bt.BmiHeader.BiSize = uint32(reflect.TypeOf(bt.BmiHeader).Size())
	bt.BmiHeader.BiWidth = int32(x)
	bt.BmiHeader.BiHeight = int32(-y)
	bt.BmiHeader.BiPlanes = 1
	bt.BmiHeader.BiBitCount = 32
	bt.BmiHeader.BiCompression = w32.BI_RGB

	ptr := unsafe.Pointer(uintptr(0))

	m_hBmp := w32.CreateDIBSection(m_hDC, &bt, w32.DIB_RGB_COLORS, &ptr, 0, 0)
	if m_hBmp == 0 {
		return nil, fmt.Errorf("Could not Create DIB Section err:%d.\n", w32.GetLastError())
	}
	if m_hBmp == w32.InvalidParameter {
		return nil, fmt.Errorf("One or more of the input parameters is invalid while calling CreateDIBSection.\n")
	}
	defer w32.DeleteObject(w32.HGDIOBJ(m_hBmp))

	obj := w32.SelectObject(m_hDC, w32.HGDIOBJ(m_hBmp))
	if obj == 0 {
		return nil, fmt.Errorf("error occurred and the selected object is not a region err:%d.\n", w32.GetLastError())
	}
	if obj == 0xffffffff { //GDI_ERROR
		return nil, fmt.Errorf("GDI_ERROR while calling SelectObject err:%d.\n", w32.GetLastError())
	}
	defer w32.DeleteObject(obj)

	//Note:BitBlt contains bad error handling, we will just assume it works and if it doesn't it will panic :x
	w32.BitBlt(m_hDC, 0, 0, x, y, hDC, rect.Min.X, rect.Min.Y, w32.SRCCOPY)

	var slice []byte
	hdrp := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	hdrp.Data = uintptr(ptr)
	hdrp.Len = x * y * 4
	hdrp.Cap = x * y * 4

	imageBytes := make([]byte, len(slice))

	for i := 0; i < len(imageBytes); i += 4 {
		imageBytes[i], imageBytes[i+2], imageBytes[i+1], imageBytes[i+3] = slice[i+2], slice[i], slice[i+1], slice[i+3]
	}

	img := &image.RGBA{imageBytes, 4 * x, image.Rect(0, 0, x, y)}
	return img, nil
}

func CaptureRectYCbCr444(rect image.Rectangle, numOfRange int64) (*image.YCbCr, error) {
	t := time.Now()
	hDC := w32.GetDC(0)
	if hDC == 0 {
		return nil, fmt.Errorf("Could not Get primary display err:%d.\n", w32.GetLastError())
	}
	defer w32.ReleaseDC(0, hDC)

	m_hDC := w32.CreateCompatibleDC(hDC)
	if m_hDC == 0 {
		return nil, fmt.Errorf("Could not Create Compatible DC err:%d.\n", w32.GetLastError())
	}
	defer w32.DeleteDC(m_hDC)

	x, y := rect.Dx(), rect.Dy()

	bt := w32.BITMAPINFO{}
	bt.BmiHeader.BiSize = uint32(reflect.TypeOf(bt.BmiHeader).Size())
	bt.BmiHeader.BiWidth = int32(x)
	bt.BmiHeader.BiHeight = int32(-y)
	bt.BmiHeader.BiPlanes = 1
	bt.BmiHeader.BiBitCount = 32
	bt.BmiHeader.BiCompression = w32.BI_RGB

	ptr := unsafe.Pointer(uintptr(0))

	m_hBmp := w32.CreateDIBSection(m_hDC, &bt, w32.DIB_RGB_COLORS, &ptr, 0, 0)
	if m_hBmp == 0 {
		return nil, fmt.Errorf("Could not Create DIB Section err:%d.\n", w32.GetLastError())
	}
	if m_hBmp == w32.InvalidParameter {
		return nil, fmt.Errorf("One or more of the input parameters is invalid while calling CreateDIBSection.\n")
	}
	defer w32.DeleteObject(w32.HGDIOBJ(m_hBmp))

	obj := w32.SelectObject(m_hDC, w32.HGDIOBJ(m_hBmp))
	if obj == 0 {
		return nil, fmt.Errorf("error occurred and the selected object is not a region err:%d.\n", w32.GetLastError())
	}
	if obj == 0xffffffff { //GDI_ERROR
		return nil, fmt.Errorf("GDI_ERROR while calling SelectObject err:%d.\n", w32.GetLastError())
	}
	defer w32.DeleteObject(obj)

	//Note:BitBlt contains bad error handling, we will just assume it works and if it doesn't it will panic :x
	w32.BitBlt(m_hDC, 0, 0, x, y, hDC, rect.Min.X, rect.Min.Y, w32.SRCCOPY)

	var slice []byte
	hdrp := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	hdrp.Data = uintptr(ptr)
	hdrp.Len = x * y * 4
	hdrp.Cap = x * y * 4

	tt := time.Now()
	if ImageCache == nil {
		ImageCache = image.NewYCbCr(image.Rect(0, 0, x, y), image.YCbCrSubsampleRatio444)
	}
	ttt := time.Now()

	// CRGBToYCbCr444(slice, ImageCache.Y, ImageCache.Cb, ImageCache.Cr)

	lenData := int64(len(slice))
	batchSize := lenData / (4 * numOfRange) * 4
	for i := int64(0); i < numOfRange-1; i++ {
		Range <- []int64{i * batchSize, batchSize, int64(ImageCache.Rect.Dx()), int64(ImageCache.Rect.Dx())}
		Data <- slice
		Y <- ImageCache.Y
		Cb <- ImageCache.Cb
		Cr <- ImageCache.Cr
	}
	start := (numOfRange - 1) * batchSize
	Range <- []int64{start, lenData - start, int64(ImageCache.Rect.Dx()), int64(ImageCache.Rect.Dx())}
	Data <- slice
	Y <- ImageCache.Y
	Cb <- ImageCache.Cb
	Cr <- ImageCache.Cr
	for i := int64(0); i < numOfRange; i++ {
		<-R
	}
	tttt := time.Now()
	log.Println(fmt.Sprintf("Shot: %v, Create: %v, Convert: %v", tt.Sub(t), ttt.Sub(tt), tttt.Sub(ttt)))
	// Shot: 15.510277ms, Create: 2.024195ms, Convert: 25.398941ms

	return ImageCache, nil
}

func CaptureWindowByte(pos *POS, size *SIZE) (*CAPTURE, error) {
	hWND := w32.GetForegroundWindow()
	hDC := w32.GetDC(hWND)
	if hDC == 0 || hWND == 0 {
		return nil, fmt.Errorf("Could not Get primary display err:%d.\n", w32.GetLastError())
	}
	defer w32.ReleaseDC(hWND, hDC)

	m_hDC := w32.CreateCompatibleDC(hDC)
	if m_hDC == 0 {
		return nil, fmt.Errorf("Could not Create Compatible DC err:%d.\n", w32.GetLastError())
	}
	defer w32.DeleteDC(m_hDC)

	rect := w32.GetClientRect(hWND)
	width := int(rect.Right - rect.Left)
	height := int(rect.Bottom - rect.Top)
	if size.W != 0 && size.H != 0 {
		width = size.W
		height = size.H
	}

	bt := w32.BITMAPINFO{}
	bt.BmiHeader.BiSize = uint32(reflect.TypeOf(bt.BmiHeader).Size())
	bt.BmiHeader.BiWidth = int32(width)
	bt.BmiHeader.BiHeight = int32(-height)
	bt.BmiHeader.BiPlanes = 1
	bt.BmiHeader.BiBitCount = 32
	bt.BmiHeader.BiCompression = w32.BI_RGB

	ptr := unsafe.Pointer(uintptr(0))

	m_hBmp := w32.CreateDIBSection(m_hDC, &bt, w32.DIB_RGB_COLORS, &ptr, 0, 0)
	if m_hBmp == 0 {
		return nil, fmt.Errorf("Could not Create DIB Section err:%d.\n", w32.GetLastError())
	}
	if m_hBmp == w32.InvalidParameter {
		return nil, fmt.Errorf("One or more of the input parameters is invalid while calling CreateDIBSection.\n")
	}
	defer w32.DeleteObject(w32.HGDIOBJ(m_hBmp))

	obj := w32.SelectObject(m_hDC, w32.HGDIOBJ(m_hBmp))
	if obj == 0 {
		return nil, fmt.Errorf("error occurred and the selected object is not a region err:%d.\n", w32.GetLastError())
	}
	if obj == 0xffffffff { //GDI_ERROR
		return nil, fmt.Errorf("GDI_ERROR while calling SelectObject err:%d.\n", w32.GetLastError())
	}
	defer w32.DeleteObject(obj)

	//Note:BitBlt contains bad error handling, we will just assume it works and if it doesn't it will panic :x
	w32.BitBlt(m_hDC, 0, 0, width, height, hDC, pos.X, pos.Y, w32.SRCCOPY)

	var slice []byte
	hdrp := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	hdrp.Data = uintptr(ptr)
	hdrp.Len = width * height * 4
	hdrp.Cap = width * height * 4

	slice_copy := make([]byte, len(slice))
	copy(slice_copy, slice)

	return &CAPTURE{width, height, &slice_copy}, nil
}

func CaptureWindowImage(capture *CAPTURE) *image.RGBA {
	imageBytes := make([]byte, len(*capture.B))

	for i := 0; i < len(imageBytes); i += 4 { // this loop take almost all time
		if i == (len(imageBytes) / 2) {
		}
		imageBytes[i], imageBytes[i+2], imageBytes[i+1], imageBytes[i+3] = (*capture.B)[i+2], (*capture.B)[i], (*capture.B)[i+1], (*capture.B)[i+3]
	}

	img := &image.RGBA{imageBytes, 4 * capture.W, image.Rect(0, 0, capture.W, capture.H)}

	return img
}

func CaptureWindowYCbCr(pos *POS, size *SIZE, resize *RESIZE, toSBS bool, cursor bool, numOfRange, windowId int64) (*image.YCbCr, error) {
	var hWND w32.HWND
	if windowId != -1 {
		hWND = w32.HWND(windowId)
	} else {
		hWND = w32.GetForegroundWindow()
	}

	hDC := w32.GetDC(hWND)
	if hDC == 0 || hWND == 0 {
		return nil, fmt.Errorf("Could not Get primary display err:%d.\n", w32.GetLastError())
	}
	defer w32.ReleaseDC(hWND, hDC)

	m_hDC := w32.CreateCompatibleDC(hDC)
	if m_hDC == 0 {
		return nil, fmt.Errorf("Could not Create Compatible DC err:%d.\n", w32.GetLastError())
	}
	defer w32.DeleteDC(m_hDC)

	rect := w32.GetClientRect(hWND)
	width := int(rect.Right - rect.Left)
	height := int(rect.Bottom - rect.Top)
	if size.W != 0 && size.H != 0 {
		width = size.W
		height = size.H
	}

	bt := w32.BITMAPINFO{}
	bt.BmiHeader.BiSize = uint32(reflect.TypeOf(bt.BmiHeader).Size())
	bt.BmiHeader.BiWidth = int32(width)
	bt.BmiHeader.BiHeight = int32(-height)
	bt.BmiHeader.BiPlanes = 1
	bt.BmiHeader.BiBitCount = 32
	bt.BmiHeader.BiCompression = w32.BI_RGB

	ptr := unsafe.Pointer(uintptr(0))

	m_hBmp := w32.CreateDIBSection(m_hDC, &bt, w32.DIB_RGB_COLORS, &ptr, 0, 0)
	if m_hBmp == 0 {
		return nil, fmt.Errorf("Could not Create DIB Section err:%d.\n", w32.GetLastError())
	}
	if m_hBmp == w32.InvalidParameter {
		return nil, fmt.Errorf("One or more of the input parameters is invalid while calling CreateDIBSection.\n")
	}
	defer w32.DeleteObject(w32.HGDIOBJ(m_hBmp))

	obj := w32.SelectObject(m_hDC, w32.HGDIOBJ(m_hBmp))
	if obj == 0 {
		return nil, fmt.Errorf("error occurred and the selected object is not a region err:%d.\n", w32.GetLastError())
	}
	if obj == 0xffffffff { //GDI_ERROR
		return nil, fmt.Errorf("GDI_ERROR while calling SelectObject err:%d.\n", w32.GetLastError())
	}
	defer w32.DeleteObject(obj)

	//Note:BitBlt contains bad error handling, we will just assume it works and if it doesn't it will panic :x
	w32.BitBlt(m_hDC, 0, 0, width, height, hDC, pos.X, pos.Y, w32.SRCCOPY)
	if cursor {
		CursorInfo := new(w32.CURSORINFO)
		_ = w32.GetCursorInfo(CursorInfo)
		cx, cy, ok := w32.ScreenToClient(hWND, int(CursorInfo.PtScreenPos.X), int(CursorInfo.PtScreenPos.Y))
		if ok {
			if int(CursorInfo.HCursor) == 65541 {
				cx -= 8
				cy -= 9
			}
			w32.DrawIcon(m_hDC, cx, cy, w32.HICON(CursorInfo.HCursor))
		}
	}

	var slice []byte
	hdrp := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	hdrp.Data = uintptr(ptr)
	hdrp.Len = width * height * 4
	hdrp.Cap = width * height * 4

	if !toSBS {
		if ImageCache == nil || (ImageCache.Rect.Dx() != width || ImageCache.Rect.Dy() != height) {
			ImageCache = image.NewYCbCr(image.Rect(0, 0, width, height), image.YCbCrSubsampleRatio444)
		}
	} else {
		if ImageCache == nil || (ImageCache.Rect.Dx() != width || ImageCache.Rect.Dy() != height) {
			ImageCache = image.NewYCbCr(image.Rect(0, 0, 2*width, height), image.YCbCrSubsampleRatio444)
		}
	}

	// CRGBToYCbCr444(slice, ImageCache.Y, ImageCache.Cb, ImageCache.Cr)
	// Shot: 15.510277ms, Create: 2.024195ms, Convert: 25.398941ms

	lenData := int64(len(slice))
	batchSize := (int64(height) / numOfRange) * int64(width) * 4
	for i := int64(0); i < numOfRange-1; i++ {
		Range <- []int64{i * batchSize, batchSize, int64(width), int64(ImageCache.Rect.Dx())}
		Data <- slice
		Y <- ImageCache.Y
		Cb <- ImageCache.Cb
		Cr <- ImageCache.Cr
	}
	start := (numOfRange - 1) * batchSize
	Range <- []int64{start, lenData - start, int64(width), int64(ImageCache.Rect.Dx())}
	Data <- slice
	Y <- ImageCache.Y
	Cb <- ImageCache.Cb
	Cr <- ImageCache.Cr
	for i := int64(0); i < numOfRange; i++ {
		<-R
	}

	return ImageCache, nil
}

func CaptureWindow(pos *POS, size *SIZE, resize *RESIZE, toSBS bool, cursor bool) (*image.RGBA, error) {
	hWND := w32.GetForegroundWindow()
	hDC := w32.GetDC(hWND)
	if hDC == 0 || hWND == 0 {
		return nil, fmt.Errorf("Could not Get primary display err:%d.\n", w32.GetLastError())
	}
	defer w32.ReleaseDC(hWND, hDC)

	m_hDC := w32.CreateCompatibleDC(hDC)
	if m_hDC == 0 {
		return nil, fmt.Errorf("Could not Create Compatible DC err:%d.\n", w32.GetLastError())
	}
	defer w32.DeleteDC(m_hDC)

	rect := w32.GetClientRect(hWND)
	width := int(rect.Right - rect.Left)
	height := int(rect.Bottom - rect.Top)
	if size.W != 0 && size.H != 0 {
		width = size.W
		height = size.H
	}

	bt := w32.BITMAPINFO{}
	bt.BmiHeader.BiSize = uint32(reflect.TypeOf(bt.BmiHeader).Size())
	bt.BmiHeader.BiWidth = int32(width)
	bt.BmiHeader.BiHeight = int32(-height)
	bt.BmiHeader.BiPlanes = 1
	bt.BmiHeader.BiBitCount = 32
	bt.BmiHeader.BiCompression = w32.BI_RGB

	ptr := unsafe.Pointer(uintptr(0))

	m_hBmp := w32.CreateDIBSection(m_hDC, &bt, w32.DIB_RGB_COLORS, &ptr, 0, 0)
	if m_hBmp == 0 {
		return nil, fmt.Errorf("Could not Create DIB Section err:%d.\n", w32.GetLastError())
	}
	if m_hBmp == w32.InvalidParameter {
		return nil, fmt.Errorf("One or more of the input parameters is invalid while calling CreateDIBSection.\n")
	}
	defer w32.DeleteObject(w32.HGDIOBJ(m_hBmp))

	obj := w32.SelectObject(m_hDC, w32.HGDIOBJ(m_hBmp))
	if obj == 0 {
		return nil, fmt.Errorf("error occurred and the selected object is not a region err:%d.\n", w32.GetLastError())
	}
	if obj == 0xffffffff { //GDI_ERROR
		return nil, fmt.Errorf("GDI_ERROR while calling SelectObject err:%d.\n", w32.GetLastError())
	}
	defer w32.DeleteObject(obj)

	//Note:BitBlt contains bad error handling, we will just assume it works and if it doesn't it will panic :x
	w32.BitBlt(m_hDC, 0, 0, width, height, hDC, pos.X, pos.Y, w32.SRCCOPY)
	if cursor {
		CursorInfo := new(w32.CURSORINFO)
		_ = w32.GetCursorInfo(CursorInfo)
		cx, cy, ok := w32.ScreenToClient(hWND, int(CursorInfo.PtScreenPos.X), int(CursorInfo.PtScreenPos.Y))
		if ok {
			if int(CursorInfo.HCursor) == 65541 {
				cx -= 8
				cy -= 9
			}
			w32.DrawIcon(m_hDC, cx, cy, w32.HICON(CursorInfo.HCursor))
		}
	}

	var slice []byte
	hdrp := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	hdrp.Data = uintptr(ptr)
	hdrp.Len = width * height * 4
	hdrp.Cap = width * height * 4

	imageBytes := make([]byte, len(slice))
	ImageToRGBAWindows(slice, imageBytes)

	var img *image.RGBA
	if toSBS {
		img = &image.RGBA{append(imageBytes, imageBytes...), 4 * width, image.Rect(0, 0, width*2-2, height-1)}
	} else {
		img = &image.RGBA{imageBytes, 4 * width, image.Rect(0, 0, width-1, height-1)}
	}

	return img, nil
}
