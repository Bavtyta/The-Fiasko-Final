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

type MainMenuState struct {
	manager     *Manager
	menuItems   []string
	selectedIdx int
	gameConfig  config.GameConfig
	shouldExit  bool
	background  *ebiten.Image

	buttonWidth  float32
	buttonHeight float32
	buttonX      float32
	buttonY      []float32
}

func NewMainMenuState(manager *Manager, gameCfg config.GameConfig) *MainMenuState {
	m := &MainMenuState{
		manager:      manager,
		menuItems:    []string{"ИГРАТЬ", "НАСТРОЙКИ", "ДОСТИЖЕНИЯ", "ВЫХОД"},
		selectedIdx:  0,
		gameConfig:   gameCfg,
		background:   asset.LoadBackgroundTexture(),
		buttonHeight: 50,
	}
	m.updateButtonPositions()
	return m
}

func (m *MainMenuState) updateButtonPositions() {
	// Вычисляем ширину по самому длинному тексту
	maxWidth := 0
	for _, item := range m.menuItems {
		if w := text.BoundString(uicommon.FontButton, item).Dx(); w > maxWidth {
			maxWidth = w
		}
	}
	m.buttonWidth = float32(maxWidth) + 80
	m.buttonX, m.buttonY = uicommon.ComputeButtonPositions(m.menuItems, m.buttonWidth, m.buttonHeight, 20, m.gameConfig.ScreenWidth, m.gameConfig.ScreenHeight, 200)
}

func (m *MainMenuState) Update() error {
	// навигация как раньше
	keyPressed := false
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		m.selectedIdx = (m.selectedIdx - 1 + len(m.menuItems)) % len(m.menuItems)
		keyPressed = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		m.selectedIdx = (m.selectedIdx + 1) % len(m.menuItems)
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

		switch m.selectedIdx {
		case 0:
			gameState := NewGameState(m.manager, m.gameConfig,
				config.DefaultCameraConfig(),
				config.DefaultPhysicsConfig(),
				config.DefaultSpeedConfig())
			m.manager.ChangeState(gameState, nil)
		case 1:
			// TODO: настройки
		case 2:
			// TODO: достижения
		case 3:
			m.shouldExit = true
		}
	}
	if m.shouldExit {
		return ebiten.Termination
	}
	return nil
}

func (m *MainMenuState) Draw(screen *ebiten.Image) {
	uicommon.DrawBackground(screen, m.background, m.gameConfig.ScreenWidth, m.gameConfig.ScreenHeight)
	uicommon.DrawTitle(screen, "THE FIASKO", 120, m.gameConfig.ScreenWidth)

	for i, label := range m.menuItems {
		uicommon.DrawButton(screen, m.buttonX, m.buttonY[i], m.buttonWidth, m.buttonHeight, label, i == m.selectedIdx)
	}
}

func (m *MainMenuState) Enter(prevState State, data interface{}) {
	m.selectedIdx = 0
	m.shouldExit = false
}
func (m *MainMenuState) Exit() {}
