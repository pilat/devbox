package state

import (
	"encoding/json"
	"fmt"
	"os"
)

type State struct {
	filename string

	// Version   int               `json:"version"`
	// CreatedAt time.Time         `json:"created_at"`
	// UpdatedAt time.Time         `json:"updated_at"`
	Mounts map[string]string `json:"mounts"`
}

func New(filename string) (*State, error) {
	state := &State{
		filename: filename,
		Mounts:   make(map[string]string),
	}

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return state, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get state file: %w", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(content, state)
	if err != nil {
		return nil, err
	}

	if state.Mounts == nil {
		state.Mounts = make(map[string]string)
	}

	return state, nil
}

func (s *State) Save() error {
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
