package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
)

func main() {
	f, e := os.Open("base.png")
	if e != nil {
		panic(e)
	}
	i, _, e := image.Decode(f)
	if e != nil {
		panic(e)
	}
	fmt.Println(i.ColorModel() == color.GrayModel)
	fmt.Println(i.ColorModel() == color.RGBAModel)
	fmt.Println(i.ColorModel() == color.NRGBAModel)
	fmt.Println(i.ColorModel() == color.Gray16Model)
	fmt.Println(i.ColorModel() == color.RGBA64Model)
}
