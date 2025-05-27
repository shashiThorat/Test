package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/user/go-git-test/pkg/logger"
	"go.uber.org/zap"
	"strings"
	"time"
)

// DiffResult represents the result of a diff operation
type DiffResult struct {
	FilePath    string
	OldFileMode string
	NewFileMode string
	OldSHA      string
	NewSHA      string
	IsBinary    bool
	Chunks      []DiffChunk
}

// DiffChunk represents a chunk in a diff
type DiffChunk struct {
	Content      string
	AddedLines   int
	DeletedLines int
}

// Repository represents a Git repository
type Repository interface {
	// Clone clones a repository from the given URL
	Clone(url, destination string, options CloneOptions) error
	
	// Diff shows the differences between two commits
	Diff(oldCommit, newCommit string) ([]DiffResult, error)
	
	// GetCommitInfo returns information about a commit
	GetCommitInfo(commitID string) (*CommitInfo, error)
	
	// ListBranches lists all branches in the repository
	ListBranches() ([]string, error)
}

// CommitInfo represents information about a commit
type CommitInfo struct {
	Hash      string
	Author    string
	Email     string
	Message   string
	Timestamp time.Time
}

// CloneOptions represents options for cloning a repository
type CloneOptions struct {
	Branch            string
	Depth             int
	Auth              interface{}
	RecurseSubmodules bool
}

// PRComment represents a comment on a pull request
type PRComment struct {
	ID        string
	Body      string
	FilePath  string
	LineNum   int
	CommentID string
	Author    string
	CreatedAt time.Time
}

// Client is the base Git client implementation
type Client struct {
	repo   *git.Repository
	logger *logger.Logger
}

// NewClient creates a new Git client
func NewClient(logger *logger.Logger) *Client {
	return &Client{
		logger: logger,
	}
}

// Clone clones a repository from the given URL
func (c *Client) Clone(url, destination string, options CloneOptions) error {
	c.logger.Info("Cloning repository",
		zap.String("url", url),
		zap.String("destination", destination),
		zap.String("branch", options.Branch))

	// Convert to go-git CloneOptions
	cloneOptions := &git.CloneOptions{
		URL:               url,
		Progress:          nil,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}

	// Set branch if specified
	if options.Branch != "" {
		cloneOptions.ReferenceName = plumbing.NewBranchReferenceName(options.Branch)
		cloneOptions.SingleBranch = true
	}

	// Set depth if specified
	if options.Depth > 0 {
		cloneOptions.Depth = options.Depth
	}

	// Set auth if provided
	if options.Auth != nil {
		cloneOptions.Auth = options.Auth.(transport.AuthMethod)
	}

	// Set submodule recursion
	if options.RecurseSubmodules {
		cloneOptions.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	}

	repo, err := git.PlainClone(destination, false, cloneOptions)
	if err != nil {
		c.logger.Error("Failed to clone repository", zap.Error(err))
		return err
	}

	c.repo = repo
	c.logger.Info("Repository cloned successfully")
	return nil
}

// OpenRepository opens an existing repository
func (c *Client) OpenRepository(path string) error {
	c.logger.Info("Opening repository", zap.String("path", path))
	
	repo, err := git.PlainOpen(path)
	if err != nil {
		c.logger.Error("Failed to open repository", zap.Error(err))
		return err
	}
	
	c.repo = repo
	c.logger.Info("Repository opened successfully")
	return nil
}

// GetCommitInfo returns information about a commit
func (c *Client) GetCommitInfo(commitID string) (*CommitInfo, error) {
	c.logger.Info("Getting commit info", zap.String("commitID", commitID))
	
	hash := plumbing.NewHash(commitID)
	commit, err := c.repo.CommitObject(hash)
	if err != nil {
		c.logger.Error("Failed to get commit", zap.Error(err))
		return nil, err
	}
	
	return &CommitInfo{
		Hash:      commit.Hash.String(),
		Author:    commit.Author.Name,
		Email:     commit.Author.Email,
		Message:   commit.Message,
		Timestamp: commit.Author.When,
	}, nil
}

// ListBranches lists all branches in the repository
func (c *Client) ListBranches() ([]string, error) {
	c.logger.Info("Listing branches")
	
	branchIter, err := c.repo.Branches()
	if err != nil {
		c.logger.Error("Failed to get branches", zap.Error(err))
		return nil, err
	}
	
	var branches []string
	err = branchIter.ForEach(func(ref *plumbing.Reference) error {
		branches = append(branches, ref.Name().Short())
		return nil
	})
	
	if err != nil {
		c.logger.Error("Failed to iterate branches", zap.Error(err))
		return nil, err
	}
	
	return branches, nil
}

// Diff shows the differences between two commits
func (c *Client) Diff(oldCommit, newCommit string) ([]DiffResult, error) {
	c.logger.Info("Getting diff",
		zap.String("oldCommit", oldCommit),
		zap.String("newCommit", newCommit))
	
	oldHash := plumbing.NewHash(oldCommit)
	newHash := plumbing.NewHash(newCommit)
	
	oldCommitObj, err := c.repo.CommitObject(oldHash)
	if err != nil {
		c.logger.Error("Failed to get old commit", zap.Error(err))
		return nil, err
	}
	
	newCommitObj, err := c.repo.CommitObject(newHash)
	if err != nil {
		c.logger.Error("Failed to get new commit", zap.Error(err))
		return nil, err
	}
	
	oldTree, err := oldCommitObj.Tree()
	if err != nil {
		c.logger.Error("Failed to get old tree", zap.Error(err))
		return nil, err
	}
	
	newTree, err := newCommitObj.Tree()
	if err != nil {
		c.logger.Error("Failed to get new tree", zap.Error(err))
		return nil, err
	}
	
	changes, err := oldTree.Diff(newTree)
	if err != nil {
		c.logger.Error("Failed to get diff", zap.Error(err))
		return nil, err
	}
	
	var results []DiffResult
	
	for _, change := range changes {
		patch, err := change.Patch()
		if err != nil {
			c.logger.Error("Failed to get patch", zap.Error(err))
			continue
		}
		
		// Get file path
		var filePath string
		action, err := change.Action()
		if err != nil {
			c.logger.Error("Failed to get action", zap.Error(err))
			continue
		}
		
		switch action {
		case merkletrie.Delete:
			filePath = change.From.Name
		case merkletrie.Insert:
			filePath = change.To.Name
		default:
			filePath = change.To.Name
		}
		
		// Create diff result
		diffResult := DiffResult{
			FilePath: filePath,
			// go-git doesn't provide a direct way to check if a file is binary
			IsBinary: false,
		}
		
		// Set file modes and SHAs
		if change.From.Name != "" {
			diffResult.OldFileMode = change.From.TreeEntry.Mode.String()
			diffResult.OldSHA = change.From.TreeEntry.Hash.String()
		}
		
		if change.To.Name != "" {
			diffResult.NewFileMode = change.To.TreeEntry.Mode.String()
			diffResult.NewSHA = change.To.TreeEntry.Hash.String()
		}
		
		// Process chunks
		for _, filePatch := range patch.FilePatches() {
			for _, chunk := range filePatch.Chunks() {
				diffChunk := DiffChunk{
					Content: chunk.Content(),
				}
				
				switch chunk.Type() {
				case diff.Add:
					diffChunk.AddedLines = len(strings.Split(chunk.Content(), "\n"))-1
				case diff.Delete:
					diffChunk.DeletedLines = len(strings.Split(chunk.Content(), "\n"))-1
				}
				
				diffResult.Chunks = append(diffResult.Chunks, diffChunk)
			}
		}
		
		results = append(results, diffResult)
	}
	
	return results, nil
}

// GetDiffAsString returns the diff as a unified diff string
func (c *Client) GetDiffAsString(oldCommit, newCommit string) (string, error) {
	c.logger.Info("Getting diff as string",
		zap.String("oldCommit", oldCommit),
		zap.String("newCommit", newCommit))
	
	oldHash := plumbing.NewHash(oldCommit)
	newHash := plumbing.NewHash(newCommit)
	
	oldCommitObj, err := c.repo.CommitObject(oldHash)
	if err != nil {
		c.logger.Error("Failed to get old commit", zap.Error(err))
		return "", err
	}
	
	newCommitObj, err := c.repo.CommitObject(newHash)
	if err != nil {
		c.logger.Error("Failed to get new commit", zap.Error(err))
		return "", err
	}
	
	patch, err := oldCommitObj.Patch(newCommitObj)
	if err != nil {
		c.logger.Error("Failed to get patch", zap.Error(err))
		return "", err
	}
	
	return patch.String(), nil
}

// GetFileAtCommit retrieves the content of a file at a specific commit
func (c *Client) GetFileAtCommit(commit, filePath string) (string, error) {
	c.logger.Info("Getting file at commit", 
		zap.String("commit", commit), 
		zap.String("filePath", filePath))
	
	hash := plumbing.NewHash(commit)
	commitObj, err := c.repo.CommitObject(hash)
	if err != nil {
		c.logger.Error("Failed to get commit", zap.Error(err))
		return "", err
	}
	
	tree, err := commitObj.Tree()
	if err != nil {
		c.logger.Error("Failed to get tree", zap.Error(err))
		return "", err
	}
	
	file, err := tree.File(filePath)
	if err != nil {
		c.logger.Error("Failed to get file", zap.Error(err))
		return "", err
	}
	
	content, err := file.Contents()
	if err != nil {
		c.logger.Error("Failed to get file contents", zap.Error(err))
		return "", err
	}
	
	return content, nil
}

// FormatPRDiff formats the PR diff in a human-readable format
func (c *Client) FormatPRDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	var formatted []string
	
	for _, line := range lines {
		if strings.HasPrefix(line, "+") {
			formatted = append(formatted, fmt.Sprintf("\033[32m%s\033[0m", line))
		} else if strings.HasPrefix(line, "-") {
			formatted = append(formatted, fmt.Sprintf("\033[31m%s\033[0m", line))
		} else if strings.HasPrefix(line, "@@") {
			formatted = append(formatted, fmt.Sprintf("\033[36m%s\033[0m", line))
		} else if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "index ") || 
			strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") {
			formatted = append(formatted, fmt.Sprintf("\033[1m%s\033[0m", line))
		} else {
			formatted = append(formatted, line)
		}
	}
	
	return strings.Join(formatted, "\n")
}