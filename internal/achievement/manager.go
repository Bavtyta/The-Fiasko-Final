package achievement

import (
	"sync"
)

type Achievement struct {
	ID          string
	Name        string
	Description string
	Condition   func(score float64, balance float64, jumps int, distance float64) bool
	Unlocked    bool
}

type Manager struct {
	mu           sync.RWMutex
	achievements []*Achievement
	score        float64
	balance      float64
	jumps        int
	distance     float64 // пройденное расстояние (можно из worldOffsetZ)
}

var (
	instance *Manager
	once     sync.Once
)

func GetManager() *Manager {
	once.Do(func() {
		instance = &Manager{
			achievements: []*Achievement{
				{
					ID:   "smile_face",
					Name: "СМАЙЛ ФЕЙС)))",
					Condition: func(score, balance float64, jumps int, distance float64) bool {
						return score >= 5
					},
				},
				{
					ID:   "score_20",
					Name: "Начинающий",
					Condition: func(score, balance float64, jumps int, distance float64) bool {
						return score >= 250
					},
				},
				{
					ID:   "score_1000",
					Name: "ЭТО УСПЕХ БРАТАН",
					Condition: func(score, balance float64, jumps int, distance float64) bool {
						return score >= 1000
					},
				},
				{
					ID:   "jump_67",
					Name: "ТЫ РЕАЛЬНО 67 БРАТАН",
					Condition: func(score, balance float64, jumps int, distance float64) bool {
						return jumps >= 67
					},
				},
				{
					ID:   "denchik_slaziet",
					Name: "ДЭНЧИК СЛАЗИЕТ",
					Condition: func(score, balance float64, jumps int, distance float64) bool {
						return balance < 5 && score > 30 // баланс почти идеален
					},
				},
			},
		}
	})
	return instance
}

func (m *Manager) UpdateStats(score, balance float64, jumps int, distance float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.score = score
	m.balance = balance
	m.jumps = jumps
	m.distance = distance

	for _, ach := range m.achievements {
		if !ach.Unlocked && ach.Condition(score, balance, jumps, distance) {
			ach.Unlocked = true
			// можно добавить звук разблокировки
		}
	}
}

func (m *Manager) GetAchievements() []*Achievement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// возвращаем копию или оригинал (только чтение)
	return m.achievements
}

// Reset для тестов или новой игры (не сбрасываем достижения при новой игре)
func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ach := range m.achievements {
		ach.Unlocked = false
	}
	m.score = 0
	m.balance = 0
	m.jumps = 0
	m.distance = 0
}
