package main

import (
    "fmt"
	"image/color"
)

func main() {
    co := color.RGBA{188, 188, 188, 255}
    y, cb, cr := color.RGBToYCbCr(co.R, co.G, co.B)

    r, g, b := color.YCbCrToRGB(y, cb, cr)
	
	fmt.Printf("r, g, b : %v, %v, %v\n", r, g, b)
}
