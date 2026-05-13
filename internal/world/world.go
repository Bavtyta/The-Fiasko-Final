package world

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"

	"TheFiaskoTest/internal/common"
	"TheFiaskoTest/internal/render"
)

const maxActiveObstacles = 100

type Layer interface {
	Update(ctx common.WorldContext, delta float64)
	Draw(screen *ebiten.Image, cam *render.Camera, ctx common.WorldContext)
}

// SurfaceProvider определяет возможность получить информацию о поверхности в точке Z.
type SurfaceProvider interface {
	SurfaceAt(z float64) (height float64, surfaceType SurfaceType, ok bool)
}

type World struct {
	worldOffsetZ float64
	speed        float64

	layers        []Layer
	obstacles     []*Obstacle
	baseSpawnDist float64 // базовое расстояние между препятствиями (в единицах Z)
	lastSpawnZ    float64 // последняя позиция Z, где было создано препятствие
}

func (w *World) AddLayer(l Layer) {
	w.layers = append(w.layers, l)
}

func New(speed float64) *World {
	return &World{
		worldOffsetZ:  0,
		speed:         speed,
		baseSpawnDist: 75.0,
		lastSpawnZ:    0.0,
		layers:        []Layer{},
		obstacles:     []*Obstacle{},
	}
}

func (w *World) AddObstacle(obs *Obstacle) {
	w.obstacles = append(w.obstacles, obs)
}

// spawnObstacle создаёт новое препятствие на самом дальнем твёрдом сегменте.
func (w *World) spawnObstacle() {
	var farthestSeg *Segment
	maxZ := -math.MaxFloat64

	// Ищем самый дальний сегмент среди всех слоёв с твёрдой поверхностью
	for _, layer := range w.layers {
		if sl, ok := layer.(*SegmentLayer); ok && sl.SurfaceType() == SurfaceSolid {
			for _, seg := range sl.Segments() {
				segZ := seg.NearZ() + seg.Length()
				if segZ > maxZ {
					maxZ = segZ
					farthestSeg = seg
				}
			}
		}
	}

	if farthestSeg == nil {
		return
	}

	// Случайная позиция вдоль сегмента с отступами
	const margin = 5.0
	segLength := farthestSeg.Length()
	var offsetZ float64
	if segLength > 2*margin {
		offsetZ = margin + rand.Float64()*(segLength-2*margin)
	} else {
		offsetZ = rand.Float64() * segLength
	}

	// Создаём статичное препятствие (без вращения)
	obs := NewObstacle(farthestSeg, offsetZ, 3, 5)
	w.AddObstacle(obs)
}

func (w *World) Update(delta float64) {
	w.worldOffsetZ += w.speed * delta

	// Обновляем слои
	for _, layer := range w.layers {
		layer.Update(w, delta)
	}

	// Препятствия статичны относительно своих брёвен, поэтому не обновляем их вращение.
	// Но они перемещаются вместе с сегментами автоматически, так как их позиция
	// вычисляется через segment.NearZ() при каждой отрисовке / запросе координат.

	// Генерация новых препятствий по мере прохождения расстояния
	maxSpawnsPerFrame := 10
	spawnsThisFrame := 0
	for w.worldOffsetZ-w.lastSpawnZ >= w.baseSpawnDist && spawnsThisFrame < maxSpawnsPerFrame {
		if len(w.obstacles) < maxActiveObstacles {
			w.spawnObstacle()
		}
		w.lastSpawnZ += w.baseSpawnDist
		spawnsThisFrame++
	}

	// Удаляем препятствия, которые полностью позади камеры
	var remaining []*Obstacle
	for _, obs := range w.obstacles {
		if obs.WorldPos().Z > -10 {
			remaining = append(remaining, obs)
		}
	}
	w.obstacles = remaining
}

func (w *World) Draw(screen *ebiten.Image, cam *render.Camera) {
	for _, layer := range w.layers {
		layer.Draw(screen, cam, w)
	}
	for _, obs := range w.obstacles {
		obs.Draw(screen, cam)
	}
}

// Obstacles возвращает список препятствий (для коллизий)
func (w *World) Obstacles() []*Obstacle {
	return w.obstacles
}

// Реализация интерфейса WorldContext
func (w *World) GetSpeed() float64        { return w.speed }
func (w *World) GetWorldOffsetZ() float64 { return w.worldOffsetZ }

// Геттеры/сеттеры
func (w *World) Speed() float64        { return w.speed }
func (w *World) WorldOffsetZ() float64 { return w.worldOffsetZ }
func (w *World) Layers() []Layer       { return w.layers }

func (w *World) SetSpeed(speed float64)         { w.speed = speed }
func (w *World) SetWorldOffsetZ(offset float64) { w.worldOffsetZ = offset }
func (w *World) SetLayers(layers []Layer)       { w.layers = layers }

// SurfaceInfo и GetSurfaceAt остаются без изменений
type SurfaceInfo struct {
	Height  float64
	Type    SurfaceType
	Segment *Segment
}

func (w *World) GetSurfaceAt(z float64) (SurfaceInfo, bool) {
	var best SurfaceInfo
	best.Height = -math.MaxFloat64
	found := false

	for _, layer := range w.layers {
		if sp, ok := layer.(SurfaceProvider); ok {
			if h, st, ok := sp.SurfaceAt(z); ok {
				var seg *Segment
				if sl, ok := layer.(*SegmentLayer); ok {
					seg = sl.SegmentAt(z)
				}
				if !found || h > best.Height {
					best.Height = h
					best.Type = st
					best.Segment = seg
					found = true
				}
			}
		}
	}
	return best, found
}
