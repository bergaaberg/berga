package cmd

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Test that root command initializes without error
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}
	
	if rootCmd.Use != "berga" {
		t.Errorf("Expected rootCmd.Use to be 'berga', got %s", rootCmd.Use)
	}
}

func TestGetConfigDir(t *testing.T) {
	// Test that GetConfigDir returns a non-empty string
	configDir := GetConfigDir()
	if configDir == "" {
		t.Error("GetConfigDir should return a non-empty string")
	}
}

func TestGetScriptsDir(t *testing.T) {
	// Test that GetScriptsDir returns a non-empty string
	scriptsDir := GetScriptsDir()
	if scriptsDir == "" {
		t.Error("GetScriptsDir should return a non-empty string")
	}
}

func TestGetTemplatesDir(t *testing.T) {
	// Test that GetTemplatesDir returns a non-empty string
	templatesDir := GetTemplatesDir()
	if templatesDir == "" {
		t.Error("GetTemplatesDir should return a non-empty string")
	}
}
