package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v54/github"
	"github.com/user/go-git-test/pkg/git"
	"github.com/user/go-git-test/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// GitHubProvider is a provider for GitHub repositories
type GitHubProvider struct {
	*BaseProvider
	githubClient *github.Client
}

// NewGitHubProvider creates a new GitHub provider
func NewGitHubProvider(logger *logger.Logger, username, password, token, sshKeyPath string) *GitHubProvider {
	base := NewBaseProvider(logger, username, password, token, sshKeyPath)

	var githubClient *github.Client

	// Create GitHub client with appropriate authentication
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(context.Background(), ts)
		githubClient = github.NewClient(tc)
	} else if username != "" && password != "" {
		tp := github.BasicAuthTransport{
			Username: username,
			Password: password,
		}
		githubClient = github.NewClient(tp.Client())
	} else {
		githubClient = github.NewClient(nil)
	}

	return &GitHubProvider{
		BaseProvider: base,
		githubClient: githubClient,
	}
}

// Clone clones a repository from GitHub
func (p *GitHubProvider) Clone(repoURL, destination, branch string) error {
	p.logger.Info("Cloning GitHub repository",
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
func (p *GitHubProvider) GetPRDiff(ctx context.Context, prURL string) (string, error) {
	p.logger.Info("Getting GitHub PR diff", zap.String("prURL", prURL))

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

	// Get the PR to determine the right commits to compare
	pr, _, err := p.githubClient.PullRequests.Get(ctx, owner, repo, prNum)
	if err != nil {
		p.logger.Error("Failed to get PR", zap.Error(err))
		return "", err
	}

	// Get the diff from GitHub API
	opts := &github.RawOptions{Type: github.Diff}
	diff, _, err := p.githubClient.PullRequests.GetRaw(ctx, owner, repo, prNum, *opts)
	if err != nil {
		p.logger.Error("Failed to get PR diff", zap.Error(err))
		return "", err
	}

	p.logger.Info("Successfully retrieved PR diff",
		zap.String("owner", owner),
		zap.String("repo", repo),
		zap.Int("prNum", prNum),
		zap.String("base", pr.GetBase().GetSHA()),
		zap.String("head", pr.GetHead().GetSHA()))

	return diff, nil
}

// AddPRComment adds a comment to a pull request
func (p *GitHubProvider) AddPRComment(ctx context.Context, prURL string, filePath string, lineNumber int, comment string) error {
	p.logger.Info("Adding comment to GitHub PR",
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

	// Get the PR to determine the commit SHA
	pr, _, err := p.githubClient.PullRequests.Get(ctx, owner, repo, prNum)
	if err != nil {
		p.logger.Error("Failed to get PR", zap.Error(err))
		return err
	}

	// Create a review comment
	reviewComment := &github.PullRequestComment{
		Body:     github.String(comment),
		Path:     github.String(filePath),
		Position: github.Int(lineNumber),
		CommitID: github.String(pr.GetHead().GetSHA()),
	}

	_, _, err = p.githubClient.PullRequests.CreateComment(ctx, owner, repo, prNum, reviewComment)
	if err != nil {
		p.logger.Error("Failed to create PR comment", zap.Error(err))
		return err
	}

	p.logger.Info("Successfully added comment to PR",
		zap.String("owner", owner),
		zap.String("repo", repo),
		zap.Int("prNum", prNum))

	return nil
}

// GetPRComments gets all comments on a pull request
func (p *GitHubProvider) GetPRComments(ctx context.Context, prURL string) ([]string, error) {
	p.logger.Info("Getting GitHub PR comments", zap.String("prURL", prURL))

	owner, repo, prNumStr, err := ParsePRURL(prURL)
	if err != nil {
		p.logger.Error("Failed to parse PR URL", zap.Error(err))
		return nil, err
	}

	prNum, err := strconv.Atoi(prNumStr)
	if err != nil {
		p.logger.Error("Invalid PR number", zap.Error(err))
		return nil, err
	}

	// Get review comments (line-specific comments)
	reviewComments, _, err := p.githubClient.PullRequests.ListComments(ctx, owner, repo, prNum, nil)
	if err != nil {
		p.logger.Error("Failed to get PR review comments", zap.Error(err))
		return nil, err
	}

	// Get issue comments (general PR comments)
	issueComments, _, err := p.githubClient.Issues.ListComments(ctx, owner, repo, prNum, nil)
	if err != nil {
		p.logger.Error("Failed to get PR issue comments", zap.Error(err))
		return nil, err
	}

	var allComments []string

	// Process review comments
	for _, comment := range reviewComments {
		commentInfo := fmt.Sprintf("File: %s, Line: %d, User: %s\nComment: %s\n",
			comment.GetPath(),
			comment.GetPosition(),
			comment.GetUser().GetLogin(),
			comment.GetBody())
		allComments = append(allComments, commentInfo)
	}

	// Process issue comments
	for _, comment := range issueComments {
		commentInfo := fmt.Sprintf("User: %s\nComment: %s\n",
			comment.GetUser().GetLogin(),
			comment.GetBody())
		allComments = append(allComments, commentInfo)
	}

	return allComments, nil
}

// GetPRCommentsByParentID gets all reply comments for a specific parent comment ID
func (p *GitHubProvider) GetPRCommentsByParentID(ctx context.Context, prURL string, parentID int64) ([]*github.PullRequestComment, error) {
	p.logger.Info("Getting GitHub PR comments by parent ID", 
		zap.String("prURL", prURL),
		zap.Int64("parentID", parentID))

	owner, repo, prNumStr, err := ParsePRURL(prURL)
	if err != nil {
		p.logger.Error("Failed to parse PR URL", zap.Error(err))
		return nil, err
	}

	prNum, err := strconv.Atoi(prNumStr)
	if err != nil {
		p.logger.Error("Invalid PR number", zap.Error(err))
		return nil, err
	}

	// Get all review comments
	reviewComments, _, err := p.githubClient.PullRequests.ListComments(ctx, owner, repo, prNum, &github.PullRequestListCommentsOptions{})
	if err != nil {
		p.logger.Error("Failed to get PR review comments", zap.Error(err))
		return nil, err
	}

	var childComments []*github.PullRequestComment
	
	// Filter comments by parent ID
	for _, comment := range reviewComments {
		if comment.GetInReplyTo() == parentID {
			childComments = append(childComments, comment)
		}
	}

	p.logger.Info("Successfully retrieved PR comments by parent ID",
		zap.String("owner", owner),
		zap.String("repo", repo),
		zap.Int("prNum", prNum),
		zap.Int64("parentID", parentID),
		zap.Int("commentCount", len(childComments)))

	return childComments, nil
}

// GetPRFiles gets all files changed in a pull request
func (p *GitHubProvider) GetPRFiles(ctx context.Context, prURL string) ([]string, error) {
	p.logger.Info("Getting GitHub PR files", zap.String("prURL", prURL))

	owner, repo, prNumStr, err := ParsePRURL(prURL)
	if err != nil {
		p.logger.Error("Failed to parse PR URL", zap.Error(err))
		return nil, err
	}

	prNum, err := strconv.Atoi(prNumStr)
	if err != nil {
		p.logger.Error("Invalid PR number", zap.Error(err))
		return nil, err
	}

	// Get files changed in the PR
	files, _, err := p.githubClient.PullRequests.ListFiles(ctx, owner, repo, prNum, nil)
	if err != nil {
		p.logger.Error("Failed to get PR files", zap.Error(err))
		return nil, err
	}

	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.GetFilename())
	}

	return fileNames, nil
}

// FormatPRDiff formats the PR diff in a human-readable format
func (p *GitHubProvider) FormatPRDiff(diff string) string {
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