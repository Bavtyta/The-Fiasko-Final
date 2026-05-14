// internal/uicommon/styles.go
package uicommon

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
)

// Цвета
var (
	ColorBgOverlay      = color.RGBA{0, 0, 0, 180}      // полупрозрачный тёмный для окон
	ColorButtonBg       = color.RGBA{205, 150, 80, 255} // шерсть Хатико
	ColorButtonText     = color.RGBA{60, 35, 15, 255}   // тёмно-коричневый
	ColorBorderNormal   = color.RGBA{80, 50, 20, 255}
	ColorBorderSelected = color.RGBA{255, 215, 0, 255} // золото
	ColorShadow         = color.RGBA{0, 0, 0, 200}
	ColorWhite          = color.RGBA{255, 255, 255, 255}
)

// Шрифты (глобальные, инициализируются один раз)
var (
	FontTitle  font.Face // 64
	FontScore  font.Face // 36
	FontButton font.Face // 24
	FontSmall  font.Face // 16
	FontHUD    font.Face // 20 – для текста во время игры
)

func init() {
	tt, err := opentype.Parse(gomono.TTF)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	FontTitle, _ = opentype.NewFace(tt, &opentype.FaceOptions{Size: 64, DPI: dpi})
	FontScore, _ = opentype.NewFace(tt, &opentype.FaceOptions{Size: 36, DPI: dpi})
	FontButton, _ = opentype.NewFace(tt, &opentype.FaceOptions{Size: 24, DPI: dpi})
	FontSmall, _ = opentype.NewFace(tt, &opentype.FaceOptions{Size: 16, DPI: dpi})
	FontHUD, _ = opentype.NewFace(tt, &opentype.FaceOptions{Size: 20, DPI: dpi})
}

// DrawBackground рисует текстуру во весь экран (если есть) или заливает чёрным
func DrawBackground(screen *ebiten.Image, bg *ebiten.Image, screenW, screenH int) {
	if bg != nil {
		op := &ebiten.DrawImageOptions{}
		bounds := bg.Bounds()
		scaleX := float64(screenW) / float64(bounds.Dx())
		scaleY := float64(screenH) / float64(bounds.Dy())
		op.GeoM.Scale(scaleX, scaleY)
		screen.DrawImage(bg, op)
	} else {
		screen.Fill(color.Black)
	}
}

// DrawButton рисует одну кнопку с рамкой, текстом и состоянием выбора
func DrawButton(screen *ebiten.Image, x, y, width, height float32, label string, selected bool) {
	vector.DrawFilledRect(screen, x, y, width, height, ColorButtonBg, true)
	borderColor := ColorBorderNormal
	if selected {
		borderColor = ColorBorderSelected
	}
	vector.StrokeRect(screen, x, y, width, height, 2, borderColor, true)

	// Центрируем текст
	bounds := text.BoundString(FontButton, label)
	textX := int(x + width/2 - float32(bounds.Dx())/2)
	textY := int(y + height/2 + 8)
	text.Draw(screen, label, FontButton, textX, textY, ColorButtonText)
}

// DrawTitle рисует заголовок с тенью по центру
func DrawTitle(screen *ebiten.Image, title string, y int, screenW int) {
	bounds := text.BoundString(FontTitle, title)
	x := (screenW - bounds.Dx()) / 2
	text.Draw(screen, title, FontTitle, x+4, y+4, ColorShadow)
	text.Draw(screen, title, FontTitle, x, y, ColorWhite)
}

// DrawScore рисует счёт (крупный, но меньше заголовка)
func DrawScore(screen *ebiten.Image, score float64, y int, screenW int) {
	scoreStr := formatScore(score)
	bounds := text.BoundString(FontScore, scoreStr)
	x := (screenW - bounds.Dx()) / 2
	text.Draw(screen, scoreStr, FontScore, x+3, y+3, ColorShadow)
	text.Draw(screen, scoreStr, FontScore, x, y, ColorWhite)
}

func formatScore(score float64) string {
	return fmt.Sprintf("%.0f", score)
}
