package asset

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	_ "image/jpeg" // Поддержка JPEG
	_ "image/png"  // Поддержка PNG
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

// Текстура бревна (JPG)
//
//go:embed image/log_texture.jpg
var logTextureJPG []byte

// Текстура игрока (PNG)
//
//go:embed image/dog_texture.png
var playerTexturePNG []byte

// Текстура игрока правая (PNG)
//
//go:embed image/dog_texture_right.png
var playerTextureRightPNG []byte

// Текстура игрока прыжок (PNG)
//
//go:embed image/dog_jump.png
var playerTextureJumpPNG []byte

// Текстура реки (JPG)
//
//go:embed image/river_texture.jpg
var riverTextureJPG []byte

// Текстура фона (PNG)
//
//go:embed image/background.png
var backgroundTexturePNG []byte

// Фон игрового уровня (PNG)
//
//go:embed image/game_background.png
var gameBackgroundTexturePNG []byte

// Текстура сугроба (PNG)
//
//go:embed image/Sugrob_png.png
var sugrobTexturePNG []byte

// LoadLogTexture загружает текстуру бревна (JPG)
func LoadLogTexture() *ebiten.Image {
	// Загружаем JPG текстуру
	if len(logTextureJPG) > 0 {
		img, _, err := image.Decode(bytes.NewReader(logTextureJPG))
		if err == nil {
			log.Printf("Info: Loaded log texture from embedded JPG")
			return ebiten.NewImageFromImage(img)
		}
		log.Printf("Warning: Failed to load log texture JPG: %v", err)
	}

	// Если не загрузился, создаём процедурную текстуру
	log.Printf("Info: Using procedural bark texture")
	return createProceduralBarkTexture()
}

// createProceduralBarkTexture создаёт простую процедурную текстуру коры
func createProceduralBarkTexture() *ebiten.Image {
	width, height := 512, 256
	img := ebiten.NewImage(width, height)

	// Заполняем коричневым цветом с вертикальными линиями
	pixels := make([]byte, width*height*4)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 4

			// Базовый коричневый цвет
			baseR := uint8(101 + (x%20)*2)
			baseG := uint8(67 + (x % 20))
			baseB := uint8(33)

			// Добавляем вертикальные линии для имитации коры
			if x%32 < 2 {
				baseR = uint8(int(baseR) * 7 / 10)
				baseG = uint8(int(baseG) * 7 / 10)
				baseB = uint8(int(baseB) * 7 / 10)
			}

			// Добавляем небольшой шум
			noise := uint8((x * y) % 20)
			baseR = uint8(clamp(int(baseR)+int(noise)-10, 0, 255))
			baseG = uint8(clamp(int(baseG)+int(noise)-10, 0, 255))
			baseB = uint8(clamp(int(baseB)+int(noise)-10, 0, 255))

			pixels[idx] = baseR
			pixels[idx+1] = baseG
			pixels[idx+2] = baseB
			pixels[idx+3] = 255 // Alpha
		}
	}

	img.WritePixels(pixels)
	return img
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// LoadPlayerTexture загружает текстуру игрока (PNG)
func LoadPlayerTexture() *ebiten.Image {
	if len(playerTexturePNG) > 0 {
		img, _, err := image.Decode(bytes.NewReader(playerTexturePNG))
		if err == nil {
			log.Printf("Info: Loaded player texture from embedded PNG")
			return ebiten.NewImageFromImage(img)
		}
		log.Printf("Warning: Failed to load player texture PNG: %v", err)
	}

	// Если не загрузился, создаём простую белую текстуру
	log.Printf("Info: Using fallback white texture for player")
	img := ebiten.NewImage(32, 32)
	img.Fill(color.White)
	return img
}

// LoadPlayerTextureRight загружает правую текстуру игрока (PNG)
func LoadPlayerTextureRight() *ebiten.Image {
	if len(playerTextureRightPNG) > 0 {
		img, _, err := image.Decode(bytes.NewReader(playerTextureRightPNG))
		if err == nil {
			log.Printf("Info: Loaded player right texture from embedded PNG")
			return ebiten.NewImageFromImage(img)
		}
		log.Printf("Warning: Failed to load player right texture PNG: %v", err)
	}

	// Если не загрузился, создаём простую белую текстуру
	log.Printf("Info: Using fallback white texture for player right")
	img := ebiten.NewImage(32, 32)
	img.Fill(color.White)
	return img
}

// LoadPlayerTextureJump загружает текстуру прыжка игрока (PNG)
func LoadPlayerTextureJump() *ebiten.Image {
	if len(playerTextureJumpPNG) > 0 {
		img, _, err := image.Decode(bytes.NewReader(playerTextureJumpPNG))
		if err == nil {
			log.Printf("Info: Loaded player jump texture from embedded PNG")
			return ebiten.NewImageFromImage(img)
		}
		log.Printf("Warning: Failed to load player jump texture PNG: %v", err)
	}

	// Если не загрузился, создаём простую белую текстуру
	log.Printf("Info: Using fallback white texture for player jump")
	img := ebiten.NewImage(32, 32)
	img.Fill(color.White)
	return img
}

// LoadRiverTexture загружает текстуру реки (JPG)
func LoadRiverTexture() *ebiten.Image {
	if len(riverTextureJPG) > 0 {
		img, _, err := image.Decode(bytes.NewReader(riverTextureJPG))
		if err == nil {
			log.Printf("Info: Loaded river texture from embedded JPG")
			return ebiten.NewImageFromImage(img)
		}
		log.Printf("Warning: Failed to load river texture JPG: %v", err)
	}

	// Если не загрузился, создаём простую синюю текстуру
	log.Printf("Info: Using fallback blue texture for river")
	img := ebiten.NewImage(512, 512)
	img.Fill(color.RGBA{0, 100, 255, 255})
	return img
}

// LoadBackgroundTexture загружает текстуру фона (PNG)
func LoadBackgroundTexture() *ebiten.Image {
	if len(backgroundTexturePNG) > 0 {
		img, _, err := image.Decode(bytes.NewReader(backgroundTexturePNG))
		if err == nil {
			log.Printf("Info: Loaded background texture from embedded PNG")
			return ebiten.NewImageFromImage(img)
		}
		log.Printf("Warning: Failed to load background texture PNG: %v", err)
	}

	// Если не загрузился, создаём простую зелёную текстуру
	log.Printf("Info: Using fallback green texture for background")
	img := ebiten.NewImage(512, 512)
	img.Fill(color.RGBA{34, 139, 34, 255})
	return img
}

// LoadGameBackgroundTexture загружает текстуру фона во время игры (PNG)
func LoadGameBackgroundTexture() *ebiten.Image {
	if len(gameBackgroundTexturePNG) > 0 {
		img, _, err := image.Decode(bytes.NewReader(gameBackgroundTexturePNG))
		if err == nil {
			log.Printf("Info: Loaded game background texture from embedded PNG")
			return ebiten.NewImageFromImage(img)
		}
		log.Printf("Warning: Failed to load game background texture PNG: %v", err)
	}

	log.Printf("Info: Using fallback green texture for game background")
	img := ebiten.NewImage(512, 512)
	img.Fill(color.RGBA{34, 139, 34, 255})
	return img
}

// LoadSugrobTexture загружает текстуру сугроба (PNG)
func LoadSugrobTexture() *ebiten.Image {
	if len(sugrobTexturePNG) > 0 {
		img, _, err := image.Decode(bytes.NewReader(sugrobTexturePNG))
		if err == nil {
			log.Printf("Info: Loaded sugrob texture from embedded PNG")
			return ebiten.NewImageFromImage(img)
		}
		log.Printf("Warning: Failed to load sugrob texture PNG: %v", err)
	}

	// Если не загрузился, создаём простую белую текстуру
	log.Printf("Info: Using fallback white texture for sugrob")
	img := ebiten.NewImage(32, 32)
	img.Fill(color.White)
	return img
}
