package world

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"TheFiaskoTest/internal/common"
	"TheFiaskoTest/internal/render"
)

// SurfaceType определяет тип поверхности сегмента.
type SurfaceType int

const (
	SurfaceSolid  SurfaceType = iota // твёрдая поверхность (бревно, земля)
	SurfaceLiquid                    // жидкость (вода)
)

// SegmentLayer — слой, состоящий из повторяющихся сегментов (река, брёвна и т.п.).
type SegmentLayer struct {
	segments       []*Segment
	parallaxFactor float64
	segmentLength  float64
	segmentCount   int
	baseX          float64
	baseY          float64
	width          float64
	color          color.Color
	surfaceType    SurfaceType
	texture        *ebiten.Image
}

// Getters
func (l *SegmentLayer) Segments() []*Segment {
	return l.segments
}

func (l *SegmentLayer) ParallaxFactor() float64 {
	return l.parallaxFactor
}

func (l *SegmentLayer) SegmentLength() float64 {
	return l.segmentLength
}

func (l *SegmentLayer) SegmentCount() int {
	return l.segmentCount
}

func (l *SegmentLayer) BaseX() float64 {
	return l.baseX
}

func (l *SegmentLayer) BaseY() float64 {
	return l.baseY
}

func (l *SegmentLayer) Width() float64 {
	return l.width
}

func (l *SegmentLayer) Color() color.Color {
	return l.color
}

func (l *SegmentLayer) SurfaceType() SurfaceType {
	return l.surfaceType
}

// Setters
func (l *SegmentLayer) SetSegments(segments []*Segment) {
	l.segments = segments
}

func (l *SegmentLayer) SetParallaxFactor(parallaxFactor float64) {
	if parallaxFactor < 0 {
		parallaxFactor = 0.0
	}
	l.parallaxFactor = parallaxFactor
}

func (l *SegmentLayer) SetSegmentLength(segmentLength float64) {
	l.segmentLength = segmentLength
}

func (l *SegmentLayer) SetSegmentCount(segmentCount int) {
	l.segmentCount = segmentCount
}

func (l *SegmentLayer) SetBaseX(baseX float64) {
	l.baseX = baseX
}

func (l *SegmentLayer) SetBaseY(baseY float64) {
	l.baseY = baseY
}

func (l *SegmentLayer) SetWidth(width float64) {
	l.width = width
}

func (l *SegmentLayer) SetColor(c color.Color) {
	l.color = c
}

func (l *SegmentLayer) SetSurfaceType(surfaceType SurfaceType) {
	l.surfaceType = surfaceType
}

func (l *SegmentLayer) SetTexture(tex *ebiten.Image) {
	l.texture = tex
}

// NewSegmentLayer создаёт новый слой сегментов.
func NewSegmentLayer(baseX, baseY, width, segmentLength float64, parallaxFactor float64, count int, slopeX, slopeY float64, col color.Color, surfaceType SurfaceType) *SegmentLayer {
	segs := make([]*Segment, count)
	for i := 0; i < count; i++ {
		nearZ := float64(i) * segmentLength
		seg := NewSegment(baseX, baseY, nearZ, width, segmentLength)
		seg.SetSlope(slopeX, slopeY)
		seg.SetColor(col)
		segs[i] = seg
	}
	return &SegmentLayer{
		segments:       segs,
		parallaxFactor: parallaxFactor,
		segmentLength:  segmentLength,
		segmentCount:   count,
		baseX:          baseX,
		baseY:          baseY,
		width:          width,
		color:          col,
		surfaceType:    surfaceType,
	}
}

// Update обновляет позиции сегментов и переставляет ушедшие за камеру.
func (l *SegmentLayer) Update(ctx common.WorldContext, delta float64) {
	effectiveSpeed := ctx.GetSpeed() * l.parallaxFactor

	// 1. Сдвигаем все сегменты в сторону камеры
	for _, seg := range l.segments {
		seg.Update(effectiveSpeed, delta)
	}

	// 2. Циклически переносим ушедшие за камеру сегменты в хвост очереди
	const maxWraps = 10
	for i := 0; i < maxWraps && len(l.segments) > 0; i++ {
		first := l.segments[0]
		// Если задний край первого сегмента ещё впереди камеры (Z > 0) – выходим
		if first.NearZ()+l.segmentLength > 0 {
			break
		}

		last := l.segments[len(l.segments)-1]

		// Новое положение первого сегмента – сразу за последним
		newNearZ := last.NearZ() + l.segmentLength

		// Вычисляем высоту стыка (базовую Y для непрерывности поверхности)
		// При одинаковых наклонах можно просто last.BaseY(), но формула универсальна
		heightAtStitch := last.BaseY() + last.SlopeY()*(last.NearZ()+l.segmentLength)
		newBaseY := heightAtStitch - first.SlopeY()*newNearZ

		first.SetNearZ(newNearZ)
		first.SetBaseY(newBaseY)

		// Сдвигаем очередь: убираем первый, добавляем в конец
		l.segments = append(l.segments[1:], first)
	}
}

// Draw отрисовывает все сегменты слоя.
func (l *SegmentLayer) Draw(screen *ebiten.Image, cam *render.Camera, ctx common.WorldContext) {
	for _, seg := range l.segments {
		seg.Draw(screen, cam, l.texture)
	}
}

// SurfaceAt возвращает высоту и тип поверхности в заданной мировой координате Z,
// если она находится в пределах какого-либо сегмента слоя.
func (l *SegmentLayer) SurfaceAt(z float64) (height float64, surfaceType SurfaceType, ok bool) {
	for _, seg := range l.segments {
		if z >= seg.NearZ() && z <= seg.NearZ()+seg.Length() {
			baseHeight := seg.BaseY() + seg.SlopeY()*z
			if l.surfaceType == SurfaceSolid {
				radius := seg.Width() / 2
				return baseHeight + radius, l.surfaceType, true
			}
			return baseHeight, l.surfaceType, true
		}
	}
	return 0, 0, false
}

// SegmentAt возвращает сегмент слоя, содержащий точку z, или nil.
func (l *SegmentLayer) SegmentAt(z float64) *Segment {
	for _, seg := range l.segments {
		if z >= seg.NearZ() && z <= seg.NearZ()+seg.Length() {
			return seg
		}
	}
	return nil
}
