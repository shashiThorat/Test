package provider

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/google/go-github/v54/github"
	"github.com/user/go-git-test/pkg/git"
	"github.com/user/go-git-test/pkg/logger"
	"go.uber.org/zap"
)

// Common errors
var (
	ErrInvalidURL       = errors.New("invalid repository URL")
	ErrInvalidProvider  = errors.New("invalid provider")
	ErrNotImplemented   = errors.New("operation not implemented for this provider")
	ErrAuthentication   = errors.New("authentication failed")
	ErrInvalidPRURL     = errors.New("invalid pull request URL")
	ErrPermissionDenied = errors.New("permission denied")
)

// Provider represents a Git provider (GitHub, GitLab, Bitbucket)
type Provider interface {
	// Clone clones a repository
	Clone(repoURL, destination, branch string) error

	// GetAuthMethod returns the authentication method for the provider
	GetAuthMethod() (transport.AuthMethod, error)

	// GetClient returns the Git client
	GetClient() *git.Client

	// GetPRDiff gets the diff for a pull/merge request
	GetPRDiff(ctx context.Context, prURL string) (string, error)

	// AddPRComment adds a comment to a pull/merge request
	AddPRComment(ctx context.Context, prURL string, filePath string, lineNumber int, comment string) error
	
	// GetPRCommentsByParentID gets all reply comments for a specific parent comment ID
	GetPRCommentsByParentID(ctx context.Context, prURL string, parentID int64) ([]*github.PullRequestComment, error)
}

// BaseProvider implements common functionality for all providers
type BaseProvider struct {
	client     *git.Client
	logger     *logger.Logger
	username   string
	password   string
	token      string
	sshKeyPath string
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(logger *logger.Logger, username, password, token, sshKeyPath string) *BaseProvider {
	return &BaseProvider{
		client:     git.NewClient(logger),
		logger:     logger,
		username:   username,
		password:   password,
		token:      token,
		sshKeyPath: sshKeyPath,
	}
}

// GetClient returns the Git client
func (p *BaseProvider) GetClient() *git.Client {
	return p.client
}

// GetAuthMethod returns the authentication method based on provided credentials
func (p *BaseProvider) GetAuthMethod() (transport.AuthMethod, error) {
	// SSH authentication
	if p.sshKeyPath != "" {
		p.logger.Info("Using SSH authentication", zap.String("keyPath", p.sshKeyPath))

		sshKey, err := os.ReadFile(p.sshKeyPath)
		if err != nil {
			p.logger.Error("Failed to read SSH key", zap.Error(err))
			return nil, err
		}

		auth, err := ssh.NewPublicKeys("git", sshKey, "")
		if err != nil {
			p.logger.Error("Failed to create SSH auth method", zap.Error(err))
			return nil, err
		}

		return auth, nil
	}

	// Token authentication
	if p.token != "" {
		p.logger.Info("Using token authentication")
		return &http.TokenAuth{Token: p.token}, nil
	}

	// Basic authentication
	if p.username != "" && p.password != "" {
		p.logger.Info("Using basic authentication", zap.String("username", p.username))
		return &http.BasicAuth{
			Username: p.username,
			Password: p.password,
		}, nil
	}

	// No authentication
	p.logger.Info("No authentication provided, using anonymous access")
	return nil, nil
}

// ParseRepoURL parses a repository URL and extracts the owner and repo name
func ParseRepoURL(repoURL string) (string, string, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", "", err
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", ErrInvalidURL
	}

	return parts[0], parts[1], nil
}

// ParsePRURL parses a PR URL and extracts the owner, repo name, and PR number
func ParsePRURL(prURL string) (string, string, string, error) {
	u, err := url.Parse(prURL)
	if err != nil {
		return "", "", "", err
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 4 {
		return "", "", "", ErrInvalidPRURL
	}

	// Different providers have different URL formats for PRs/MRs
	// GitHub: /owner/repo/pull/123
	// GitLab: /owner/repo/-/merge_requests/123
	// Bitbucket: /owner/repo/pull-requests/123

	owner := parts[0]
	repo := parts[1]
	var prNumber string

	if strings.Contains(u.Host, "github.com") {
		if parts[2] != "pull" {
			return "", "", "", ErrInvalidPRURL
		}
		prNumber = parts[3]
	} else if strings.Contains(u.Host, "gitlab.com") {
		if parts[2] != "-" || parts[3] != "merge_requests" {
			return "", "", "", ErrInvalidPRURL
		}
		prNumber = parts[4]
	} else if strings.Contains(u.Host, "bitbucket.org") {
		if parts[2] != "pull-requests" {
			return "", "", "", ErrInvalidPRURL
		}
		prNumber = parts[3]
	} else {
		return "", "", "", ErrInvalidPRURL
	}

	return owner, repo, prNumber, nil
}

// GetPRDiff gets the diff for a pull/merge request (base implementation)
func (p *BaseProvider) GetPRDiff(ctx context.Context, prURL string) (string, error) {
	return "", ErrNotImplemented
}

// AddPRComment adds a comment to a pull/merge request (base implementation)
func (p *BaseProvider) AddPRComment(ctx context.Context, prURL string, filePath string, lineNumber int, comment string) error {
	return ErrNotImplemented
}

// GetPRCommentsByParentID gets all reply comments for a specific parent comment ID (base implementation)
func (p *BaseProvider) GetPRCommentsByParentID(ctx context.Context, prURL string, parentID int64) ([]*github.PullRequestComment, error) {
	return nil, ErrNotImplemented
}