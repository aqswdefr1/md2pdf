package main

import "image/color"

type FontStyle struct {
	Size float64
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
	"github": {
		Name:            "github",
		BgColor:         color.RGBA{R: 255, G: 255, B: 255, A: 255}, // White
		TextColor:       color.RGBA{R: 36, G: 41, B: 47, A: 255},    // Dark Grey
		H1:              FontStyle{Size: 24, Color: color.RGBA{R: 36, G: 41, B: 47, A: 255}},
		H2:              FontStyle{Size: 20, Color: color.RGBA{R: 36, G: 41, B: 47, A: 255}},
		H3:              FontStyle{Size: 16, Color: color.RGBA{R: 36, G: 41, B: 47, A: 255}},
		H4:              FontStyle{Size: 14, Color: color.RGBA{R: 36, G: 41, B: 47, A: 255}},
		Body:            FontStyle{Size: 10.5, Color: color.RGBA{R: 36, G: 41, B: 47, A: 255}},
		Code:            FontStyle{Size: 9.5, Color: color.RGBA{R: 36, G: 41, B: 47, A: 255}},
		CodeBg:          color.RGBA{R: 246, G: 248, B: 250, A: 255}, // Very light grey
		BlockquoteColor: color.RGBA{R: 101, G: 109, B: 118, A: 255}, // Muted grey
		LinkColor:       color.RGBA{R: 9, G: 105, B: 218, A: 255},   // GitHub Blue
		BorderColor:     color.RGBA{R: 208, G: 215, B: 222, A: 255}, // Light border
	},
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
	"academic": {
		Name:            "academic",
		BgColor:         color.RGBA{R: 255, G: 255, B: 255, A: 255},
		TextColor:       color.RGBA{R: 0, G: 0, B: 0, A: 255}, // Black
		H1:              FontStyle{Size: 22, Color: color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		H2:              FontStyle{Size: 17, Color: color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		H3:              FontStyle{Size: 14, Color: color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		H4:              FontStyle{Size: 12, Color: color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		Body:            FontStyle{Size: 11, Color: color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		Code:            FontStyle{Size: 10, Color: color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		CodeBg:          color.RGBA{R: 245, G: 245, B: 245, A: 255},
		BlockquoteColor: color.RGBA{R: 80, G: 80, B: 80, A: 255},
		LinkColor:       color.RGBA{R: 0, G: 0, B: 120, A: 255},
		BorderColor:     color.RGBA{R: 150, G: 150, B: 150, A: 255},
	},
}
