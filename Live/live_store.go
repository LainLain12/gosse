package Live

import (
	"sync"
)

var (
	liveDataStore []Live
	liveDataMu    sync.Mutex
)
