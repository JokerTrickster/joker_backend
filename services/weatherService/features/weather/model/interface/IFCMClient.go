package _interface

import (
	"context"

	"firebase.google.com/go/v4/messaging"
)

// IFCMClient defines the interface for Firebase Cloud Messaging client
// This interface allows mocking FCM client for testing purposes
type IFCMClient interface {
	// SendMulticast sends a message to multiple devices
	// Returns BatchResponse containing results for each token
	SendMulticast(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error)
}
