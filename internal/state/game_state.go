package state

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"

	"TheFiaskoTest/internal/achievement"
	"TheFiaskoTest/internal/asset"
	"TheFiaskoTest/internal/audio"
	"TheFiaskoTest/internal/config"
	"TheFiaskoTest/internal/entity"
	"TheFiaskoTest/internal/render"
	"TheFiaskoTest/internal/ui"
	"TheFiaskoTest/internal/world"

	uicommon "TheFiaskoTest/internal/ui/menu"
)

type GameState struct {
	manager     *Manager
	world       *world.World
	camera      *render.Camera
	player      *entity.Player
	balanceBar  *ui.BalanceBarLayer
	score       float64
	driftDir    int // 1 - вправо, -1 - влево
	gameConfig  config.GameConfig
	speedConfig config.SpeedConfig
	logTexture  *ebiten.Image // Текстура бревна
	hintAlpha   float64       // 1 = полностью видно, 0 = не видно
	hintFading  bool          // идёт ли затухание

	countdownActive bool    // идёт ли отсчёт после паузы
	countdownTime   float64 // оставшееся время отсчёта (сек)
	countdownValue  int     // отображаемая цифра (2 или 1)
	countdownAlpha  float64 // прозрачность текущей цифры
}

func NewGameState(manager *Manager, gameCfg config.GameConfig, cameraCfg config.CameraConfig, physicsCfg config.PhysicsConfig, speedCfg config.SpeedConfig) *GameState {
	w := world.New(speedCfg.InitialSpeed)

	skyLayer := world.NewSkyLayer(gameCfg.ScreenWidth, 300, 0.1)
	w.AddLayer(skyLayer)

	// Загружаем текстуру фона уровня и создаём слой дальнего берега
	backgroundTexture := asset.LoadGameBackgroundTexture()
	farBankLayer := world.NewFarBankLayer(0, gameCfg.ScreenHeight, 0.2) // от 0 до высоты экрана
	farBankLayer.SetTexture(backgroundTexture)
	w.AddLayer(farBankLayer)

	// Загружаем текстуру реки
	riverTexture := asset.LoadRiverTexture()

	riverLayer := world.NewSegmentLayer(
		0, -25, 500, 40, 0.3, 19,
		0.0, 0.15,
		color.RGBA{0, 100, 255, 255}, world.SurfaceLiquid,
	)
	riverLayer.SetTexture(riverTexture) // Устанавливаем текстуру для реки
	w.AddLayer(riverLayer)

	// Загружаем текстуру бревна
	logTexture := asset.LoadLogTexture()

	logLayer := world.NewSegmentLayer(0, -27, 20, 40, 1.0, 19, 0.0, 0.15, color.RGBA{139, 69, 19, 255}, world.SurfaceSolid)
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
		manager:     manager,
		world:       w,
		camera:      render.NewCamera(float64(gameCfg.ScreenWidth), float64(gameCfg.ScreenHeight), cameraCfg),
		player:      player,
		balanceBar:  balanceBar,
		score:       0,
		driftDir:    1,
		gameConfig:  gameCfg,
		speedConfig: speedCfg,
		logTexture:  logTexture,
		hintAlpha:   1.0,
		hintFading:  false,

		countdownActive: false,
		countdownTime:   0,
		countdownValue:  2,
		countdownAlpha:  1.0,
	}
}

func (g *GameState) Update() error {
	delta := 1.0 / 60.0

	if g.countdownActive {
		g.countdownTime -= delta
		if g.countdownTime <= 0 {
			g.countdownActive = false
			return nil
		}
		newValue := 1
		if g.countdownTime > 1 {
			newValue = 2
		}
		if newValue != g.countdownValue {
			g.countdownValue = newValue
			g.countdownAlpha = 1.0
			if newValue == 1 {
				if snd, err := asset.LoadCountdownSound(); err == nil {
					_ = audio.GetSoundManager().PlayEffect(snd)
				}
			}
		} else {
			if newValue == 2 {
				g.countdownAlpha -= delta * 0.5 // «2» на экране ~2 с — затухание за это время
			} else {
				g.countdownAlpha -= delta
			}
			if g.countdownAlpha < 0 {
				g.countdownAlpha = 0
			}
		}

		// В конце Update (перед return nil)
		achManager := achievement.GetManager()
		achManager.UpdateStats(g.score, g.player.Balance(), 0, g.world.WorldOffsetZ())
		// количество прыжков нужно увеличивать при прыжке:
		// добавьте поле jumpCount в GameState и увеличивайте при g.player.Jump()
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		pauseState := NewPauseState(g.manager, g, g.gameConfig)
		g.manager.ChangeState(pauseState, nil)
		return nil // важно: прекращаем обновление игры на этом кадре
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		soundMgr := audio.GetSoundManager()
		if err := soundMgr.NextTrack(); err != nil {
			log.Printf("Warning: could not switch track: %v", err)
		}
	}

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

	g.player.Update(g.world) // Исправленный вызов без delta

	// Проверка столкновений с препятствиями (эллипс игрока × эллипс препятствия)
	const collisionZThreshold = 5.0
	pc, pSemiT, pSemiR, pAng := g.player.CollisionEllipse()
	collided := false

	for _, obs := range g.world.Obstacles() {
		if obs.CollidesPlayerEllipse(pc.X, pc.Y, pc.Z, pSemiT, pSemiR, pAng, collisionZThreshold) {
			collided = true
			break
		}
	}

	if collided || g.player.IsFalling() {
		// Воспроизводим звук проигрыша
		soundMgr := audio.GetSoundManager()
		loseSound, err := asset.LoadLoseSound()
		if err != nil {
			// Логируем ошибку, но не прерываем игру
			// log.Printf("Warning: could not load lose sound: %v", err)
		} else {
			if err := soundMgr.PlayEffect(loseSound); err != nil {
				// log.Printf("Warning: could not play lose sound: %v", err)
			}
		}

		gameOver := NewGameOverState(g.manager, g.score, g.gameConfig)
		g.manager.ChangeState(gameOver, nil)
		return nil
	}

	// Обновляем скорость в зависимости от очков
	newSpeed := g.speedConfig.InitialSpeed + g.score*g.speedConfig.SpeedIncreasePerScore
	if newSpeed > g.speedConfig.MaxSpeed {
		newSpeed = g.speedConfig.MaxSpeed
	}
	g.world.SetSpeed(newSpeed)

	g.score += 10.0 / 60.0    // 10 очков в секунду
	g.world.SetScore(g.score) // передаём счёт в мир

	// Плавное исчезновение подсказки после достижения порога дрифта
	if !g.hintFading && g.score >= g.gameConfig.DriftThreshold {
		g.hintFading = true
	}
	if g.hintFading && g.hintAlpha > 0 {
		g.hintAlpha -= 0.02 // скорость затухания (~3 секунды при 60 FPS)
		if g.hintAlpha < 0 {
			g.hintAlpha = 0
		}
	}

	soundMgr := audio.GetSoundManager()
	if soundMgr.IsMusicFinished() {
		if err := soundMgr.NextTrack(); err != nil {
			log.Printf("Warning: could not auto-advance track: %v", err)
		}
	}

	return nil
}

func (g *GameState) Draw(screen *ebiten.Image) {
	g.world.Draw(screen, g.camera)
	g.player.Draw(screen, g.camera, g.world)

	// Баланс-бар над игроком
	g.balanceBar.Draw(screen, g.camera, g.player.TiltedUpperWorldPos())

	// HUD: только счёт в левом верхнем углу
	x, y := 25, 40
	shadowOffset := 2

	scoreStr := fmt.Sprintf("Score: %.0f", g.score)
	text.Draw(screen, scoreStr, uicommon.FontHUD, x+shadowOffset, y+shadowOffset, color.RGBA{0, 0, 0, 180})
	text.Draw(screen, scoreStr, uicommon.FontHUD, x, y, color.White)

	// Подсказка по управлению балансом (исчезает после порога очков)
	if g.hintAlpha > 0 {
		hintStr := "Press A/D to change balance"
		bounds := text.BoundString(uicommon.FontHUD, hintStr)
		hintX := (g.gameConfig.ScreenWidth - bounds.Dx()) / 2
		hintY := g.gameConfig.ScreenHeight - 580

		col := color.RGBA{255, 255, 255, uint8(g.hintAlpha * 255)}
		shadowCol := color.RGBA{0, 0, 0, uint8(g.hintAlpha * 180)}

		text.Draw(screen, hintStr, uicommon.FontHUD,
			hintX+shadowOffset, hintY+shadowOffset, shadowCol)
		text.Draw(screen, hintStr, uicommon.FontHUD,
			hintX, hintY, col)
	}

	if g.countdownActive {
		numStr := fmt.Sprintf("%d", g.countdownValue)
		bounds := text.BoundString(uicommon.FontTitle, numStr)
		x := (g.gameConfig.ScreenWidth - bounds.Dx()) / 2
		y := g.gameConfig.ScreenHeight / 2

		alpha := uint8(g.countdownAlpha * 255)
		col := color.RGBA{255, 255, 255, alpha}
		shadowCol := color.RGBA{0, 0, 0, uint8(float64(alpha) * 0.7)}

		text.Draw(screen, numStr, uicommon.FontTitle, x+4, y+4, shadowCol)
		text.Draw(screen, numStr, uicommon.FontTitle, x, y, col)
	}
}

func (g *GameState) Enter(prevState State, data interface{}) {
	g.hintAlpha = 1.0
	g.hintFading = false

	if data == "resume" {
		g.countdownActive = true
		g.countdownTime = 3.0 // 2 с на «2» + 1 с на «1»
		g.countdownValue = 2
		g.countdownAlpha = 1.0
	} else {
		g.countdownActive = false
	}
}
func (g *GameState) Exit() {}
