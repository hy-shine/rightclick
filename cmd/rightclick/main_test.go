package main

import "testing"

func TestShouldOpenWindow(t *testing.T) {
	tests := []struct {
		name           string
		openedAtLogin  bool
		launchedForURL bool
		want           bool
	}{
		{"manual start", false, false, true},
		{"started at login", true, false, false},
		{"started by a link", false, true, false},
		{"login item asked to open a link", true, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldOpenWindow(tt.openedAtLogin, tt.launchedForURL); got != tt.want {
				t.Errorf("shouldOpenWindow(%v, %v) = %v, want %v",
					tt.openedAtLogin, tt.launchedForURL, got, tt.want)
			}
		})
	}
}
