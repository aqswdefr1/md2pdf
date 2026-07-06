package main

import "image/color"

type FontStyle struct {
	Size  float64
	Color color.RGBA
}

type Theme struct {
	Name            string
	BgColor         color.RGBA
	TextColor       color.RGBA
	H1              FontStyle
	H2              FontStyle
	H3              FontStyle
	H4              FontStyle
	Body            FontStyle
	Code            FontStyle
	CodeBg          color.RGBA
	BlockquoteColor color.RGBA
	LinkColor       color.RGBA
	BorderColor     color.RGBA
}

var Themes = map[string]Theme{
	"modern": {
		Name:            "modern",
		BgColor:         color.RGBA{R: 255, G: 255, B: 255, A: 255},
		TextColor:       color.RGBA{R: 30, G: 41, B: 59, A: 255}, // Slate 800
		H1:              FontStyle{Size: 26, Color: color.RGBA{R: 15, G: 23, B: 42, A: 255}}, // Slate 900
		H2:              FontStyle{Size: 20, Color: color.RGBA{R: 15, G: 23, B: 42, A: 255}},
		H3:              FontStyle{Size: 16, Color: color.RGBA{R: 71, G: 85, B: 105, A: 255}}, // Slate 600
		H4:              FontStyle{Size: 13, Color: color.RGBA{R: 71, G: 85, B: 105, A: 255}},
		Body:            FontStyle{Size: 11, Color: color.RGBA{R: 51, G: 65, B: 85, A: 255}},  // Slate 700
		Code:            FontStyle{Size: 9.5, Color: color.RGBA{R: 225, G: 29, B: 72, A: 255}}, // Rose 600
		CodeBg:          color.RGBA{R: 248, G: 250, B: 252, A: 255}, // Slate 50
		BlockquoteColor: color.RGBA{R: 79, G: 70, B: 229, A: 255},  // Indigo 600
		LinkColor:       color.RGBA{R: 79, G: 70, B: 229, A: 255},   // Indigo 600
		BorderColor:     color.RGBA{R: 226, G: 232, B: 240, A: 255}, // Slate 200
	},
}
