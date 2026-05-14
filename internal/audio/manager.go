package audio

import (
	"fmt"
	"io"
	"log"
	"sync"

	"TheFiaskoTest/internal/config"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

type SoundManager struct {
	musicPlayer  *audio.Player
	musicVolume  float64
	effectsVol   float64
	ctx          *audio.Context
	mu           sync.Mutex
	currentState string
}

var (
	instance     *SoundManager
	instanceOnce sync.Once
)

func GetSoundManager() *SoundManager {
	instanceOnce.Do(func() {
		menuVol := config.DefaultSoundConfig().MusicVolumeMenu
		effVol := config.DefaultSoundConfig().EffectsVolume
		ctx := audio.NewContext(48000)
		if ctx == nil {
			log.Println("Warning: failed to create audio context, audio will be disabled")
			// Создаем менеджер с nil контекстом, но не падаем
			instance = &SoundManager{
				ctx:          nil,
				musicVolume:  menuVol,
				effectsVol:   effVol,
				currentState: "menu",
			}
		} else {
			instance = &SoundManager{
				ctx:          ctx,
				musicVolume:  menuVol,
				effectsVol:   effVol,
				currentState: "menu",
			}
		}
	})
	return instance
}

func (s *SoundManager) SetMusicVolume(vol float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.musicVolume = vol
	if s.musicPlayer != nil && s.musicPlayer.IsPlaying() {
		s.musicPlayer.SetVolume(vol)
	}
}

func (s *SoundManager) SetEffectsVolume(vol float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.effectsVol = vol
}

func (s *SoundManager) PlayMusic(r io.ReadSeeker) error {
	if r == nil {
		return fmt.Errorf("cannot play music: nil reader provided")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, что аудио контекст существует
	if s.ctx == nil {
		log.Println("Warning: audio context is nil, cannot play music")
		return fmt.Errorf("audio context is nil")
	}

	if s.musicPlayer != nil {
		_ = s.musicPlayer.Close()
	}
	stream, err := mp3.Decode(s.ctx, r)
	if err != nil {
		return err
	}
	loop := audio.NewInfiniteLoop(stream, stream.Length())
	player, err := s.ctx.NewPlayer(loop)
	if err != nil {
		return err
	}
	s.musicPlayer = player
	s.musicPlayer.SetVolume(s.musicVolume)
	s.musicPlayer.Play()

	// Логируем для отладки
	log.Printf("Info: Music started playing, volume: %.2f", s.musicVolume)
	return nil
}

func (s *SoundManager) PlayEffect(r io.ReadSeeker) error {
	if r == nil {
		// Просто логируем и выходим, чтобы не прерывать игру из-за отсутствия звука.
		log.Println("Warning: attempt to play nil sound effect, skipping")
		return nil
	}

	// Проверяем, что аудио контекст существует
	if s.ctx == nil {
		log.Println("Warning: audio context is nil, cannot play sound effect")
		return fmt.Errorf("audio context is nil")
	}

	stream, err := mp3.Decode(s.ctx, r)
	if err != nil {
		return err
	}
	player, err := s.ctx.NewPlayer(stream)
	if err != nil {
		return err
	}
	player.SetVolume(s.effectsVol)
	player.Play()

	// Логируем для отладки
	log.Printf("Info: Sound effect played, volume: %.2f", s.effectsVol)

	// Не закрываем – объект будет собран GC после завершения
	return nil
}

func (s *SoundManager) OnStateChange(state string) {
	s.currentState = state
	sc := config.DefaultSoundConfig()
	switch state {
	case "menu":
		s.SetMusicVolume(sc.MusicVolumeMenu)
	case "game":
		s.SetMusicVolume(sc.MusicVolumeGame)
	}
}
