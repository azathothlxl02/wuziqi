package utils

import (
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var MplusFont font.Face

func InitFont() {
	fontData, err := os.ReadFile("assets/MPLUS1p-Regular.ttf")
	if err != nil {
		log.Fatal(err)
	}
	tt, err := opentype.Parse(fontData)
	if err != nil {
		log.Fatal(err)
	}
	MplusFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func DrawCenteredText(screen *ebiten.Image, line1, line2 string, face font.Face, width int) {
	y1 := width/2 - 20
	y2 := width/2 + 30

	bounds1 := text.BoundString(face, line1)
	x1 := (width - bounds1.Dx()) / 2
	text.Draw(screen, line1, face, x1, y1, color.White)

	bounds2 := text.BoundString(face, line2)
	x2 := (width - bounds2.Dx()) / 2
	text.Draw(screen, line2, face, x2, y2, color.White)
}
