package state

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"

	"TheFiaskoTest/internal/config"
)

type MenuItem struct {
	Label    string
	Icon     string
	Action   func()
	Selected bool
}

type MainMenuState struct {
	manager     *Manager
	menuItems   []MenuItem
	selectedIdx int
	font        font.Face
	fontBold    font.Face
	titleFont   font.Face
	windowTitle string
	version     string
	statusText  string
	statusHint  string
	animFrame   int
	gameConfig  config.GameConfig
}

func NewMainMenuState(manager *Manager, gameCfg config.GameConfig) *MainMenuState {
	// Загрузка шрифтов
	tt, err := opentype.Parse(gomono.TTF)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72

	// Обычный шрифт
	fontFace, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    14,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Жирный шрифт для заголовков
	fontBold, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    16,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Шрифт для заголовка игры
	titleFont, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    32,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	m := &MainMenuState{
		manager:     manager,
		font:        fontFace,
		fontBold:    fontBold,
		titleFont:   titleFont,
		windowTitle: "THE FIASKO.EXE",
		gameConfig:  gameCfg,
	}

	// Инициализация пунктов меню
	m.menuItems = []MenuItem{
		{Label: "ИГРАТЬ", Icon: "▶", Action: m.onPlay},
		{Label: "НАСТРОЙКИ", Icon: "⚙", Action: m.onSettings},
		{Label: "ДОСТИЖЕНИЯ", Icon: "★", Action: m.onAchievements},
		{Label: "ВЫХОД", Icon: "×", Action: m.onExit},
	}
	m.menuItems[0].Selected = true

	return m
}

func (m *MainMenuState) onPlay() {
	m.statusText = "ЗАПУСК ИГРЫ..."
	// Переключаемся на игровое состояние
	gameState := NewGameState(m.manager, m.gameConfig, config.DefaultCameraConfig(), config.DefaultPhysicsConfig())
	m.manager.ChangeState(gameState, nil)
}

func (m *MainMenuState) onSettings() {
	m.statusText = "НАСТРОЙКИ"
}

func (m *MainMenuState) onAchievements() {
	m.statusText = "ДОСТИЖЕНИЯ"
}

func (m *MainMenuState) onExit() {
	m.statusText = "ВЫХОД..."
	// Можно добавить логику выхода
}

func (m *MainMenuState) Update() error {
	m.animFrame++

	// Навигация стрелками
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		m.menuItems[m.selectedIdx].Selected = false
		m.selectedIdx = (m.selectedIdx + 1) % len(m.menuItems)
		m.menuItems[m.selectedIdx].Selected = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		m.menuItems[m.selectedIdx].Selected = false
		m.selectedIdx = (m.selectedIdx - 1 + len(m.menuItems)) % len(m.menuItems)
		m.menuItems[m.selectedIdx].Selected = true
	}

	// Выбор Enter
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if m.selectedIdx < len(m.menuItems) {
			m.menuItems[m.selectedIdx].Action()
		}
	}

	return nil
}

func (m *MainMenuState) Draw(screen *ebiten.Image) {
	// Фон — чёрный
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Рисуем фоновый узор (scanlines)
	m.drawScanlines(screen)

	// Рисуем окно меню
	m.drawWindow(screen)
}

func (m *MainMenuState) drawScanlines(screen *ebiten.Image) {
	scanlineColor := color.RGBA{10, 10, 10, 255}
	screenW := float32(m.gameConfig.ScreenWidth)

	for y := 0; y < m.gameConfig.ScreenHeight; y += 4 {
		vector.StrokeLine(screen, 0, float32(y), screenW, float32(y), 2, scanlineColor, true)
	}

	// Градиентные блики (Windows style)
	for i := 0; i < 20; i++ {
		alpha := uint8(20 - i)
		blueTint := color.RGBA{0, 120, 212, alpha}
		vector.DrawFilledRect(screen, 0, float32(i*20), screenW, 20, blueTint, true)
	}
}

func (m *MainMenuState) drawWindow(screen *ebiten.Image) {
	windowX := float32(120)
	windowY := float32(80)
	windowW := float32(400)
	windowH := float32(320)

	// Тень окна
	shadowColor := color.RGBA{255, 255, 255, 30}
	vector.DrawFilledRect(screen, windowX+8, windowY+8, windowW, windowH, shadowColor, true)

	// Рамка окна (3px белая)
	borderColor := color.RGBA{255, 255, 255, 255}
	vector.StrokeRect(screen, windowX, windowY, windowW, windowH, 6, borderColor, true)
	vector.StrokeRect(screen, windowX+3, windowY+3, windowW-6, windowH-6, 2, color.RGBA{0, 0, 0, 255}, true)

	// Title bar (белый)
	titleBarH := float32(36)
	vector.DrawFilledRect(screen, windowX+3, windowY+3, windowW-6, titleBarH, borderColor, true)

	// Текст заголовка окна
	titleColor := color.RGBA{0, 0, 0, 255}
	text.Draw(screen, m.windowTitle, m.font, int(windowX)+16, int(windowY)+24, titleColor)

	// Кнопки управления окном (символы)
	btnY := int(windowY) + 10
	btnX := int(windowX) + int(windowW) - 80
	buttons := []string{"−", "□", "×"}
	for i, btn := range buttons {
		// Рамка кнопки
		vector.StrokeRect(screen, float32(btnX+i*26), float32(btnY), 22, 22, 2, color.RGBA{0, 0, 0, 255}, true)
		text.Draw(screen, btn, m.fontBold, btnX+i*26+6, btnY+16, titleColor)
	}

	// Контент окна
	contentY := windowY + titleBarH + 20

	// Заголовок игры
	titleStr := "THE FIASKO"
	titleX := windowX + windowW/2 - float32(len(titleStr)*8)
	text.Draw(screen, titleStr, m.titleFont, int(titleX), int(contentY), borderColor)

	// Подзаголовок версии
	versionY := int(contentY) + 30
	text.Draw(screen, m.version, m.font, int(windowX)+int(windowW)/2-50, versionY, color.RGBA{200, 200, 200, 255})

	// Разделитель
	sepY := float32(versionY + 20)
	for x := windowX + 40; x < windowX+windowW-40; x += 20 {
		vector.DrawFilledRect(screen, x, sepY, 12, 2, color.RGBA{255, 255, 255, 128}, true)
	}

	// Кнопки меню
	buttonStartY := sepY + 30
	buttonHeight := float32(44)
	buttonGap := float32(16)

	for i, item := range m.menuItems {
		btnY := buttonStartY + float32(i)*(buttonHeight+buttonGap)

		// Пропуск для разделителя перед Exit
		if i == 3 {
			btnY += 20
		}

		m.drawMenuButton(screen, windowX+40, btnY, windowW-80, buttonHeight, item, i == m.selectedIdx)
	}

	// Status bar
	statusY := windowY + windowH - 30
	vector.DrawFilledRect(screen, windowX+3, statusY, windowW-6, 27, borderColor, true)

	statusTextColor := color.RGBA{0, 0, 0, 255}
	text.Draw(screen, m.statusText, m.font, int(windowX)+16, int(statusY)+18, statusTextColor)

	hintWidth := len(m.statusHint) * 7
	text.Draw(screen, m.statusHint, m.font, int(windowX+windowW)-hintWidth-16, int(statusY)+18, color.RGBA{100, 100, 100, 255})
}

func (m *MainMenuState) drawMenuButton(screen *ebiten.Image, x, y, w, h float32, item MenuItem, selected bool) {
	// Цвета в зависимости от состояния
	bgColor := color.RGBA{0, 0, 0, 255}
	textColor := color.RGBA{255, 255, 255, 255}
	borderColor := color.RGBA{255, 255, 255, 255}

	if selected {
		bgColor = color.RGBA{0, 120, 212, 255} // Windows blue
		textColor = color.RGBA{255, 255, 255, 255}

		// Сдвиг при выборе (hover effect)
		x += 8
		w -= 8
	}

	// Фон кнопки
	vector.DrawFilledRect(screen, x, y, w, h, bgColor, true)

	// Рамка (3px)
	vector.StrokeRect(screen, x, y, w, h, 6, borderColor, true)
	vector.StrokeRect(screen, x+3, y+3, w-6, h-6, 2, color.RGBA{0, 0, 0, 255}, true)

	// Индикатор фокуса (треугольник слева)
	if selected {
		focusX := x - 30
		focusY := y + h/2
		vector.DrawFilledRect(screen, focusX, focusY-8, 16, 16, color.RGBA{0, 120, 212, 255}, true)
	}

	// Иконка
	iconX := int(x) + 20
	textY := int(y) + int(h/2) + 6
	text.Draw(screen, item.Icon, m.fontBold, iconX, textY, textColor)

	// Текст кнопки (центрированный)
	textWidth := len(item.Label) * 9
	textX := int(x) + int(w)/2 - textWidth/2
	text.Draw(screen, item.Label, m.fontBold, textX, textY, textColor)
}

func (m *MainMenuState) Enter(prevState State, data interface{}) {}
func (m *MainMenuState) Exit()                                   {}
