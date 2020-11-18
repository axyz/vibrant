package vibrant

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/nfnt/resize"
)

// type bitmap is a simple wrapper for an image.Image
type bitmap struct {
	Width  int
	Height int
	Source image.Image
}

func newBitmap(input image.Image) *bitmap {
	bounds := input.Bounds()
	return &bitmap{bounds.Dx(), bounds.Dy(), input}
}

func newCropBitmap(input image.Image, crop image.Rectangle) (*bitmap, error) {
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}

	// img is an Image interface. This checks if the underlying value has a
	// method called SubImage. If it does, then we can use SubImage to crop the
	// image.
	simg, ok := input.(subImager)
	if !ok {
		return nil, fmt.Errorf("image does not support cropping")
	}

	subImage := simg.SubImage(crop)

	return &bitmap{subImage.Bounds().Dx(), subImage.Bounds().Dy(), subImage}, nil
}

// Scales input image.Image by aspect ratio using https://github.com/nfnt/resize
func newScaledBitmap(input image.Image, ratio float64) *bitmap {
	bounds := input.Bounds()
	w := math.Ceil(float64(bounds.Dx()) * ratio)
	h := math.Ceil(float64(bounds.Dy()) * ratio)
	return &bitmap{int(w), int(h), resize.Resize(uint(w), uint(h), input, resize.NearestNeighbor)}
}

// Returns all of the pixels of this bitmap.Source as a 1D array of image/color.Color
func (b *bitmap) Pixels() []color.Color {
	c := make([]color.Color, 0)
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			c = append(c, b.Source.At(x, y))
		}
	}
	return c
}
