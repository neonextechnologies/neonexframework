package cms

import (
	"time"

	"gorm.io/gorm"
)

// Page represents a CMS page
type Page struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	Slug        string         `json:"slug" gorm:"size:255;uniqueIndex;not null"`
	Content     string         `json:"content" gorm:"type:text"`
	Excerpt     string         `json:"excerpt" gorm:"type:text"`
	Template    string         `json:"template" gorm:"size:100;default:'default'"`
	Status      string         `json:"status" gorm:"size:20;default:'draft'"` // draft, published, archived
	AuthorID    uint           `json:"author_id"`
	ParentID    *uint          `json:"parent_id"` // For nested pages
	Order       int            `json:"order" gorm:"default:0"`
	FeaturedImg string         `json:"featured_image" gorm:"size:500"`
	SEOTitle    string         `json:"seo_title" gorm:"size:255"`
	SEODesc     string         `json:"seo_description" gorm:"size:500"`
	SEOKeywords string         `json:"seo_keywords" gorm:"size:500"`
	ViewCount   int            `json:"view_count" gorm:"default:0"`
	PublishedAt *time.Time     `json:"published_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for Page model
func (Page) TableName() string {
	return "cms_pages"
}

// Post represents a blog post
type Post struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	Slug        string         `json:"slug" gorm:"size:255;uniqueIndex;not null"`
	Content     string         `json:"content" gorm:"type:text"`
	Excerpt     string         `json:"excerpt" gorm:"type:text"`
	Status      string         `json:"status" gorm:"size:20;default:'draft'"` // draft, published, archived
	AuthorID    uint           `json:"author_id"`
	CategoryID  *uint          `json:"category_id"`
	FeaturedImg string         `json:"featured_image" gorm:"size:500"`
	Tags        string         `json:"tags" gorm:"type:text"` // JSON array
	SEOTitle    string         `json:"seo_title" gorm:"size:255"`
	SEODesc     string         `json:"seo_description" gorm:"size:500"`
	SEOKeywords string         `json:"seo_keywords" gorm:"size:500"`
	ViewCount   int            `json:"view_count" gorm:"default:0"`
	PublishedAt *time.Time     `json:"published_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for Post model
func (Post) TableName() string {
	return "cms_posts"
}

// Category represents a content category
type Category struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:100;not null"`
	Slug        string         `json:"slug" gorm:"size:100;uniqueIndex;not null"`
	Description string         `json:"description" gorm:"type:text"`
	ParentID    *uint          `json:"parent_id"`
	Order       int            `json:"order" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for Category model
func (Category) TableName() string {
	return "cms_categories"
}

// Tag represents a content tag
type Tag struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:100;uniqueIndex;not null"`
	Slug      string         `json:"slug" gorm:"size:100;uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for Tag model
func (Tag) TableName() string {
	return "cms_tags"
}

// Media represents uploaded media files
type Media struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	FileName   string         `json:"file_name" gorm:"size:255;not null"`
	FilePath   string         `json:"file_path" gorm:"size:500;not null"`
	FileType   string         `json:"file_type" gorm:"size:100"` // image, video, document, etc.
	MimeType   string         `json:"mime_type" gorm:"size:100"`
	FileSize   int64          `json:"file_size"`
	Width      int            `json:"width"`  // For images
	Height     int            `json:"height"` // For images
	Alt        string         `json:"alt" gorm:"size:255"`
	Caption    string         `json:"caption" gorm:"type:text"`
	UploadedBy uint           `json:"uploaded_by"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for Media model
func (Media) TableName() string {
	return "cms_media"
}

// Menu represents a navigation menu
type Menu struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	Location  string         `json:"location" gorm:"size:50"` // header, footer, sidebar
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for Menu model
func (Menu) TableName() string {
	return "cms_menus"
}

// MenuItem represents a menu item
type MenuItem struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	MenuID    uint           `json:"menu_id"`
	ParentID  *uint          `json:"parent_id"`
	Title     string         `json:"title" gorm:"size:255;not null"`
	URL       string         `json:"url" gorm:"size:500"`
	Target    string         `json:"target" gorm:"size:20;default:'_self'"` // _self, _blank
	Icon      string         `json:"icon" gorm:"size:100"`
	Order     int            `json:"order" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name for MenuItem model
func (MenuItem) TableName() string {
	return "cms_menu_items"
}
