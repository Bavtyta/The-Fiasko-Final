package world

import (
	"math"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"TheFiaskoTest/internal/asset"
	"TheFiaskoTest/internal/common"
	"TheFiaskoTest/internal/render"
)

const maxActiveObstacles = 6

var (
	sugrobTexture     *ebiten.Image
	sugrobTextureOnce sync.Once
)

type Layer interface {
	Update(ctx common.WorldContext, delta float64)
	Draw(screen *ebiten.Image, cam *render.Camera, ctx common.WorldContext)
}

// SurfaceProvider определяет возможность получить информацию о поверхности в точке Z.
type SurfaceProvider interface {
	SurfaceAt(z float64) (height float64, surfaceType SurfaceType, ok bool)
}

type World struct {
	worldOffsetZ  float64
	speed         float64
	initialSpeed  float64
	score         float64
	layers        []Layer
	obstacles     []*Obstacle
	baseSpawnDist float64
	lastSpawnZ    float64
}

func New(speed float64) *World {
	return &World{
		worldOffsetZ:  0,
		speed:         speed,
		initialSpeed:  speed,
		score:         0,
		baseSpawnDist: 50.0,
		lastSpawnZ:    -1, // -1 означает "ещё не инициализировано"
		layers:        []Layer{},
		obstacles:     []*Obstacle{},
	}
}

// AddObstacle добавляет препятствие в мир
func (w *World) AddObstacle(obs *Obstacle) {
	w.obstacles = append(w.obstacles, obs)
}

// getSugrobTexture возвращает текстуру сугроба (ленивая загрузка)
func getSugrobTexture() *ebiten.Image {
	sugrobTextureOnce.Do(func() {
		sugrobTexture = asset.LoadSugrobTexture()
	})
	return sugrobTexture
}

// effectiveSpawnDist возвращает дистанцию между препятствиями в единицах Z.
// Плотность по времени постоянна: чем выше скорость, тем больше дистанция.
func (w *World) effectiveSpawnDist() float64 {
	if w.initialSpeed == 0 {
		return w.baseSpawnDist
	}
	return w.baseSpawnDist * (w.speed / w.initialSpeed)
}

// initialSpawnOffset возвращает абсолютную Z для первого препятствия:
// середина первого сегмента твёрдого слоя + 100 единиц.
func (w *World) initialSpawnOffset(solidLayer *SegmentLayer) float64 {
	segs := solidLayer.Segments()
	if len(segs) == 0 {
		return 0
	}
	seg := segs[0] // первый (ближайший к камере) сегмент
	if seg == nil {
		return 0
	}
	midOfFirst := seg.NearZ() + seg.Length()/2
	const extraOffset = 250.0
	return midOfFirst + extraOffset
}

// spawnObstacleAt создаёт препятствие на заданной абсолютной Z
func (w *World) spawnObstacleAt(spawnZ float64) {
	const minAngle = math.Pi * -60 / 180.0
	const maxAngle = math.Pi * 60.0 / 180.0
	angle := minAngle + rand.Float64()*(maxAngle-minAngle)

	texture := getSugrobTexture()
	obs := NewObstacleAtZ(spawnZ, angle, 5, 8, texture)
	obs.UpdateSurface(w) // сразу кэшируем поверхность
	w.AddObstacle(obs)
}

func (w *World) Update(delta float64) {
	w.worldOffsetZ += w.speed * delta

	for _, layer := range w.layers {
		layer.Update(w, delta)
	}

	// Обновляем позиции и поверхности всех препятствий
	for _, obs := range w.obstacles {
		obs.Update(w.speed, delta)
		obs.UpdateSurface(w)
	}

	// Инициализация lastSpawnZ при первом вызове Update
	if w.lastSpawnZ < 0 {
		var solidLayer *SegmentLayer
		for _, layer := range w.layers {
			if sl, ok := layer.(*SegmentLayer); ok && sl.SurfaceType() == SurfaceSolid {
				solidLayer = sl
				break
			}
		}
		if solidLayer != nil {
			firstSpawnZ := w.initialSpawnOffset(solidLayer)
			// Ставим lastSpawnZ так, чтобы первое препятствие создалось на firstSpawnZ
			w.lastSpawnZ = firstSpawnZ - w.effectiveSpawnDist()
		} else {
			w.lastSpawnZ = 0
		}
	}

	// Спавн новых препятствий
	spawnDist := w.effectiveSpawnDist()
	for w.worldOffsetZ-w.lastSpawnZ >= spawnDist {
		targetZ := w.lastSpawnZ + spawnDist
		if len(w.obstacles) < maxActiveObstacles {
			w.spawnObstacleAt(targetZ)
		}
		w.lastSpawnZ = targetZ
	}

	// Удаляем препятствия, ушедшие за камеру (Z < -10)
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
func (w *World) SetScore(score float64) {
	w.score = score
}

// SurfaceInfo и GetSurfaceAt
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

func (w *World) AddLayer(l Layer) {
	w.layers = append(w.layers, l)
}
