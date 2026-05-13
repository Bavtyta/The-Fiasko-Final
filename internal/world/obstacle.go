package world

import (
	"image/color"
	"math"
	"math/rand"

	"TheFiaskoTest/internal/core"
	"TheFiaskoTest/internal/render"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Obstacle struct {
	segment *Segment
	offsetZ float64
	width   float64
	height  float64
	angle   float64 // статический угол размещения на окружности бревна
	color   color.Color
	radius  float64
}

// NewObstacle создаёт новое препятствие. width и height – размеры спрайта.
// Препятствие размещается на дуге окружности бревна в случайном месте
// в пределах углов от π/3 до 5π/3 (исключая нижнюю часть бревна).
func NewObstacle(segment *Segment, offsetZ, width, height float64) *Obstacle {
	// Случайный угол в радианах: от 60° до 300° (π/3 .. 5π/3)
	angle := math.Pi/3 + rand.Float64()*(4*math.Pi/3)

	return &Obstacle{
		segment: segment,
		offsetZ: offsetZ,
		width:   width,
		height:  height,
		angle:   angle,
		color:   color.RGBA{255, 0, 0, 255},
		radius:  segment.Width() / 2,
	}
}

// WorldPos возвращает мировые координаты центра бревна (оси вращения)
func (o *Obstacle) WorldPos() core.Vec3 {
	z := o.segment.NearZ() + o.offsetZ
	x := o.segment.X() + o.segment.SlopeX()*z
	y := o.segment.BaseY() + o.segment.SlopeY()*z // центр бревна
	return core.Vec3{X: x, Y: y, Z: z}
}

// Draw рисует препятствие как прямоугольник, повёрнутый на статический угол o.angle
func (o *Obstacle) Draw(screen *ebiten.Image, cam *render.Camera) {
	center := o.WorldPos()
	halfW := o.width / 2
	h := o.height
	r := o.radius

	// Локальные координаты до поворота: основание на расстоянии r от центра бревна
	local := [][2]float64{
		{-halfW, r},
		{halfW, r},
		{-halfW, r + h},
		{halfW, r + h},
	}

	cosT := math.Cos(o.angle)
	sinT := math.Sin(o.angle)

	var worldPts [4]core.Vec3
	for i, l := range local {
		rx := l[0]*cosT + l[1]*sinT
		ry := -l[0]*sinT + l[1]*cosT
		worldPts[i] = core.Vec3{
			X: center.X + rx,
			Y: center.Y + ry,
			Z: center.Z,
		}
	}

	var screenPts [4][2]float64
	for i, wp := range worldPts {
		sx, sy, scale := cam.Project(wp)
		if scale <= 0 {
			return
		}
		screenPts[i] = [2]float64{sx, sy}
	}

	col := o.color
	ebitenutil.DrawLine(screen, screenPts[0][0], screenPts[0][1], screenPts[1][0], screenPts[1][1], col)
	ebitenutil.DrawLine(screen, screenPts[1][0], screenPts[1][1], screenPts[3][0], screenPts[3][1], col)
	ebitenutil.DrawLine(screen, screenPts[3][0], screenPts[3][1], screenPts[2][0], screenPts[2][1], col)
	ebitenutil.DrawLine(screen, screenPts[2][0], screenPts[2][1], screenPts[0][0], screenPts[0][1], col)
}

// Radius возвращает приблизительный радиус ограничивающей сферы для коллизий
func (o *Obstacle) Radius() float64 {
	return math.Sqrt(math.Pow(o.width/2, 2) + math.Pow(o.height/2, 2))
}

// SpriteCenter возвращает мировые координаты центра спрайта препятствия
func (o *Obstacle) SpriteCenter() core.Vec3 {
	r := o.radius
	h := o.height
	sinT := math.Sin(o.angle)
	cosT := math.Cos(o.angle)
	wp := o.WorldPos()
	return core.Vec3{
		X: wp.X + (r+h/2)*sinT,
		Y: wp.Y + (r+h/2)*cosT,
		Z: wp.Z,
	}
}
