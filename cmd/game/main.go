// cmd/game/main.go
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"TheFiaskoTest/internal/asset"
	"TheFiaskoTest/internal/audio"

	"TheFiaskoTest/internal/config"
	"TheFiaskoTest/internal/state"
)

func main() {
	// Создаём конфигурацию игры
	gameCfg := config.DefaultGameConfig()

	// Создаём менеджер состояний с начальным состоянием MainMenu
	manager := state.NewManager(nil, gameCfg) // временно nil
	mainMenuState := state.NewMainMenuState(manager, gameCfg)
	manager.ChangeState(mainMenuState, nil)

	// Инициализируем звуковой менеджер
	soundMgr := audio.GetSoundManager()

	// Пытаемся загрузить и запустить музыку
	musicReader, err := asset.LoadMusicTrack()
	if err != nil {
		// Логируем ошибку, но не паникуем. Игра продолжит работу без фоновой музыки.
		log.Printf("Warning: could not load music track: %v", err)
	} else {
		// Музыка загружена, пытаемся её проиграть
		if err := soundMgr.PlayMusic(musicReader); err != nil {
			log.Printf("Warning: could not play music: %v", err)
		} else {
			log.Printf("Info: Music started playing")
		}
	}

	game := &Game{manager: manager}

	ebiten.SetWindowSize(gameCfg.ScreenWidth, gameCfg.ScreenHeight)
	ebiten.SetWindowTitle("The Fiasko")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	manager *state.Manager
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	return g.manager.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.manager.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	cfg := g.manager.GameConfig()
	return cfg.ScreenWidth, cfg.ScreenHeight
}
