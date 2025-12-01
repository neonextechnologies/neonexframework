# Request and Response Handling

NeonEx Framework provides a powerful and elegant API for handling HTTP requests and responses, built on top of Fiber's fast HTTP router. This guide covers parsing requests, formatting responses, handling file uploads, managing cookies, and working with headers.

## Table of Contents

- [Request Parsing](#request-parsing)
- [Response Formatting](#response-formatting)
- [File Uploads](#file-uploads)
- [Cookies](#cookies)
- [Headers](#headers)
- [Query Parameters](#query-parameters)
- [Path Parameters](#path-parameters)
- [Best Practices](#best-practices)

## Request Parsing

### JSON Request Body

Parse JSON payloads from request body:

```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

func createUser(c *fiber.Ctx) error {
    var req CreateUserRequest
    
    // Parse JSON body
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
            "details": err.Error(),
        })
    }
    
    // Validate request
    if errors := validator.Validate(req); errors != nil {
        return api.ValidationError(c, errors)
    }
    
    // Create user
    user, err := userService.Create(req)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    
    return api.Created(c, "User created successfully", user)
}
```

### Form Data

Parse form-encoded data:

```go
type LoginForm struct {
    Email    string `form:"email"`
    Password string `form:"password"`
    Remember bool   `form:"remember"`
}

func handleLogin(c *fiber.Ctx) error {
    var form LoginForm
    
    // Parse form data
    if err := c.BodyParser(&form); err != nil {
        return api.BadRequest(c, "Invalid form data", nil)
    }
    
    // Process login
    token, err := authService.Login(form.Email, form.Password)
    if err != nil {
        return api.Unauthorized(c, "Invalid credentials")
    }
    
    return api.Success(c, fiber.Map{
        "token": token,
    })
}
```

### Multipart Form Data

Handle multipart forms with files:

```go
func updateProfile(c *fiber.Ctx) error {
    // Get form fields
    name := c.FormValue("name")
    bio := c.FormValue("bio")
    
    // Get file
    file, err := c.FormFile("avatar")
    if err == nil {
        // Process file upload
        filename := fmt.Sprintf("avatars/%d_%s", userID, file.Filename)
        if err := c.SaveFile(file, filename); err != nil {
            return api.InternalError(c, "Failed to save file")
        }
    }
    
    // Update user
    user, err := userService.Update(userID, name, bio, filename)
    if err != nil {
        return api.InternalError(c, err.Error())
    }
    
    return api.Success(c, user)
}
```

### Raw Request Body

Access raw request body:

```go
func handleWebhook(c *fiber.Ctx) error {
    // Get raw body as bytes
    body := c.Body()
    
    // Get raw body as string
    bodyString := string(c.Body())
    
    // Verify webhook signature
    signature := c.Get("X-Webhook-Signature")
    if !verifySignature(body, signature) {
        return api.Unauthorized(c, "Invalid signature")
    }
    
    // Process webhook
    return api.Success(c, fiber.Map{"received": true})
}
```

### Request Context

Access request context:

```go
func myHandler(c *fiber.Ctx) error {
    // Get request context
    ctx := c.Context()
    
    // Use with database operations
    user, err := userRepo.FindByID(ctx, userID)
    if err != nil {
        return api.NotFound(c, "User not found")
    }
    
    return api.Success(c, user)
}
```

## Response Formatting

NeonEx provides standardized response helpers for consistent API responses.

### Success Responses

```go
// Simple success
func getUsers(c *fiber.Ctx) error {
    users, err := userService.GetAll()
    if err != nil {
        return api.InternalError(c, err.Error())
    }
    
    return api.Success(c, users)
}

// Success with message
func deleteUser(c *fiber.Ctx) error {
    id, _ := strconv.Atoi(c.Params("id"))
    
    if err := userService.Delete(uint(id)); err != nil {
        return api.NotFound(c, "User not found")
    }
    
    return api.SuccessWithMessage(c, "User deleted successfully", nil)
}

// Created response (201)
func createPost(c *fiber.Ctx) error {
    var req CreatePostRequest
    if err := c.BodyParser(&req); err != nil {
        return api.BadRequest(c, "Invalid request", nil)
    }
    
    post, err := postService.Create(req)
    if err != nil {
        return api.InternalError(c, err.Error())
    }
    
    return api.Created(c, "Post created successfully", post)
}

// No content response (204)
func updateSettings(c *fiber.Ctx) error {
    // Update settings
    if err := settingsService.Update(); err != nil {
        return api.InternalError(c, err.Error())
    }
    
    return api.NoContent(c)
}
```

### Error Responses

```go
// Bad request (400)
func handler(c *fiber.Ctx) error {
    return api.BadRequest(c, "Invalid input", fiber.Map{
        "field": "email",
        "error": "Invalid email format",
    })
}

// Unauthorized (401)
func protectedRoute(c *fiber.Ctx) error {
    return api.Unauthorized(c, "Authentication required")
}

// Forbidden (403)
func adminOnly(c *fiber.Ctx) error {
    return api.Forbidden(c, "Admin access required")
}

// Not found (404)
func getUser(c *fiber.Ctx) error {
    return api.NotFound(c, "User not found")
}

// Conflict (409)
func createUser(c *fiber.Ctx) error {
    return api.Conflict(c, "Email already exists")
}

// Validation error (422)
func validateRequest(c *fiber.Ctx) error {
    errors := map[string]string{
        "email": "Email is required",
        "password": "Password must be at least 8 characters",
    }
    return api.ValidationError(c, errors)
}

// Internal error (500)
func handler(c *fiber.Ctx) error {
    return api.InternalError(c, "Something went wrong")
}
```

### Paginated Responses

```go
func listUsers(c *fiber.Ctx) error {
    // Get pagination parameters
    pagination := api.GetPagination(c)
    
    // Fetch data
    users, total, err := userService.Paginate(
        pagination.Page,
        pagination.Limit,
    )
    if err != nil {
        return api.InternalError(c, err.Error())
    }
    
    // Return paginated response
    return api.Paginated(c, users, pagination.Page, pagination.Limit, total)
}

// Response format:
// {
//   "success": true,
//   "data": [...],
//   "meta": {
//     "page": 1,
//     "limit": 10,
//     "total": 100,
//     "total_pages": 10,
//     "has_next_page": true,
//     "has_prev_page": false,
//     "next_page": 2,
//     "prev_page": null
//   },
//   "timestamp": 1638360000
// }
```

### Custom Response Formats

```go
// Custom JSON response
func customResponse(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "status": "success",
        "data": fiber.Map{
            "message": "Custom response",
        },
        "meta": fiber.Map{
            "version": "1.0",
        },
    })
}

// Custom status code
func customStatus(c *fiber.Ctx) error {
    return c.Status(418).JSON(fiber.Map{
        "error": "I'm a teapot",
    })
}

// XML response
func xmlResponse(c *fiber.Ctx) error {
    type User struct {
        XMLName xml.Name `xml:"user"`
        ID      int      `xml:"id"`
        Name    string   `xml:"name"`
    }
    
    c.Set("Content-Type", "application/xml")
    return c.XML(User{ID: 1, Name: "John"})
}

// Plain text response
func textResponse(c *fiber.Ctx) error {
    return c.SendString("Hello, World!")
}

// Stream response
func streamResponse(c *fiber.Ctx) error {
    c.Set("Content-Type", "text/event-stream")
    c.Set("Cache-Control", "no-cache")
    c.Set("Connection", "keep-alive")
    
    // Stream data
    for i := 0; i < 10; i++ {
        fmt.Fprintf(c, "data: Message %d\n\n", i)
        c.Context().Flush()
        time.Sleep(time.Second)
    }
    
    return nil
}
```

## File Uploads

### Single File Upload

```go
func uploadFile(c *fiber.Ctx) error {
    // Get uploaded file
    file, err := c.FormFile("file")
    if err != nil {
        return api.BadRequest(c, "No file uploaded", nil)
    }
    
    // Validate file size (5MB max)
    maxSize := int64(5 * 1024 * 1024)
    if file.Size > maxSize {
        return api.BadRequest(c, "File too large", fiber.Map{
            "max_size": "5MB",
        })
    }
    
    // Validate file type
    allowedTypes := map[string]bool{
        "image/jpeg": true,
        "image/png":  true,
        "image/gif":  true,
    }
    
    contentType := file.Header.Get("Content-Type")
    if !allowedTypes[contentType] {
        return api.BadRequest(c, "Invalid file type", fiber.Map{
            "allowed_types": []string{"image/jpeg", "image/png", "image/gif"},
        })
    }
    
    // Generate unique filename
    ext := filepath.Ext(file.Filename)
    filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), generateRandomString(10), ext)
    savePath := filepath.Join("storage/uploads", filename)
    
    // Save file
    if err := c.SaveFile(file, savePath); err != nil {
        return api.InternalError(c, "Failed to save file")
    }
    
    return api.Created(c, "File uploaded successfully", fiber.Map{
        "filename": filename,
        "size":     file.Size,
        "url":      "/uploads/" + filename,
    })
}
```

### Multiple File Uploads

```go
func uploadMultiple(c *fiber.Ctx) error {
    // Get multipart form
    form, err := c.MultipartForm()
    if err != nil {
        return api.BadRequest(c, "Invalid form data", nil)
    }
    
    files := form.File["files"]
    if len(files) == 0 {
        return api.BadRequest(c, "No files uploaded", nil)
    }
    
    // Limit number of files
    if len(files) > 10 {
        return api.BadRequest(c, "Too many files", fiber.Map{
            "max_files": 10,
        })
    }
    
    uploadedFiles := []fiber.Map{}
    
    for _, file := range files {
        // Validate each file
        if file.Size > 5*1024*1024 {
            continue // Skip files larger than 5MB
        }
        
        // Save file
        filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
        savePath := filepath.Join("storage/uploads", filename)
        
        if err := c.SaveFile(file, savePath); err != nil {
            continue
        }
        
        uploadedFiles = append(uploadedFiles, fiber.Map{
            "filename": filename,
            "size":     file.Size,
            "url":      "/uploads/" + filename,
        })
    }
    
    return api.Created(c, "Files uploaded successfully", uploadedFiles)
}
```

### Streaming File Downloads

```go
func downloadFile(c *fiber.Ctx) error {
    filename := c.Params("filename")
    filePath := filepath.Join("storage/uploads", filename)
    
    // Check if file exists
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        return api.NotFound(c, "File not found")
    }
    
    // Set headers for download
    c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
    c.Set("Content-Type", "application/octet-stream")
    
    return c.SendFile(filePath)
}

// Stream large file
func streamLargeFile(c *fiber.Ctx) error {
    filePath := "storage/large-file.zip"
    
    file, err := os.Open(filePath)
    if err != nil {
        return api.NotFound(c, "File not found")
    }
    defer file.Close()
    
    stat, _ := file.Stat()
    c.Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
    c.Set("Content-Type", "application/zip")
    
    return c.SendStream(file)
}
```

## Cookies

### Setting Cookies

```go
func setCookie(c *fiber.Ctx) error {
    // Simple cookie
    c.Cookie(&fiber.Cookie{
        Name:  "session_id",
        Value: "abc123",
    })
    
    // Secure cookie with options
    c.Cookie(&fiber.Cookie{
        Name:     "auth_token",
        Value:    "xyz789",
        Path:     "/",
        Domain:   "example.com",
        MaxAge:   3600, // 1 hour
        Secure:   true,
        HTTPOnly: true,
        SameSite: "Lax",
    })
    
    return api.Success(c, fiber.Map{"message": "Cookie set"})
}

// Session cookie (expires when browser closes)
func setSessionCookie(c *fiber.Ctx) error {
    c.Cookie(&fiber.Cookie{
        Name:     "temp_token",
        Value:    "temp123",
        HTTPOnly: true,
    })
    
    return c.SendString("Session cookie set")
}

// Long-lived cookie
func setRememberMe(c *fiber.Ctx) error {
    c.Cookie(&fiber.Cookie{
        Name:     "remember_token",
        Value:    "long_lived_token",
        MaxAge:   30 * 24 * 3600, // 30 days
        HTTPOnly: true,
        Secure:   true,
    })
    
    return c.SendString("Remember me cookie set")
}
```

### Reading Cookies

```go
func getCookie(c *fiber.Ctx) error {
    // Get single cookie
    sessionID := c.Cookies("session_id")
    if sessionID == "" {
        return api.Unauthorized(c, "No session found")
    }
    
    // Get cookie with default value
    theme := c.Cookies("theme", "light")
    
    return api.Success(c, fiber.Map{
        "session_id": sessionID,
        "theme":      theme,
    })
}
```

### Deleting Cookies

```go
func logout(c *fiber.Ctx) error {
    // Delete cookie by setting MaxAge to -1
    c.Cookie(&fiber.Cookie{
        Name:   "auth_token",
        Value:  "",
        MaxAge: -1,
    })
    
    // Alternative: ClearCookie
    c.ClearCookie("session_id")
    
    return api.Success(c, fiber.Map{"message": "Logged out"})
}
```

## Headers

### Reading Headers

```go
func getHeaders(c *fiber.Ctx) error {
    // Get single header
    userAgent := c.Get("User-Agent")
    authorization := c.Get("Authorization")
    contentType := c.Get("Content-Type")
    
    // Get all headers
    headers := c.GetReqHeaders()
    
    // Get specific headers
    apiKey := c.Get("X-API-Key")
    requestID := c.Get("X-Request-ID")
    
    return api.Success(c, fiber.Map{
        "user_agent": userAgent,
        "headers":    headers,
    })
}
```

### Setting Response Headers

```go
func setHeaders(c *fiber.Ctx) error {
    // Set single header
    c.Set("X-Custom-Header", "value")
    c.Set("X-API-Version", "1.0")
    
    // Set multiple headers
    c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
    c.Set("Pragma", "no-cache")
    c.Set("Expires", "0")
    
    return api.Success(c, fiber.Map{"message": "Headers set"})
}

// Content negotiation
func handleContentType(c *fiber.Ctx) error {
    accept := c.Get("Accept")
    
    data := fiber.Map{"message": "Hello"}
    
    switch {
    case strings.Contains(accept, "application/json"):
        return c.JSON(data)
    case strings.Contains(accept, "application/xml"):
        return c.XML(data)
    default:
        return c.SendString("Hello")
    }
}
```

## Query Parameters

### Reading Query Parameters

```go
func searchUsers(c *fiber.Ctx) error {
    // Get single query parameter
    query := c.Query("q")
    sortBy := c.Query("sort", "created_at") // with default
    
    // Get integer query parameter
    page := c.QueryInt("page", 1)
    limit := c.QueryInt("limit", 10)
    
    // Get boolean query parameter
    active := c.QueryBool("active", true)
    
    // Get array query parameters (?tags=go&tags=fiber)
    tags := c.QueryArray("tags")
    
    // Get all query parameters
    queries := c.Queries()
    
    // Search users
    users, total, err := userService.Search(query, page, limit, sortBy, active, tags)
    if err != nil {
        return api.InternalError(c, err.Error())
    }
    
    return api.Paginated(c, users, page, limit, total)
}
```

## Path Parameters

### Reading Path Parameters

```go
func getUser(c *fiber.Ctx) error {
    // Get path parameter
    id := c.Params("id")
    
    // Convert to integer
    userID, err := strconv.Atoi(id)
    if err != nil {
        return api.BadRequest(c, "Invalid user ID", nil)
    }
    
    user, err := userService.GetByID(uint(userID))
    if err != nil {
        return api.NotFound(c, "User not found")
    }
    
    return api.Success(c, user)
}

// Multiple parameters
func getPostComment(c *fiber.Ctx) error {
    postID := c.Params("postId")
    commentID := c.Params("commentId")
    
    comment, err := commentService.Get(postID, commentID)
    if err != nil {
        return api.NotFound(c, "Comment not found")
    }
    
    return api.Success(c, comment)
}

// Wildcard parameters
func serveFiles(c *fiber.Ctx) error {
    // Route: /files/*
    filepath := c.Params("*")
    return c.SendFile(filepath)
}
```

## Best Practices

### 1. Always Validate Input

```go
func createUser(c *fiber.Ctx) error {
    var req CreateUserRequest
    
    // Parse request
    if err := c.BodyParser(&req); err != nil {
        return api.BadRequest(c, "Invalid request body", nil)
    }
    
    // Validate request
    if errors := validator.Validate(req); errors != nil {
        return api.ValidationError(c, errors)
    }
    
    // Process request
    user, err := userService.Create(req)
    if err != nil {
        return api.InternalError(c, err.Error())
    }
    
    return api.Created(c, "User created", user)
}
```

### 2. Use Consistent Response Format

Always use the standardized response helpers:

```go
// Good
return api.Success(c, data)
return api.Created(c, "Created", data)
return api.NotFound(c, "Not found")

// Avoid
return c.JSON(fiber.Map{"data": data})
return c.Status(404).SendString("Not found")
```

### 3. Handle File Uploads Securely

```go
func secureUpload(c *fiber.Ctx) error {
    file, err := c.FormFile("file")
    if err != nil {
        return api.BadRequest(c, "No file uploaded", nil)
    }
    
    // 1. Validate file size
    if file.Size > maxFileSize {
        return api.BadRequest(c, "File too large", nil)
    }
    
    // 2. Validate file type
    if !isAllowedType(file) {
        return api.BadRequest(c, "Invalid file type", nil)
    }
    
    // 3. Generate secure filename (prevent path traversal)
    filename := generateSecureFilename(file.Filename)
    
    // 4. Save to secure location
    savePath := filepath.Join("storage/uploads", filename)
    if err := c.SaveFile(file, savePath); err != nil {
        return api.InternalError(c, "Failed to save file")
    }
    
    return api.Created(c, "File uploaded", fiber.Map{"filename": filename})
}
```

### 4. Use Context for Database Operations

```go
func getUser(c *fiber.Ctx) error {
    ctx := c.Context()
    
    user, err := userRepo.FindByID(ctx, userID)
    if err != nil {
        return api.NotFound(c, "User not found")
    }
    
    return api.Success(c, user)
}
```

### 5. Implement Proper Error Handling

```go
func handler(c *fiber.Ctx) error {
    result, err := service.DoSomething()
    if err != nil {
        // Log the error
        logger.Error("Operation failed", logger.Fields{
            "error": err.Error(),
            "user_id": c.Locals("user_id"),
        })
        
        // Return appropriate response
        if errors.Is(err, ErrNotFound) {
            return api.NotFound(c, "Resource not found")
        }
        return api.InternalError(c, "Operation failed")
    }
    
    return api.Success(c, result)
}
```

This comprehensive guide provides everything you need to handle requests and responses effectively in NeonEx Framework!
