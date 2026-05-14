package asset

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// MusicFolder – путь к папке с музыкой относительно рабочей директории
var MusicFolder = "asset/music"

// SoundsFolder – путь к папке со звуками
var SoundsFolder = "asset/sounds"

// LoadMusicTrack загружает первый попавшийся MP3 из MusicFolder.
// Возвращает ошибку, если папка пуста или файлы не найдены.
func LoadMusicTrack() (io.ReadSeeker, error) {
	files, err := filepath.Glob(filepath.Join(MusicFolder, "*.mp3"))
	if err != nil {
		return nil, fmt.Errorf("failed to scan music folder: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no MP3 files found in %s", MusicFolder)
	}
	// Берём первый файл (можно расширить до выбора случайного или плейлиста)
	return os.Open(files[0])
}

// LoadSound загружает звук по имени файла.
func LoadSound(filename string) (io.ReadSeeker, error) {
	path := filepath.Join(SoundsFolder, filename)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sound file %s: %w", path, err)
	}
	return file, nil
}

// Конкретные функции-обёртки
func LoadLoseSound() (io.ReadSeeker, error)      { return LoadSound("lose.mp3") }
func LoadJumpSound() (io.ReadSeeker, error)      { return LoadSound("jump.mp3") }
func LoadKeypressSound() (io.ReadSeeker, error)  { return LoadSound("keypress.mp3") }
func LoadCountdownSound() (io.ReadSeeker, error) { return LoadSound("countdown.mp3") }
