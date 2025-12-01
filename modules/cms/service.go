package cms

import (
	"context"
	"time"
)

// PageService handles page business logic
type PageService struct {
	repo *PageRepository
}

// NewPageService creates a new page service
func NewPageService(repo *PageRepository) *PageService {
	return &PageService{repo: repo}
}

// CreatePage creates a new page
func (s *PageService) CreatePage(ctx context.Context, page *Page) error {
	now := time.Now()
	if page.Status == "published" && page.PublishedAt == nil {
		page.PublishedAt = &now
	}
	return s.repo.Create(ctx, page)
}

// UpdatePage updates an existing page
func (s *PageService) UpdatePage(ctx context.Context, page *Page) error {
	if page.Status == "published" && page.PublishedAt == nil {
		now := time.Now()
		page.PublishedAt = &now
	}
	return s.repo.Update(ctx, page)
}

// GetPageBySlug gets a page by slug
func (s *PageService) GetPageBySlug(ctx context.Context, slug string) (*Page, error) {
	return s.repo.FindBySlug(ctx, slug)
}

// IncrementViewCount increments the view count of a page
func (s *PageService) IncrementViewCount(ctx context.Context, pageID uint) error {
	page, err := s.repo.FindByID(ctx, pageID)
	if err != nil {
		return err
	}
	page.ViewCount++
	return s.repo.Update(ctx, page)
}

// PostService handles post business logic
type PostService struct {
	repo *PostRepository
}

// NewPostService creates a new post service
func NewPostService(repo *PostRepository) *PostService {
	return &PostService{repo: repo}
}

// CreatePost creates a new post
func (s *PostService) CreatePost(ctx context.Context, post *Post) error {
	now := time.Now()
	if post.Status == "published" && post.PublishedAt == nil {
		post.PublishedAt = &now
	}
	return s.repo.Create(ctx, post)
}

// UpdatePost updates an existing post
func (s *PostService) UpdatePost(ctx context.Context, post *Post) error {
	if post.Status == "published" && post.PublishedAt == nil {
		now := time.Now()
		post.PublishedAt = &now
	}
	return s.repo.Update(ctx, post)
}

// GetPostBySlug gets a post by slug
func (s *PostService) GetPostBySlug(ctx context.Context, slug string) (*Post, error) {
	return s.repo.FindBySlug(ctx, slug)
}

// GetPostsByCategory gets posts by category
func (s *PostService) GetPostsByCategory(ctx context.Context, categoryID uint) ([]*Post, error) {
	return s.repo.FindByCategory(ctx, categoryID)
}

// CategoryService handles category business logic
type CategoryService struct {
	repo *CategoryRepository
}

// NewCategoryService creates a new category service
func NewCategoryService(repo *CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(ctx context.Context, category *Category) error {
	return s.repo.Create(ctx, category)
}

// GetCategoryBySlug gets a category by slug
func (s *CategoryService) GetCategoryBySlug(ctx context.Context, slug string) (*Category, error) {
	return s.repo.FindBySlug(ctx, slug)
}

// MediaService handles media business logic
type MediaService struct {
	repo *MediaRepository
}

// NewMediaService creates a new media service
func NewMediaService(repo *MediaRepository) *MediaService {
	return &MediaService{repo: repo}
}

// CreateMedia creates a new media record
func (s *MediaService) CreateMedia(ctx context.Context, media *Media) error {
	return s.repo.Create(ctx, media)
}

// GetMediaByType gets media by file type
func (s *MediaService) GetMediaByType(ctx context.Context, fileType string) ([]*Media, error) {
	return s.repo.FindByType(ctx, fileType)
}
