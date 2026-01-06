package service

import (
	"assistant-sf/internal/service"
	"testing"
)

func TestValidateSyncPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		ok   bool
	}{
		// Linux
		{"linux fs root", "/", false},
		{"linux home parent", "/home", false},
		{"linux home", "/home/user", false},
		{"linux home subdir", "/home/user/folder", true},

		// macOS
		{"darwin fs root", "/", false},
		{"darwin home parent", "/Users", false},
		{"darwin home", "/Users/user", false},
		{"darwin home subdir", "/Users/user/folder", true},

		// Windows
		{"windows fs root", `C:\`, false},
		{"windows home parent", `C:\Users`, false},
		{"windows home", `C:\Users\user`, false},
		{"windows home subdir", `C:\Users\user\folder`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateSyncPath(tt.path)

			if tt.ok && err != nil {
				t.Fatalf("expected OK, got error: %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("expected error, got OK")
			}
		})
	}
}
