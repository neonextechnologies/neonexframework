package settings

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Setting represents a setting entry
type Setting struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Key       string         `gorm:"uniqueIndex;size:100;not null" json:"key"`
	Value     string         `gorm:"type:text" json:"value"`
	Type      string         `gorm:"size:20;default:'string'" json:"type"` // string, int, bool, json
	Module    string         `gorm:"size:50;index" json:"module"`
	IsPublic  bool           `gorm:"default:false" json:"is_public"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Manager manages settings
type Manager struct {
	db    *gorm.DB
	cache map[string]interface{}
}

// NewManager creates a new settings manager
func NewManager(db *gorm.DB) *Manager {
	return &Manager{
		db:    db,
		cache: make(map[string]interface{}),
	}
}

// Get gets a setting value
func (m *Manager) Get(ctx context.Context, key string, defaultValue interface{}) (interface{}, error) {
	// Check cache
	if val, ok := m.cache[key]; ok {
		return val, nil
	}

	var setting Setting
	err := m.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error
	if err == gorm.ErrRecordNotFound {
		return defaultValue, nil
	}
	if err != nil {
		return nil, err
	}

	value := m.parseValue(setting)
	m.cache[key] = value
	return value, nil
}

// Set sets a setting value
func (m *Manager) Set(ctx context.Context, key string, value interface{}, module string) error {
	valueStr, valueType := m.serializeValue(value)

	setting := Setting{
		Key:    key,
		Value:  valueStr,
		Type:   valueType,
		Module: module,
	}

	err := m.db.WithContext(ctx).
		Where("key = ?", key).
		Assign(setting).
		FirstOrCreate(&setting).Error

	if err == nil {
		m.cache[key] = value
	}

	return err
}

// GetString gets string setting
func (m *Manager) GetString(ctx context.Context, key string, defaultValue string) string {
	val, _ := m.Get(ctx, key, defaultValue)
	if str, ok := val.(string); ok {
		return str
	}
	return defaultValue
}

// GetInt gets int setting
func (m *Manager) GetInt(ctx context.Context, key string, defaultValue int) int {
	val, _ := m.Get(ctx, key, defaultValue)
	if num, ok := val.(int); ok {
		return num
	}
	return defaultValue
}

// GetBool gets bool setting
func (m *Manager) GetBool(ctx context.Context, key string, defaultValue bool) bool {
	val, _ := m.Get(ctx, key, defaultValue)
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultValue
}

// GetByModule gets all settings for a module
func (m *Manager) GetByModule(ctx context.Context, module string) (map[string]interface{}, error) {
	var settings []Setting
	err := m.db.WithContext(ctx).Where("module = ?", module).Find(&settings).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, s := range settings {
		result[s.Key] = m.parseValue(s)
	}

	return result, nil
}

// Delete deletes a setting
func (m *Manager) Delete(ctx context.Context, key string) error {
	delete(m.cache, key)
	return m.db.WithContext(ctx).Where("key = ?", key).Delete(&Setting{}).Error
}

// ClearCache clears settings cache
func (m *Manager) ClearCache() {
	m.cache = make(map[string]interface{})
}

// parseValue converts stored value to appropriate type
func (m *Manager) parseValue(setting Setting) interface{} {
	switch setting.Type {
	case "int":
		var num int
		json.Unmarshal([]byte(setting.Value), &num)
		return num
	case "bool":
		var b bool
		json.Unmarshal([]byte(setting.Value), &b)
		return b
	case "json":
		var data interface{}
		json.Unmarshal([]byte(setting.Value), &data)
		return data
	default:
		return setting.Value
	}
}

// serializeValue converts value to string and determines type
func (m *Manager) serializeValue(value interface{}) (string, string) {
	switch v := value.(type) {
	case string:
		return v, "string"
	case int, int64, int32:
		bytes, _ := json.Marshal(v)
		return string(bytes), "int"
	case bool:
		bytes, _ := json.Marshal(v)
		return string(bytes), "bool"
	default:
		bytes, _ := json.Marshal(v)
		return string(bytes), "json"
	}
}
