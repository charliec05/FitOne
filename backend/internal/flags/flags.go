package flags

import (
	"hash/fnv"
)

type Manager struct {
	flags map[string]float64
}

func New(flags map[string]float64) *Manager {
	copyMap := make(map[string]float64, len(flags))
	for k, v := range flags {
		copyMap[k] = v
	}
	return &Manager{flags: copyMap}
}

// IsEnabled determines if a feature flag percentage rollout allows the given user.
func (m *Manager) IsEnabled(flag string, userID string) bool {
	if m == nil {
		return false
	}
	percent, ok := m.flags[flag]
	if !ok {
		return false
	}
	if percent >= 100 {
		return true
	}
	if percent <= 0 {
		return false
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(userID))
	value := hash.Sum32() % 100
	return float64(value) < percent
}

// All returns a copy of the current flag map.
func (m *Manager) All() map[string]float64 {
	result := make(map[string]float64, len(m.flags))
	for k, v := range m.flags {
		result[k] = v
	}
	return result
}
