package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/ktrysmt/go-bitbucket"
	"github.com/user/go-git-test/pkg/git"
	"github.com/user/go-git-test/pkg/logger"
	"go.uber.org/zap"
)

// BitbucketProvider is a provider for Bitbucket repositories
type BitbucketProvider struct {
	*BaseProvider
	bitbucketClient *bitbucket.Client
}

// NewBitbucketProvider creates a new Bitbucket provider
func NewBitbucketProvider(logger *logger.Logger, username, password, token, sshKeyPath string) *BitbucketProvider {
	base := NewBaseProvider(logger, username, password, token, sshKeyPath)
	
	var bitbucketClient *bitbucket.Client
	
	// Create Bitbucket client with appropriate authentication
	if username != "" && password != "" {
		bitbucketClient = bitbucket.NewBasicAuth(username, password)
	} else if token != "" {
		// Bitbucket uses app passwords, not tokens, but we'll use the token field for compatibility
		bitbucketClient = bitbucket.NewBasicAuth(username, token)
	} else {
		// Anonymous client
		bitbucketClient = &bitbucket.Client{}
	}
	
	return &BitbucketProvider{
		BaseProvider:    base,
		bitbucketClient: bitbucketClient,
	}
}

// Clone clones a repository from Bitbucket
func (p *BitbucketProvider) Clone(repoURL, destination, branch string) error {
	p.logger.Info("Cloning Bitbucket repository",
		zap.String("url", repoURL),
		zap.String("destination", destination),
		zap.String("branch", branch))
	
	auth, err := p.GetAuthMethod()
	if err != nil {
		return err
	}
	
	options := git.CloneOptions{
		Branch:            branch,
		Auth:              auth,
		RecurseSubmodules: true,
	}
	
	return p.client.Clone(repoURL, destination, options)
}

// GetPRDiff gets the diff for a pull request
func (p *BitbucketProvider) GetPRDiff(ctx context.Context, prURL string) (string, error) {
	p.logger.Info("Getting Bitbucket PR diff", zap.String("prURL", prURL))
	
	owner, repo, prNumStr, err := ParsePRURL(prURL)
	if err != nil {
		p.logger.Error("Failed to parse PR URL", zap.Error(err))
		return "", err
	}
	
	prNum, err := strconv.Atoi(prNumStr)
	if err != nil {
		p.logger.Error("Invalid PR number", zap.Error(err))
		return "", err
	}
	
	// Bitbucket API endpoint for getting a PR diff
	diffURL := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/%d/diff", 
		owner, repo, prNum)
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", diffURL, nil)
	if err != nil {
		p.logger.Error("Failed to create request", zap.Error(err))
		return "", err
	}
	
	// Add authentication if available
	if p.username != "" && (p.password != "" || p.token != "") {
		password := p.password
		if password == "" {
			password = p.token
		}
		req.SetBasicAuth(p.username, password)
	}
	
	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		p.logger.Error("Failed to execute request", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		p.logger.Error("Bitbucket API returned non-200 status", 
			zap.Int("statusCode", resp.StatusCode))
		return "", fmt.Errorf("bitbucket API returned status %d", resp.StatusCode)
	}
	
	// Read the diff content
	diffBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.Error("Failed to read response body", zap.Error(err))
		return "", err
	}
	
	p.logger.Info("Successfully retrieved PR diff",
		zap.String("owner", owner),
		zap.String("repo", repo),
		zap.Int("prNum", prNum))
	
	return string(diffBytes), nil
}

// AddPRComment adds a comment to a pull request
func (p *BitbucketProvider) AddPRComment(ctx context.Context, prURL string, filePath string, lineNumber int, comment string) error {
	p.logger.Info("Adding comment to Bitbucket PR",
		zap.String("prURL", prURL),
		zap.String("filePath", filePath),
		zap.Int("lineNumber", lineNumber))
	
	owner, repo, prNumStr, err := ParsePRURL(prURL)
	if err != nil {
		p.logger.Error("Failed to parse PR URL", zap.Error(err))
		return err
	}
	
	prNum, err := strconv.Atoi(prNumStr)
	if err != nil {
		p.logger.Error("Invalid PR number", zap.Error(err))
		return err
	}
	
	// Get the PR to determine the latest commit
	opt := &bitbucket.PullRequestsOptions{
		Owner:    owner,
		RepoSlug: repo,
		ID:       prNumStr,
	}
	
	pr, err := p.bitbucketClient.Repositories.PullRequests.Get(opt)
	if err != nil {
		p.logger.Error("Failed to get PR", zap.Error(err))
		return err
	}
	
	// Extract the commit hash from the PR
	var prData map[string]interface{}
	err = json.Unmarshal(pr.([]byte), &prData)
	if err != nil {
		p.logger.Error("Failed to parse PR data", zap.Error(err))
		return err
	}
	
	source, ok := prData["source"].(map[string]interface{})
	if !ok {
		p.logger.Error("Invalid PR data format")
		return fmt.Errorf("invalid PR data format")
	}
	
	commit, ok := source["commit"].(map[string]interface{})
	if !ok {
		p.logger.Error("Invalid PR data format")
		return fmt.Errorf("invalid PR data format")
	}
	
	hash, ok := commit["hash"].(string)
	if !ok {
		p.logger.Error("Invalid PR data format")
		return fmt.Errorf("invalid PR data format")
	}
	
	// Bitbucket API endpoint for adding a comment to a specific line
	commentURL := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/commit/%s/comments", 
		owner, repo, hash)
	
	// Prepare the comment data
	commentData := map[string]interface{}{
		"content": map[string]string{
			"raw": comment,
		},
		"inline": map[string]interface{}{
			"path": filePath,
			"to":   lineNumber,
		},
	}
	
	commentJSON, err := json.Marshal(commentData)
	if err != nil {
		p.logger.Error("Failed to marshal comment data", zap.Error(err))
		return err
	}
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", commentURL, strings.NewReader(string(commentJSON)))
	if err != nil {
		p.logger.Error("Failed to create request", zap.Error(err))
		return err
	}
	
	// Set content type
	req.Header.Set("Content-Type", "application/json")
	
	// Add authentication if available
	if p.username != "" && (p.password != "" || p.token != "") {
		password := p.password
		if password == "" {
			password = p.token
		}
		req.SetBasicAuth(p.username, password)
	}
	
	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		p.logger.Error("Failed to execute request", zap.Error(err))
		return err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		p.logger.Error("Bitbucket API returned non-2xx status", 
			zap.Int("statusCode", resp.StatusCode))
		
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bitbucket API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	p.logger.Info("Successfully added comment to PR",
		zap.String("owner", owner),
		zap.String("repo", repo),
		zap.Int("prNum", prNum))
	
	return nil
}