package vibrant

import "container/heap"

const (
	blackMaxLightness float64 = 0.05
	whiteMinLightness float64 = 0.95
)

// A color quantizer based on the Median-cut algorithm, optimized for
// picking out distinct colors rather than representation colors.
//
// The color space is represented as a 3-dimensional cube with each
// dimension being an RGB component. The cube is then repeatedly divided
// until we have reduced the color space to the requested number of colors.
// An average color is then generated from each cube.
//
// Whereas median-cut divides cubes so they all have roughly the same
// population, this quantizer divides boxes based on their color volume.
type colorCutQuantizer struct {
	Colors           []int
	ColorPopulations map[int]int
	ColorRatios      map[int]float64
	QuantizedColors  []*Swatch
	PixelCount       int
}

// true if the color is close to pure black, pure white
func shouldIgnoreColor(color int) bool {
	_, _, l := rgbToHsl(color)
	return l <= blackMaxLightness || l >= whiteMinLightness
}

func shouldIgnoreColorSwatch(sw *Swatch) bool {
	return shouldIgnoreColor(int(sw.Color))
}

func newColorCutQuantizer(bitmap bitmap, maxColors int) *colorCutQuantizer {
	pixels := bitmap.Pixels()
	pixelCount := bitmap.Width * bitmap.Height
	histo := newColorHistogram(pixels)
	colorPopulations := make(map[int]int, histo.NumberColors)
	colorRatios := make(map[int]float64, histo.NumberColors)
	for i, c := range histo.Colors {
		colorPopulations[c] = histo.ColorCounts[i]
		colorRatios[c] = float64(histo.ColorCounts[i]) / float64(pixelCount)
	}
	validColors := make([]int, 0)
	i := 0
	for _, c := range histo.Colors {
		if !shouldIgnoreColor(c) {
			validColors = append(validColors, c)
			i++
		}
	}
	validCount := len(validColors)
	ccq := &colorCutQuantizer{Colors: validColors, ColorPopulations: colorPopulations, ColorRatios: colorRatios, PixelCount: pixelCount}
	if validCount <= maxColors {
		// note: no quantization actually occurs
		for _, c := range validColors {
			ccq.QuantizedColors = append(ccq.QuantizedColors, &Swatch{Color: Color(c), Population: colorPopulations[c], Ratio: colorRatios[c]})
		}
	} else {
		ccq.quantizePixels(validCount-1, maxColors)
	}
	return ccq
}

// see also vbox.go
func (ccq *colorCutQuantizer) quantizePixels(maxColorIndex, maxColors int) {
	pq := make(priorityQueue, 0)
	heap.Init(&pq)
	heap.Push(&pq, newVbox(0, maxColorIndex, ccq.Colors, ccq.ColorPopulations))
	for pq.Len() < maxColors {
		v := heap.Pop(&pq).(*vbox)
		if v.CanSplit() {
			heap.Push(&pq, v.Split())
			heap.Push(&pq, v)
		} else {
			break
		}
	}
	for pq.Len() > 0 {
		v := heap.Pop(&pq).(*vbox)
		swatch := v.AverageColor(ccq.PixelCount)
		if !shouldIgnoreColorSwatch(swatch) {
			ccq.QuantizedColors = append(ccq.QuantizedColors, swatch)
		}
	}
}
