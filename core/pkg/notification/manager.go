package notification

import (
	"context"
	"fmt"
)

// Channel represents notification channel
type Channel string

const (
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"
	ChannelPush  Channel = "push"
)

// Notification represents a notification
type Notification struct {
	Channel Channel
	To      string
	Subject string
	Body    string
	Data    map[string]interface{}
}

// Sender interface for notification senders
type Sender interface {
	Send(ctx context.Context, notification *Notification) error
}

// Manager manages notifications
type Manager struct {
	senders map[Channel]Sender
}

// NewManager creates a new notification manager
func NewManager() *Manager {
	return &Manager{
		senders: make(map[Channel]Sender),
	}
}

// RegisterSender registers a sender for a channel
func (m *Manager) RegisterSender(channel Channel, sender Sender) {
	m.senders[channel] = sender
}

// Send sends a notification
func (m *Manager) Send(ctx context.Context, notification *Notification) error {
	sender, ok := m.senders[notification.Channel]
	if !ok {
		return fmt.Errorf("no sender registered for channel: %s", notification.Channel)
	}

	return sender.Send(ctx, notification)
}

// SendEmail sends an email notification
func (m *Manager) SendEmail(ctx context.Context, to, subject, body string) error {
	return m.Send(ctx, &Notification{
		Channel: ChannelEmail,
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

// SendSMS sends an SMS notification
func (m *Manager) SendSMS(ctx context.Context, to, body string) error {
	return m.Send(ctx, &Notification{
		Channel: ChannelSMS,
		To:      to,
		Body:    body,
	})
}
