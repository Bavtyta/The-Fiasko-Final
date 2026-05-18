package asset

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// MusicFolder – путь к папке с музыкой относительно рабочей директории
var MusicFolder = "asset/music"

// SoundsFolder – путь к папке со звуками
var SoundsFolder = "asset/sounds"

// StoriesFolder – путь к папке с историями
var StoriesFolder = "asset/storyes"

// allMusicData хранит все MP3 в памяти (порядок — отсортированные имена файлов).
var allMusicData [][]byte

// allStoriesData хранит все истории MP3 в памяти
var allStoriesData [][]byte

// LoadAllMusicTracks загружает все MP3 из MusicFolder в память.
// Возвращает количество успешно загруженных треков или ошибку, если нет ни одного.
func LoadAllMusicTracks() (int, error) {
	pattern := filepath.Join(MusicFolder, "*.mp3")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return 0, err
	}
	if len(files) == 0 {
		return 0, fmt.Errorf("no MP3 files found in %s", MusicFolder)
	}
	sort.Strings(files)

	allMusicData = nil
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		allMusicData = append(allMusicData, data)
	}
	if len(allMusicData) == 0 {
		return 0, fmt.Errorf("no valid MP3 files could be read from %s", MusicFolder)
	}
	return len(allMusicData), nil
}

// GetMusicTrackData возвращает данные трека по индексу.
func GetMusicTrackData(index int) ([]byte, error) {
	if index < 0 || index >= len(allMusicData) {
		return nil, fmt.Errorf("track index %d out of range [0..%d]", index, len(allMusicData)-1)
	}
	return allMusicData[index], nil
}

// GetMusicTrackCount возвращает количество загруженных треков.
func GetMusicTrackCount() int {
	return len(allMusicData)
}

// LoadAllStories загружает все истории из StoriesFolder в память.
func LoadAllStories() (int, error) {
	pattern := filepath.Join(StoriesFolder, "*.mp3")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return 0, err
	}
	if len(files) == 0 {
		return 0, fmt.Errorf("no story MP3 files found in %s", StoriesFolder)
	}
	sort.Strings(files)

	allStoriesData = nil
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		allStoriesData = append(allStoriesData, data)
	}
	if len(allStoriesData) == 0 {
		return 0, fmt.Errorf("no valid story MP3 files could be read from %s", StoriesFolder)
	}
	return len(allStoriesData), nil
}

// GetStoryData возвращает данные истории по индексу.
func GetStoryData(index int) ([]byte, error) {
	if index < 0 || index >= len(allStoriesData) {
		return nil, fmt.Errorf("story index %d out of range [0..%d]", index, len(allStoriesData)-1)
	}
	return allStoriesData[index], nil
}

// GetStoryCount возвращает количество загруженных историй.
func GetStoryCount() int {
	return len(allStoriesData)
}

// LoadMusicTrack возвращает первый трек как ReadSeeker (после LoadAllMusicTracks — из памяти).
func LoadMusicTrack() (io.ReadSeeker, error) {
	if len(allMusicData) == 0 {
		return nil, fmt.Errorf("no music tracks loaded")
	}
	return bytes.NewReader(allMusicData[0]), nil
}

// LoadSound загружает звук по имени файла (каждый раз открывает файл на диске).
func LoadSound(filename string) (io.ReadSeeker, error) {
	path := filepath.Join(SoundsFolder, filename)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sound file %s: %w", path, err)
	}
	return file, nil
}

// soundCache — один раз читает MP3 в []byte, дальше только bytes.NewReader (без повторных Open).
type soundCache struct {
	name string
	once sync.Once
	data []byte
	err  error
}

func (c *soundCache) newReader() (io.ReadSeeker, error) {
	c.once.Do(func() {
		path := filepath.Join(SoundsFolder, c.name)
		var readErr error
		c.data, readErr = os.ReadFile(path)
		if readErr != nil {
			c.err = fmt.Errorf("failed to read sound file %s: %w", path, readErr)
		}
	})
	if c.err != nil {
		return nil, c.err
	}
	return bytes.NewReader(c.data), nil
}

var (
	loseSoundCache      = &soundCache{name: "lose.mp3"}
	jumpSoundCache      = &soundCache{name: "jump.mp3"}
	keypressSoundCache  = &soundCache{name: "keypress.mp3"}
	countdownSoundCache = &soundCache{name: "countdown.mp3"}
)

func LoadLoseSound() (io.ReadSeeker, error)      { return loseSoundCache.newReader() }
func LoadJumpSound() (io.ReadSeeker, error)      { return jumpSoundCache.newReader() }
func LoadKeypressSound() (io.ReadSeeker, error)  { return keypressSoundCache.newReader() }
func LoadCountdownSound() (io.ReadSeeker, error) { return countdownSoundCache.newReader() }
