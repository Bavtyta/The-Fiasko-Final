package audio

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

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

	playlist     [][]byte
	currentTrack int
	trackCount   int
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
			instance = &SoundManager{
				ctx:          nil,
				musicVolume:  menuVol,
				effectsVol:   effVol,
				currentState: "menu",
				currentTrack: 0,
				trackCount:   0,
			}
		} else {
			instance = &SoundManager{
				ctx:          ctx,
				musicVolume:  menuVol,
				effectsVol:   effVol,
				currentState: "menu",
				currentTrack: 0,
				trackCount:   0,
			}
		}
	})
	return instance
}

// SetPlaylist задаёт плейлист (копия ссылок на []byte).
func (s *SoundManager) SetPlaylist(tracks [][]byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playlist = tracks
	s.trackCount = len(tracks)
	s.currentTrack = 0
}

// PlayCurrentTrack запускает текущий трек без зацикливания (по окончании можно переключить).
func (s *SoundManager) PlayCurrentTrack() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ctx == nil {
		log.Println("Warning: audio context is nil, cannot play music")
		return fmt.Errorf("audio context is nil")
	}
	if s.trackCount == 0 {
		return nil
	}
	if s.musicPlayer != nil {
		_ = s.musicPlayer.Close()
		s.musicPlayer = nil
	}

	data := s.playlist[s.currentTrack]
	reader := bytes.NewReader(data)
	stream, err := mp3.Decode(s.ctx, reader)
	if err != nil {
		return err
	}
	player, err := s.ctx.NewPlayer(stream)
	if err != nil {
		return err
	}
	s.musicPlayer = player
	s.musicPlayer.SetVolume(s.musicVolume)
	s.musicPlayer.Play()
	log.Printf("Info: music track %d/%d started, volume %.2f", s.currentTrack+1, s.trackCount, s.musicVolume)
	return nil
}

// NextTrack переключает на следующий трек по кругу и запускает воспроизведение.
func (s *SoundManager) NextTrack() error {
	s.mu.Lock()
	if s.ctx == nil {
		s.mu.Unlock()
		return fmt.Errorf("audio context is nil")
	}
	if s.trackCount == 0 {
		s.mu.Unlock()
		return nil
	}
	s.currentTrack = (s.currentTrack + 1) % s.trackCount
	if s.musicPlayer != nil {
		_ = s.musicPlayer.Close()
		s.musicPlayer = nil
	}
	data := s.playlist[s.currentTrack]
	s.mu.Unlock()

	reader := bytes.NewReader(data)
	stream, err := mp3.Decode(s.ctx, reader)
	if err != nil {
		return err
	}
	player, err := s.ctx.NewPlayer(stream)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.musicPlayer = player
	s.musicPlayer.SetVolume(s.musicVolume)
	s.musicPlayer.Play()
	s.mu.Unlock()
	log.Printf("Info: switched to music track %d/%d", s.currentTrack+1, s.trackCount)
	return nil
}

// IsMusicFinished — true, если есть плейлист и текущий плеер существует, но уже не играет.
func (s *SoundManager) IsMusicFinished() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.trackCount == 0 || s.ctx == nil {
		return false
	}
	if s.musicPlayer == nil {
		return false
	}
	return !s.musicPlayer.IsPlaying()
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

// PlayMusic воспроизводит один поток до конца (без зацикливания).
func (s *SoundManager) PlayMusic(r io.ReadSeeker) error {
	if r == nil {
		return fmt.Errorf("cannot play music: nil reader provided")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

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
	player, err := s.ctx.NewPlayer(stream)
	if err != nil {
		return err
	}
	s.musicPlayer = player
	s.musicPlayer.SetVolume(s.musicVolume)
	s.musicPlayer.Play()

	log.Printf("Info: music started playing, volume: %.2f", s.musicVolume)
	return nil
}

func (s *SoundManager) PlayEffect(r io.ReadSeeker) error {
	if r == nil {
		log.Println("Warning: attempt to play nil sound effect, skipping")
		return nil
	}

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

	log.Printf("Info: sound effect played, volume: %.2f", s.effectsVol)

	// Плеер не хранится в менеджере: после окончания нужно Close(), иначе накапливаются
	// декодеры/буферы oto и растёт потребление памяти при каждом эффекте.
	go closeEffectPlayerWhenIdle(player)

	return nil
}

func closeEffectPlayerWhenIdle(p *audio.Player) {
	if !p.IsPlaying() {
		_ = p.Close()
		return
	}
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(2 * time.Minute)
	defer deadline.Stop()
	for {
		select {
		case <-deadline.C:
			_ = p.Close()
			return
		case <-ticker.C:
			if !p.IsPlaying() {
				_ = p.Close()
				return
			}
		}
	}
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
