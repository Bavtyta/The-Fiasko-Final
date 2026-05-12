package state

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"TheFiaskoTest/internal/asset"
	"TheFiaskoTest/internal/config"
	"TheFiaskoTest/internal/entity"
	"TheFiaskoTest/internal/render"
	"TheFiaskoTest/internal/ui"
	"TheFiaskoTest/internal/world"
)

type GameState struct {
	manager    *Manager
	world      *world.World
	camera     *render.Camera
	player     *entity.Player
	balanceBar *ui.BalanceBarLayer
	score      float64
	driftDir   int // 1 - вправо, -1 - влево
	gameConfig config.GameConfig
	logTexture *ebiten.Image // Текстура бревна
}

func NewGameState(manager *Manager, gameCfg config.GameConfig, cameraCfg config.CameraConfig, physicsCfg config.PhysicsConfig) *GameState {
	w := world.New(50)

	skyLayer := world.NewSkyLayer(gameCfg.ScreenWidth, 300, 0.1)
	w.AddLayer(skyLayer)

	// Загружаем текстуру фона и создаём слой дальнего берега
	backgroundTexture := asset.LoadBackgroundTexture()
	farBankLayer := world.NewFarBankLayer(0, gameCfg.ScreenHeight, 0.2) // от 0 до высоты экрана
	farBankLayer.SetTexture(backgroundTexture)
	w.AddLayer(farBankLayer)

	// Загружаем текстуру реки
	riverTexture := asset.LoadRiverTexture()

	riverLayer := world.NewSegmentLayer(
		0, -25, 2000, 40, 0.3, 20,
		0.0, 0.25,
		color.RGBA{0, 100, 255, 255}, world.SurfaceLiquid,
	)
	riverLayer.SetTexture(riverTexture) // Устанавливаем текстуру для реки
	w.AddLayer(riverLayer)

	// Загружаем текстуру бревна
	logTexture := asset.LoadLogTexture()

	logLayer := world.NewSegmentLayer(0, -20, 10, 40, 1.0, 20, 0.0, 0.30, color.RGBA{139, 69, 19, 255}, world.SurfaceSolid)
	logLayer.SetTexture(logTexture) // Устанавливаем текстуру для слоя
	for _, seg := range logLayer.Segments() {
		seg.SetHeight(seg.Width())
		seg.SetRadialSegments(16)
	}
	w.AddLayer(logLayer)

	player := entity.NewPlayer(w, entity.PlayerConfig{
		StartX:       0,
		StartZ:       50,
		Width:        20.0,
		Height:       16.0,
		BalanceSpeed: 0.2,
		Physics:      physicsCfg,
		MaxTiltAngle: 0.8,
	})

	// Загружаем и устанавливаем текстуры игрока
	playerTexture := asset.LoadPlayerTexture()
	playerTextureRight := asset.LoadPlayerTextureRight()
	playerTextureJump := asset.LoadPlayerTextureJump()
	player.SetTexture(playerTexture)
	player.SetTextureRight(playerTextureRight)
	player.SetTextureJump(playerTextureJump)

	balanceBar := ui.NewBalanceBarLayer(
		func() float64 { return player.Balance() },
		func() float64 { return player.MaxBalance() },
		func() bool { return player.IsFalling() },
	)

	return &GameState{
		manager:    manager,
		world:      w,
		camera:     render.NewCamera(float64(gameCfg.ScreenWidth), float64(gameCfg.ScreenHeight), cameraCfg),
		player:     player,
		balanceBar: balanceBar,
		score:      0,
		driftDir:   1,
		gameConfig: gameCfg,
		logTexture: logTexture,
	}
}

func (g *GameState) Update() error {
	delta := 1.0 / 60.0
	g.world.Update(delta)

	if g.score >= g.gameConfig.DriftThreshold {
		if inpututil.IsKeyJustPressed(ebiten.KeyA) {
			g.driftDir = -1
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyD) {
			g.driftDir = 1
		}
	}

	effectiveDrift := 0
	if g.score >= g.gameConfig.DriftThreshold {
		effectiveDrift = g.driftDir
	}
	g.player.ApplyBalanceInput(effectiveDrift, delta)

	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.player.Jump(2.3)
	}

	g.player.Update(g.world, delta)

	// Проверка столкновений с препятствиями

	if !g.player.IsFalling() {
		g.score += 10.0 / 60.0
	} else {
		gameOver := NewGameOverState(g.manager, g.score, g.gameConfig)
		g.manager.ChangeState(gameOver, nil)
	}

	return nil
}

func (g *GameState) Draw(screen *ebiten.Image) {
	g.world.Draw(screen, g.camera)
	g.player.Draw(screen, g.camera, g.world)

	// Баланс-бар над игроком
	g.balanceBar.Draw(screen, g.camera, g.player.TiltedUpperWorldPos())

	// Отладочная информация
	balanceText := fmt.Sprintf("Balance: %.2f / %.0f", g.player.Balance(), g.player.MaxBalance())
	ebitenutil.DebugPrintAt(screen, balanceText, 10, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %.0f", g.score), 10, 30)

	if g.score >= g.gameConfig.DriftThreshold {
		dirStr := "RIGHT"
		if g.driftDir == -1 {
			dirStr = "LEFT"
		}
		ebitenutil.DebugPrintAt(screen, "Drift direction: "+dirStr+" (A/D to change)", 10, 50)
	} else {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Stable (need %.0f points)", g.gameConfig.DriftThreshold), 10, 50)
	}
}

func (g *GameState) Enter(prevState State, data interface{}) {}
func (g *GameState) Exit()                                   {}
