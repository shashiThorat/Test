# Generic Git Project

A Go library for working with Git repositories across multiple providers (GitHub, GitLab, and Bitbucket).

## Features

- Clone repositories from GitHub, GitLab, and Bitbucket
- View diffs between commits
- View diffs for pull requests/merge requests
- Add comments to specific lines in pull requests/merge requests
- Get comments by parent ID (threaded discussions) in pull requests
- Support for multiple authentication methods (Basic, Token, SSH)
- Structured logging with Zap
- Clean architecture using design patterns (Adapter, Factory)
- RESTful API using Echo framework
- Interactive API documentation with Swagger

## Installation

```bash
go get github.com/user/go-git-test
```

## CLI Usage

### Clone a Repository

```bash
# Clone a GitHub repository
go run cmd/gitclient/main.go -command=clone -provider=github -url=https://github.com/user/repo.git -dest=./repo

# Clone a GitLab repository with a specific branch
go run cmd/gitclient/main.go -command=clone -provider=gitlab -url=https://gitlab.com/user/repo.git -dest=./repo -branch=develop

# Clone a Bitbucket repository with token authentication
go run cmd/gitclient/main.go -command=clone -provider=bitbucket -url=https://bitbucket.org/user/repo.git -dest=./repo -token=your_token
```

### View Diff Between Commits

```bash
# View diff between two commits in a cloned repository
go run cmd/gitclient/main.go -command=diff -dest=./repo -old-commit=abc123 -new-commit=def456
```

### View PR/MR Diff

```bash
# View diff for a GitHub pull request
go run cmd/gitclient/main.go -command=prdiff -provider=github -pr-url=https://github.com/user/repo/pull/123 -token=your_token

# View diff for a GitLab merge request
go run cmd/gitclient/main.go -command=prdiff -provider=gitlab -pr-url=https://gitlab.com/user/repo/-/merge_requests/123 -token=your_token

# View diff for a Bitbucket pull request
go run cmd/gitclient/main.go -command=prdiff -provider=bitbucket -pr-url=https://bitbucket.org/user/repo/pull-requests/123 -username=your_username -password=your_password
```

### Add Comment to PR/MR

```bash
# Add a comment to a GitHub pull request
go run cmd/gitclient/main.go -command=prcomment -provider=github -pr-url=https://github.com/user/repo/pull/123 -file-path=path/to/file.go -line-number=42 -comment="This looks good!" -token=your_token

# Add a comment to a GitLab merge request
go run cmd/gitclient/main.go -command=prcomment -provider=gitlab -pr-url=https://gitlab.com/user/repo/-/merge_requests/123 -file-path=path/to/file.go -line-number=42 -comment="This looks good!" -token=your_token

# Add a comment to a Bitbucket pull request
go run cmd/gitclient/main.go -command=prcomment -provider=bitbucket -pr-url=https://bitbucket.org/user/repo/pull-requests/123 -file-path=path/to/file.go -line-number=42 -comment="This looks good!" -username=your_username -password=your_password
```

### Get Comments by Parent ID (GitHub only)

```bash
# Get comments by parent ID from a GitHub pull request
go run cmd/gitclient/main.go -command=prcomments -provider=github -pr-url=https://github.com/user/repo/pull/123 -parent-id=123456789 -token=your_token
```

### Start API Server

```bash
# Start the API server on default port 8080
go run cmd/gitclient/main.go -command=serve

# Start the API server on a custom port
go run cmd/gitclient/main.go -command=serve -port=3000

# Start the API server with debug logging
go run cmd/gitclient/main.go -command=serve -debug
```

## API Usage

The API provides RESTful endpoints for Git operations. You can explore the API using the Swagger UI at `http://localhost:8080/swagger/index.html`.

### Health Check

```
GET /api/v1/health
```

Response:
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

### Clone Repository

```
POST /api/v1/clone
```

Request:
```json
{
  "provider": "github",
  "url": "https://github.com/user/repo.git",
  "destination": "./repo",
  "branch": "main",
  "token": "your_token"
}
```

Response:
```json
{
  "success": true,
  "message": "Repository cloned successfully",
  "data": {
    "destination": "./repo"
  }
}
```

### Get PR Diff

```
GET /api/v1/prdiff?provider=github&pr_url=https://github.com/user/repo/pull/123&token=your_token
```

Response:
```json
{
  "success": true,
  "data": {
    "diff": "diff --git a/file.go b/file.go\nindex abc..def 100644\n--- a/file.go\n+++ b/file.go\n@@ -10,7 +10,7 @@\n func main() {\n-  fmt.Println(\"Hello, World!\")\n+  fmt.Println(\"Hello, Go!\")\n }"
  }
}
```

### Add PR Comment

```
POST /api/v1/prcomment
```

Request:
```json
{
  "provider": "github",
  "pr_url": "https://github.com/user/repo/pull/123",
  "file_path": "file.go",
  "line_number": 42,
  "comment": "This looks good!",
  "token": "your_token"
}
```

Response:
```json
{
  "success": true,
  "message": "Comment added successfully to PR",
  "data": {
    "pr_url": "https://github.com/user/repo/pull/123",
    "file_path": "file.go",
    "line_number": 42
  }
}
```

### Get PR Comments by Parent ID

```
GET /api/v1/prcomments/parent?provider=github&pr_url=https://github.com/user/repo/pull/123&parent_id=123456789&token=your_token
```

Response:
```json
{
  "success": true,
  "message": "Comments retrieved successfully",
  "data": {
    "pr_url": "https://github.com/user/repo/pull/123",
    "parent_id": 123456789,
    "comments": [
      {
        "id": 234567890,
        "body": "I agree with your comment!",
        "user": "another-user",
        "created_at": "2023-01-02T12:00:00Z",
        "updated_at": "2023-01-02T12:00:00Z",
        "path": "file.go",
        "position": 42,
        "commit_id": "abc123def456",
        "html_url": "https://github.com/user/repo/pull/123#discussion_r234567890"
      }
    ]
  }
}
```

## Swagger Documentation

The API is documented using Swagger. To access the Swagger UI:

1. Start the API server:
   ```bash
   go run cmd/gitclient/main.go -command=serve
   ```

2. Open your browser and navigate to:
   ```
   http://localhost:8080/swagger/index.html
   ```

3. Use the Swagger UI to explore and test the API endpoints.

To generate or update the Swagger documentation:

```bash
# Install Swagger
go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger docs
make swagger
```

## Authentication

The tool supports multiple authentication methods:

- Basic authentication (username/password)
- Token authentication
- SSH key authentication

Example with SSH:

```bash
go run cmd/gitclient/main.go -command=clone -provider=github -url=git@github.com:user/repo.git -dest=./repo -ssh-key=~/.ssh/id_rsa
```

## Command Line Options

| Flag        | Description                                                      | Default  |
|-------------|------------------------------------------------------------------|----------|
| provider    | Git provider (github, gitlab, bitbucket)                         | github   |
| url         | Repository URL                                                   |          |
| dest        | Destination directory                                            | .        |
| branch      | Branch to clone                                                  |          |
| username    | Username for authentication                                      |          |
| password    | Password for authentication                                      |          |
| token       | Token for authentication                                         |          |
| ssh-key     | Path to SSH key                                                  |          |
| command     | Command to execute (clone, diff, prdiff, prcomment, prcomments, serve) | clone    |
| old-commit  | Old commit for diff                                              |          |
| new-commit  | New commit for diff                                              |          |
| pr-url      | PR/MR URL for prdiff, prcomment, and prcomments commands         |          |
| file-path   | File path for prcomment command                                  |          |
| line-number | Line number for prcomment command                                |          |
| comment     | Comment text for prcomment command                               |          |
| parent-id   | Parent comment ID for prcomments command                         |          |
| debug       | Enable debug logging                                             | false    |
| port        | Port for API server (when using serve command)                   | 8080     |

## Architecture

The project uses several design patterns:

1. **Adapter Pattern**: Provides a common interface for different Git providers
2. **Factory Pattern**: Creates provider instances based on the provider type
3. **Repository Pattern**: Abstracts the Git operations

### API Layer

The API layer uses the Echo framework and follows these principles:

1. **Clean Architecture**: Separation of concerns between routes, handlers, and models
2. **Middleware**: Custom middleware for logging, error handling, and validation
3. **Validation**: Request validation using the validator package
4. **Error Handling**: Consistent error responses across all endpoints
5. **Documentation**: API documentation using Swagger

## Adding a New Provider

To add a new Git provider:

1. Create a new file in `pkg/provider/` implementing the Provider interface
2. Add the provider type to the ProviderType enum in `pkg/provider/factory.go`
3. Update the factory's CreateProvider method to support the new provider

## Dependencies

- [go-git](https://github.com/go-git/go-git): A highly extensible Git implementation in pure Go
- [go-github](https://github.com/google/go-github): Go library for accessing the GitHub API
- [go-gitlab](https://github.com/xanzy/go-gitlab): Go client library for the GitLab API
- [go-bitbucket](https://github.com/ktrysmt/go-bitbucket): Go client library for the Bitbucket API
- [zap](https://github.com/uber-go/zap): Blazing fast, structured, leveled logging in Go
- [echo](https://github.com/labstack/echo): High performance, extensible, minimalist Go web framework
- [validator](https://github.com/go-playground/validator): Go Struct and Field validation
- [swag](https://github.com/swaggo/swag): Converts Go annotations to Swagger Documentation
- [echo-swagger](https://github.com/swaggo/echo-swagger): Swagger UI for Echo framework