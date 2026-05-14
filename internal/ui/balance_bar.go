package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"TheFiaskoTest/internal/core"
	"TheFiaskoTest/internal/render"
)

type BalanceBarLayer struct {
	getBalance    func() float64
	getMaxBalance func() float64
	isFalling     func() bool
}

func NewBalanceBarLayer(getBalance, getMaxBalance func() float64, isFalling func() bool) *BalanceBarLayer {
	return &BalanceBarLayer{
		getBalance:    getBalance,
		getMaxBalance: getMaxBalance,
		isFalling:     isFalling,
	}
}

func (b *BalanceBarLayer) Update() {}

// Draw рисует горизонтальную полоску баланса над игроком.
// Цвет меняется от зелёного (в центре) до красного (на границах).
func (b *BalanceBarLayer) Draw(screen *ebiten.Image, cam *render.Camera, upperCenter core.Vec3) {
	if b.isFalling() {
		return
	}

	// Проецируем центр верхней грани
	sx, sy, scale := cam.Project(upperCenter)
	if scale <= 0 {
		return
	}

	// Отступ вверх от спроецированной точки
	const barOffsetY = 25
	barY := sy - barOffsetY
	barHeight := 20.0
	maxBarLength := 100.0 // половина длины полоски (от центра в одну сторону)

	balance := b.getBalance()
	maxBal := b.getMaxBalance()
	if maxBal == 0 {
		return
	}

	// Нормализованное значение от 0 до 1 (модуль)
	norm := math.Abs(balance) / maxBal
	// Ограничиваем для цветов
	if norm > 1 {
		norm = 1
	}

	// Цвет: от зелёного (0) до красного (1)
	barColor := interpolateColor(color.RGBA{0, 255, 0, 255}, color.RGBA{255, 0, 0, 255}, norm)

	// Небольшая пульсация при приближении к критической границе (норм > 0.7)
	if norm > 0.7 {
		// Простая пульсация: альфа-канал меняется со временем
		// Для простоты используем синус от счётчика кадров (но у нас нет time, можно добавить поле animationFrame)
		// Реализуем через синус от реального времени? Лучше просто сделать мигание.
		// Для упрощения оставим только цвет, либо добавим отдельный эффект.
		// Сделаем яркость выше:
		barColor = color.RGBA{255, uint8(float64(255) * (1 - norm) * 0.5), uint8(float64(255) * (1 - norm) * 0.5), 255}
	}

	if balance >= 0 {
		length := norm * maxBarLength
		// Рисуем правую часть (положительный баланс)
		ebitenutil.DrawRect(screen, sx, barY, length, barHeight, barColor)
	} else {
		length := norm * maxBarLength
		// Рисуем левую часть (отрицательный баланс)
		ebitenutil.DrawRect(screen, sx-length, barY, length, barHeight, barColor)
	}

}

// interpolateColor линейно интерполирует между двумя цветами (0..1)
func interpolateColor(c1, c2 color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c1.R)*(1-t) + float64(c2.R)*t),
		G: uint8(float64(c1.G)*(1-t) + float64(c2.G)*t),
		B: uint8(float64(c1.B)*(1-t) + float64(c2.B)*t),
		A: 255,
	}
}
