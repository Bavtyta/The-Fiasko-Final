package state

import (
	"image/color"
	"math"
	"strings"

	"TheFiaskoTest/internal/achievement"
	"TheFiaskoTest/internal/asset"
	"TheFiaskoTest/internal/config"
	uicommon "TheFiaskoTest/internal/ui/menu"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

const (
	achievementsTitleBaselineY = 120
	achievementLineHeight      = 76
	achievementRowCount        = 5
	separatorStrokeWidth       = 1.0
)

type AchievementsState struct {
	manager      *Manager
	gameConfig   config.GameConfig
	background   *ebiten.Image
	panelX       float32
	panelY       float32
	panelWidth   float32
	panelHeight  float32
	achievements []*achievement.Achievement
	selectedIdx  int
	scrollOffset int // если больше 5, можно скроллить, но пока 5
}

func NewAchievementsState(manager *Manager, gameCfg config.GameConfig) *AchievementsState {
	// Получаем список достижений из менеджера
	achManager := achievement.GetManager()
	achList := achManager.GetAchievements()

	return &AchievementsState{
		manager:      manager,
		gameConfig:   gameCfg,
		background:   asset.LoadBackgroundTexture(),
		achievements: achList,
		selectedIdx:  0,
	}
}

func (a *AchievementsState) updatePanelDimensions() {
	screenW := float32(a.gameConfig.ScreenWidth)
	screenH := float32(a.gameConfig.ScreenHeight)
	a.panelWidth = screenW / 3
	a.panelHeight = screenH
	a.panelX = (screenW - a.panelWidth) / 2
	a.panelY = 0
}

func (a *AchievementsState) Update() error {
	a.updatePanelDimensions()

	// Навигация по списку достижений (только если их больше 1)
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		a.selectedIdx = (a.selectedIdx - 1 + len(a.achievements)) % len(a.achievements)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		a.selectedIdx = (a.selectedIdx + 1) % len(a.achievements)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		mainMenu := NewMainMenuState(a.manager, a.gameConfig)
		a.manager.ChangeState(mainMenu, nil)
		return nil
	}

	return nil
}

func (a *AchievementsState) Draw(screen *ebiten.Image) {
	// Фон
	uicommon.DrawBackground(screen, a.background, a.gameConfig.ScreenWidth, a.gameConfig.ScreenHeight)

	// Полупрозрачная серая панель
	panelRect := ebiten.NewImage(int(a.panelWidth), int(a.panelHeight))
	panelRect.Fill(color.RGBA{80, 80, 80, 200}) // полупрозрачный серый
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(a.panelX), float64(a.panelY))
	screen.DrawImage(panelRect, op)

	// Заголовок как в главном меню / паузе (FontTitle + тень, по центру экрана)
	title := "ДОСТИЖЕНИЯ"
	uicommon.DrawTitle(screen, title, achievementsTitleBaselineY, a.gameConfig.ScreenWidth)

	// Отступ под заголовком = отступу от верха экрана до верхней границы текста
	titleBounds := text.BoundString(uicommon.FontTitle, title)
	marginTop := achievementsTitleBaselineY + titleBounds.Min.Y
	startY := achievementsTitleBaselineY + titleBounds.Max.Y + marginTop

	circleRadius := 15
	visible := len(a.achievements)
	if visible > achievementRowCount {
		visible = achievementRowCount
	}

	// Тонкие разделители между строками: длина 1/5 ширины экрана, по центру серой панели
	lineLen := float32(a.gameConfig.ScreenWidth) / 5
	centerX := a.panelX + a.panelWidth/2
	x0 := centerX - lineLen/2
	x1 := centerX + lineLen/2
	sepColor := color.RGBA{0, 0, 0, 255}
	for i := 0; i < visible-1; i++ {
		sepY := float32(startY + (i+1)*achievementLineHeight)
		vector.StrokeLine(screen, x0, sepY, x1, sepY, separatorStrokeWidth, sepColor, true)
	}

	for i := 0; i < visible; i++ {
		ach := a.achievements[i]
		y := startY + i*achievementLineHeight

		nameX := int(a.panelX) + 20
		circleX := int(a.panelX + a.panelWidth - float32(circleRadius) - 20)
		nameMaxW := circleX - circleRadius - 12 - nameX
		if nameMaxW < 24 {
			nameMaxW = 24
		}
		rowCenterY := y + achievementLineHeight/2
		drawAchievementNameCenteredInRow(screen, ach.Name, nameX, rowCenterY, nameMaxW)

		// Круг по правой части панели, по вертикали по центру строки
		circleY := y + achievementLineHeight/2

		if ach.Unlocked {
			// Золотой круг (заливка)
			vector.DrawFilledCircle(screen, float32(circleX), float32(circleY), float32(circleRadius), color.RGBA{255, 215, 0, 255}, true)
		} else {
			// Только белая обводка
			vector.StrokeCircle(screen, float32(circleX), float32(circleY), float32(circleRadius), 2, color.RGBA{255, 255, 255, 255}, true)
		}
	}
}

func (a *AchievementsState) Enter(prevState State, data interface{}) {
	// Обновляем список достижений на случай, если за время в другом состоянии они изменились
	a.achievements = achievement.GetManager().GetAchievements()
	a.selectedIdx = 0
}

func (a *AchievementsState) Exit() {}

func achievementHUDLineSpacing(face font.Face) int {
	m := face.Metrics()
	h := m.Height.Ceil()
	if h < 1 {
		return text.BoundString(face, "M").Dy()
	}
	return h
}

func wrapAchievementLines(face font.Face, s string, maxW int) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if maxW < 1 {
		maxW = 1
	}
	var lines []string
	var cur strings.Builder

	pushCur := func() {
		if cur.Len() == 0 {
			return
		}
		lines = append(lines, strings.TrimSpace(cur.String()))
		cur.Reset()
	}

	breakLongWord := func(word string) {
		frag := ""
		for _, r := range word {
			ch := string(r)
			next := frag + ch
			if frag == "" && text.BoundString(face, next).Dx() > maxW {
				lines = append(lines, ch)
				continue
			}
			if text.BoundString(face, next).Dx() <= maxW {
				frag = next
				continue
			}
			if frag != "" {
				lines = append(lines, frag)
			}
			frag = ch
		}
		if frag != "" {
			cur.WriteString(frag)
		}
	}

	for _, word := range strings.Fields(s) {
		try := word
		if cur.Len() > 0 {
			try = cur.String() + " " + word
		}
		if text.BoundString(face, try).Dx() <= maxW {
			if cur.Len() > 0 {
				cur.WriteString(" ")
			}
			cur.WriteString(word)
			continue
		}
		pushCur()
		if text.BoundString(face, word).Dx() <= maxW {
			cur.WriteString(word)
			continue
		}
		breakLongWord(word)
	}
	pushCur()
	return lines
}

func achievementFirstBaseline(face font.Face, lines []string, lineSpacing, rowCenterY int) int {
	if len(lines) == 0 {
		return rowCenterY
	}
	minA, maxC := math.MaxInt, math.MinInt
	for i, line := range lines {
		b := text.BoundString(face, line)
		ai := i*lineSpacing + b.Min.Y
		ci := i*lineSpacing + b.Max.Y
		if ai < minA {
			minA = ai
		}
		if ci > maxC {
			maxC = ci
		}
	}
	return rowCenterY - (minA+maxC)/2
}

func drawAchievementNameCenteredInRow(screen *ebiten.Image, name string, leftX, rowCenterY, maxW int) {
	face := uicommon.FontHUD
	lines := wrapAchievementLines(face, name, maxW)
	if len(lines) == 0 {
		return
	}
	sp := achievementHUDLineSpacing(face)
	base := achievementFirstBaseline(face, lines, sp, rowCenterY)
	for i, line := range lines {
		text.Draw(screen, line, face, leftX, base+i*sp, uicommon.ColorWhite)
	}
}
