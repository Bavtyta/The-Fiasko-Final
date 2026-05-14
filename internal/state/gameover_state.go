package state

import (
	"TheFiaskoTest/internal/asset"
	"TheFiaskoTest/internal/audio"
	"TheFiaskoTest/internal/config"
	uicommon "TheFiaskoTest/internal/ui/menu"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type GameOverState struct {
	manager    *Manager
	score      float64
	gameConfig config.GameConfig
	background *ebiten.Image

	menuItems   []string
	selectedIdx int

	buttonWidth  float32
	buttonHeight float32
	buttonX      float32
	buttonY      []float32
}

func NewGameOverState(manager *Manager, score float64, gameCfg config.GameConfig) *GameOverState {
	g := &GameOverState{
		manager:      manager,
		score:        score,
		gameConfig:   gameCfg,
		background:   asset.LoadBackgroundTexture(),
		menuItems:    []string{"ЗАНОВО", "НАСТРОЙКИ", "ГЛАВНОЕ МЕНЮ"},
		selectedIdx:  0,
		buttonHeight: 50,
	}
	g.updateButtonPositions()
	return g
}

func (g *GameOverState) updateButtonPositions() {
	maxWidth := 0
	for _, item := range g.menuItems {
		if w := text.BoundString(uicommon.FontButton, item).Dx(); w > maxWidth {
			maxWidth = w
		}
	}
	g.buttonWidth = float32(maxWidth) + 80
	g.buttonX, g.buttonY = uicommon.ComputeButtonPositions(g.menuItems, g.buttonWidth, g.buttonHeight, 20, g.gameConfig.ScreenWidth, g.gameConfig.ScreenHeight, 350)
}

func (g *GameOverState) Update() error {
	keyPressed := false
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.selectedIdx = (g.selectedIdx - 1 + len(g.menuItems)) % len(g.menuItems)
		keyPressed = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.selectedIdx = (g.selectedIdx + 1) % len(g.menuItems)
		keyPressed = true
	}

	// Воспроизводим звук нажатия клавиши
	if keyPressed {
		soundMgr := audio.GetSoundManager()
		keySound, err := asset.LoadKeypressSound()
		if err == nil {
			soundMgr.PlayEffect(keySound)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		// Воспроизводим звук выбора
		soundMgr := audio.GetSoundManager()
		keySound, err := asset.LoadKeypressSound()
		if err == nil {
			soundMgr.PlayEffect(keySound)
		}

		switch g.selectedIdx {
		case 0: // ЗАНОВО
			gameState := NewGameState(g.manager, g.gameConfig,
				config.DefaultCameraConfig(),
				config.DefaultPhysicsConfig(),
				config.DefaultSpeedConfig())
			g.manager.ChangeState(gameState, nil)
		case 2: // ГЛАВНОЕ МЕНЮ
			mainMenu := NewMainMenuState(g.manager, g.gameConfig)
			g.manager.ChangeState(mainMenu, nil)
		}
	}
	return nil
}

func (g *GameOverState) Draw(screen *ebiten.Image) {
	uicommon.DrawBackground(screen, g.background, g.gameConfig.ScreenWidth, g.gameConfig.ScreenHeight)
	uicommon.DrawTitle(screen, "ЭТО ФИАСКО БРАТАН)))", 120, g.gameConfig.ScreenWidth)
	uicommon.DrawScore(screen, g.score, 200, g.gameConfig.ScreenWidth)

	for i, label := range g.menuItems {
		uicommon.DrawButton(screen, g.buttonX, g.buttonY[i], g.buttonWidth, g.buttonHeight, label, i == g.selectedIdx)
	}
}

func (g *GameOverState) Enter(prevState State, data interface{}) { g.selectedIdx = 0 }
func (g *GameOverState) Exit()                                   {}
