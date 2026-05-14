package config

// CameraConfig содержит конфигурационные константы для камеры
type CameraConfig struct {
	DefaultPositionY float64
	DefaultPositionZ float64
	HorizonRatio     float64
	CurveStrength    float64 // Сила закругления перспективы влево (0 = нет закругления)
	CurveDepth       float64 // Глубина, на которой закругление максимально

	// Вертикальное искривление (эффект горизонта)
	VerticalCurveStrength float64 // Сила вертикального подъёма на средней дистанции
	VerticalCurvePeak     float64 // Точка максимального подъёма (0.0-1.0, где 1.0 = CurveDepth)
	VerticalCurveDrop     float64 // Сила опускания на дальней дистанции
}

// GameConfig содержит конфигурационные константы для игры
type GameConfig struct {
	ScreenWidth          int
	ScreenHeight         int
	DriftThreshold       float64
	ShowCollisionCircles bool // Показывать ли круги столкновения вокруг препятствий
}

// PhysicsConfig содержит конфигурационные константы для физики
type PhysicsConfig struct {
	Gravity float64
}

// SpeedConfig содержит конфигурацию скорости мира
type SpeedConfig struct {
	InitialSpeed          float64 // начальная скорость (единиц Z в секунду)
	MaxSpeed              float64 // максимальная скорость
	SpeedIncreasePerScore float64 // увеличение скорости за 1 очко
}

// DefaultCameraConfig возвращает конфигурацию камеры по умолчанию
func DefaultCameraConfig() CameraConfig {
	return CameraConfig{
		DefaultPositionY: -1.0,
		DefaultPositionZ: -2.0,
		HorizonRatio:     2.0 / 3.0,
		CurveStrength:    0.045,  // Умеренное закругление влево
		CurveDepth:       1000.0, // Максимальное закругление на глубине 1000 единиц

		// Вертикальное искривление
		VerticalCurveStrength: 0.1,  // Сила подъёма поверхности
		VerticalCurvePeak:     0.65, // Пик подъёма на 65% глубины
		VerticalCurveDrop:     0.02, // Опускание на последней трети
	}
}

// DefaultGameConfig возвращает конфигурацию игры по умолчанию
func DefaultGameConfig() GameConfig {
	return GameConfig{
		ScreenWidth:          1266,
		ScreenHeight:         768,
		DriftThreshold:       15.0,
		ShowCollisionCircles: true, // По умолчанию показываем круги столкновения
	}
}

// DefaultPhysicsConfig возвращает конфигурацию физики по умолчанию
func DefaultPhysicsConfig() PhysicsConfig {
	return PhysicsConfig{
		Gravity: 0.145,
	}
}

// DefaultSpeedConfig возвращает конфигурацию по умолчанию
func DefaultSpeedConfig() SpeedConfig {
	return SpeedConfig{
		InitialSpeed:          50.0,
		MaxSpeed:              150.0,
		SpeedIncreasePerScore: 0.05,
	}
}
