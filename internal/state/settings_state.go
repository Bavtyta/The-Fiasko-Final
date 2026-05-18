package state

import (
	"TheFiaskoTest/internal/asset"
	"TheFiaskoTest/internal/audio"
	"TheFiaskoTest/internal/config"
	uicommon "TheFiaskoTest/internal/ui/menu"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type SettingsState struct {
	manager     *Manager
	gameConfig  config.GameConfig
	background  *ebiten.Image
	returnState State // состояние, в которое вернуться

	menuItems   []string // "Режим звуков", "Громкость музыки", "Громкость эффектов"
	selectedIdx int
	// Значения настроек
	modeIndex        int // 0=музыка, 1..5 = история 1..5 (всего 6 режимов)
	musicVolumeIdx   int // 0..5: 0%,20%,40%,55%,80%,100%
	effectsVolumeIdx int

	// Для отображения
	modes        []string
	volumeLabels []string // "0%","20%","40%","55%","80%","100%"
}

func NewSettingsState(manager *Manager, gameCfg config.GameConfig, returnTo State) *SettingsState {
	modes := []string{"Музыка", "История 1", "История 2", "История 3", "История 4", "История 5"}
	volLabels := []string{"0%", "20%", "40%", "55%", "80%", "100%"}

	// Загружаем сохранённые настройки (если есть, иначе по умолчанию)
	musicIdx := loadMusicVolumeIdx() // реализуем временное сохранение в переменные
	effectsIdx := loadEffectsVolumeIdx()
	mode := loadSoundMode()

	return &SettingsState{
		manager:          manager,
		gameConfig:       gameCfg,
		background:       asset.LoadBackgroundTexture(),
		returnState:      returnTo,
		menuItems:        []string{"Режим звуков", "Громкость музыки", "Громкость эффектов", "НАЗАД"},
		selectedIdx:      0,
		modeIndex:        mode,
		musicVolumeIdx:   musicIdx,
		effectsVolumeIdx: effectsIdx,
		modes:            modes,
		volumeLabels:     volLabels,
	}
}

// временное глобальное хранилище настроек (в реальном проекте сохраняйте в JSON)
var (
	settingsMusicVolumeIdx   = 4 // 80% по умолчанию
	settingsEffectsVolumeIdx = 4
	settingsSoundMode        = 0
)

func loadMusicVolumeIdx() int      { return settingsMusicVolumeIdx }
func loadEffectsVolumeIdx() int    { return settingsEffectsVolumeIdx }
func loadSoundMode() int           { return settingsSoundMode }
func saveMusicVolumeIdx(idx int)   { settingsMusicVolumeIdx = idx }
func saveEffectsVolumeIdx(idx int) { settingsEffectsVolumeIdx = idx }
func saveSoundMode(idx int)        { settingsSoundMode = idx }

func (s *SettingsState) Update() error {
	keyPressed := false
	valueChanged := false

	// Навигация по пунктам меню
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		s.selectedIdx = (s.selectedIdx - 1 + len(s.menuItems)) % len(s.menuItems)
		keyPressed = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.selectedIdx = (s.selectedIdx + 1) % len(s.menuItems)
		keyPressed = true
	}

	if keyPressed {
		if snd, err := asset.LoadKeypressSound(); err == nil {
			audio.GetSoundManager().PlayEffect(snd)
		}
	}

	// Изменение значения выбранного пункта (влево/вправо)
	if s.selectedIdx == 0 { // Режим звуков
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
			s.modeIndex = (s.modeIndex - 1 + len(s.modes)) % len(s.modes)
			saveSoundMode(s.modeIndex)
			s.applyModeChange()
			valueChanged = true
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
			s.modeIndex = (s.modeIndex + 1) % len(s.modes)
			saveSoundMode(s.modeIndex)
			s.applyModeChange()
			valueChanged = true
		}
	} else if s.selectedIdx == 1 { // Громкость музыки
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
			s.musicVolumeIdx = (s.musicVolumeIdx - 1 + 6) % 6
			saveMusicVolumeIdx(s.musicVolumeIdx)
			s.applyMusicVolume()
			valueChanged = true
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
			s.musicVolumeIdx = (s.musicVolumeIdx + 1) % 6
			saveMusicVolumeIdx(s.musicVolumeIdx) // димон здарова от ромы
			s.applyMusicVolume()
			valueChanged = true
		}
	} else if s.selectedIdx == 2 { // Громкость эффектов
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
			s.effectsVolumeIdx = (s.effectsVolumeIdx - 1 + 6) % 6
			saveEffectsVolumeIdx(s.effectsVolumeIdx)
			s.applyEffectsVolume()
			valueChanged = true
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
			s.effectsVolumeIdx = (s.effectsVolumeIdx + 1) % 6
			saveEffectsVolumeIdx(s.effectsVolumeIdx)
			s.applyEffectsVolume()
			valueChanged = true
		}
	}

	if valueChanged {
		if snd, err := asset.LoadKeypressSound(); err == nil {
			audio.GetSoundManager().PlayEffect(snd)
		}
	}

	// Обработка выхода (Enter на "НАЗАД" или Escape)
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && s.selectedIdx == 3 {
		if snd, err := asset.LoadKeypressSound(); err == nil {
			audio.GetSoundManager().PlayEffect(snd)
		}
		s.manager.ChangeState(s.returnState, nil)
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if snd, err := asset.LoadKeypressSound(); err == nil {
			audio.GetSoundManager().PlayEffect(snd)
		}
		s.manager.ChangeState(s.returnState, nil)
		return nil
	}
	return nil
}

func (s *SettingsState) applyModeChange() {
	// Устанавливаем режим звука в аудио-менеджере
	soundMgr := audio.GetSoundManager()
	if err := soundMgr.SetSoundMode(s.modeIndex); err != nil {
		log.Printf("Warning: could not set sound mode to %d: %v", s.modeIndex, err)
	}
}

func (s *SettingsState) applyMusicVolume() {
	// Получаем множитель от 0 до 1
	mult := s.getVolumeMultiplier(s.musicVolumeIdx)
	// Устанавливаем громкость музыки в зависимости от текущего состояния (меню или игра)
	soundMgr := audio.GetSoundManager()
	// Сохраняем множитель в конфиг или применяем к текущей громкости
	// Текущая громкость уже установлена в OnStateChange, нужно её умножить на mult.
	// Лучше хранить базовые громкости в soundMgr и умножать на множитель из настроек.
	// Для простоты сделаем так: получаем текущую целевую громкость (0.4 или 0.75) и умножаем на mult.
	currentBase := 0.4
	if soundMgr.CurrentState() == "game" {
		currentBase = 0.75
	}
	soundMgr.SetUserMusicVolume(mult)
	soundMgr.SetMusicVolume(currentBase * mult)
}

func (s *SettingsState) applyEffectsVolume() {
	mult := s.getVolumeMultiplier(s.effectsVolumeIdx)
	audio.GetSoundManager().SetEffectsVolume(mult * 0.8) // базовый эффект 0.8
}

func (s *SettingsState) getVolumeMultiplier(idx int) float64 {
	switch idx {
	case 0:
		return 0.0
	case 1:
		return 0.2
	case 2:
		return 0.4
	case 3:
		return 0.55
	case 4:
		return 0.8
	case 5:
		return 1.0
	default:
		return 1.0
	}
}

func (s *SettingsState) Draw(screen *ebiten.Image) {
	uicommon.DrawBackground(screen, s.background, s.gameConfig.ScreenWidth, s.gameConfig.ScreenHeight)
	uicommon.DrawTitle(screen, "НАСТРОЙКИ", 80, s.gameConfig.ScreenWidth)

	// Подготовим тексты кнопок с учётом значений
	buttonTexts := make([]string, len(s.menuItems))
	for i, label := range s.menuItems {
		if i == 0 {
			buttonTexts[i] = label + ": " + s.modes[s.modeIndex]
		} else if i == 1 {
			buttonTexts[i] = label + ": " + s.volumeLabels[s.musicVolumeIdx]
		} else if i == 2 {
			buttonTexts[i] = label + ": " + s.volumeLabels[s.effectsVolumeIdx]
		} else {
			buttonTexts[i] = label
		}
	}

	// Вычисляем максимальную ширину текста, чтобы установить ширину кнопок
	maxWidth := 0
	for _, txt := range buttonTexts {
		if w := text.BoundString(uicommon.FontButton, txt).Dx(); w > maxWidth {
			maxWidth = w
		}
	}
	buttonWidth := float32(maxWidth) + 80
	buttonHeight := float32(50)
	gap := float32(20)
	startY := float32(180)
	screenW := float32(s.gameConfig.ScreenWidth)
	buttonX := (screenW - buttonWidth) / 2

	for i, txt := range buttonTexts {
		y := startY + float32(i)*(buttonHeight+gap)
		uicommon.DrawButton(screen, buttonX, y, buttonWidth, buttonHeight, txt, i == s.selectedIdx)
	}
}

func (s *SettingsState) Enter(prevState State, data interface{}) {
	// Обновляем текущие настройки из глобальных переменных
	s.musicVolumeIdx = loadMusicVolumeIdx()
	s.effectsVolumeIdx = loadEffectsVolumeIdx()
	s.modeIndex = loadSoundMode()
	s.applyMusicVolume()
	s.applyEffectsVolume()
}

func (s *SettingsState) Exit() {}
