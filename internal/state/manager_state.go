package state

import (
	"log"

	"TheFiaskoTest/internal/audio"
	"TheFiaskoTest/internal/config"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Manager struct {
	current    State
	next       State
	data       interface{}
	gameConfig config.GameConfig
}

func NewManager(initial State, gameCfg config.GameConfig) *Manager {
	return &Manager{
		current:    initial,
		gameConfig: gameCfg,
	}
}

func (m *Manager) Update() error {
	// Глобальная обработка кнопки N для переключения музыки
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		soundMgr := audio.GetSoundManager()
		if err := soundMgr.NextTrack(); err != nil {
			log.Printf("Warning: could not switch track: %v", err)
		}
	}

	if m.next != nil {
		if m.current != nil { // проверка, чтобы не вызвать Exit на nil
			m.current.Exit()
		}
		m.current = m.next
		m.syncMusicVolumeForCurrentState()
		m.current.Enter(nil, m.data)
		m.next = nil
		m.data = nil
	}
	if m.current == nil {
		return nil // или вернуть ошибку, но лучше не допускать такого состояния
	}
	return m.current.Update()
}

func (m *Manager) Draw(screen *ebiten.Image) {
	m.current.Draw(screen)
}

func (m *Manager) ChangeState(state State, data interface{}) {
	m.next = state
	m.data = data
}

// syncMusicVolumeForCurrentState выставляет громкость музыки один раз при фактической смене состояния:
// тихо везде, кроме активного геймплея (GameState).
func (m *Manager) syncMusicVolumeForCurrentState() {
	if m.current == nil {
		return
	}
	switch m.current.(type) {
	case *GameState:
		audio.GetSoundManager().OnStateChange("game")
	default:
		audio.GetSoundManager().OnStateChange("menu")
	}
}

// GameConfig возвращает конфигурацию игры
func (m *Manager) GameConfig() config.GameConfig {
	return m.gameConfig
}
