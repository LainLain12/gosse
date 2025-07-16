// TwodData represents a row in the twoddata table
package twoddata

type TwodData struct {
	ID       int
	MSet     string
	MValue   string
	MResult  string
	ESet     string
	EValue   string
	EResult  string
	TModern  string
	TInernet string
	NModern  string
	NInernet string
	Date     string // or time.Time if you want to parse
}

// UpdateTime and Status fields removed

const (
	ColID       = "id"
	ColMSet     = "mset"
	ColMValue   = "mvalue"
	ColMResult  = "mresult"
	ColESet     = "eset"
	ColEValue   = "evalue"
	ColEResult  = "eresult"
	ColTModern  = "tmodern"
	ColTInernet = "tinernet"
	ColNModern  = "nmodern"
	ColNInernet = "ninternet"
	ColDate     = "date"
	// ColUpdateTime and ColStatus removed
)
