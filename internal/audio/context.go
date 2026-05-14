package audio

import (
	"log"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

var (
	audioContext     *audio.Context
	audioContextOnce sync.Once
)

// GetAudioContext возвращает глобальный аудиоконтекст (инициализируется один раз)
func GetAudioContext() *audio.Context {
	audioContextOnce.Do(func() {
		// Частота дискретизации 48000 Гц – стандарт для MP3
		audioContext = audio.NewContext(48000)
		if audioContext == nil {
			log.Fatal("failed to create audio context")
		}
	})
	return audioContext
}
