package chat

import (
	"sync"
)

var (
	chatMessages []any
	chatMu       sync.Mutex
)

// AddChatMessage adds a message to the chatMessages slice, keeping only the last 50
func AddChatMessage(msg any) {
	chatMu.Lock()
	defer chatMu.Unlock()
	chatMessages = append(chatMessages, msg)
	if len(chatMessages) > 50 {
		chatMessages = chatMessages[len(chatMessages)-50:]
	}
}

// GetLatestChatMessage returns the latest message (or nil if none)
func GetLatestChatMessage() any {
	chatMu.Lock()
	defer chatMu.Unlock()
	if len(chatMessages) == 0 {
		return nil
	}
	return chatMessages[len(chatMessages)-1]
}
