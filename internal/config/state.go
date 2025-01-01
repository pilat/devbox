package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type state struct {
	filename string

	Version   int               `json:"version"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Mounts    map[string]string `json:"mounts"`
}

func (s *state) Save() error {
	json, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	err = os.WriteFile(s.filename, json, 0644)
	if err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}
