package chat

import (
	"reflect"
	"sync"
)

var (
	chatMessages []any
	chatMu       sync.Mutex
)

// RemoveMessagesByID removes all messages whose map[string]any has id == given id.
// Returns number of removed messages.
func RemoveMessagesByID(id string) int {
	chatMu.Lock()
	defer chatMu.Unlock()
	if id == "" || len(chatMessages) == 0 {
		return 0
	}
	kept := make([]any, 0, len(chatMessages))
	removed := 0
	for _, m := range chatMessages {
		if mm, ok := m.(map[string]any); ok {
			if vid, ok2 := mm["id"].(string); ok2 && vid == id {
				removed++
				continue
			}
		}
		kept = append(kept, m)
	}
	if removed > 0 {
		chatMessages = kept
	}
	return removed
}

// AddChatMessage adds a message to the chatMessages slice, keeping only the last 50.
// Returns true if the message was appended, false if it was considered a duplicate of the last.
func AddChatMessage(msg any) bool {
	chatMu.Lock()
	defer chatMu.Unlock()
	if n := len(chatMessages); n > 0 {
		last := chatMessages[n-1]
		if reflect.DeepEqual(last, msg) { // exact duplicate of last message
			return false
		}
	}
	chatMessages = append(chatMessages, msg)
	if len(chatMessages) > 50 {
		chatMessages = chatMessages[len(chatMessages)-50:]
	}
	return true
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
