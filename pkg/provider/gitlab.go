package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/user/go-git-test/pkg/git"
	"github.com/user/go-git-test/pkg/logger"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/zap"
)

// GitLabProvider is a provider for GitLab repositories
type GitLabProvider struct {
	*BaseProvider
	gitlabClient *gitlab.Client
}

// NewGitLabProvider creates a new GitLab provider
func NewGitLabProvider(logger *logger.Logger, username, password, token, sshKeyPath string) *GitLabProvider {
	base := NewBaseProvider(logger, username, password, token, sshKeyPath)

	var gitlabClient *gitlab.Client
	var err error

	// Create GitLab client with appropriate authentication
	if token != "" {
		gitlabClient, err = gitlab.NewClient(token)
		if err != nil {
			logger.Error("Failed to create GitLab client", zap.Error(err))
			return nil
		}
	} else if username != "" && password != "" {
		gitlabClient, err = gitlab.NewBasicAuthClient(username, password, gitlab.WithBaseURL("https://gitlab.com/api/v4"))
		if err != nil {
			logger.Error("Failed to create GitLab client", zap.Error(err))
			return nil
		}
	} else {
		// Anonymous client
		gitlabClient, err = gitlab.NewClient("")
		if err != nil {
			logger.Error("Failed to create GitLab client", zap.Error(err))
			return nil
		}
	}

	return &GitLabProvider{
		BaseProvider: base,
		gitlabClient: gitlabClient,
	}
}

// Clone clones a repository from GitLab
func (p *GitLabProvider) Clone(repoURL, destination, branch string) error {
	p.logger.Info("Cloning GitLab repository",
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

// GetPRDiff gets the diff for a merge request
func (p *GitLabProvider) GetPRDiff(ctx context.Context, prURL string) (string, error) {
	p.logger.Info("Getting GitLab MR diff", zap.String("prURL", prURL))

	owner, repo, mrNumStr, err := ParsePRURL(prURL)
	if err != nil {
		p.logger.Error("Failed to parse MR URL", zap.Error(err))
		return "", err
	}

	mrNum, err := strconv.Atoi(mrNumStr)
	if err != nil {
		p.logger.Error("Invalid MR number", zap.Error(err))
		return "", err
	}

	// In GitLab, we need the project ID (which can be the path with namespace)
	projectPath := fmt.Sprintf("%s/%s", owner, repo)

	// Get the merge request changes
	changes, _, err := p.gitlabClient.MergeRequests.GetMergeRequestChanges(projectPath, mrNum, nil)
	if err != nil {
		p.logger.Error("Failed to get MR changes", zap.Error(err))
		return "", err
	}

	// Build a unified diff format
	var diffBuilder strings.Builder

	for _, change := range changes.Changes {
		diffBuilder.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", change.OldPath, change.NewPath))
		diffBuilder.WriteString(fmt.Sprintf("--- a/%s\n", change.OldPath))
		diffBuilder.WriteString(fmt.Sprintf("+++ b/%s\n", change.NewPath))
		diffBuilder.WriteString(change.Diff)
		diffBuilder.WriteString("\n")
	}

	p.logger.Info("Successfully retrieved MR diff",
		zap.String("projectPath", projectPath),
		zap.Int("mrNum", mrNum))

	return diffBuilder.String(), nil
}

// AddPRComment adds a comment to a merge request
func (p *GitLabProvider) AddPRComment(ctx context.Context, prURL string, filePath string, lineNumber int, comment string) error {
	p.logger.Info("Adding comment to GitLab MR",
		zap.String("prURL", prURL),
		zap.String("filePath", filePath),
		zap.Int("lineNumber", lineNumber))

	owner, repo, mrNumStr, err := ParsePRURL(prURL)
	if err != nil {
		p.logger.Error("Failed to parse MR URL", zap.Error(err))
		return err
	}

	mrNum, err := strconv.Atoi(mrNumStr)
	if err != nil {
		p.logger.Error("Invalid MR number", zap.Error(err))
		return err
	}

	// In GitLab, we need the project ID (which can be the path with namespace)
	projectPath := fmt.Sprintf("%s/%s", owner, repo)

	// Get the merge request to determine the commit SHA
	mr, _, err := p.gitlabClient.MergeRequests.GetMergeRequest(projectPath, mrNum, nil)
	if err != nil {
		p.logger.Error("Failed to get MR", zap.Error(err))
		return err
	}
	// Create a new discussion (comment) on the merge request
	discussionOpts := &gitlab.CreateMergeRequestDiscussionOptions{
		Body: gitlab.String(comment),
		Position: &gitlab.PositionOptions{ //todo:--Position: &gitlab.NotePosition{
			BaseSHA:      gitlab.String(mr.DiffRefs.BaseSha),
			StartSHA:     gitlab.String(mr.DiffRefs.StartSha),
			HeadSHA:      gitlab.String(mr.DiffRefs.HeadSha),
			PositionType: gitlab.String("text"),
			NewPath:      gitlab.String(filePath),
			NewLine:      gitlab.Int(lineNumber),
		},
	}

	_, _, err = p.gitlabClient.Discussions.CreateMergeRequestDiscussion(projectPath, mrNum, discussionOpts)
	if err != nil {
		p.logger.Error("Failed to create MR comment", zap.Error(err))
		return err
	}

	p.logger.Info("Successfully added comment to MR",
		zap.String("projectPath", projectPath),
		zap.Int("mrNum", mrNum))

	return nil
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
