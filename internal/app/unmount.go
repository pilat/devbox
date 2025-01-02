package app

import "fmt"

func (a *app) Unmount(sourceName string) error {
	if sourceName == "" {
		_, s, err := a.autodetect()
		if err != nil {
			return fmt.Errorf("failed to autodetect source name: %w", err)
		} else {
			sourceName = s
		}
	}

	if _, ok := a.state.Mounts[sourceName]; !ok {
		return fmt.Errorf("source %s is not mounted", sourceName)
	}

	delete(a.state.Mounts, sourceName)

	if err := a.state.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return a.Info()
}
