package img

import (
	"image"
	"image/color"
)

func Pad(img image.Image, padding int, bgColor color.Color) *image.RGBA64 {
	if padding < 0 {
		panic("padding should be positive")
	}

	imgBounds := img.Bounds()

	res := image.NewRGBA64(image.Rect(0, 0, imgBounds.Dx()+2*padding, imgBounds.Dy()+2*padding))

	resBounds := res.Bounds()

	for y := resBounds.Min.Y; y < resBounds.Max.Y; y++ {
		for x := resBounds.Min.X; x < resBounds.Max.X; x++ {
			if padding <= x && x < resBounds.Max.X-padding &&
				padding <= y && y < resBounds.Max.Y-padding {
				res.Set(x, y, img.At(x+imgBounds.Min.X-padding, y+imgBounds.Min.Y-padding))
			} else {
				res.Set(x, y, bgColor)
			}
		}
	}

	return res
}
