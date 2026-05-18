package config

type SoundConfig struct {
	// MusicVolumeMenu — громкость вне активного геймплея (главное меню, пауза, гейм овер и т.д.).
	MusicVolumeMenu float64
	// MusicVolumeGame — только во время GameState (активный забег).
	MusicVolumeGame   float64
	EffectsVolume     float64 // общая громкость эффектов
	CountdownDuration float64 // длительность периода "без баланса" в секундах
}

func DefaultSoundConfig() SoundConfig {
	return SoundConfig{
		MusicVolumeMenu:   0.2,
		MusicVolumeGame:   1,
		EffectsVolume:     0.8,
		CountdownDuration: 3.0,
	}
}
