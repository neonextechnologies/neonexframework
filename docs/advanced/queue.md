# Background Queue System

Master asynchronous job processing with NeonEx Framework's queue system. Learn to handle background tasks, schedule jobs, implement retry logic, and build scalable async workflows.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [Queue Configuration](#queue-configuration)
- [Job Definition](#job-definition)
- [Job Dispatching](#job-dispatching)
- [Workers](#workers)
- [Job Scheduling](#job-scheduling)
- [Retry Logic](#retry-logic)
- [Failed Jobs](#failed-jobs)
- [Job Priority](#job-priority)
- [Integration Examples](#integration-examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Introduction

NeonEx provides a powerful queue system for handling background jobs and async processing. Key features include:

- **Multiple Drivers**: Redis, Database, Memory queues
- **Job Scheduling**: Delayed and scheduled jobs
- **Retry Logic**: Automatic retry with exponential backoff
- **Failed Job Tracking**: Monitor and retry failed jobs
- **Job Priority**: High/Normal/Low priority queues
- **Worker Pools**: Multiple concurrent workers
- **Job Chaining**: Sequential job execution
- **Progress Tracking**: Monitor job status

## Quick Start

### Basic Job Processing

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "neonex/core/pkg/queue"
)

// Define a job
type SendEmailJob struct {
    To      string
    Subject string
    Body    string
}

func (j *SendEmailJob) Handle(ctx context.Context) error {
    fmt.Printf("Sending email to %s: %s\n", j.To, j.Subject)
    // Send email logic here
    time.Sleep(2 * time.Second) // Simulate work
    return nil
}

func main() {
    // Create queue manager
    qm := queue.NewManager(queue.Config{
        Driver:     "redis",
        Connection: "localhost:6379",
    })
    
    // Dispatch job
    job := &SendEmailJob{
        To:      "user@example.com",
        Subject: "Welcome!",
        Body:    "Welcome to our platform!",
    }
    
    qm.Dispatch(context.Background(), "emails", job)
    
    // Start worker
    worker := qm.Worker("emails", 5) // 5 concurrent workers
    worker.Start(context.Background())
}
```

## Queue Configuration

### Redis Queue

```go
import "neonex/core/pkg/queue"

config := queue.Config{
    Driver:         "redis",
    Connection:     "localhost:6379",
    Password:       "",
    Database:       0,
    
    // Queue settings
    DefaultQueue:   "default",
    MaxRetries:     3,
    RetryDelay:     5 * time.Second,
    Timeout:        60 * time.Second,
    
    // Worker settings
    Workers:        5,
    MaxJobs:        100,
    PollInterval:   1 * time.Second,
}

manager := queue.NewManager(config)
```

### Database Queue

```go
config := queue.Config{
    Driver:         "database",
    Connection:     db, // *gorm.DB instance
    DefaultQueue:   "default",
    TableName:      "jobs",
    MaxRetries:     3,
    Workers:        5,
}

manager := queue.NewManager(config)
```

### Memory Queue (Development)

```go
config := queue.Config{
    Driver:       "memory",
    DefaultQueue: "default",
    Workers:      3,
}

manager := queue.NewManager(config)
```

### Environment Configuration

```yaml
# config/queue.yaml
queue:
  driver: redis
  connection: ${REDIS_URL}
  password: ${REDIS_PASSWORD}
  database: 0
  
  default_queue: default
  max_retries: 3
  retry_delay: 5s
  timeout: 60s
  
  workers: 5
  max_jobs: 100
  poll_interval: 1s
```

## Job Definition

### Simple Job

```go
type ProcessImageJob struct {
    ImageID int
    UserID  int
}

func (j *ProcessImageJob) Handle(ctx context.Context) error {
    log.Info("Processing image", logger.Fields{
        "image_id": j.ImageID,
        "user_id":  j.UserID,
    })
    
    // Image processing logic
    return processImage(ctx, j.ImageID)
}
```

### Job with Dependencies

```go
type GenerateReportJob struct {
    ReportID int
    db       *gorm.DB
    storage  storage.Storage
    mailer   *notification.Manager
}

func NewGenerateReportJob(reportID int, db *gorm.DB, storage storage.Storage, mailer *notification.Manager) *GenerateReportJob {
    return &GenerateReportJob{
        ReportID: reportID,
        db:       db,
        storage:  storage,
        mailer:   mailer,
    }
}

func (j *GenerateReportJob) Handle(ctx context.Context) error {
    // Fetch data
    var report Report
    if err := j.db.First(&report, j.ReportID).Error; err != nil {
        return err
    }
    
    // Generate PDF
    pdf := generatePDF(&report)
    
    // Upload to storage
    url, err := j.storage.Upload(ctx, "reports", pdf)
    if err != nil {
        return err
    }
    
    // Send email
    return j.mailer.SendEmail(ctx, report.Email, "Report Ready", 
        fmt.Sprintf("Your report is ready: %s", url))
}
```

### Job Interface

```go
// Job interface that all jobs must implement
type Job interface {
    Handle(ctx context.Context) error
}

// Optional interfaces
type Retryable interface {
    ShouldRetry(err error) bool
    MaxAttempts() int
}

type Delayable interface {
    Delay() time.Duration
}

type Prioritizable interface {
    Priority() Priority
}
```

### Advanced Job

```go
type ComplexJob struct {
    Data map[string]interface{}
}

func (j *ComplexJob) Handle(ctx context.Context) error {
    // Main logic
    return nil
}

// Implement retry logic
func (j *ComplexJob) ShouldRetry(err error) bool {
    // Don't retry validation errors
    if _, ok := err.(*ValidationError); ok {
        return false
    }
    return true
}

func (j *ComplexJob) MaxAttempts() int {
    return 5
}

// Implement delay
func (j *ComplexJob) Delay() time.Duration {
    return 10 * time.Second
}

// Implement priority
func (j *ComplexJob) Priority() queue.Priority {
    if urgent, ok := j.Data["urgent"].(bool); ok && urgent {
        return queue.PriorityHigh
    }
    return queue.PriorityNormal
}
```

## Job Dispatching

### Immediate Dispatch

```go
ctx := context.Background()

// Dispatch to default queue
qm.Dispatch(ctx, "default", &SendEmailJob{
    To:      "user@example.com",
    Subject: "Hello",
    Body:    "Message body",
})

// Dispatch to specific queue
qm.Dispatch(ctx, "emails", job)
```

### Delayed Dispatch

```go
// Dispatch with delay
qm.DispatchAfter(ctx, "emails", job, 5*time.Minute)

// Dispatch at specific time
executeAt := time.Now().Add(1 * time.Hour)
qm.DispatchAt(ctx, "emails", job, executeAt)
```

### Batch Dispatch

```go
jobs := []queue.Job{
    &SendEmailJob{To: "user1@example.com", Subject: "Hello"},
    &SendEmailJob{To: "user2@example.com", Subject: "Hello"},
    &SendEmailJob{To: "user3@example.com", Subject: "Hello"},
}

qm.DispatchBatch(ctx, "emails", jobs)
```

### Conditional Dispatch

```go
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    user := &User{
        Email: req.Email,
        Name:  req.Name,
    }
    
    if err := s.db.Create(user).Error; err != nil {
        return err
    }
    
    // Dispatch welcome email job
    if req.SendWelcomeEmail {
        s.qm.Dispatch(ctx, "emails", &SendWelcomeEmailJob{
            UserID: user.ID,
            Email:  user.Email,
        })
    }
    
    return nil
}
```

## Workers

### Starting Workers

```go
// Start single worker
worker := qm.Worker("default", 1)
go worker.Start(context.Background())

// Start multiple workers
for i := 0; i < 5; i++ {
    worker := qm.Worker("default", 1)
    go worker.Start(context.Background())
}

// Worker with custom configuration
worker := qm.Worker("emails", 3, queue.WorkerOptions{
    MaxJobs:      100,
    Timeout:      5 * time.Minute,
    PollInterval: 2 * time.Second,
})
```

### Worker Lifecycle

```go
type QueueWorker struct {
    manager *queue.Manager
    queue   string
    workers int
    ctx     context.Context
    cancel  context.CancelFunc
}

func NewQueueWorker(manager *queue.Manager, queue string, workers int) *QueueWorker {
    ctx, cancel := context.WithCancel(context.Background())
    return &QueueWorker{
        manager: manager,
        queue:   queue,
        workers: workers,
        ctx:     ctx,
        cancel:  cancel,
    }
}

func (qw *QueueWorker) Start() {
    log.Info("Starting queue workers", logger.Fields{
        "queue":   qw.queue,
        "workers": qw.workers,
    })
    
    for i := 0; i < qw.workers; i++ {
        go qw.work(i)
    }
}

func (qw *QueueWorker) work(id int) {
    worker := qw.manager.Worker(qw.queue, 1)
    
    log.Info("Worker started", logger.Fields{
        "worker_id": id,
        "queue":     qw.queue,
    })
    
    if err := worker.Start(qw.ctx); err != nil {
        log.Error("Worker stopped", logger.Fields{
            "worker_id": id,
            "error":     err,
        })
    }
}

func (qw *QueueWorker) Stop() {
    log.Info("Stopping queue workers")
    qw.cancel()
}
```

### Graceful Shutdown

```go
func main() {
    qm := queue.NewManager(config)
    
    // Start workers
    ctx, cancel := context.WithCancel(context.Background())
    worker := qm.Worker("default", 5)
    go worker.Start(ctx)
    
    // Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    <-sigChan
    log.Info("Shutting down workers...")
    
    // Cancel context (graceful shutdown)
    cancel()
    
    // Wait for workers to finish current jobs
    time.Sleep(5 * time.Second)
    
    log.Info("Shutdown complete")
}
```

## Job Scheduling

### Cron Jobs

```go
type ScheduledJob struct {
    Name     string
    Schedule string // Cron expression
    Job      queue.Job
}

type Scheduler struct {
    manager *queue.Manager
    jobs    []ScheduledJob
}

func NewScheduler(manager *queue.Manager) *Scheduler {
    return &Scheduler{
        manager: manager,
        jobs:    make([]ScheduledJob, 0),
    }
}

func (s *Scheduler) Register(name, schedule string, job queue.Job) {
    s.jobs = append(s.jobs, ScheduledJob{
        Name:     name,
        Schedule: schedule,
        Job:      job,
    })
}

func (s *Scheduler) Start(ctx context.Context) {
    for _, scheduled := range s.jobs {
        go s.runScheduled(ctx, scheduled)
    }
}

func (s *Scheduler) runScheduled(ctx context.Context, scheduled ScheduledJob) {
    ticker := cron.New(scheduled.Schedule)
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            log.Info("Running scheduled job", logger.Fields{
                "job": scheduled.Name,
            })
            s.manager.Dispatch(ctx, "scheduled", scheduled.Job)
        }
    }
}

// Usage
scheduler := NewScheduler(qm)

// Run every day at midnight
scheduler.Register("daily-report", "0 0 * * *", &GenerateReportJob{})

// Run every hour
scheduler.Register("cleanup", "0 * * * *", &CleanupJob{})

// Run every 5 minutes
scheduler.Register("sync", "*/5 * * * *", &SyncDataJob{})

scheduler.Start(context.Background())
```

### Recurring Jobs

```go
type RecurringJob struct {
    Job      queue.Job
    Interval time.Duration
}

func (s *Scheduler) RegisterRecurring(name string, interval time.Duration, job queue.Job) {
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        
        for range ticker.C {
            log.Info("Running recurring job", logger.Fields{
                "job":      name,
                "interval": interval,
            })
            s.manager.Dispatch(context.Background(), "recurring", job)
        }
    }()
}

// Usage
scheduler.RegisterRecurring("health-check", 30*time.Second, &HealthCheckJob{})
scheduler.RegisterRecurring("metrics-export", 5*time.Minute, &ExportMetricsJob{})
```

## Retry Logic

### Automatic Retry

```go
type RetryableJob struct {
    Data     string
    Attempts int
}

func (j *RetryableJob) Handle(ctx context.Context) error {
    j.Attempts++
    
    // Simulate failure
    if j.Attempts < 3 {
        return fmt.Errorf("temporary failure (attempt %d)", j.Attempts)
    }
    
    // Success on 3rd attempt
    log.Info("Job succeeded", logger.Fields{
        "attempts": j.Attempts,
    })
    return nil
}

func (j *RetryableJob) MaxAttempts() int {
    return 5
}

func (j *RetryableJob) ShouldRetry(err error) bool {
    // Don't retry permanent errors
    if errors.Is(err, ErrPermanent) {
        return false
    }
    return true
}
```

### Exponential Backoff

```go
func calculateBackoff(attempt int) time.Duration {
    // Exponential backoff: 1s, 2s, 4s, 8s, 16s...
    delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
    
    // Cap at 1 hour
    if delay > time.Hour {
        delay = time.Hour
    }
    
    return delay
}

type BackoffJob struct {
    Attempts int
}

func (j *BackoffJob) Handle(ctx context.Context) error {
    // Try to process
    if err := process(); err != nil {
        return err
    }
    return nil
}

func (j *BackoffJob) RetryDelay() time.Duration {
    return calculateBackoff(j.Attempts)
}
```

### Custom Retry Strategy

```go
type CustomRetryJob struct {
    Data     interface{}
    Attempts int
    LastErr  error
}

func (j *CustomRetryJob) ShouldRetry(err error) bool {
    // Network errors are retryable
    if isNetworkError(err) {
        return true
    }
    
    // Rate limit errors - retry with longer delay
    if isRateLimitError(err) {
        return true
    }
    
    // Validation errors - don't retry
    if isValidationError(err) {
        return false
    }
    
    return true
}

func (j *CustomRetryJob) RetryDelay() time.Duration {
    // Custom delay based on error type
    if isRateLimitError(j.LastErr) {
        return 5 * time.Minute
    }
    
    return calculateBackoff(j.Attempts)
}
```

## Failed Jobs

### Tracking Failed Jobs

```go
type FailedJob struct {
    ID          int
    Queue       string
    Job         []byte // Serialized job
    Error       string
    FailedAt    time.Time
    Attempts    int
    LastAttempt time.Time
}

type FailedJobRepository struct {
    db *gorm.DB
}

func (r *FailedJobRepository) Store(job queue.Job, err error, attempts int) error {
    jobData, _ := json.Marshal(job)
    
    failedJob := &FailedJob{
        Queue:       "default",
        Job:         jobData,
        Error:       err.Error(),
        FailedAt:    time.Now(),
        Attempts:    attempts,
        LastAttempt: time.Now(),
    }
    
    return r.db.Create(failedJob).Error
}

func (r *FailedJobRepository) GetAll(limit, offset int) ([]FailedJob, error) {
    var jobs []FailedJob
    err := r.db.Order("failed_at DESC").
        Limit(limit).
        Offset(offset).
        Find(&jobs).Error
    return jobs, err
}

func (r *FailedJobRepository) Retry(id int) error {
    var failedJob FailedJob
    if err := r.db.First(&failedJob, id).Error; err != nil {
        return err
    }
    
    // Deserialize and re-dispatch
    var job queue.Job
    json.Unmarshal(failedJob.Job, &job)
    
    // Delete from failed jobs
    r.db.Delete(&failedJob)
    
    // Re-dispatch
    return qm.Dispatch(context.Background(), failedJob.Queue, job)
}
```

### Failed Job Dashboard

```go
func (h *JobHandler) FailedJobs(c echo.Context) error {
    limit := 50
    offset := getOffset(c)
    
    jobs, err := h.failedJobRepo.GetAll(limit, offset)
    if err != nil {
        return err
    }
    
    return c.JSON(http.StatusOK, map[string]interface{}{
        "jobs":  jobs,
        "total": len(jobs),
    })
}

func (h *JobHandler) RetryJob(c echo.Context) error {
    id, _ := strconv.Atoi(c.Param("id"))
    
    if err := h.failedJobRepo.Retry(id); err != nil {
        return err
    }
    
    return c.JSON(http.StatusOK, map[string]string{
        "message": "Job queued for retry",
    })
}
```

## Job Priority

### Priority Queues

```go
const (
    PriorityHigh   Priority = "high"
    PriorityNormal Priority = "normal"
    PriorityLow    Priority = "low"
)

// Dispatch with priority
qm.DispatchWithPriority(ctx, "emails", job, queue.PriorityHigh)

// Start workers for each priority
highWorker := qm.Worker("emails:high", 5)
normalWorker := qm.Worker("emails:normal", 3)
lowWorker := qm.Worker("emails:low", 1)

go highWorker.Start(ctx)
go normalWorker.Start(ctx)
go lowWorker.Start(ctx)
```

### Dynamic Priority

```go
type PriorityJob struct {
    UserID   int
    IsUrgent bool
    IsPaid   bool
}

func (j *PriorityJob) Priority() queue.Priority {
    if j.IsUrgent {
        return queue.PriorityHigh
    }
    
    if j.IsPaid {
        return queue.PriorityNormal
    }
    
    return queue.PriorityLow
}
```

## Integration Examples

### Complete Email Queue

```go
type EmailQueue struct {
    qm     *queue.Manager
    mailer *notification.Manager
}

func NewEmailQueue(qm *queue.Manager, mailer *notification.Manager) *EmailQueue {
    return &EmailQueue{qm: qm, mailer: mailer}
}

type SendEmailJob struct {
    To       string
    Subject  string
    Body     string
    Template string
    Data     map[string]interface{}
}

func (j *SendEmailJob) Handle(ctx context.Context) error {
    if j.Template != "" {
        return mailer.SendTemplate(ctx, j.To, j.Subject, j.Template, j.Data)
    }
    return mailer.SendEmail(ctx, j.To, j.Subject, j.Body)
}

func (eq *EmailQueue) SendWelcomeEmail(userID int, email string) {
    eq.qm.Dispatch(context.Background(), "emails", &SendEmailJob{
        To:       email,
        Template: "welcome",
        Data: map[string]interface{}{
            "user_id": userID,
        },
    })
}

func (eq *EmailQueue) SendPasswordReset(email, token string) {
    eq.qm.DispatchWithPriority(context.Background(), "emails", 
        &SendEmailJob{
            To:       email,
            Template: "password-reset",
            Data: map[string]interface{}{
                "reset_token": token,
            },
        },
        queue.PriorityHigh,
    )
}
```

### Image Processing Queue

```go
type ImageProcessingJob struct {
    ImageID   int
    UserID    int
    storage   storage.Storage
    db        *gorm.DB
}

func (j *ImageProcessingJob) Handle(ctx context.Context) error {
    // Get image
    var image Image
    if err := j.db.First(&image, j.ImageID).Error; err != nil {
        return err
    }
    
    // Download original
    data, err := j.storage.Download(ctx, image.Path)
    if err != nil {
        return err
    }
    
    // Process: resize, compress, watermark
    thumbnails := []struct {
        size int
        name string
    }{
        {800, "large"},
        {400, "medium"},
        {200, "small"},
        {100, "thumbnail"},
    }
    
    for _, thumb := range thumbnails {
        processed := resizeImage(data, thumb.size)
        path := fmt.Sprintf("images/%d/%s.jpg", j.ImageID, thumb.name)
        
        if _, err := j.storage.Upload(ctx, path, processed); err != nil {
            return err
        }
        
        // Update database
        j.db.Model(&image).Update(thumb.name+"_url", path)
    }
    
    // Mark as processed
    image.Status = "processed"
    return j.db.Save(&image).Error
}
```

## Best Practices

### 1. Job Serialization

```go
// Make jobs JSON-serializable
type GoodJob struct {
    UserID int               `json:"user_id"`
    Data   map[string]string `json:"data"`
}

// Avoid
type BadJob struct {
    DB     *gorm.DB          // Can't serialize
    Cache  cache.Cache       // Can't serialize
    Logger logger.Logger     // Can't serialize
}

// Instead, inject dependencies in Handle()
func (j *GoodJob) Handle(ctx context.Context) error {
    db := getDB(ctx)           // Get from context
    cache := getCache(ctx)     // Get from context
    logger := getLogger(ctx)   // Get from context
    
    // Use dependencies
    return nil
}
```

### 2. Idempotent Jobs

```go
// Make jobs idempotent (safe to run multiple times)
type IdempotentJob struct {
    OrderID int
}

func (j *IdempotentJob) Handle(ctx context.Context) error {
    // Check if already processed
    var order Order
    if err := db.First(&order, j.OrderID).Error; err != nil {
        return err
    }
    
    if order.Status == "processed" {
        return nil // Already processed
    }
    
    // Process order
    if err := processOrder(&order); err != nil {
        return err
    }
    
    // Mark as processed
    order.Status = "processed"
    return db.Save(&order).Error
}
```

### 3. Timeout Handling

```go
func (j *LongRunningJob) Handle(ctx context.Context) error {
    // Add timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()
    
    // Check context in long-running operations
    for i := 0; i < 1000; i++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Continue processing
            processItem(i)
        }
    }
    
    return nil
}
```

### 4. Error Logging

```go
func (j *MyJob) Handle(ctx context.Context) error {
    err := doWork()
    if err != nil {
        log.Error("Job failed", logger.Fields{
            "job":   "MyJob",
            "error": err,
            "data":  j.Data,
        })
        return err
    }
    
    log.Info("Job completed", logger.Fields{
        "job": "MyJob",
    })
    
    return nil
}
```

### 5. Resource Cleanup

```go
func (j *ResourceJob) Handle(ctx context.Context) error {
    // Acquire resources
    file, err := os.Open(j.FilePath)
    if err != nil {
        return err
    }
    defer file.Close() // Always clean up
    
    conn, err := grpc.Dial(j.ServiceURL)
    if err != nil {
        return err
    }
    defer conn.Close()
    
    // Use resources
    return processFile(file, conn)
}
```

## Troubleshooting

### Jobs Not Processing

```go
// Check queue size
size := qm.Size("default")
log.Info("Queue size", logger.Fields{"size": size})

// Check worker status
if !worker.IsRunning() {
    log.Warn("Worker not running")
    worker.Start(ctx)
}
```

### Memory Leaks

```go
// Limit concurrent jobs
worker := qm.Worker("default", 5, queue.WorkerOptions{
    MaxJobs: 100, // Process max 100 jobs then restart
})

// Monitor memory
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        log.Info("Memory stats", logger.Fields{
            "alloc_mb": m.Alloc / 1024 / 1024,
        })
    }
}()
```

### Stuck Jobs

```go
// Add timeout to all jobs
type TimeoutJob struct {
    Job     queue.Job
    Timeout time.Duration
}

func (tj *TimeoutJob) Handle(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, tj.Timeout)
    defer cancel()
    
    done := make(chan error, 1)
    go func() {
        done <- tj.Job.Handle(ctx)
    }()
    
    select {
    case err := <-done:
        return err
    case <-ctx.Done():
        return fmt.Errorf("job timeout after %v", tj.Timeout)
    }
}
```

---

**Next Steps:**
- Learn about [Events](events.md) for triggering jobs
- Explore [Email](email.md) for email job processing
- See [Logging](logging.md) for job monitoring

**Related Topics:**
- [Background Processing](../core-concepts/background-jobs.md)
- [Cron Jobs](../deployment/scheduling.md)
- [Performance Optimization](../deployment/performance.md)
