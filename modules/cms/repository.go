package cms

import (
	"context"

	"neonexcore/pkg/database"

	"gorm.io/gorm"
)

// PageRepository handles page data access
type PageRepository struct {
	*database.BaseRepository[Page]
}

// NewPageRepository creates a new page repository
func NewPageRepository(db *gorm.DB) *PageRepository {
	return &PageRepository{
		BaseRepository: database.NewBaseRepository[Page](db),
	}
}

// FindBySlug finds a page by slug
func (r *PageRepository) FindBySlug(ctx context.Context, slug string) (*Page, error) {
	var page Page
	err := r.GetDB().WithContext(ctx).Where("slug = ?", slug).First(&page).Error
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// FindPublished finds all published pages
func (r *PageRepository) FindPublished(ctx context.Context) ([]*Page, error) {
	var pages []*Page
	err := r.GetDB().WithContext(ctx).
		Where("status = ?", "published").
		Order("order ASC").
		Find(&pages).Error
	return pages, err
}

// PostRepository handles post data access
type PostRepository struct {
	*database.BaseRepository[Post]
}

// NewPostRepository creates a new post repository
func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{
		BaseRepository: database.NewBaseRepository[Post](db),
	}
}

// FindBySlug finds a post by slug
func (r *PostRepository) FindBySlug(ctx context.Context, slug string) (*Post, error) {
	var post Post
	err := r.GetDB().WithContext(ctx).Where("slug = ?", slug).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// FindByCategory finds posts by category
func (r *PostRepository) FindByCategory(ctx context.Context, categoryID uint) ([]*Post, error) {
	var posts []*Post
	err := r.GetDB().WithContext(ctx).
		Where("category_id = ? AND status = ?", categoryID, "published").
		Order("published_at DESC").
		Find(&posts).Error
	return posts, err
}

// CategoryRepository handles category data access
type CategoryRepository struct {
	*database.BaseRepository[Category]
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{
		BaseRepository: database.NewBaseRepository[Category](db),
	}
}

// FindBySlug finds a category by slug
func (r *CategoryRepository) FindBySlug(ctx context.Context, slug string) (*Category, error) {
	var category Category
	err := r.GetDB().WithContext(ctx).Where("slug = ?", slug).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// MediaRepository handles media data access
type MediaRepository struct {
	*database.BaseRepository[Media]
}

// NewMediaRepository creates a new media repository
func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{
		BaseRepository: database.NewBaseRepository[Media](db),
	}
}

// FindByType finds media by file type
func (r *MediaRepository) FindByType(ctx context.Context, fileType string) ([]*Media, error) {
	var media []*Media
	err := r.GetDB().WithContext(ctx).
		Where("file_type = ?", fileType).
		Order("created_at DESC").
		Find(&media).Error
	return media, err
}
