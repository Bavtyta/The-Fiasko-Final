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

	// Режим воспроизведения: 0 = музыка (плейлист), 1-5 = история (одна история)
	soundMode    int
	playlist     [][]byte
	stories      [][]byte
	currentTrack int
	trackCount   int
	storyCount   int

	userMusicVolume   float64 // 0..1, из настроек
	userEffectsVolume float64
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
				ctx:               nil,
				musicVolume:       menuVol,
				effectsVol:        effVol,
				currentState:      "menu",
				currentTrack:      0,
				trackCount:        0,
				userMusicVolume:   1.0,
				userEffectsVolume: 1.0,
			}
		} else {
			instance = &SoundManager{
				ctx:               ctx,
				musicVolume:       menuVol,
				effectsVol:        effVol,
				currentState:      "menu",
				soundMode:         0, // 0 = музыка по умолчанию
				currentTrack:      0,
				trackCount:        0,
				storyCount:        0,
				userMusicVolume:   1.0,
				userEffectsVolume: 1.0,
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

// SetStories задаёт список историй.
func (s *SoundManager) SetStories(stories [][]byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stories = stories
	s.storyCount = len(stories)
}

// SetSoundMode устанавливает режим звука: 0 = музыка, 1-5 = история (индекс 0-4)
func (s *SoundManager) SetSoundMode(mode int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if mode < 0 || mode > 5 {
		return fmt.Errorf("invalid sound mode: %d (must be 0-5)", mode)
	}

	s.soundMode = mode

	// Если переключаемся в режим истории, останавливаем текущую музыку
	if mode > 0 {
		if s.musicPlayer != nil {
			_ = s.musicPlayer.Close()
			s.musicPlayer = nil
		}
		// Воспроизводим выбранную историю
		storyIndex := mode - 1 // конвертируем 1-5 в 0-4
		if storyIndex < s.storyCount {
			data := s.stories[storyIndex]
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
			s.updateMusicVolume()
			s.musicPlayer.Play()
			log.Printf("Info: started story %d, volume %.2f", storyIndex+1, s.musicVolume)
		}
	} else {
		// Переключаемся в режим музыки, воспроизводим текущий трек
		if s.trackCount > 0 {
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
			s.updateMusicVolume()
			s.musicPlayer.Play()
			log.Printf("Info: switched to music mode, track %d/%d", s.currentTrack+1, s.trackCount)
		}
	}

	return nil
}

// PlayCurrentTrack запускает текущий трек без зацикливания (по окончании можно переключить).
// В режиме музыки: воспроизводит текущий трек плейлиста.
// В режиме истории: воспроизводит выбранную историю.
func (s *SoundManager) PlayCurrentTrack() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ctx == nil {
		log.Println("Warning: audio context is nil, cannot play music")
		return fmt.Errorf("audio context is nil")
	}

	if s.soundMode > 0 {
		// Режим истории
		storyIndex := s.soundMode - 1
		if storyIndex >= s.storyCount {
			return nil // нет такой истории
		}
		if s.musicPlayer != nil {
			_ = s.musicPlayer.Close()
			s.musicPlayer = nil
		}

		data := s.stories[storyIndex]
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
		s.updateMusicVolume()
		s.musicPlayer.Play()
		log.Printf("Info: story %d/%d started, volume %.2f", storyIndex+1, s.storyCount, s.musicVolume)
		return nil
	} else {
		// Режим музыки
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
		s.updateMusicVolume()
		s.musicPlayer.Play()
		log.Printf("Info: music track %d/%d started, volume %.2f", s.currentTrack+1, s.trackCount, s.musicVolume)
		return nil
	}
}

// NextTrack переключает на следующий трек по кругу и запускает воспроизведение.
// В режиме музыки переключает между треками плейлиста.
// В режиме истории ничего не делает (история играет одна).
func (s *SoundManager) NextTrack() error {
	s.mu.Lock()

	// В режиме истории не переключаем треки
	if s.soundMode > 0 {
		s.mu.Unlock()
		return nil
	}

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
	s.updateMusicVolume()
	s.musicPlayer.Play()
	s.mu.Unlock()
	log.Printf("Info: switched to music track %d/%d", s.currentTrack+1, s.trackCount)
	return nil
}

// IsMusicFinished — true, если есть плейлист и текущий плеер существует, но уже не играет.
// В режиме музыки: проверяет окончание трека для автопереключения.
// В режиме истории: всегда возвращает false (не переключаем автоматически).
func (s *SoundManager) IsMusicFinished() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// В режиме истории не переключаем автоматически
	if s.soundMode > 0 {
		return false
	}

	if s.trackCount == 0 || s.ctx == nil {
		return false
	}
	if s.musicPlayer == nil {
		return false
	}
	return !s.musicPlayer.IsPlaying()
}

func (s *SoundManager) CurrentState() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentState
}

// SetMusicVolume теперь просто устанавливает громкость плеера, без привязки к состоянию.
// А состояние хранит базовую громкость (0.4 или 0.75), а умножать будем в настройках.

func (s *SoundManager) SetMusicVolume(vol float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.musicVolume = vol
	s.updateMusicVolume()
}

func (s *SoundManager) SetEffectsVolume(vol float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.effectsVol = vol
}

func (s *SoundManager) SetUserMusicVolume(vol float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userMusicVolume = vol
	s.updateMusicVolume()
}

func (s *SoundManager) SetUserEffectsVolume(vol float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userEffectsVolume = vol
	// обновляем плеер эффектов, но эффекты каждый раз создаются новые, так что просто сохраняем
}

func (s *SoundManager) updateMusicVolume() {
	base := 0.4
	if s.currentState == "game" {
		base = 0.75
	}
	finalVol := base * s.userMusicVolume
	if s.musicPlayer != nil {
		s.musicPlayer.SetVolume(finalVol)
	}
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
	s.updateMusicVolume()
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
	s.updateMusicVolume()
}
