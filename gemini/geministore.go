package gemini

import (
	"gosse/Live"
	"sync"
)

type (
	LiveDatastore []Live.Live
	liveDataMu    sync.Mutex
)
