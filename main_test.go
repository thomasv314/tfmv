package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(homeDir, "test")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", filepath.Join(getCwd(t), "relative/path")},
	}

	for _, test := range tests {
		result := resolvePath(test.input)
		if result != test.expected {
			t.Errorf("resolvePath(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestRunTerraformMove_DryRun(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")

	// Prepare temp directories
	sourceDir := prepareTempDir(t)
	targetDir := prepareTempDir(t)
	defer os.RemoveAll(sourceDir)
	defer os.RemoveAll(targetDir)

	// Run the command with dry-run enabled
	cmd.Env = append(cmd.Env,
		"SOURCE_DIR="+sourceDir,
		"TARGET_DIR="+targetDir,
		"RESOURCE_ADDR=test-resource",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("cmd.Run() failed with %s\n", err)
	}

	expectedOutput := "[dry-run] enabled. Not actually making moves."
	if !strings.Contains(string(output), expectedOutput) {
		t.Errorf("Expected output %q to contain %q", string(output), expectedOutput)
	}
}

func prepareTempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "terraform-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return dir
}

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	sourceDir := os.Getenv("SOURCE_DIR")
	targetDir := os.Getenv("TARGET_DIR")
	resourceAddr := os.Getenv("RESOURCE_ADDR")

	cmd := exec.Command("go", "run", ".", resourceAddr,
		"--source-dir", sourceDir,
		"--target-dir", targetDir,
		"--dry-run")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func getCwd(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	return cwd
}
