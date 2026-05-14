package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"TheFiaskoTest/internal/asset"
	"TheFiaskoTest/internal/audio"
	"TheFiaskoTest/internal/common"
	"TheFiaskoTest/internal/config"
	"TheFiaskoTest/internal/core"
	"TheFiaskoTest/internal/render"
	"TheFiaskoTest/internal/world"
)

type PlayerConfig struct {
	StartX, StartZ, Width, Height float64
	BalanceSpeed                  float64
	Physics                       config.PhysicsConfig
	MaxTiltAngle                  float64 // максимальный угол наклона в радианах
}

type Player struct {
	world          *world.World
	position       core.Vec3
	width          float64
	height         float64
	color          color.Color
	texture        *ebiten.Image // текстура игрока (левая)
	textureRight   *ebiten.Image // текстура игрока (правая)
	textureJump    *ebiten.Image // текстура игрока (прыжок)
	animationFrame int           // счётчик кадров для анимации
	currentTexture int           // 0 = левая, 1 = правая
	isJumping      bool
	jumpVelocity   float64
	groundY        float64
	balance        float64
	maxBalance     float64
	balanceSpeed   float64
	isFalling      bool
	physics        config.PhysicsConfig
	maxTiltAngle   float64
	currentSegment *world.Segment // текущее бревно, на котором стоит игрок
	standingRadius float64        // радиус текущего бревна
	jumpOffset     float64        // вертикальное смещение при прыжке
}

func NewPlayer(world *world.World, cfg PlayerConfig) *Player {
	return &Player{
		world:        world,
		position:     core.Vec3{X: cfg.StartX, Y: 0, Z: cfg.StartZ},
		width:        cfg.Width,
		height:       cfg.Height,
		color:        color.White,
		groundY:      0,
		balance:      0,
		maxBalance:   10,
		balanceSpeed: cfg.BalanceSpeed,
		isFalling:    false,
		physics:      cfg.Physics,
		maxTiltAngle: cfg.MaxTiltAngle,
	}
}

// Update реализует интерфейс Entity (только ctx), использует фиксированный delta = 1/60.
func (p *Player) Update(ctx common.WorldContext) {
	p.update(ctx, 1.0/60.0)
}

// Внутренний метод обновления с переданным delta.
func (p *Player) update(ctx common.WorldContext, delta float64) {
	if p.isFalling {
		return
	}

	// Анимация: переключаем текстуру каждые 20 кадров
	p.animationFrame++
	if p.animationFrame >= 20 {
		p.animationFrame = 0
		p.currentTexture = 1 - p.currentTexture // переключаем между 0 и 1
	}

	z := p.position.Z
	info, ok := p.world.GetSurfaceAt(z)
	if ok {
		p.groundY = info.Height
		if info.Segment != nil {
			p.currentSegment = info.Segment
			p.standingRadius = info.Segment.Width() / 2
			// Центр вращения в центре бревна
			p.position.X = info.Segment.X() + info.Segment.SlopeX()*z
			p.position.Y = info.Height - p.standingRadius
		} else {
			// Плоская поверхность (без сегмента)
			p.currentSegment = nil
			p.standingRadius = 0
			p.position.Y = info.Height
		}
	} else {
		p.isFalling = true
		return
	}

	// Прыжок
	if p.isJumping {
		p.jumpOffset += p.jumpVelocity * delta * 60
		p.jumpVelocity -= p.physics.Gravity * delta * 60
		if p.jumpOffset <= 0 {
			p.jumpOffset = 0
			p.isJumping = false
			p.jumpVelocity = 0
		}
	} else {
		p.jumpOffset = 0
	}
}

func (p *Player) Jump(initialVelocity float64) {
	if !p.isJumping && p.position.Y <= p.groundY+0.01 {
		p.isJumping = true
		p.jumpVelocity = initialVelocity

		// Воспроизводим звук прыжка
		soundMgr := audio.GetSoundManager()
		jumpSound, err := asset.LoadJumpSound()
		if err != nil {
			// Логируем ошибку, но не прерываем игру
			// log.Printf("Warning: could not load jump sound: %v", err)
		} else {
			if err := soundMgr.PlayEffect(jumpSound); err != nil {
				// log.Printf("Warning: could not play jump sound: %v", err)
			}
		}
	}
}

func (p *Player) ApplyBalanceInput(driftDir int, delta float64) {
	if p.isFalling {
		return
	}
	deltaBalance := p.balanceSpeed * float64(driftDir) * delta * 60
	p.balance += deltaBalance
	if p.balance > p.maxBalance {
		p.balance = p.maxBalance
		p.isFalling = true
	} else if p.balance < -p.maxBalance {
		p.balance = -p.maxBalance
		p.isFalling = true
	}
}

// Draw отрисовывает игрока с учётом наклона (вращение вокруг центра нижней стороны).
func (p *Player) Draw(screen *ebiten.Image, cam *render.Camera, ctx common.WorldContext) {
	if p.isFalling {
		return
	}

	factor := p.balance / p.maxBalance
	if factor < -1 {
		factor = -1
	} else if factor > 1 {
		factor = 1
	}
	theta := factor * p.maxTiltAngle

	halfW := p.width / 2
	h := p.height
	r := p.standingRadius

	// Локальные координаты относительно центра вращения (p.position)
	local := [][2]float64{
		{-halfW, r + p.jumpOffset},     // левый нижний
		{halfW, r + p.jumpOffset},      // правый нижний
		{-halfW, r + h + p.jumpOffset}, // левый верхний
		{halfW, r + h + p.jumpOffset},  // правый верхний
	}

	cosT := math.Cos(theta)
	sinT := math.Sin(theta)

	var worldPts [4]core.Vec3
	for i, l := range local {
		// Поворот по часовой стрелке
		rx := l[0]*cosT + l[1]*sinT
		ry := -l[0]*sinT + l[1]*cosT
		worldPts[i] = core.Vec3{
			X: p.position.X + rx,
			Y: p.position.Y + ry,
			Z: p.position.Z,
		}
	}

	// Проецируем на экран
	var screenPts [4][2]float64
	for i, wp := range worldPts {
		sx, sy, scale := cam.Project(wp)
		if scale <= 0 {
			return
		}
		screenPts[i] = [2]float64{sx, sy}
	}

	// Если есть текстура, рисуем с текстурой
	currentTex := p.texture
	if p.isJumping && p.textureJump != nil {
		currentTex = p.textureJump
	} else if p.currentTexture == 1 && p.textureRight != nil {
		currentTex = p.textureRight
	}

	if currentTex != nil {
		bounds := currentTex.Bounds()
		texW := float32(bounds.Dx())
		texH := float32(bounds.Dy())

		vertices := []ebiten.Vertex{
			{DstX: float32(screenPts[0][0]), DstY: float32(screenPts[0][1]), SrcX: 0, SrcY: texH, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
			{DstX: float32(screenPts[1][0]), DstY: float32(screenPts[1][1]), SrcX: texW, SrcY: texH, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
			{DstX: float32(screenPts[3][0]), DstY: float32(screenPts[3][1]), SrcX: texW, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
			{DstX: float32(screenPts[2][0]), DstY: float32(screenPts[2][1]), SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		}
		indices := []uint16{0, 1, 2, 0, 2, 3}
		opts := &ebiten.DrawTrianglesOptions{
			Filter: ebiten.FilterLinear,
		}
		screen.DrawTriangles(vertices, indices, currentTex, opts)
	} else {
		col := p.color
		ebitenutil.DrawLine(screen, screenPts[0][0], screenPts[0][1], screenPts[1][0], screenPts[1][1], col)
		ebitenutil.DrawLine(screen, screenPts[1][0], screenPts[1][1], screenPts[3][0], screenPts[3][1], col)
		ebitenutil.DrawLine(screen, screenPts[3][0], screenPts[3][1], screenPts[2][0], screenPts[2][1], col)
		ebitenutil.DrawLine(screen, screenPts[2][0], screenPts[2][1], screenPts[0][0], screenPts[0][1], col)
	}

	// Рисуем круг зоны столкновения для игрока
	p.drawCollisionCircle(screen, cam)
}

// TiltedUpperWorldPos возвращает мировые координаты центра верхней грани после наклона.
func (p *Player) TiltedUpperWorldPos() core.Vec3 {
	if p.isFalling {
		return core.Vec3{}
	}
	factor := p.balance / p.maxBalance
	if factor < -1 {
		factor = -1
	} else if factor > 1 {
		factor = 1
	}
	theta := factor * p.maxTiltAngle
	cosT := math.Cos(theta)
	sinT := math.Sin(theta)

	r := p.standingRadius
	rx := 0*cosT + (r+p.height+p.jumpOffset)*sinT
	ry := -0*sinT + (r+p.height+p.jumpOffset)*cosT
	return core.Vec3{
		X: p.position.X + rx,
		Y: p.position.Y + ry,
		Z: p.position.Z,
	}
}

// UpperWorldPos возвращает центр верхней грани без наклона (если нужно).
func (p *Player) UpperWorldPos() core.Vec3 {
	return core.Vec3{
		X: p.position.X,
		Y: p.position.Y + p.height,
		Z: p.position.Z,
	}
}

// GetZ возвращает глубину для сортировки.
func (p *Player) GetZ() float64 {
	return p.position.Z
}

func (p *Player) SetZ(z float64) {
	p.position.Z = z
}

func (p *Player) Balance() float64    { return p.balance }
func (p *Player) MaxBalance() float64 { return p.maxBalance }
func (p *Player) IsFalling() bool     { return p.isFalling }
func (p *Player) Position() core.Vec3 { return p.position }
func (p *Player) Width() float64      { return p.width }
func (p *Player) Height() float64     { return p.height }

// SetTexture устанавливает текстуру для игрока
func (p *Player) SetTexture(texture *ebiten.Image) {
	p.texture = texture
}

// SetTextureRight устанавливает правую текстуру для игрока
func (p *Player) SetTextureRight(texture *ebiten.Image) {
	p.textureRight = texture
}

// SetTextureJump устанавливает текстуру прыжка для игрока
func (p *Player) SetTextureJump(texture *ebiten.Image) {
	p.textureJump = texture
}

// CollisionRadius возвращает радиус аппроксимирующей сферы для коллизий.
func (p *Player) CollisionRadius() float64 {
	// Фиксированный радиус 6, как requested
	return 6.0
}

// drawCollisionCircle рисует круг зоны столкновения вокруг игрока
func (p *Player) drawCollisionCircle(screen *ebiten.Image, cam *render.Camera) {
	// Получаем центр круга столкновения в мировых координатах
	center := p.CollisionCircleCenter()

	// Проецируем центр на экран
	sx, sy, scale := cam.Project(center)
	if scale <= 0 {
		return
	}

	// Вычисляем радиус столкновения в мировых координатах
	collisionRadius := p.CollisionRadius()

	// Масштабируем радиус для экранных координат
	screenRadius := collisionRadius * scale

	// Цвет круга: зеленый с прозрачностью для игрока
	circleColor := color.RGBA{0, 255, 0, 150} // Зеленый для игрока

	// Рисуем круг с помощью нескольких линий (аппроксимация круга)
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

	// Также рисуем тонкий контур для лучшей видимости
	outlineColor := color.RGBA{0, 255, 0, 255} // Полностью зеленый
	const outlineSegments = 32
	for i := 0; i < outlineSegments; i++ {
		angle1 := 2 * math.Pi * float64(i) / float64(outlineSegments)
		angle2 := 2 * math.Pi * float64(i+1) / float64(outlineSegments)

		x1 := sx + (screenRadius+0.5)*math.Cos(angle1) // Немного больше радиуса
		y1 := sy + (screenRadius+0.5)*math.Sin(angle1)
		x2 := sx + (screenRadius+0.5)*math.Cos(angle2)
		y2 := sy + (screenRadius+0.5)*math.Sin(angle2)

		ebitenutil.DrawLine(screen, x1, y1, x2, y2, outlineColor)
	}
}

// SpriteCenter возвращает мировые координаты центра спрайта игрока,
// учитывая наклон, дрифт, прыжок и положение на бревне.
func (p *Player) SpriteCenter() core.Vec3 {
	r := p.standingRadius
	theta := p.balance / p.maxBalance * p.maxTiltAngle
	// Ограничение угла (на всякий случай)
	if theta > p.maxTiltAngle {
		theta = p.maxTiltAngle
	} else if theta < -p.maxTiltAngle {
		theta = -p.maxTiltAngle
	}
	effectiveR := r + p.jumpOffset
	return core.Vec3{
		X: p.position.X + (effectiveR+p.height/2)*math.Sin(theta),
		Y: p.position.Y + (effectiveR+p.height/2)*math.Cos(theta),
		Z: p.position.Z,
	}
}

// CollisionCircleCenter возвращает мировые координаты центра круга столкновения.
// Центр расположен так, что нижняя точка круга совпадает с нижней границей текстуры игрока.
func (p *Player) CollisionCircleCenter() core.Vec3 {
	r := p.standingRadius
	theta := p.balance / p.maxBalance * p.maxTiltAngle
	// Ограничение угла (на всякий случай)
	if theta > p.maxTiltAngle {
		theta = p.maxTiltAngle
	} else if theta < -p.maxTiltAngle {
		theta = -p.maxTiltAngle
	}
	effectiveR := r + p.jumpOffset
	collisionRadius := p.CollisionRadius()

	// Центр круга находится на высоте (effectiveR + collisionRadius) от центра бревна
	// чтобы нижняя точка круга была на уровне effectiveR
	return core.Vec3{
		X: p.position.X + (effectiveR+collisionRadius)*math.Sin(theta),
		Y: p.position.Y + (effectiveR+collisionRadius)*math.Cos(theta),
		Z: p.position.Z,
	}
}
