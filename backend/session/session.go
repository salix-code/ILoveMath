package session

import "sync"

// Config holds per-session state.
type Config struct {
	ID            string
	ProblemID     int
	Difficulty    int
	OrderMode     string // "sequential" or "random" (default "random")
	QuestionIndex int    // used when OrderMode == "sequential"
	Score         int    // number of correctly answered questions
	Total         int    // number of questions submitted
	CurrentGUID   string // GUID of the question currently being answered
	CurrentAnswer string // expected answer for validation on next request (stored as JSON string)
}

var store sync.Map

// GetOrCreate returns the existing session for id, or creates a new one.
// If id is empty a new Config is created without storing it — the caller
// must supply a freshly generated ID via the second return value.
func GetOrCreate(id string) *Config {
	if v, ok := store.Load(id); ok {
		return v.(*Config)
	}
	cfg := &Config{ID: id}
	store.Store(id, cfg)
	return cfg
}

// Get returns the session for id, if it exists.
func Get(id string) (*Config, bool) {
	v, ok := store.Load(id)
	if !ok {
		return nil, false
	}
	return v.(*Config), true
}
