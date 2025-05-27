package provider

import (
	"fmt"
	"strings"
	"github.com/user/go-git-test/pkg/logger"
)

// ProviderType represents the type of Git provider
type ProviderType string

const (
	// GitHub provider type
	GitHub ProviderType = "github"
	// GitLab provider type
	GitLab ProviderType = "gitlab"
	// Bitbucket provider type
	Bitbucket ProviderType = "bitbucket"
)

// Factory creates provider instances
type Factory struct {
	logger *logger.Logger
}

// NewFactory creates a new provider factory
func NewFactory(logger *logger.Logger) *Factory {
	return &Factory{
		logger: logger,
	}
}

// CreateProvider creates a provider of the given type
func (f *Factory) CreateProvider(
	providerType ProviderType,
	username, password, token, sshKeyPath string,
) (Provider, error) {
	switch providerType {
	case GitHub:
		return NewGitHubProvider(f.logger, username, password, token, sshKeyPath), nil
	case GitLab:
		return NewGitLabProvider(f.logger, username, password, token, sshKeyPath), nil
	case Bitbucket:
		return NewBitbucketProvider(f.logger, username, password, token, sshKeyPath), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

// DetectProviderFromURL detects the provider type from a repository URL
func (f *Factory) DetectProviderFromURL(repoURL string) (ProviderType, error) {
	repoURL = strings.ToLower(repoURL)
	
	if strings.Contains(repoURL, "github.com") {
		return GitHub, nil
	} else if strings.Contains(repoURL, "gitlab.com") {
		return GitLab, nil
	} else if strings.Contains(repoURL, "bitbucket.org") {
		return Bitbucket, nil
	}
	
	return "", fmt.Errorf("unable to detect provider from URL: %s", repoURL)
}

// CreateProviderFromURL creates a provider based on the repository URL
func (f *Factory) CreateProviderFromURL(repoURL, username, password, token, sshKeyPath string) (Provider, error) {
	providerType, err := f.DetectProviderFromURL(repoURL)
	if err != nil {
		return nil, err
	}
	
	return f.CreateProvider(providerType, username, password, token, sshKeyPath)
}