package screenshot

import (
	"fmt"
	"image"
	"reflect"
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

func CaptureScreen() (*image.RGBA, error) {
	r, e := ScreenRect()
	if e != nil {
		return nil, e
	}
	return CaptureRect(r)
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

func CaptureWindow(pos *POS, size *SIZE, resize *RESIZE, toSBS bool) (*image.RGBA, error) {
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

	imageBytes := make([]byte, len(slice))

	for i := 0; i < len(imageBytes); i += 4 { // this loop take almost all time
		imageBytes[i], imageBytes[i+2], imageBytes[i+1], imageBytes[i+3] = slice[i+2], slice[i], slice[i+1], slice[i+3]
	}

	var img *image.RGBA
	if toSBS {
		img = &image.RGBA{append(imageBytes, imageBytes...), 4 * width, image.Rect(0, 0, width*2, height)}
	} else {
		img = &image.RGBA{imageBytes, 4 * width, image.Rect(0, 0, width, height)}
	}

	return img, nil
}

func CaptureWindowMust(pos *POS, size *SIZE, resize *RESIZE, toSBS bool) *image.RGBA {
	img, err := CaptureWindow(pos, size, resize, toSBS)
	for err != nil {
		img, err = CaptureWindow(pos, size, resize, toSBS)
	}
	return img
}
