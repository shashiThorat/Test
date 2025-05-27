package git

import (
	"github.com/user/go-git-test/pkg/logger"
	"go.uber.org/zap"
	"os"
	"testing"
)

func TestClient_Clone(t *testing.T) {
	// Skip test if not running in CI environment
	if os.Getenv("CI") == "" {
		t.Skip("Skipping test in local environment")
	}
	
	// Create logger
	log, err := logger.NewLogger(true)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	
	// Create client
	client := NewClient(log)
	
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Clone repository
	err = client.Clone("https://github.com/go-git/go-git", tempDir, CloneOptions{
		Branch: "master",
	})
	
	if err != nil {
		t.Fatalf("Failed to clone repository: %v", err)
	}
	
	// Verify repository was cloned
	_, err = os.Stat(tempDir + "/.git")
	if err != nil {
		t.Fatalf("Repository was not cloned correctly: %v", err)
	}
}

func TestClient_OpenRepository(t *testing.T) {
	// Skip test if not running in CI environment
	if os.Getenv("CI") == "" {
		t.Skip("Skipping test in local environment")
	}
	
	// Create logger
	log, err := logger.NewLogger(true)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	
	// Create client
	client := NewClient(log)
	
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Clone repository
	err = client.Clone("https://github.com/go-git/go-git", tempDir, CloneOptions{
		Branch: "master",
	})
	
	if err != nil {
		t.Fatalf("Failed to clone repository: %v", err)
	}
	
	// Create new client
	client2 := NewClient(log)
	
	// Open repository
	err = client2.OpenRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}
	
	// Verify branches can be listed
	branches, err := client2.ListBranches()
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}
	
	if len(branches) == 0 {
		t.Fatal("No branches found")
	}
	
	log.Info("Branches found", zap.Strings("branches", branches))
}