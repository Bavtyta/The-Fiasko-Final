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

// CollidesPlayerEllipse — столкновение эллипса игрока с эллипсом препятствия (оба в плоскости XY), порог по Z.
func (o *Obstacle) CollidesPlayerEllipse(playerCX, playerCY, playerCZ, pSemiT, pSemiR, pAng float64, zThreshold float64) bool {
	obsC := o.SpriteCenter()
	if math.Abs(playerCZ-obsC.Z) > zThreshold {
		return false
	}
	return orientedEllipsesOverlap(
		playerCX, playerCY, pSemiT, pSemiR, pAng,
		obsC.X, obsC.Y, o.width*0.5, o.height*0.5, o.angle,
	)
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

	if o.texture != nil {
		bounds := o.texture.Bounds()
		texW := float32(bounds.Dx())
		texH := float32(bounds.Dy())
		// Два треугольника с UV, как у игрока (квад в мировых координатах → экран)
		vertices := []ebiten.Vertex{
			{DstX: float32(screenPts[0][0]), DstY: float32(screenPts[0][1]), SrcX: 0, SrcY: texH, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
			{DstX: float32(screenPts[1][0]), DstY: float32(screenPts[1][1]), SrcX: texW, SrcY: texH, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
			{DstX: float32(screenPts[3][0]), DstY: float32(screenPts[3][1]), SrcX: texW, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
			{DstX: float32(screenPts[2][0]), DstY: float32(screenPts[2][1]), SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		}
		indices := []uint16{0, 1, 2, 0, 2, 3}
		opts := &ebiten.DrawTrianglesOptions{Filter: ebiten.FilterLinear}
		screen.DrawTriangles(vertices, indices, o.texture, opts)
	} else {
		col := o.color
		for i := 0; i < 4; i++ {
			next := (i + 1) % 4
			ebitenutil.DrawLine(screen, screenPts[i][0], screenPts[i][1], screenPts[next][0], screenPts[next][1], col)
		}
	}

	o.drawCollisionEllipse(screen, cam)
}

func (o *Obstacle) drawCollisionEllipse(screen *ebiten.Image, cam *render.Camera) {
	c := o.SpriteCenter()
	sinT, cosT := math.Sin(o.angle), math.Cos(o.angle)
	semiT := o.width * 0.5
	semiR := o.height * 0.5
	lineColor := color.RGBA{255, 255, 255, 150}
	const segments = 48
	for i := 0; i < segments; i++ {
		phi1 := 2 * math.Pi * float64(i) / float64(segments)
		phi2 := 2 * math.Pi * float64(i+1) / float64(segments)
		wx1 := c.X + semiT*math.Cos(phi1)*cosT + semiR*math.Sin(phi1)*sinT
		wy1 := c.Y - semiT*math.Cos(phi1)*sinT + semiR*math.Sin(phi1)*cosT
		wx2 := c.X + semiT*math.Cos(phi2)*cosT + semiR*math.Sin(phi2)*sinT
		wy2 := c.Y - semiT*math.Cos(phi2)*sinT + semiR*math.Sin(phi2)*cosT
		sx1, sy1, s1 := cam.Project(core.Vec3{X: wx1, Y: wy1, Z: c.Z})
		sx2, sy2, s2 := cam.Project(core.Vec3{X: wx2, Y: wy2, Z: c.Z})
		if s1 > 0 && s2 > 0 {
			ebitenutil.DrawLine(screen, sx1, sy1, sx2, sy2, lineColor)
		}
	}
}
