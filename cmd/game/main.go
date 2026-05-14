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

	trackCount, err := asset.LoadAllMusicTracks()
	if err != nil {
		log.Printf("Warning: could not load music tracks: %v", err)
	} else {
		tracks := make([][]byte, 0, trackCount)
		for i := 0; i < trackCount; i++ {
			data, err := asset.GetMusicTrackData(i)
			if err == nil {
				tracks = append(tracks, data)
			}
		}
		soundMgr.SetPlaylist(tracks)
		if err := soundMgr.PlayCurrentTrack(); err != nil {
			log.Printf("Warning: could not play first track: %v", err)
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
