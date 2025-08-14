package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/user/go-git-test/pkg/api"
	"github.com/user/go-git-test/pkg/git"
	"github.com/user/go-git-test/pkg/logger"
	"github.com/user/go-git-test/pkg/provider"
	"go.uber.org/zap"
)

// Version information
var (
	version = "1.0.0"
)

// Command line flags
var (
	destination  = flag.String("dest", ".", "Destination directory")
	branch       = flag.String("branch", "", "Branch to clone")
	username     = flag.String("username", "", "Username for authentication")
	password     = flag.String("password", "", "Password for authentication")
	token        = flag.String("token", "", "Token for authentication")
	sshKeyPath   = flag.String("ssh-key", "", "Path to SSH key")
	command      = flag.String("command", "clone", "Command to execute (clone, diff, prdiff, prcomment, prcomments, serve)")
	oldCommit    = flag.String("old-commit", "", "Old commit for diff")
	newCommit    = flag.String("new-commit", "", "New commit for diff")
	prURL        = flag.String("pr-url", "", "PR/MR URL for prdiff and prcomment commands")
	filePath     = flag.String("file-path", "", "File path for prcomment command")
	lineNumber   = flag.Int("line-number", 0, "Line number for prcomment command")
	comment      = flag.String("comment", "", "Comment text for prcomment command")
	parentID     = flag.Int64("parent-id", 0, "Parent comment ID for prcomments command")
	debug        = flag.Bool("debug", false, "Enable debug logging")
	port         = flag.Int("port", 8080, "Port for API server (when using serve command)")
)

func main() {
	flag.Parse()

	// Initialize logger
	log, err := logger.NewLogger(*debug)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Create provider factory
	factory := provider.NewFactory(log)

	// Execute command
	switch *command {
	case "serve":
		// Start API server
		runAPIServer(log, factory)

	case "clone", "diff", "prdiff", "prcomment", "prcomments":
		runGitCommand(log, factory)

	default:
		log.Fatal("Unsupported command", zap.String("command", *command))
	}
}

// runAPIServer starts the API server
func runAPIServer(log *logger.Logger, factory *provider.Factory) {
	// Create API server config
	config := api.DefaultConfig()
	config.Port = *port
	config.Version = version

	// Create and start API server
	server := api.NewServer(config, log, factory)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("Failed to start API server", zap.Error(err))
		}
	}()

	log.Info("API server started", zap.Int("port", config.Port))

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Create a deadline for server shutdown
	_, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := server.Stop(); err != nil {
		log.Error("Failed to stop API server", zap.Error(err))
	}

	log.Info("API server stopped")
}

// runGitCommand runs Git commands (clone, diff, prdiff, prcomment, prcomments)
func runGitCommand(log *logger.Logger, factory *provider.Factory) {
	// Determine provider type
	var pt provider.ProviderType
	switch strings.ToLower(*providerType) {
	case "github":
		pt = provider.GitHub
	case "gitlab":
		pt = provider.GitLab
	case "bitbucket":
		pt = provider.Bitbucket
	default:
		log.Fatal("Unsupported provider type", zap.String("provider", *providerType))
	}

	// Create provider
	p, err := factory.CreateProvider(pt, *username, *password, *token, *sshKeyPath)
	if err != nil {
		log.Fatal("Failed to create provider", zap.Error(err))
	}

	// Create context
	ctx := context.Background()

	// Execute command
	switch *command {
	case "clone":
		err = p.Clone(*repoURL, *destination, *branch)
		if err != nil {
			log.Fatal("Failed to clone repository", zap.Error(err))
		}
		log.Info("Repository cloned successfully", zap.String("destination", *destination))

	case "diff":
		// First, we need to open the repository
		err = p.GetClient().OpenRepository(*destination)
		if err != nil {
			log.Fatal("Failed to open repository", zap.Error(err))
		}

		// Get diff
		diffs, err := p.GetClient().Diff(*oldCommit, *newCommit)
		if err != nil {
			log.Fatal("Failed to get diff", zap.Error(err))
		}

		// Print diff
		fmt.Printf("Diff between %s and %s:\n\n", *oldCommit, *newCommit)
		for _, diff := range diffs {
			fmt.Printf("File: %s\n", diff.FilePath)
			if diff.IsBinary {
				fmt.Println("Binary file")
			} else {
				fmt.Printf("Changes: +%d -%d\n", countAddedLines(diff), countDeletedLines(diff))
				for _, chunk := range diff.Chunks {
					fmt.Print(chunk.Content)
				}
			}
			fmt.Println()
		}

	case "prdiff":
		// Get PR diff
		diff, err := p.GetPRDiff(ctx, *prURL)
		if err != nil {
			log.Fatal("Failed to get PR diff", zap.Error(err))
		}

		// Print diff
		fmt.Printf("Diff for PR: %s\n\n", *prURL)
		fmt.Println(diff)

	case "prcomment":
		// Add comment to PR
		err = p.AddPRComment(ctx, *prURL, *filePath, *lineNumber, *comment)
		if err != nil {
			log.Fatal("Failed to add comment to PR", zap.Error(err))
		}
		log.Info("Comment added successfully to PR",
			zap.String("prURL", *prURL),
			zap.String("filePath", *filePath),
			zap.Int("lineNumber", *lineNumber))
			
	case "prcomments":
		// Only GitHub supports this feature currently
		if strings.ToLower(*providerType) != "github" {
			log.Fatal("This feature is currently only supported for GitHub")
		}
		
		// Get GitHub provider
		githubProvider, ok := p.(*provider.GitHubProvider)
		if !ok {
			log.Fatal("Failed to cast to GitHub provider")
		}
		
		// Get comments by parent ID
		comments, err := githubProvider.GetPRCommentsByParentID(ctx, *prURL, *parentID)
		if err != nil {
			log.Fatal("Failed to get PR comments by parent ID", zap.Error(err))
		}
		
		// Print comments
		fmt.Printf("Comments for PR: %s with parent ID: %d\n\n", *prURL, *parentID)
		fmt.Printf("Found %d comment(s)\n\n", len(comments))
		
		for i, comment := range comments {
			fmt.Printf("Comment #%d:\n", i+1)
			fmt.Printf("ID: %d\n", comment.GetID())
			fmt.Printf("User: %s\n", comment.GetUser().GetLogin())
			fmt.Printf("Created: %s\n", comment.GetCreatedAt().Format("2006-01-02 15:04:05"))
			fmt.Printf("Updated: %s\n", comment.GetUpdatedAt().Format("2006-01-02 15:04:05"))
			if comment.GetPath() != "" {
				fmt.Printf("File: %s\n", comment.GetPath())
				fmt.Printf("Position: %d\n", comment.GetPosition())
			}
			fmt.Printf("URL: %s\n", comment.GetHTMLURL())
			fmt.Printf("Comment: %s\n\n", comment.GetBody())
		}
	}
}

// validateFlags validates that the required flags are present based on the command
func validateFlags(log *logger.Logger) {
	switch *command {
	case "clone":
		if *repoURL == "" {
			log.Fatal("Repository URL is required for clone command")
		}

	case "diff":
		if *oldCommit == "" || *newCommit == "" {
			log.Fatal("Both old-commit and new-commit are required for diff command")
		}

	case "prdiff":
		if *prURL == "" {
			log.Fatal("PR URL is required for prdiff command")
		}

	case "prcomment":
		if *prURL == "" {
			log.Fatal("PR URL is required for prcomment command")
		}
		if *filePath == "" {
			log.Fatal("File path is required for prcomment command")
		}
		if *lineNumber == 0 {
			log.Fatal("Line number is required for prcomment command")
		}
		if *comment == "" {
			log.Fatal("Comment text is required for prcomment command")
		}
		
	case "prcomments":
		if *prURL == "" {
			log.Fatal("PR URL is required for prcomments command")
		}
		if *parentID == 0 {
			log.Fatal("Parent comment ID is required for prcomments command")
		}

	case "serve":
		// No validation needed for serve command
	}
}

// countAddedLines counts the total added lines in a diff
func countAddedLines(diff git.DiffResult) int {
	total := 0
	for _, chunk := range diff.Chunks {
		total += chunk.AddedLines
	}
	return total
}

// countDeletedLines counts the total deleted lines in a diff
func countDeletedLines(diff git.DiffResult) int {
	total := 0
	for _, chunk := range diff.Chunks {
		total += chunk.DeletedLines
	}
	return total
}
