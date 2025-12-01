package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

// FeatureStore stores and manages ML features
type FeatureStore struct {
	db         *gorm.DB
	cache      map[string]*Feature
	cacheTTL   time.Duration
	mu         sync.RWMutex
}

// Feature represents a machine learning feature
type Feature struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"index"`
	EntityType  string                 `json:"entity_type"` // user, product, etc.
	EntityID    string                 `json:"entity_id" gorm:"index"`
	Values      map[string]interface{} `json:"values" gorm:"type:jsonb"`
	Version     int                    `json:"version"`
	ComputedAt  time.Time              `json:"computed_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Metadata    map[string]string      `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// FeatureGroup groups related features
type FeatureGroup struct {
	ID          string            `json:"id" gorm:"primaryKey"`
	Name        string            `json:"name" gorm:"uniqueIndex"`
	Description string            `json:"description"`
	Features    []string          `json:"features" gorm:"type:jsonb"`
	EntityType  string            `json:"entity_type"`
	Version     int               `json:"version"`
	Metadata    map[string]string `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// NewFeatureStore creates a new feature store
func NewFeatureStore(db *gorm.DB) *FeatureStore {
	store := &FeatureStore{
		db:       db,
		cache:    make(map[string]*Feature),
		cacheTTL: 5 * time.Minute,
	}

	// Auto-migrate
	db.AutoMigrate(&Feature{}, &FeatureGroup{})

	// Start cleanup goroutine
	go store.cleanupLoop()

	return store
}

// SetFeature sets a feature value
func (fs *FeatureStore) SetFeature(ctx context.Context, feature *Feature) error {
	if feature.ID == "" {
		feature.ID = fmt.Sprintf("%s:%s:%s", feature.EntityType, feature.EntityID, feature.Name)
	}

	feature.ComputedAt = time.Now()

	// Save to database
	if err := fs.db.WithContext(ctx).Save(feature).Error; err != nil {
		return fmt.Errorf("failed to save feature: %w", err)
	}

	// Update cache
	fs.mu.Lock()
	fs.cache[feature.ID] = feature
	fs.mu.Unlock()

	return nil
}

// GetFeature gets a feature by ID
func (fs *FeatureStore) GetFeature(ctx context.Context, featureID string) (*Feature, error) {
	// Check cache first
	fs.mu.RLock()
	if cached, exists := fs.cache[featureID]; exists {
		if cached.ExpiresAt == nil || time.Now().Before(*cached.ExpiresAt) {
			fs.mu.RUnlock()
			return cached, nil
		}
	}
	fs.mu.RUnlock()

	// Get from database
	var feature Feature
	if err := fs.db.WithContext(ctx).Where("id = ?", featureID).First(&feature).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("feature not found: %s", featureID)
		}
		return nil, err
	}

	// Update cache
	fs.mu.Lock()
	fs.cache[featureID] = &feature
	fs.mu.Unlock()

	return &feature, nil
}

// GetFeaturesByEntity gets all features for an entity
func (fs *FeatureStore) GetFeaturesByEntity(ctx context.Context, entityType, entityID string) ([]*Feature, error) {
	var features []*Feature
	if err := fs.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Find(&features).Error; err != nil {
		return nil, err
	}
	return features, nil
}

// GetFeatureVector gets feature vector for entity
func (fs *FeatureStore) GetFeatureVector(ctx context.Context, entityType, entityID string, featureNames []string) (map[string]interface{}, error) {
	features, err := fs.GetFeaturesByEntity(ctx, entityType, entityID)
	if err != nil {
		return nil, err
	}

	vector := make(map[string]interface{})
	featureMap := make(map[string]*Feature)
	for _, f := range features {
		featureMap[f.Name] = f
	}

	for _, name := range featureNames {
		if feature, exists := featureMap[name]; exists {
			for k, v := range feature.Values {
				vector[k] = v
			}
		}
	}

	return vector, nil
}

// CreateFeatureGroup creates a feature group
func (fs *FeatureStore) CreateFeatureGroup(ctx context.Context, group *FeatureGroup) error {
	return fs.db.WithContext(ctx).Create(group).Error
}

// GetFeatureGroup gets a feature group
func (fs *FeatureStore) GetFeatureGroup(ctx context.Context, name string) (*FeatureGroup, error) {
	var group FeatureGroup
	if err := fs.db.WithContext(ctx).Where("name = ?", name).First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("feature group not found: %s", name)
		}
		return nil, err
	}
	return &group, nil
}

// GetFeatureGroupVector gets all features in a group for an entity
func (fs *FeatureStore) GetFeatureGroupVector(ctx context.Context, groupName, entityType, entityID string) (map[string]interface{}, error) {
	group, err := fs.GetFeatureGroup(ctx, groupName)
	if err != nil {
		return nil, err
	}

	return fs.GetFeatureVector(ctx, entityType, entityID, group.Features)
}

// BatchSetFeatures sets multiple features at once
func (fs *FeatureStore) BatchSetFeatures(ctx context.Context, features []*Feature) error {
	tx := fs.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, feature := range features {
		if feature.ID == "" {
			feature.ID = fmt.Sprintf("%s:%s:%s", feature.EntityType, feature.EntityID, feature.Name)
		}
		feature.ComputedAt = time.Now()

		if err := tx.Save(feature).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Update cache
		fs.mu.Lock()
		fs.cache[feature.ID] = feature
		fs.mu.Unlock()
	}

	return tx.Commit().Error
}

// DeleteExpiredFeatures deletes expired features
func (fs *FeatureStore) DeleteExpiredFeatures(ctx context.Context) error {
	return fs.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&Feature{}).Error
}

// ComputeFeatures computes features using a function
func (fs *FeatureStore) ComputeFeatures(ctx context.Context, entityType, entityID string, computeFunc func(context.Context, string, string) (map[string]interface{}, error)) error {
	values, err := computeFunc(ctx, entityType, entityID)
	if err != nil {
		return err
	}

	features := make([]*Feature, 0)
	for name, value := range values {
		feature := &Feature{
			Name:       name,
			EntityType: entityType,
			EntityID:   entityID,
			Values: map[string]interface{}{
				name: value,
			},
			Version: 1,
		}
		features = append(features, feature)
	}

	return fs.BatchSetFeatures(ctx, features)
}

// ExportFeatures exports features to JSON
func (fs *FeatureStore) ExportFeatures(ctx context.Context, entityType string) ([]byte, error) {
	var features []*Feature
	if err := fs.db.WithContext(ctx).
		Where("entity_type = ?", entityType).
		Find(&features).Error; err != nil {
		return nil, err
	}

	return json.Marshal(features)
}

// ImportFeatures imports features from JSON
func (fs *FeatureStore) ImportFeatures(ctx context.Context, data []byte) error {
	var features []*Feature
	if err := json.Unmarshal(data, &features); err != nil {
		return err
	}

	return fs.BatchSetFeatures(ctx, features)
}

// cleanupLoop periodically cleans up expired features
func (fs *FeatureStore) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		fs.DeleteExpiredFeatures(ctx)

		// Clean cache
		fs.mu.Lock()
		for id, feature := range fs.cache {
			if feature.ExpiresAt != nil && time.Now().After(*feature.ExpiresAt) {
				delete(fs.cache, id)
			}
		}
		fs.mu.Unlock()
	}
}

// GetStats returns feature store statistics
func (fs *FeatureStore) GetStats(ctx context.Context) (map[string]interface{}, error) {
	var totalFeatures int64
	fs.db.WithContext(ctx).Model(&Feature{}).Count(&totalFeatures)

	var totalGroups int64
	fs.db.WithContext(ctx).Model(&FeatureGroup{}).Count(&totalGroups)

	fs.mu.RLock()
	cacheSize := len(fs.cache)
	fs.mu.RUnlock()

	return map[string]interface{}{
		"total_features": totalFeatures,
		"total_groups":   totalGroups,
		"cache_size":     cacheSize,
	}, nil
}
