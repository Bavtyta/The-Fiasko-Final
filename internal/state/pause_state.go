package state

import (
	"image/color"

	"TheFiaskoTest/internal/asset"
	"TheFiaskoTest/internal/audio"
	"TheFiaskoTest/internal/config"
	uicommon "TheFiaskoTest/internal/ui/menu"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type PauseState struct {
	manager      *Manager
	gameState    *GameState // сохранённое игровое состояние
	gameConfig   config.GameConfig
	background   *ebiten.Image
	overlay      *ebiten.Image // один раз на время паузы; не создавать каждый кадр в Draw
	menuItems    []string
	selectedIdx  int
	buttonWidth  float32
	buttonHeight float32
	buttonX      float32
	buttonY      []float32
}

func NewPauseState(manager *Manager, gameState *GameState, gameCfg config.GameConfig) *PauseState {
	p := &PauseState{
		manager:      manager,
		gameState:    gameState,
		gameConfig:   gameCfg,
		background:   asset.LoadBackgroundTexture(),
		menuItems:    []string{"ПРОДОЛЖИТЬ", "НАСТРОЙКИ", "ГЛАВНОЕ МЕНЮ"},
		selectedIdx:  0,
		buttonHeight: 50,
	}
	p.updateButtonPositions()
	return p
}

func (p *PauseState) updateButtonPositions() {
	maxWidth := 0
	for _, item := range p.menuItems {
		if w := text.BoundString(uicommon.FontButton, item).Dx(); w > maxWidth {
			maxWidth = w
		}
	}
	p.buttonWidth = float32(maxWidth) + 80
	p.buttonX, p.buttonY = uicommon.ComputeButtonPositions(p.menuItems, p.buttonWidth, p.buttonHeight, 20, p.gameConfig.ScreenWidth, p.gameConfig.ScreenHeight, 200)
}

func (p *PauseState) Update() error {
	keyPressed := false

	// Навигация
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		p.selectedIdx = (p.selectedIdx - 1 + len(p.menuItems)) % len(p.menuItems)
		keyPressed = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		p.selectedIdx = (p.selectedIdx + 1) % len(p.menuItems)
		keyPressed = true
	}

	// Звук навигации
	if keyPressed {
		if snd, err := asset.LoadKeypressSound(); err == nil {
			audio.GetSoundManager().PlayEffect(snd)
		}
	}

	// Обработка выбора
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if snd, err := asset.LoadKeypressSound(); err == nil {
			audio.GetSoundManager().PlayEffect(snd)
		}
		switch p.selectedIdx {
		case 0: // ПРОДОЛЖИТЬ
			p.manager.ChangeState(p.gameState, "resume")
		case 1: // НАСТРОЙКИ
			settingsState := NewSettingsState(p.manager, p.gameConfig, p)
			p.manager.ChangeState(settingsState, nil)
		case 2: // ГЛАВНОЕ МЕНЮ
			mainMenu := NewMainMenuState(p.manager, p.gameConfig)
			p.manager.ChangeState(mainMenu, nil)
		}
	}

	// Escape тоже выходит из паузы
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if snd, err := asset.LoadKeypressSound(); err == nil {
			audio.GetSoundManager().PlayEffect(snd)
		}
		p.manager.ChangeState(p.gameState, "resume")
	}
	return nil
}

func (p *PauseState) Draw(screen *ebiten.Image) {
	// Рисуем фон и затемнение
	uicommon.DrawBackground(screen, p.background, p.gameConfig.ScreenWidth, p.gameConfig.ScreenHeight)
	if p.overlay != nil {
		p.overlay.Fill(color.RGBA{0, 0, 0, 180})
		screen.DrawImage(p.overlay, nil)
	}

	uicommon.DrawTitle(screen, "PAUSE", 100, p.gameConfig.ScreenWidth)

	for i, label := range p.menuItems {
		uicommon.DrawButton(screen, p.buttonX, p.buttonY[i], p.buttonWidth, p.buttonHeight, label, i == p.selectedIdx)
	}
}

func (p *PauseState) Enter(prevState State, data interface{}) {
	p.selectedIdx = 0
	p.ensurePauseOverlay()
}

func (p *PauseState) Exit() {
	if p.overlay != nil {
		p.overlay.Dispose()
		p.overlay = nil
	}
}

func (p *PauseState) ensurePauseOverlay() {
	w, h := p.gameConfig.ScreenWidth, p.gameConfig.ScreenHeight
	if p.overlay != nil {
		b := p.overlay.Bounds()
		if b.Dx() == w && b.Dy() == h {
			return
		}
		p.overlay.Dispose()
		p.overlay = nil
	}
	p.overlay = ebiten.NewImage(w, h)
}
