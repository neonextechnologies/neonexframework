# File Storage & Upload Management

Master file handling with NeonEx Framework's storage system. Learn local storage, S3 integration, file validation, image processing, and security best practices.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [Storage Drivers](#storage-drivers)
- [File Uploads](#file-uploads)
- [File Validation](#file-validation)
- [Image Processing](#image-processing)
- [Serving Files](#serving-files)
- [Security](#security)
- [Integration Examples](#integration-examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Introduction

NeonEx provides a flexible file storage system with support for multiple backends. Key features:

- **Multiple Drivers**: Local, S3, GCS, Azure Blob
- **File Uploads**: Multipart upload handling
- **Validation**: Size, type, dimension validation
- **Image Processing**: Resize, crop, watermark
- **Streaming**: Memory-efficient large file handling
- **CDN Integration**: CloudFront, CloudFlare support
- **Security**: Signed URLs, access control

## Quick Start

### Basic File Upload

```go
package main

import (
    "net/http"
    "neonex/core/pkg/storage"
    "github.com/labstack/echo/v4"
)

func UploadHandler(c echo.Context) error {
    // Get uploaded file
    file, err := c.FormFile("file")
    if err != nil {
        return err
    }
    
    // Open file
    src, err := file.Open()
    if err != nil {
        return err
    }
    defer src.Close()
    
    // Create storage
    store := storage.NewLocalStorage("./uploads")
    
    // Save file
    path, err := store.Put(c.Request().Context(), "documents", file.Filename, src)
    if err != nil {
        return err
    }
    
    return c.JSON(http.StatusOK, map[string]string{
        "path": path,
        "url":  "/uploads/" + path,
    })
}
```

## Storage Drivers

### Local Storage

```go
import "neonex/core/pkg/storage"

config := storage.LocalConfig{
    RootPath:    "./storage",
    PublicURL:   "https://example.com/storage",
    Permissions: 0755,
}

store := storage.NewLocalStorage(config)
```

### AWS S3 Storage

```go
import (
    "neonex/core/pkg/storage"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)

config := storage.S3Config{
    Region:      "us-east-1",
    Bucket:      "my-bucket",
    AccessKeyID: "your-access-key",
    SecretKey:   "your-secret-key",
    Endpoint:    "", // Optional, for S3-compatible services
    PublicURL:   "https://my-bucket.s3.amazonaws.com",
}

store := storage.NewS3Storage(config)
```

### Google Cloud Storage

```go
config := storage.GCSConfig{
    ProjectID:   "my-project",
    Bucket:      "my-bucket",
    Credentials: "path/to/credentials.json",
    PublicURL:   "https://storage.googleapis.com/my-bucket",
}

store := storage.NewGCSStorage(config)
```

### Azure Blob Storage

```go
config := storage.AzureConfig{
    AccountName:   "myaccount",
    AccountKey:    "your-account-key",
    ContainerName: "my-container",
    PublicURL:     "https://myaccount.blob.core.windows.net/my-container",
}

store := storage.NewAzureStorage(config)
```

### Storage Manager (Multi-Driver)

```go
type StorageManager struct {
    drivers map[string]storage.Storage
    default string
}

func NewStorageManager() *StorageManager {
    return &StorageManager{
        drivers: make(map[string]storage.Storage),
        default: "local",
    }
}

func (sm *StorageManager) Register(name string, driver storage.Storage) {
    sm.drivers[name] = driver
}

func (sm *StorageManager) SetDefault(name string) {
    sm.default = name
}

func (sm *StorageManager) Disk(name ...string) storage.Storage {
    diskName := sm.default
    if len(name) > 0 {
        diskName = name[0]
    }
    return sm.drivers[diskName]
}

// Usage
manager := NewStorageManager()
manager.Register("local", storage.NewLocalStorage("./storage"))
manager.Register("s3", storage.NewS3Storage(s3Config))
manager.Register("gcs", storage.NewGCSStorage(gcsConfig))

// Use default disk
manager.Disk().Put(ctx, "path", "file.txt", data)

// Use specific disk
manager.Disk("s3").Put(ctx, "images", "photo.jpg", imageData)
```

## File Uploads

### Single File Upload

```go
func (h *FileHandler) Upload(c echo.Context) error {
    ctx := c.Request().Context()
    
    // Get file from request
    file, err := c.FormFile("file")
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "No file uploaded")
    }
    
    // Validate file
    if err := validateFile(file); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }
    
    // Open file
    src, err := file.Open()
    if err != nil {
        return err
    }
    defer src.Close()
    
    // Generate unique filename
    filename := generateUniqueFilename(file.Filename)
    
    // Upload to storage
    path, err := h.storage.Put(ctx, "uploads", filename, src)
    if err != nil {
        return err
    }
    
    // Save to database
    fileRecord := &File{
        UserID:       getUserID(c),
        OriginalName: file.Filename,
        Path:         path,
        Size:         file.Size,
        MimeType:     file.Header.Get("Content-Type"),
        UploadedAt:   time.Now(),
    }
    
    h.db.Create(fileRecord)
    
    return c.JSON(http.StatusOK, map[string]interface{}{
        "id":   fileRecord.ID,
        "url":  h.storage.URL(path),
        "size": file.Size,
    })
}
```

### Multiple File Upload

```go
func (h *FileHandler) UploadMultiple(c echo.Context) error {
    ctx := c.Request().Context()
    
    // Get multiple files
    form, err := c.MultipartForm()
    if err != nil {
        return err
    }
    
    files := form.File["files"]
    if len(files) == 0 {
        return echo.NewHTTPError(http.StatusBadRequest, "No files uploaded")
    }
    
    uploadedFiles := make([]map[string]interface{}, 0)
    
    for _, file := range files {
        // Validate each file
        if err := validateFile(file); err != nil {
            log.Warn("File validation failed", logger.Fields{
                "filename": file.Filename,
                "error":    err,
            })
            continue
        }
        
        // Open file
        src, err := file.Open()
        if err != nil {
            continue
        }
        
        // Upload
        filename := generateUniqueFilename(file.Filename)
        path, err := h.storage.Put(ctx, "uploads", filename, src)
        src.Close()
        
        if err != nil {
            log.Error("Upload failed", logger.Fields{
                "filename": file.Filename,
                "error":    err,
            })
            continue
        }
        
        uploadedFiles = append(uploadedFiles, map[string]interface{}{
            "filename": file.Filename,
            "path":     path,
            "url":      h.storage.URL(path),
            "size":     file.Size,
        })
    }
    
    return c.JSON(http.StatusOK, map[string]interface{}{
        "files": uploadedFiles,
        "count": len(uploadedFiles),
    })
}
```

### Chunked Upload (Large Files)

```go
type ChunkedUpload struct {
    ID          string
    Filename    string
    TotalChunks int
    Chunks      map[int]string
    CreatedAt   time.Time
}

func (h *FileHandler) UploadChunk(c echo.Context) error {
    ctx := c.Request().Context()
    
    uploadID := c.FormValue("upload_id")
    chunkIndex, _ := strconv.Atoi(c.FormValue("chunk_index"))
    totalChunks, _ := strconv.Atoi(c.FormValue("total_chunks"))
    
    // Get chunk data
    file, err := c.FormFile("chunk")
    if err != nil {
        return err
    }
    
    src, err := file.Open()
    if err != nil {
        return err
    }
    defer src.Close()
    
    // Save chunk
    chunkPath := fmt.Sprintf("chunks/%s/%d", uploadID, chunkIndex)
    _, err = h.storage.Put(ctx, "temp", chunkPath, src)
    if err != nil {
        return err
    }
    
    // Track upload progress
    upload := getOrCreateUpload(uploadID)
    upload.Chunks[chunkIndex] = chunkPath
    
    // Check if all chunks uploaded
    if len(upload.Chunks) == totalChunks {
        // Combine chunks
        finalPath, err := h.combineChunks(ctx, upload)
        if err != nil {
            return err
        }
        
        // Clean up chunks
        h.cleanupChunks(ctx, upload)
        
        return c.JSON(http.StatusOK, map[string]interface{}{
            "complete": true,
            "path":     finalPath,
            "url":      h.storage.URL(finalPath),
        })
    }
    
    return c.JSON(http.StatusOK, map[string]interface{}{
        "complete":       false,
        "chunks_uploaded": len(upload.Chunks),
        "total_chunks":    totalChunks,
    })
}

func (h *FileHandler) combineChunks(ctx context.Context, upload *ChunkedUpload) (string, error) {
    // Create final file
    finalPath := fmt.Sprintf("uploads/%s", upload.Filename)
    var chunks []io.Reader
    
    // Read all chunks in order
    for i := 0; i < upload.TotalChunks; i++ {
        chunkPath := upload.Chunks[i]
        reader, err := h.storage.Get(ctx, chunkPath)
        if err != nil {
            return "", err
        }
        chunks = append(chunks, reader)
    }
    
    // Combine chunks
    combined := io.MultiReader(chunks...)
    
    // Upload final file
    _, err := h.storage.Put(ctx, "", finalPath, combined)
    return finalPath, err
}
```

### Direct Upload to S3 (Presigned URL)

```go
func (h *FileHandler) GetUploadURL(c echo.Context) error {
    filename := c.QueryParam("filename")
    contentType := c.QueryParam("content_type")
    
    // Generate presigned URL
    s3Store := h.storage.(*storage.S3Storage)
    url, err := s3Store.PresignedPutURL(
        "uploads/"+filename,
        contentType,
        15*time.Minute, // URL valid for 15 minutes
    )
    
    if err != nil {
        return err
    }
    
    return c.JSON(http.StatusOK, map[string]interface{}{
        "upload_url": url,
        "method":     "PUT",
        "headers": map[string]string{
            "Content-Type": contentType,
        },
    })
}

// Client-side usage:
// fetch(uploadURL, {
//   method: 'PUT',
//   body: file,
//   headers: { 'Content-Type': file.type }
// })
```

## File Validation

### Size Validation

```go
const (
    MaxImageSize    = 5 * 1024 * 1024  // 5MB
    MaxVideoSize    = 100 * 1024 * 1024 // 100MB
    MaxDocumentSize = 10 * 1024 * 1024  // 10MB
)

func validateFileSize(file *multipart.FileHeader, maxSize int64) error {
    if file.Size > maxSize {
        return fmt.Errorf("file size %d exceeds maximum %d bytes", file.Size, maxSize)
    }
    return nil
}
```

### MIME Type Validation

```go
var (
    AllowedImageTypes = []string{
        "image/jpeg",
        "image/png",
        "image/gif",
        "image/webp",
    }
    
    AllowedDocumentTypes = []string{
        "application/pdf",
        "application/msword",
        "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
        "text/plain",
    }
)

func validateMimeType(file *multipart.FileHeader, allowedTypes []string) error {
    contentType := file.Header.Get("Content-Type")
    
    for _, allowed := range allowedTypes {
        if contentType == allowed {
            return nil
        }
    }
    
    return fmt.Errorf("file type %s not allowed", contentType)
}

// Validate by file content (more secure)
func detectMimeType(file io.Reader) (string, error) {
    buffer := make([]byte, 512)
    _, err := file.Read(buffer)
    if err != nil {
        return "", err
    }
    
    return http.DetectContentType(buffer), nil
}
```

### Image Dimension Validation

```go
import "image"

func validateImageDimensions(file io.Reader, maxWidth, maxHeight int) error {
    img, _, err := image.Decode(file)
    if err != nil {
        return fmt.Errorf("invalid image: %w", err)
    }
    
    bounds := img.Bounds()
    width := bounds.Dx()
    height := bounds.Dy()
    
    if width > maxWidth || height > maxHeight {
        return fmt.Errorf("image dimensions %dx%d exceed maximum %dx%d",
            width, height, maxWidth, maxHeight)
    }
    
    return nil
}
```

### Complete File Validator

```go
type FileValidator struct {
    MaxSize       int64
    AllowedTypes  []string
    MaxWidth      int
    MaxHeight     int
}

func (v *FileValidator) Validate(file *multipart.FileHeader) error {
    // Size validation
    if err := validateFileSize(file, v.MaxSize); err != nil {
        return err
    }
    
    // MIME type validation
    if err := validateMimeType(file, v.AllowedTypes); err != nil {
        return err
    }
    
    // For images, validate dimensions
    if isImage(file) {
        src, err := file.Open()
        if err != nil {
            return err
        }
        defer src.Close()
        
        if err := validateImageDimensions(src, v.MaxWidth, v.MaxHeight); err != nil {
            return err
        }
    }
    
    return nil
}

// Usage
imageValidator := &FileValidator{
    MaxSize:      5 * 1024 * 1024,
    AllowedTypes: AllowedImageTypes,
    MaxWidth:     4096,
    MaxHeight:    4096,
}

if err := imageValidator.Validate(file); err != nil {
    return echo.NewHTTPError(http.StatusBadRequest, err.Error())
}
```

## Image Processing

### Image Resize

```go
import (
    "image"
    "image/jpeg"
    "github.com/nfnt/resize"
)

func resizeImage(src io.Reader, width, height uint) (io.Reader, error) {
    // Decode image
    img, _, err := image.Decode(src)
    if err != nil {
        return nil, err
    }
    
    // Resize
    resized := resize.Resize(width, height, img, resize.Lanczos3)
    
    // Encode to buffer
    var buf bytes.Buffer
    if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 90}); err != nil {
        return nil, err
    }
    
    return &buf, nil
}

// Usage
func (h *FileHandler) UploadImage(c echo.Context) error {
    file, _ := c.FormFile("image")
    src, _ := file.Open()
    defer src.Close()
    
    // Create thumbnails
    sizes := []struct {
        name   string
        width  uint
        height uint
    }{
        {"large", 1200, 0},
        {"medium", 600, 0},
        {"small", 300, 0},
        {"thumb", 150, 150},
    }
    
    urls := make(map[string]string)
    
    for _, size := range sizes {
        resized, err := resizeImage(src, size.width, size.height)
        if err != nil {
            continue
        }
        
        filename := fmt.Sprintf("%s_%s.jpg", generateID(), size.name)
        path, _ := h.storage.Put(ctx, "images", filename, resized)
        urls[size.name] = h.storage.URL(path)
        
        // Reset reader for next iteration
        src.Seek(0, 0)
    }
    
    return c.JSON(http.StatusOK, urls)
}
```

### Image Crop

```go
import "image"

func cropImage(src io.Reader, x, y, width, height int) (io.Reader, error) {
    img, _, err := image.Decode(src)
    if err != nil {
        return nil, err
    }
    
    // Create crop rectangle
    rect := image.Rect(x, y, x+width, y+height)
    
    // Crop image
    cropped := img.(interface {
        SubImage(r image.Rectangle) image.Image
    }).SubImage(rect)
    
    // Encode
    var buf bytes.Buffer
    if err := jpeg.Encode(&buf, cropped, nil); err != nil {
        return nil, err
    }
    
    return &buf, nil
}
```

### Watermark

```go
import (
    "image"
    "image/draw"
)

func addWatermark(baseImage, watermark io.Reader) (io.Reader, error) {
    // Decode base image
    base, _, err := image.Decode(baseImage)
    if err != nil {
        return nil, err
    }
    
    // Decode watermark
    mark, _, err := image.Decode(watermark)
    if err != nil {
        return nil, err
    }
    
    // Create new image
    bounds := base.Bounds()
    result := image.NewRGBA(bounds)
    
    // Draw base image
    draw.Draw(result, bounds, base, image.Point{}, draw.Src)
    
    // Position watermark (bottom right corner)
    offset := image.Pt(
        bounds.Max.X-mark.Bounds().Dx()-10,
        bounds.Max.Y-mark.Bounds().Dy()-10,
    )
    
    // Draw watermark
    draw.Draw(result, mark.Bounds().Add(offset), mark, image.Point{}, draw.Over)
    
    // Encode
    var buf bytes.Buffer
    if err := jpeg.Encode(&buf, result, nil); err != nil {
        return nil, err
    }
    
    return &buf, nil
}
```

## Serving Files

### Public Files

```go
func (h *FileHandler) ServeFile(c echo.Context) error {
    path := c.Param("path")
    
    // Get file from storage
    reader, err := h.storage.Get(c.Request().Context(), path)
    if err != nil {
        return echo.NewHTTPError(http.StatusNotFound)
    }
    defer reader.Close()
    
    // Set content type
    contentType := mime.TypeByExtension(filepath.Ext(path))
    if contentType == "" {
        contentType = "application/octet-stream"
    }
    
    c.Response().Header().Set("Content-Type", contentType)
    
    // Stream file
    _, err = io.Copy(c.Response(), reader)
    return err
}
```

### Private Files with Authentication

```go
func (h *FileHandler) ServePrivateFile(c echo.Context) error {
    fileID, _ := strconv.Atoi(c.Param("id"))
    userID := getUserID(c)
    
    // Get file record
    var file File
    if err := h.db.First(&file, fileID).Error; err != nil {
        return echo.NewHTTPError(http.StatusNotFound)
    }
    
    // Check access permission
    if file.UserID != userID && !isAdmin(c) {
        return echo.NewHTTPError(http.StatusForbidden)
    }
    
    // Serve file
    reader, err := h.storage.Get(c.Request().Context(), file.Path)
    if err != nil {
        return err
    }
    defer reader.Close()
    
    // Set headers
    c.Response().Header().Set("Content-Type", file.MimeType)
    c.Response().Header().Set("Content-Disposition", 
        fmt.Sprintf("attachment; filename=%s", file.OriginalName))
    
    io.Copy(c.Response(), reader)
    return nil
}
```

### Signed URLs (Temporary Access)

```go
func (h *FileHandler) GetSignedURL(c echo.Context) error {
    fileID, _ := strconv.Atoi(c.Param("id"))
    userID := getUserID(c)
    
    // Get file and verify permission
    var file File
    if err := h.db.First(&file, fileID).Error; err != nil {
        return echo.NewHTTPError(http.StatusNotFound)
    }
    
    if file.UserID != userID {
        return echo.NewHTTPError(http.StatusForbidden)
    }
    
    // Generate signed URL (valid for 1 hour)
    url := h.storage.SignedURL(file.Path, 1*time.Hour)
    
    return c.JSON(http.StatusOK, map[string]interface{}{
        "url":        url,
        "expires_at": time.Now().Add(1 * time.Hour),
    })
}
```

### CDN Integration

```go
type CDNStorage struct {
    storage storage.Storage
    cdnURL  string
}

func (cdn *CDNStorage) URL(path string) string {
    // Return CDN URL instead of storage URL
    return fmt.Sprintf("%s/%s", cdn.cdnURL, path)
}

// Usage
cdnStorage := &CDNStorage{
    storage: s3Storage,
    cdnURL:  "https://cdn.example.com",
}

url := cdnStorage.URL("images/photo.jpg")
// Returns: https://cdn.example.com/images/photo.jpg
```

## Security

### Filename Sanitization

```go
import (
    "path/filepath"
    "regexp"
    "strings"
)

var (
    illegalChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
    multiSpaces  = regexp.MustCompile(`\s+`)
)

func sanitizeFilename(filename string) string {
    // Get extension
    ext := filepath.Ext(filename)
    name := strings.TrimSuffix(filename, ext)
    
    // Remove illegal characters
    name = illegalChars.ReplaceAllString(name, "")
    
    // Replace multiple spaces with single space
    name = multiSpaces.ReplaceAllString(name, " ")
    
    // Trim spaces
    name = strings.TrimSpace(name)
    
    // Add unique ID to prevent collisions
    name = fmt.Sprintf("%s_%s", generateID(), name)
    
    return name + ext
}
```

### Virus Scanning

```go
import "github.com/dutchcoders/go-clamd"

func scanFile(file io.Reader) error {
    c := clamd.NewClamd("/var/run/clamav/clamd.sock")
    
    response, err := c.ScanStream(file, make(chan bool))
    if err != nil {
        return err
    }
    
    for s := range response {
        if s.Status == clamd.RES_FOUND {
            return fmt.Errorf("virus detected: %s", s.Description)
        }
    }
    
    return nil
}

// Usage in upload handler
if err := scanFile(src); err != nil {
    return echo.NewHTTPError(http.StatusBadRequest, "File rejected: virus detected")
}
```

### Access Control

```go
type FileACL struct {
    FileID     int
    UserID     int
    Permission string // "read", "write", "delete"
}

func (h *FileHandler) CheckAccess(fileID, userID int, permission string) bool {
    var acl FileACL
    err := h.db.Where("file_id = ? AND user_id = ? AND permission = ?",
        fileID, userID, permission).First(&acl).Error
    
    return err == nil
}

func (h *FileHandler) Download(c echo.Context) error {
    fileID, _ := strconv.Atoi(c.Param("id"))
    userID := getUserID(c)
    
    if !h.CheckAccess(fileID, userID, "read") {
        return echo.NewHTTPError(http.StatusForbidden)
    }
    
    // Serve file...
}
```

## Best Practices

### 1. Unique Filenames

```go
import (
    "crypto/rand"
    "encoding/hex"
)

func generateUniqueFilename(original string) string {
    ext := filepath.Ext(original)
    id := generateID()
    timestamp := time.Now().Unix()
    
    return fmt.Sprintf("%d_%s%s", timestamp, id, ext)
}

func generateID() string {
    b := make([]byte, 16)
    rand.Read(b)
    return hex.EncodeToString(b)
}
```

### 2. Organized Directory Structure

```go
func getUploadPath(userID int, fileType string) string {
    now := time.Now()
    return fmt.Sprintf("%s/%d/%d/%d",
        fileType,
        userID,
        now.Year(),
        now.Month(),
    )
}

// Result: images/123/2024/12/filename.jpg
```

### 3. Cleanup Old Files

```go
func (s *StorageService) CleanupOldFiles(ctx context.Context, days int) error {
    cutoff := time.Now().AddDate(0, 0, -days)
    
    var files []File
    s.db.Where("created_at < ? AND deleted_at IS NULL", cutoff).Find(&files)
    
    for _, file := range files {
        // Delete from storage
        s.storage.Delete(ctx, file.Path)
        
        // Mark as deleted in database
        s.db.Delete(&file)
    }
    
    return nil
}
```

### 4. Progress Tracking

```go
type ProgressReader struct {
    reader   io.Reader
    total    int64
    read     int64
    callback func(read, total int64)
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
    n, err := pr.reader.Read(p)
    pr.read += int64(n)
    
    if pr.callback != nil {
        pr.callback(pr.read, pr.total)
    }
    
    return n, err
}

// Usage
progressReader := &ProgressReader{
    reader: file,
    total:  file.Size,
    callback: func(read, total int64) {
        percent := float64(read) / float64(total) * 100
        log.Info("Upload progress", logger.Fields{
            "percent": fmt.Sprintf("%.2f%%", percent),
        })
    },
}

storage.Put(ctx, "uploads", filename, progressReader)
```

### 5. Error Handling

```go
func (h *FileHandler) Upload(c echo.Context) error {
    file, err := c.FormFile("file")
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "No file uploaded")
    }
    
    // Validate
    if err := h.validator.Validate(file); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }
    
    src, err := file.Open()
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to open file")
    }
    defer src.Close()
    
    // Upload
    path, err := h.storage.Put(c.Request().Context(), "uploads", filename, src)
    if err != nil {
        log.Error("Upload failed", logger.Fields{
            "filename": file.Filename,
            "error":    err,
        })
        return echo.NewHTTPError(http.StatusInternalServerError, "Upload failed")
    }
    
    return c.JSON(http.StatusOK, map[string]string{
        "path": path,
        "url":  h.storage.URL(path),
    })
}
```

## Troubleshooting

### Upload Size Limit

```go
// In main.go
e := echo.New()
e.Use(middleware.BodyLimit("10M")) // Set max body size

// Or in nginx
// client_max_body_size 10M;
```

### S3 Permission Issues

```go
// Check IAM policy
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject",
                "s3:DeleteObject"
            ],
            "Resource": "arn:aws:s3:::your-bucket/*"
        }
    ]
}
```

### Memory Issues with Large Files

```go
// Use streaming instead of loading entire file
func (s *S3Storage) PutStream(ctx context.Context, path string, reader io.Reader) error {
    // Stream directly to S3 without buffering
    uploader := s3manager.NewUploader(s.session)
    
    _, err := uploader.Upload(&s3manager.UploadInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(path),
        Body:   reader,
    })
    
    return err
}
```

---

**Next Steps:**
- Learn about [Image Processing](image-processing.md) for advanced manipulation
- Explore [CDN Setup](../deployment/cdn.md) for performance
- See [Security](../security/file-security.md) for access control

**Related Topics:**
- [File Validation](../core-concepts/validation.md)
- [API Endpoints](../api-reference/files.md)
- [S3 Configuration](../deployment/s3-setup.md)
