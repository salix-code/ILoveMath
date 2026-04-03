package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ProblemItem represents a single question within a problem set.
type ProblemItem struct {
	Difficulty int               `json:"Difficulty"`
	Question   string            `json:"Question"`
	Answer     string            `json:"Answer"`
	Input      map[string]string `json:"Input"` // values are expressions: literals or random(min,max)
	Tips       string            `json:"Tips"`
}

// ProblemConfig represents a full problem-set JSON file.
type ProblemConfig struct {
	ID    int           `json:"ID"`
	Title string        `json:"Title"`
	Items []ProblemItem `json:"Items"`
}

// ProblemTypes holds all loaded problem sets, indexed by ID.
var ProblemTypes = map[int]ProblemConfig{}

// LoadAll reads every *.json file in dir and populates ProblemTypes.
func LoadAll(dir string) error {
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return fmt.Errorf("config glob: %w", err)
	}
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		var cfg ProblemConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		ProblemTypes[cfg.ID] = cfg
	}
	return nil
}
