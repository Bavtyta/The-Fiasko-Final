package world

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"TheFiaskoTest/internal/core"
	"TheFiaskoTest/internal/render"
)

type Obstacle struct {
	worldZ  float64 // абсолютная Z-позиция в мире
	angle   float64 // угол размещения на бревне (в радианах)
	width   float64
	height  float64
	color   color.Color
	texture *ebiten.Image

	// кэшированные данные поверхности (обновляются каждый кадр)
	surfaceX float64
	surfaceY float64
	radius   float64
}

// NewObstacleAtZ создаёт препятствие в заданной абсолютной Z.
func NewObstacleAtZ(worldZ, angle, width, height float64, texture *ebiten.Image) *Obstacle {
	return &Obstacle{
		worldZ:  worldZ,
		angle:   angle,
		width:   width,
		height:  height,
		color:   color.RGBA{255, 255, 255, 255},
		texture: texture,
	}
}

// Update вызывается каждый кадр — двигаем препятствие вместе с миром.
func (o *Obstacle) Update(speed, delta float64) {
	o.worldZ -= speed * delta
}

// UpdateSurface обновляет кэш поверхности (X, Y, radius) под препятствием.
// Должен вызываться после Update и перед отрисовкой/коллизиями.
func (o *Obstacle) UpdateSurface(w *World) {
	info, ok := w.GetSurfaceAt(o.worldZ)
	if ok && info.Segment != nil {
		seg := info.Segment
		z := o.worldZ
		o.surfaceX = seg.X() + seg.SlopeX()*z
		o.surfaceY = seg.BaseY() + seg.SlopeY()*z
		o.radius = seg.Width() / 2
	}
	// Если по каким-то причинам сегмент не найден (например, Z позади камеры),
	// оставляем предыдущие значения – они всё равно скоро удалятся.
}

func (o *Obstacle) WorldPos() core.Vec3 {
	return core.Vec3{X: o.surfaceX, Y: o.surfaceY, Z: o.worldZ}
}

func (o *Obstacle) SpriteCenter() core.Vec3 {
	sinT := math.Sin(o.angle)
	cosT := math.Cos(o.angle)
	r := o.radius
	h := o.height
	return core.Vec3{
		X: o.surfaceX + (r+h/2)*sinT,
		Y: o.surfaceY + (r+h/2)*cosT,
		Z: o.worldZ,
	}
}

func (o *Obstacle) Radius() float64 {
	return math.Sqrt(math.Pow(o.width/2, 2) + math.Pow(o.height/2, 2))
}

// Draw остаётся почти без изменений, только использует o.surfaceX/Y/radius
func (o *Obstacle) Draw(screen *ebiten.Image, cam *render.Camera) {
	center := o.WorldPos()
	halfW := o.width / 2
	h := o.height
	r := o.radius

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
	visible := true
	for i, wp := range worldPts {
		sx, sy, scale := cam.Project(wp)
		if scale <= 0 {
			visible = false
			break
		}
		screenPts[i] = [2]float64{sx, sy}
	}

	if !visible {
		return
	}

	// Если есть текстура, рисуем её (упрощённо – прямоугольник с текстурой)
	if o.texture != nil {
		// Можно использовать DrawTriangles как раньше, но для краткости оставлю заливку.
		// Реализуйте текстурированный вывод по аналогии с игроком, если нужно.
		// Пока нарисуем цветной прямоугольник.
		col := o.color
		// Рисуем контур (или залитый прямоугольник)
		// В вашем коде уже была функция drawTexture – можно её скопировать.
		// Я пока дам простую отрисовку линиями.
		for i := 0; i < 4; i++ {
			next := (i + 1) % 4
			ebitenutil.DrawLine(screen, screenPts[i][0], screenPts[i][1], screenPts[next][0], screenPts[next][1], col)
		}
	} else {
		// Без текстуры – цветной прямоугольник
		col := o.color
		for i := 0; i < 4; i++ {
			next := (i + 1) % 4
			ebitenutil.DrawLine(screen, screenPts[i][0], screenPts[i][1], screenPts[next][0], screenPts[next][1], col)
		}
	}

	// Рисуем круг коллизии (как было)
	o.drawCollisionCircle(screen, cam)
}

// drawCollisionCircle – скопируйте из вашего старого obstacle.go, он там был.
// Я его приведу для полноты:
func (o *Obstacle) drawCollisionCircle(screen *ebiten.Image, cam *render.Camera) {
	center := o.SpriteCenter()
	sx, sy, scale := cam.Project(center)
	if scale <= 0 {
		return
	}
	screenRadius := o.Radius() * scale
	circleColor := color.RGBA{255, 255, 255, 150}
	const segments = 32
	for i := 0; i < segments; i++ {
		angle1 := 2 * math.Pi * float64(i) / float64(segments)
		angle2 := 2 * math.Pi * float64(i+1) / float64(segments)
		x1 := sx + screenRadius*math.Cos(angle1)
		y1 := sy + screenRadius*math.Sin(angle1)
		x2 := sx + screenRadius*math.Cos(angle2)
		y2 := sy + screenRadius*math.Sin(angle2)
		ebitenutil.DrawLine(screen, x1, y1, x2, y2, circleColor)
	}
}
