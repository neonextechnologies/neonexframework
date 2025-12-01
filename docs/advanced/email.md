# Email & Notifications

Master email sending and notification management with NeonEx Framework. Learn SMTP configuration, template rendering, async sending, and multi-channel notifications.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [SMTP Configuration](#smtp-configuration)
- [Sending Emails](#sending-emails)
- [Email Templates](#email-templates)
- [Attachments](#attachments)
- [Async Email Sending](#async-email-sending)
- [Multi-Channel Notifications](#multi-channel-notifications)
- [Email Verification Flow](#email-verification-flow)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Introduction

NeonEx provides a comprehensive notification system with support for multiple channels including email, SMS, and push notifications. Key features:

- **Multiple Channels**: Email, SMS, Push notifications
- **SMTP Support**: Gmail, SendGrid, Mailgun, custom SMTP
- **Template Engine**: HTML/text email templates
- **Attachments**: Send files with emails
- **Async Sending**: Queue-based email delivery
- **Track & Monitor**: Delivery status tracking
- **Rate Limiting**: Prevent spam and respect limits

## Quick Start

### Basic Email Setup

```go
package main

import (
    "context"
    "neonex/core/pkg/notification"
)

func main() {
    // Create notification manager
    manager := notification.NewManager()
    
    // Register email sender
    emailSender := notification.NewSMTPSender(notification.SMTPConfig{
        Host:     "smtp.gmail.com",
        Port:     587,
        Username: "your-email@gmail.com",
        Password: "your-app-password",
        From:     "noreply@example.com",
    })
    
    manager.RegisterSender(notification.ChannelEmail, emailSender)
    
    // Send email
    ctx := context.Background()
    err := manager.SendEmail(ctx, 
        "user@example.com",
        "Welcome to NeonEx!",
        "Thank you for signing up. We're excited to have you!",
    )
    
    if err != nil {
        panic(err)
    }
}
```

## SMTP Configuration

### Gmail SMTP

```go
config := notification.SMTPConfig{
    Host:     "smtp.gmail.com",
    Port:     587,
    Username: "your-email@gmail.com",
    Password: "your-app-password", // Use App Password, not regular password
    From:     "noreply@example.com",
    FromName: "NeonEx Team",
    
    // TLS settings
    TLS:      true,
    
    // Connection settings
    Timeout:  10 * time.Second,
    KeepAlive: true,
}

sender := notification.NewSMTPSender(config)
```

### SendGrid

```go
config := notification.SMTPConfig{
    Host:     "smtp.sendgrid.net",
    Port:     587,
    Username: "apikey",
    Password: "your-sendgrid-api-key",
    From:     "noreply@yourdomain.com",
    FromName: "Your App Name",
}

sender := notification.NewSMTPSender(config)
```

### Mailgun

```go
config := notification.SMTPConfig{
    Host:     "smtp.mailgun.org",
    Port:     587,
    Username: "postmaster@yourdomain.mailgun.org",
    Password: "your-mailgun-password",
    From:     "noreply@yourdomain.com",
}

sender := notification.NewSMTPSender(config)
```

### Custom SMTP

```go
config := notification.SMTPConfig{
    Host:     "mail.yourserver.com",
    Port:     465,
    Username: "smtp-user",
    Password: "smtp-password",
    From:     "noreply@yourdomain.com",
    TLS:      true,
}

sender := notification.NewSMTPSender(config)
```

### Environment Configuration

```yaml
# config/mail.yaml
mail:
  driver: smtp
  host: ${SMTP_HOST}
  port: ${SMTP_PORT}
  username: ${SMTP_USERNAME}
  password: ${SMTP_PASSWORD}
  from_address: ${MAIL_FROM_ADDRESS}
  from_name: ${MAIL_FROM_NAME}
  tls: true
```

```go
// Load from environment
config := notification.SMTPConfig{
    Host:     os.Getenv("SMTP_HOST"),
    Port:     getEnvInt("SMTP_PORT", 587),
    Username: os.Getenv("SMTP_USERNAME"),
    Password: os.Getenv("SMTP_PASSWORD"),
    From:     os.Getenv("MAIL_FROM_ADDRESS"),
    FromName: os.Getenv("MAIL_FROM_NAME"),
    TLS:      true,
}
```

## Sending Emails

### Simple Email

```go
ctx := context.Background()

err := manager.SendEmail(ctx,
    "user@example.com",
    "Welcome!",
    "Thank you for signing up!",
)
```

### HTML Email

```go
htmlBody := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; }
        .button { background-color: #4CAF50; color: white; padding: 10px 20px; }
    </style>
</head>
<body>
    <h1>Welcome to NeonEx!</h1>
    <p>Thank you for signing up.</p>
    <a href="https://example.com/verify" class="button">Verify Email</a>
</body>
</html>
`

err := manager.Send(ctx, &notification.Notification{
    Channel: notification.ChannelEmail,
    To:      "user@example.com",
    Subject: "Welcome!",
    Body:    htmlBody,
    Data: map[string]interface{}{
        "html": true,
    },
})
```

### Email with CC and BCC

```go
err := manager.Send(ctx, &notification.Notification{
    Channel: notification.ChannelEmail,
    To:      "user@example.com",
    Subject: "Team Update",
    Body:    "Important team announcement...",
    Data: map[string]interface{}{
        "cc":  []string{"manager@example.com"},
        "bcc": []string{"admin@example.com"},
    },
})
```

### Bulk Email

```go
recipients := []string{
    "user1@example.com",
    "user2@example.com",
    "user3@example.com",
}

for _, recipient := range recipients {
    go manager.SendEmail(ctx, recipient, "Newsletter", newsletterBody)
}
```

## Email Templates

### Template System

```go
type EmailTemplate struct {
    Name    string
    Subject string
    HTML    string
    Text    string
}

type TemplateRenderer struct {
    templates map[string]*EmailTemplate
}

func NewTemplateRenderer() *TemplateRenderer {
    return &TemplateRenderer{
        templates: make(map[string]*EmailTemplate),
    }
}

func (tr *TemplateRenderer) Register(template *EmailTemplate) {
    tr.templates[template.Name] = template
}

func (tr *TemplateRenderer) Render(name string, data map[string]interface{}) (string, string, error) {
    template, exists := tr.templates[name]
    if !exists {
        return "", "", fmt.Errorf("template not found: %s", name)
    }
    
    // Render HTML
    htmlTmpl, _ := htmlTemplate.New("html").Parse(template.HTML)
    var htmlBuf bytes.Buffer
    htmlTmpl.Execute(&htmlBuf, data)
    
    // Render text
    textTmpl, _ := textTemplate.New("text").Parse(template.Text)
    var textBuf bytes.Buffer
    textTmpl.Execute(&textBuf, data)
    
    return htmlBuf.String(), textBuf.String(), nil
}
```

### Welcome Email Template

```go
welcomeTemplate := &EmailTemplate{
    Name:    "welcome",
    Subject: "Welcome to {{.AppName}}!",
    HTML: `
<!DOCTYPE html>
<html>
<body>
    <h1>Welcome {{.Name}}!</h1>
    <p>Thank you for signing up for {{.AppName}}.</p>
    <p>Click the button below to verify your email:</p>
    <a href="{{.VerifyURL}}" style="background: #4CAF50; color: white; padding: 10px 20px; text-decoration: none;">
        Verify Email
    </a>
    <p>If you didn't create this account, please ignore this email.</p>
</body>
</html>
    `,
    Text: `
Welcome {{.Name}}!

Thank you for signing up for {{.AppName}}.

Please verify your email by clicking this link:
{{.VerifyURL}}

If you didn't create this account, please ignore this email.
    `,
}

templateRenderer.Register(welcomeTemplate)
```

### Password Reset Template

```go
resetTemplate := &EmailTemplate{
    Name:    "password-reset",
    Subject: "Reset Your Password",
    HTML: `
<!DOCTYPE html>
<html>
<body>
    <h1>Password Reset Request</h1>
    <p>Hi {{.Name}},</p>
    <p>You requested to reset your password. Click the button below:</p>
    <a href="{{.ResetURL}}" style="background: #2196F3; color: white; padding: 10px 20px; text-decoration: none;">
        Reset Password
    </a>
    <p>This link will expire in {{.ExpiresIn}} hours.</p>
    <p>If you didn't request this, please ignore this email.</p>
</body>
</html>
    `,
    Text: `
Password Reset Request

Hi {{.Name}},

You requested to reset your password. Click this link:
{{.ResetURL}}

This link will expire in {{.ExpiresIn}} hours.

If you didn't request this, please ignore this email.
    `,
}
```

### Using Templates

```go
type TemplatedEmailService struct {
    manager  *notification.Manager
    renderer *TemplateRenderer
}

func (s *TemplatedEmailService) SendWelcomeEmail(ctx context.Context, user *User, verifyToken string) error {
    htmlBody, textBody, err := s.renderer.Render("welcome", map[string]interface{}{
        "Name":      user.Name,
        "AppName":   "NeonEx",
        "VerifyURL": fmt.Sprintf("https://example.com/verify?token=%s", verifyToken),
    })
    
    if err != nil {
        return err
    }
    
    return s.manager.Send(ctx, &notification.Notification{
        Channel: notification.ChannelEmail,
        To:      user.Email,
        Subject: "Welcome to NeonEx!",
        Body:    htmlBody,
        Data: map[string]interface{}{
            "html":      true,
            "text_body": textBody,
        },
    })
}

func (s *TemplatedEmailService) SendPasswordReset(ctx context.Context, user *User, resetToken string) error {
    htmlBody, textBody, err := s.renderer.Render("password-reset", map[string]interface{}{
        "Name":      user.Name,
        "ResetURL":  fmt.Sprintf("https://example.com/reset?token=%s", resetToken),
        "ExpiresIn": 24,
    })
    
    if err != nil {
        return err
    }
    
    return s.manager.Send(ctx, &notification.Notification{
        Channel: notification.ChannelEmail,
        To:      user.Email,
        Subject: "Reset Your Password",
        Body:    htmlBody,
        Data: map[string]interface{}{
            "html":      true,
            "text_body": textBody,
        },
    })
}
```

## Attachments

### Single Attachment

```go
// Read file
data, err := os.ReadFile("report.pdf")
if err != nil {
    return err
}

// Send with attachment
err = manager.Send(ctx, &notification.Notification{
    Channel: notification.ChannelEmail,
    To:      "user@example.com",
    Subject: "Monthly Report",
    Body:    "Please find attached your monthly report.",
    Data: map[string]interface{}{
        "attachments": []notification.Attachment{
            {
                Filename:    "report.pdf",
                ContentType: "application/pdf",
                Data:        data,
            },
        },
    },
})
```

### Multiple Attachments

```go
attachments := []notification.Attachment{
    {
        Filename:    "invoice.pdf",
        ContentType: "application/pdf",
        Data:        invoiceData,
    },
    {
        Filename:    "receipt.pdf",
        ContentType: "application/pdf",
        Data:        receiptData,
    },
    {
        Filename:    "logo.png",
        ContentType: "image/png",
        Data:        logoData,
    },
}

err := manager.Send(ctx, &notification.Notification{
    Channel: notification.ChannelEmail,
    To:      "customer@example.com",
    Subject: "Your Order Documents",
    Body:    "Thank you for your order. Please find attached your invoice and receipt.",
    Data: map[string]interface{}{
        "attachments": attachments,
    },
})
```

### Inline Images

```go
htmlBody := `
<!DOCTYPE html>
<html>
<body>
    <h1>Welcome!</h1>
    <img src="cid:logo" alt="Logo" />
    <p>Thank you for joining us!</p>
</body>
</html>
`

logoData, _ := os.ReadFile("logo.png")

err := manager.Send(ctx, &notification.Notification{
    Channel: notification.ChannelEmail,
    To:      "user@example.com",
    Subject: "Welcome!",
    Body:    htmlBody,
    Data: map[string]interface{}{
        "html": true,
        "inline_images": []notification.InlineImage{
            {
                CID:         "logo",
                ContentType: "image/png",
                Data:        logoData,
            },
        },
    },
})
```

## Async Email Sending

### Queue-Based Sending

```go
import "neonex/core/pkg/queue"

type SendEmailJob struct {
    To      string
    Subject string
    Body    string
    manager *notification.Manager
}

func (j *SendEmailJob) Handle(ctx context.Context) error {
    return j.manager.SendEmail(ctx, j.To, j.Subject, j.Body)
}

// Dispatch email jobs
func (s *EmailService) SendAsync(to, subject, body string) error {
    job := &SendEmailJob{
        To:      to,
        Subject: subject,
        Body:    body,
        manager: s.manager,
    }
    
    return s.queue.Dispatch(context.Background(), "emails", job)
}
```

### Background Email Worker

```go
type EmailWorker struct {
    queue   *queue.Manager
    manager *notification.Manager
}

func NewEmailWorker(qm *queue.Manager, nm *notification.Manager) *EmailWorker {
    return &EmailWorker{
        queue:   qm,
        manager: nm,
    }
}

func (ew *EmailWorker) Start(ctx context.Context) {
    worker := ew.queue.Worker("emails", 5) // 5 concurrent email workers
    
    log.Info("Starting email workers")
    
    if err := worker.Start(ctx); err != nil {
        log.Error("Email worker error", logger.Fields{"error": err})
    }
}
```

### Batch Email Processing

```go
type BatchEmailJob struct {
    Recipients []string
    Subject    string
    Body       string
    manager    *notification.Manager
}

func (j *BatchEmailJob) Handle(ctx context.Context) error {
    // Send to all recipients with rate limiting
    rateLimiter := time.NewTicker(100 * time.Millisecond) // 10 emails/second
    defer rateLimiter.Stop()
    
    for _, recipient := range j.Recipients {
        <-rateLimiter.C // Wait for rate limit
        
        err := j.manager.SendEmail(ctx, recipient, j.Subject, j.Body)
        if err != nil {
            log.Error("Failed to send email", logger.Fields{
                "recipient": recipient,
                "error":     err,
            })
            // Continue with other recipients
        }
    }
    
    return nil
}
```

## Multi-Channel Notifications

### SMS Notifications

```go
// Implement SMS sender
type TwilioSender struct {
    accountSID string
    authToken  string
    fromNumber string
}

func (ts *TwilioSender) Send(ctx context.Context, notification *notification.Notification) error {
    // Twilio API integration
    return sendSMS(ts.fromNumber, notification.To, notification.Body)
}

// Register SMS sender
smsSender := &TwilioSender{
    accountSID: "your-twilio-sid",
    authToken:  "your-twilio-token",
    fromNumber: "+1234567890",
}

manager.RegisterSender(notification.ChannelSMS, smsSender)

// Send SMS
manager.SendSMS(ctx, "+1234567890", "Your verification code is: 123456")
```

### Push Notifications

```go
// Implement push sender
type FCMSender struct {
    serverKey string
}

func (fs *FCMSender) Send(ctx context.Context, notification *notification.Notification) error {
    // Firebase Cloud Messaging integration
    return sendPushNotification(notification.To, notification.Subject, notification.Body)
}

// Register push sender
pushSender := &FCMSender{
    serverKey: "your-fcm-server-key",
}

manager.RegisterSender(notification.ChannelPush, pushSender)

// Send push notification
manager.Send(ctx, &notification.Notification{
    Channel: notification.ChannelPush,
    To:      "device-token",
    Subject: "New Message",
    Body:    "You have a new message from John",
})
```

### Multi-Channel Strategy

```go
type NotificationService struct {
    manager *notification.Manager
}

func (ns *NotificationService) NotifyUser(ctx context.Context, userID int, title, message string) error {
    user, err := getUser(userID)
    if err != nil {
        return err
    }
    
    // Send email
    if user.EmailNotifications {
        ns.manager.SendEmail(ctx, user.Email, title, message)
    }
    
    // Send SMS for urgent notifications
    if user.SMSNotifications && isUrgent(title) {
        ns.manager.SendSMS(ctx, user.Phone, message)
    }
    
    // Send push notification
    if user.PushToken != "" {
        ns.manager.Send(ctx, &notification.Notification{
            Channel: notification.ChannelPush,
            To:      user.PushToken,
            Subject: title,
            Body:    message,
        })
    }
    
    return nil
}
```

## Email Verification Flow

### Generate Verification Token

```go
func GenerateVerificationToken() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}
```

### Send Verification Email

```go
func (s *AuthService) SendVerificationEmail(ctx context.Context, user *User) error {
    // Generate token
    token := GenerateVerificationToken()
    
    // Store token in database
    verification := &EmailVerification{
        UserID:    user.ID,
        Token:     token,
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
    s.db.Create(verification)
    
    // Send email
    verifyURL := fmt.Sprintf("https://example.com/verify?token=%s", token)
    
    return s.emailService.Send(ctx, &notification.Notification{
        Channel: notification.ChannelEmail,
        To:      user.Email,
        Subject: "Verify Your Email",
        Body: fmt.Sprintf(`
            <h1>Verify Your Email</h1>
            <p>Click the link below to verify your email address:</p>
            <a href="%s">Verify Email</a>
            <p>This link expires in 24 hours.</p>
        `, verifyURL),
        Data: map[string]interface{}{"html": true},
    })
}
```

### Verify Email Endpoint

```go
func (h *AuthHandler) VerifyEmail(c echo.Context) error {
    token := c.QueryParam("token")
    
    // Find verification record
    var verification EmailVerification
    err := h.db.Where("token = ? AND expires_at > ?", token, time.Now()).
        First(&verification).Error
    
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid or expired token",
        })
    }
    
    // Update user as verified
    h.db.Model(&User{}).
        Where("id = ?", verification.UserID).
        Update("email_verified_at", time.Now())
    
    // Delete verification record
    h.db.Delete(&verification)
    
    return c.JSON(http.StatusOK, map[string]string{
        "message": "Email verified successfully",
    })
}
```

### Resend Verification

```go
func (h *AuthHandler) ResendVerification(c echo.Context) error {
    userID := getUserID(c)
    
    var user User
    if err := h.db.First(&user, userID).Error; err != nil {
        return err
    }
    
    if user.EmailVerifiedAt != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Email already verified",
        })
    }
    
    // Send new verification email
    if err := h.authService.SendVerificationEmail(c.Request().Context(), &user); err != nil {
        return err
    }
    
    return c.JSON(http.StatusOK, map[string]string{
        "message": "Verification email sent",
    })
}
```

## Best Practices

### 1. Rate Limiting

```go
type RateLimitedEmailService struct {
    manager     *notification.Manager
    rateLimiter *time.Ticker
}

func NewRateLimitedEmailService(manager *notification.Manager, ratePerSecond int) *RateLimitedEmailService {
    interval := time.Second / time.Duration(ratePerSecond)
    return &RateLimitedEmailService{
        manager:     manager,
        rateLimiter: time.NewTicker(interval),
    }
}

func (s *RateLimitedEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
    <-s.rateLimiter.C // Wait for rate limit
    return s.manager.SendEmail(ctx, to, subject, body)
}
```

### 2. Error Handling

```go
func (s *EmailService) SendWithRetry(ctx context.Context, to, subject, body string, maxRetries int) error {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        err := s.manager.SendEmail(ctx, to, subject, body)
        if err == nil {
            return nil
        }
        
        lastErr = err
        log.Warn("Email send failed, retrying", logger.Fields{
            "attempt": i + 1,
            "error":   err,
        })
        
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    
    return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
```

### 3. Email Validation

```go
import "net/mail"

func ValidateEmail(email string) bool {
    _, err := mail.ParseAddress(email)
    return err == nil
}

func (s *EmailService) SendEmail(ctx context.Context, to, subject, body string) error {
    if !ValidateEmail(to) {
        return fmt.Errorf("invalid email address: %s", to)
    }
    
    return s.manager.SendEmail(ctx, to, subject, body)
}
```

### 4. Unsubscribe Handling

```go
func (s *EmailService) SendMarketing(ctx context.Context, to, subject, body string) error {
    // Check if user has unsubscribed
    var unsubscribe Unsubscribe
    err := s.db.Where("email = ?", to).First(&unsubscribe).Error
    
    if err == nil {
        log.Info("User unsubscribed, skipping email", logger.Fields{"email": to})
        return nil
    }
    
    // Add unsubscribe link
    unsubscribeToken := generateToken(to)
    unsubscribeURL := fmt.Sprintf("https://example.com/unsubscribe?token=%s", unsubscribeToken)
    
    body += fmt.Sprintf(`
        <hr>
        <p style="font-size: 12px; color: #666;">
            <a href="%s">Unsubscribe</a> from these emails
        </p>
    `, unsubscribeURL)
    
    return s.manager.SendEmail(ctx, to, subject, body)
}
```

### 5. Logging and Monitoring

```go
type LoggingEmailService struct {
    manager *notification.Manager
    logger  logger.Logger
}

func (s *LoggingEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
    start := time.Now()
    
    err := s.manager.SendEmail(ctx, to, subject, body)
    
    duration := time.Since(start)
    
    if err != nil {
        s.logger.Error("Email send failed", logger.Fields{
            "to":       to,
            "subject":  subject,
            "duration": duration,
            "error":    err,
        })
    } else {
        s.logger.Info("Email sent", logger.Fields{
            "to":       to,
            "subject":  subject,
            "duration": duration,
        })
    }
    
    return err
}
```

## Troubleshooting

### SMTP Connection Issues

```go
func TestSMTPConnection(config notification.SMTPConfig) error {
    // Test connection
    conn, err := net.DialTimeout("tcp", 
        fmt.Sprintf("%s:%d", config.Host, config.Port),
        5*time.Second)
    
    if err != nil {
        return fmt.Errorf("cannot connect to SMTP server: %w", err)
    }
    defer conn.Close()
    
    log.Info("SMTP connection successful")
    return nil
}
```

### Email Not Arriving

```go
// Check spam folder
// Verify SPF, DKIM, DMARC records
// Test with mail-tester.com

func (s *EmailService) SendTestEmail() error {
    return s.SendEmail(
        context.Background(),
        "your-email@example.com",
        "Test Email",
        "If you receive this, email is working!",
    )
}
```

### Rate Limit Errors

```go
// Implement exponential backoff
func (s *EmailService) SendWithBackoff(ctx context.Context, to, subject, body string) error {
    maxRetries := 5
    baseDelay := 1 * time.Second
    
    for i := 0; i < maxRetries; i++ {
        err := s.manager.SendEmail(ctx, to, subject, body)
        
        if err == nil {
            return nil
        }
        
        if strings.Contains(err.Error(), "rate limit") {
            delay := baseDelay * time.Duration(math.Pow(2, float64(i)))
            log.Warn("Rate limited, backing off", logger.Fields{
                "delay": delay,
            })
            time.Sleep(delay)
            continue
        }
        
        return err
    }
    
    return fmt.Errorf("failed after %d retries", maxRetries)
}
```

---

**Next Steps:**
- Learn about [Queue System](queue.md) for async email processing
- Explore [Events](events.md) for triggering notifications
- See [Templates](../frontend/templates.md) for email design

**Related Topics:**
- [Authentication](../security/authentication.md)
- [Background Jobs](queue.md)
- [SMS Integration](sms.md)
