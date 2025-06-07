package img

import (
	"image"
	"image/color"
)

func Pad(img image.Image, padding int) *image.RGBA64 {
	if padding < 0 {
		panic("padding should be positive")
	}

	imgBounds := img.Bounds()

	res := image.NewRGBA64(image.Rectangle{
		Min: image.Point{
			X: imgBounds.Min.X - padding,
			Y: imgBounds.Min.Y - padding,
		},
		Max: image.Point{
			X: imgBounds.Max.X + padding,
			Y: imgBounds.Max.Y + padding,
		},
	})

	resBounds := res.Bounds()

	for y := resBounds.Min.Y; y < resBounds.Max.Y; y++ {
		for x := resBounds.Min.X; x < resBounds.Max.X; x++ {
			if imgBounds.Min.X <= x && x < imgBounds.Max.X &&
				imgBounds.Min.Y <= y && y < imgBounds.Max.Y {
				res.Set(x, y, img.At(x, y))
			} else {
				res.Set(x, y, color.Transparent)
			}
		}
	}

	return res
}
