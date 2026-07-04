package main

import (
	"slices"
	"testing"
)

func TestSplitScenarioArgs(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		wantScenario    string
		wantPassthrough []string
	}{
		{name: "flag only", args: []string{"e2e", "--tag", "foo"}, wantScenario: "e2e", wantPassthrough: []string{"--tag", "foo"}},
		{name: "separator stripped", args: []string{"e2e", "--", "--tag", "foo"}, wantScenario: "e2e", wantPassthrough: []string{"--tag", "foo"}},
		{name: "no args", args: []string{"e2e"}, wantScenario: "e2e", wantPassthrough: nil},
		{name: "bare separator", args: []string{"e2e", "--"}, wantScenario: "e2e", wantPassthrough: nil},
		{name: "double separator keeps second", args: []string{"e2e", "--", "--"}, wantScenario: "e2e", wantPassthrough: []string{"--"}},
		{name: "later separator untouched", args: []string{"e2e", "--tag", "--", "foo"}, wantScenario: "e2e", wantPassthrough: []string{"--tag", "--", "foo"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario, passthrough := splitScenarioArgs(tt.args)
			if scenario != tt.wantScenario {
				t.Errorf("scenario = %q, want %q", scenario, tt.wantScenario)
			}
			if !slices.Equal(passthrough, tt.wantPassthrough) {
				t.Errorf("passthrough = %v, want %v", passthrough, tt.wantPassthrough)
			}
		})
	}
}

// TestRunCmdArgsContract guards the real command's validator + SetInterspersed(false)
// against a revert to ExactArgs, which would break scenario argument pass-through.
func TestRunCmdArgsContract(t *testing.T) {
	tests := []struct {
		name     string
		argv     []string
		wantArgs []string
		wantErr  bool
	}{
		{name: "flag only", argv: []string{"e2e", "--tag", "foo"}, wantArgs: []string{"e2e", "--tag", "foo"}},
		{name: "explicit separator", argv: []string{"e2e", "--", "--tag", "foo"}, wantArgs: []string{"e2e", "--", "--tag", "foo"}},
		{name: "scenario only", argv: []string{"e2e"}, wantArgs: []string{"e2e"}},
		{name: "no scenario", argv: []string{}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRunCmd()
			if err := cmd.ParseFlags(tt.argv); err != nil {
				t.Fatalf("ParseFlags(%v) failed: %v", tt.argv, err)
			}

			got := cmd.Flags().Args()
			err := cmd.ValidateArgs(got)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("ValidateArgs(%v) = nil, want error", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("ValidateArgs(%v) failed: %v", got, err)
			}
			if !slices.Equal(got, tt.wantArgs) {
				t.Errorf("Flags().Args() = %v, want %v", got, tt.wantArgs)
			}
		})
	}
}
