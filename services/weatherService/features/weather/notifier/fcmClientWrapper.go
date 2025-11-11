package notifier

import (
	"context"

	"firebase.google.com/go/v4/messaging"
)

// FCMClientWrapper wraps Firebase messaging.Client to implement IFCMClient interface
type FCMClientWrapper struct {
	client *messaging.Client
}

// NewFCMClientWrapper creates a new FCM client wrapper
func NewFCMClientWrapper(client *messaging.Client) *FCMClientWrapper {
	return &FCMClientWrapper{client: client}
}

// SendMulticast sends a message to multiple devices
func (w *FCMClientWrapper) SendMulticast(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	return w.client.SendMulticast(ctx, message)
}
