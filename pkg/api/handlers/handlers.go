package handlers

import (
	"context"
	"net/http"

	"github.com/google/go-github/v54/github"
	"github.com/labstack/echo/v4"
	"github.com/user/go-git-test/pkg/api/models"
	"github.com/user/go-git-test/pkg/logger"
	"github.com/user/go-git-test/pkg/provider"
	"go.uber.org/zap"
)

// Handlers contains all API handlers
type Handlers struct {
	logger          *logger.Logger
	providerFactory *provider.Factory
	version         string
}

// NewHandlers creates a new handlers instance
func NewHandlers(logger *logger.Logger, providerFactory *provider.Factory, version string) *Handlers {
	return &Handlers{
		logger:          logger,
		providerFactory: providerFactory,
		version:         version,
	}
}

// HealthCheck godoc
// @Summary Check API health
// @Description Get the health status of the API
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Router /health [get]
func (h *Handlers) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, models.HealthResponse{
		Status:  "ok",
		Version: h.version,
	})
}

// Clone godoc
// @Summary Clone a repository
// @Description Clone a Git repository from GitHub, GitLab, or Bitbucket
// @Tags git
// @Accept json
// @Produce json
// @Param request body models.CloneRequest true "Clone Request"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /clone [post]
func (h *Handlers) Clone(c echo.Context) error {
	req := new(models.CloneRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	// Determine provider type
	var pt provider.ProviderType
	switch req.Provider {
	case "github":
		pt = provider.GitHub
	case "gitlab":
		pt = provider.GitLab
	case "bitbucket":
		pt = provider.Bitbucket
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Unsupported provider type")
	}

	// Create provider
	p, err := h.providerFactory.CreateProvider(pt, req.Username, req.Password, req.Token, req.SSHKeyPath)
	if err != nil {
		h.logger.Error("Failed to create provider", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create provider")
	}

	// Clone repository
	err = p.Clone(req.URL, req.Destination, req.Branch)
	if err != nil {
		h.logger.Error("Failed to clone repository", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to clone repository: "+err.Error())
	}

	return c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Repository cloned successfully",
		Data: map[string]string{
			"destination": req.Destination,
		},
	})
}

// GetPRDiff godoc
// @Summary Get pull/merge request diff
// @Description Get the diff for a pull/merge request from GitHub, GitLab, or Bitbucket
// @Tags git
// @Accept json
// @Produce json
// @Param provider query string true "Provider (github, gitlab, bitbucket)"
// @Param pr_url query string true "Pull/Merge Request URL"
// @Param token query string false "Authentication token"
// @Param username query string false "Username for authentication"
// @Param password query string false "Password for authentication"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /prdiff [get]
func (h *Handlers) GetPRDiff(c echo.Context) error {
	req := new(models.PRDiffRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	// Determine provider type
	var pt provider.ProviderType
	switch req.Provider {
	case "github":
		pt = provider.GitHub
	case "gitlab":
		pt = provider.GitLab
	case "bitbucket":
		pt = provider.Bitbucket
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Unsupported provider type")
	}

	// Create provider
	p, err := h.providerFactory.CreateProvider(pt, req.Username, req.Password, req.Token, "")
	if err != nil {
		h.logger.Error("Failed to create provider", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create provider")
	}

	// Get PR diff
	diff, err := p.GetPRDiff(context.Background(), req.PRURL)
	if err != nil {
		h.logger.Error("Failed to get PR diff", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get PR diff: "+err.Error())
	}

	return c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data: map[string]string{
			"diff": diff,
		},
	})
}

// AddPRComment godoc
// @Summary Add a comment to a pull/merge request
// @Description Add a comment to a specific line in a pull/merge request from GitHub, GitLab, or Bitbucket
// @Tags git
// @Accept json
// @Produce json
// @Param request body models.PRCommentRequest true "PR Comment Request"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /prcomment [post]
func (h *Handlers) AddPRComment(c echo.Context) error {
	req := new(models.PRCommentRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	// Determine provider type
	var pt provider.ProviderType
	switch req.Provider {
	case "github":
		pt = provider.GitHub
	case "gitlab":
		pt = provider.GitLab
	case "bitbucket":
		pt = provider.Bitbucket
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Unsupported provider type")
	}

	// Create provider
	p, err := h.providerFactory.CreateProvider(pt, req.Username, req.Password, req.Token, "")
	if err != nil {
		h.logger.Error("Failed to create provider", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create provider")
	}

	// Add PR comment
	err = p.AddPRComment(context.Background(), req.PRURL, req.FilePath, req.LineNumber, req.Comment)
	if err != nil {
		h.logger.Error("Failed to add PR comment", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add PR comment: "+err.Error())
	}

	return c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Comment added successfully to PR",
		Data: map[string]interface{}{
			"pr_url":      req.PRURL,
			"file_path":   req.FilePath,
			"line_number": req.LineNumber,
		},
	})
}

// GetPRCommentsByParentID godoc
// @Summary Get PR comments by parent comment ID
// @Description Get all reply comments for a specific parent comment ID from GitHub
// @Tags git
// @Accept json
// @Produce json
// @Param provider query string true "Provider (github, gitlab, bitbucket)"
// @Param pr_url query string true "Pull/Merge Request URL"
// @Param parent_id query int true "Parent Comment ID"
// @Param token query string false "Authentication token"
// @Param username query string false "Username for authentication"
// @Param password query string false "Password for authentication"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /prcomments/parent [get]
func (h *Handlers) GetPRCommentsByParentID(c echo.Context) error {
	req := new(models.PRCommentsByParentRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	// Currently, only GitHub supports this feature
	if req.Provider != "github" {
		return echo.NewHTTPError(http.StatusBadRequest, "This feature is currently only supported for GitHub")
	}

	// Create provider
	p, err := h.providerFactory.CreateProvider(provider.GitHub, req.Username, req.Password, req.Token, "")
	if err != nil {
		h.logger.Error("Failed to create provider", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create provider")
	}

	// Get PR comments by parent ID
	githubProvider, ok := p.(*provider.GitHubProvider)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to cast to GitHub provider")
	}

	comments, err := githubProvider.GetPRCommentsByParentID(context.Background(), req.PRURL, req.ParentID)
	if err != nil {
		h.logger.Error("Failed to get PR comments by parent ID", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get PR comments: "+err.Error())
	}

	// Convert GitHub comments to our response model
	responseComments := make([]models.PRComment, 0, len(comments))
	for _, comment := range comments {
		responseComments = append(responseComments, models.PRComment{
			ID:        comment.GetID(),
			Body:      comment.GetBody(),
			User:      comment.GetUser().GetLogin(),
			CreatedAt: comment.GetCreatedAt().Format("2006-01-02T15:04:05Z"),
			UpdatedAt: comment.GetUpdatedAt().Format("2006-01-02T15:04:05Z"),
			Path:      comment.GetPath(),
			Position:  comment.GetPosition(),
			CommitID:  comment.GetCommitID(),
			HTMLURL:   comment.GetHTMLURL(),
		})
	}

	return c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Comments retrieved successfully",
		Data: map[string]interface{}{
			"pr_url":    req.PRURL,
			"parent_id": req.ParentID,
			"comments":  responseComments,
		},
	})
}