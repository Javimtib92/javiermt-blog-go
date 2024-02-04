package color

import (
	"errors"
	"fmt"
	"log"
	"math"
)

type RGB struct {
	R, G, B float64
}

type HSL struct {
	H, S, L float64
}

func HexToRgb(in string) (RGB, error) {
	if in[0] == '#' {
		in = in[1:]
	}

	if len(in) != 6 {
		return RGB{}, errors.New("Invalid string length")
	}

	var r, g, b byte
	if n, err := fmt.Sscanf(in, "%2x%2x%2x", &r, &g, &b); err != nil || n != 3 {
		return RGB{}, err
	}

	return RGB{float64(r), float64(g), float64(b)}, nil
}

func (c RGB) ToHSL() HSL {
	var h, s, l float64

	r := c.R
	g := c.G
	b := c.B

	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)

	// Luminosity is the average of the max and min rgb color intensities.
	l = (max + min) / 2

	// saturation
	delta := max - min
	if delta == 0 {
		// it's gray
		return HSL{0, 0, l}
	}

	// it's not gray
	if l < 0.5 {
		s = delta / (max + min)
	} else {
		s = delta / (2 - max - min)
	}

	// hue
	switch max {
	case r:
		h = (g - b) / delta
		if g < b {
			h += 2
		}
	case g:
		h = (b - r) / delta + 2
	case b:
		h = (r - g) / delta + 4
	}

	// normalize hue to [0, 1]
	h /= 6

	// convert hue to degrees
	h *= 360

	return HSL{h, s, l}
}

func HextoHSL(c string) (HSL, error) {
	rgb, err := HexToRgb(c);

	log.Print("rgb", rgb)
	if err != nil {
		return HSL{}, err
	}

	return rgb.ToHSL(), nil
}