package main

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"os"
	"time"
)

var (
	Base     *image.Gray
	Nums     map[int]*image.Gray
	Military bool
)

func TimeNowImg() *image.Gray {
	currentTime := time.Now()
	hour, min := (currentTime.Hour())%24, currentTime.Minute()
	if !Military {
		hour = hour % 12
	}

	bigRect, smallRect := image.Rect(0, 0, 160, 43), image.Rect(0, 0, 25, 43)
	img := image.NewGray(bigRect)

	// draw ':' in middle
	draw.Draw(img, bigRect, Base, image.Pt(0, 0), 0)

	//draw hour digits
	switch {
	case hour >= 20:
		draw.Draw(img, smallRect.Add(image.Pt(3, 0)), Nums[2], image.Pt(0, 0), 0)
	case hour >= 10:
		draw.Draw(img, smallRect.Add(image.Pt(3, 0)), Nums[1], image.Pt(0, 0), 0)
	}
	draw.Draw(img, smallRect.Add(image.Pt(32, 0)), Nums[hour%10], image.Pt(0, 0), 0)

	//draw minute digits
	draw.Draw(img, smallRect.Add(image.Pt(75, 0)), Nums[min/10], image.Pt(0, 0), 0)
	draw.Draw(img, smallRect.Add(image.Pt(103, 0)), Nums[min%10], image.Pt(0, 0), 0)

	return img
}

func init() {
	Base, _ = LoadGrayImg("/usr/share/go510/base.png")
	Nums = make(map[int]*image.Gray)
	for i := 0; i < 10; i++ {
		Nums[i], _ = LoadGrayImg(fmt.Sprintf("/usr/share/go510/%v.png", i))
	}
}

func LoadGrayImg(name string) (*image.Gray, error) {
	f, e := os.Open(name)
	if e == nil {
		i, _, e := image.Decode(f)
		if e == nil {
			gray := image.NewGray(i.Bounds())
			for x := 0; x < i.Bounds().Max.X; x++ {
				for y := 0; y < i.Bounds().Max.Y; y++ {
					// since all the images are b&w, we'll just grab the red channel
					r, _, _, _ := i.At(x, y).RGBA()
					gray.Pix[y*gray.Stride+x] = uint8(255 - r)
				}
			}
			return gray, nil

		} else {
			return nil, e
		}
	} else {
		return nil, e
	}
	return nil, errors.New("You shouldn't hit this")
}
