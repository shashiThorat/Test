package models

// CloneRequest represents a repository clone request
type CloneRequest struct {
	Provider    string `json:"provider" validate:"required,oneof=github gitlab bitbucket" example:"github"`
	URL         string `json:"url" validate:"required,url" example:"https://github.com/user/repo.git"`
	Destination string `json:"destination" validate:"required" example:"./repo"`
	Branch      string `json:"branch" example:"main"`
	Token       string `json:"token" example:"ghp_xxxxxxxxxxxx"`
	Username    string `json:"username" example:"username"`
	Password    string `json:"password" example:"password"`
	SSHKeyPath  string `json:"ssh_key_path" example:"~/.ssh/id_rsa"`
}

// PRDiffRequest represents a PR diff request
type PRDiffRequest struct {
	Provider string `json:"provider" validate:"required,oneof=github gitlab bitbucket" example:"github"`
	PRURL    string `json:"pr_url" validate:"required,url" example:"https://github.com/user/repo/pull/123"`
	Token    string `json:"token" example:"ghp_xxxxxxxxxxxx"`
	Username string `json:"username" example:"username"`
	Password string `json:"password" example:"password"`
}

// PRCommentRequest represents a PR comment request
type PRCommentRequest struct {
	Provider   string `json:"provider" validate:"required,oneof=github gitlab bitbucket" example:"github"`
	PRURL      string `json:"pr_url" validate:"required,url" example:"https://github.com/user/repo/pull/123"`
	FilePath   string `json:"file_path" validate:"required" example:"src/main.go"`
	LineNumber int    `json:"line_number" validate:"required,min=1" example:"42"`
	Comment    string `json:"comment" validate:"required" example:"This looks good!"`
	Token      string `json:"token" example:"ghp_xxxxxxxxxxxx"`
	Username   string `json:"username" example:"username"`
	Password   string `json:"password" example:"password"`
}

// PRCommentsByParentRequest represents a request to get PR comments by parent ID
type PRCommentsByParentRequest struct {
	Provider string `json:"provider" validate:"required,oneof=github gitlab bitbucket" example:"github"`
	PRURL    string `json:"pr_url" validate:"required,url" example:"https://github.com/user/repo/pull/123"`
	ParentID int64  `json:"parent_id" validate:"required" example:"123456789"`
	Token    string `json:"token" example:"ghp_xxxxxxxxxxxx"`
	Username string `json:"username" example:"username"`
	Password string `json:"password" example:"password"`
}

// PRComment represents a PR comment in the response
type PRComment struct {
	ID        int64  `json:"id" example:"123456789"`
	Body      string `json:"body" example:"This is a comment"`
	User      string `json:"user" example:"username"`
	CreatedAt string `json:"created_at" example:"2023-01-01T12:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2023-01-01T12:00:00Z"`
	Path      string `json:"path,omitempty" example:"src/main.go"`
	Position  int    `json:"position,omitempty" example:"42"`
	CommitID  string `json:"commit_id,omitempty" example:"abc123def456"`
	HTMLURL   string `json:"html_url" example:"https://github.com/user/repo/pull/123#discussion_r123456789"`
}

// Response represents a generic API response
type Response struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message,omitempty" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty" example:"An error occurred"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request parameters"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Version string `json:"version" example:"1.0.0"`
}