# Claude Code Instructions for WebEncode

This file contains project-specific instructions for Claude Code (the AI coding assistant) when working on the WebEncode codebase.

## Code Quality Standards

### Testing Requirements
- **100% test coverage target**: All new code must include comprehensive unit tests
- Test both success and error paths
- Use table-driven tests where appropriate
- Mock external dependencies (database, S3, NATS, etc.)
- Integration tests should be clearly marked and skippable without external services

### Documentation Requirements
- All exported functions, types, and constants must have Go doc comments
- Complex algorithms and business logic must include inline comments
- Keep SPECIFICATION.md updated with architectural decisions and system behavior
- Track all features and capabilities in SPECIFICATION.md
- Document API changes in docs/API_REFERENCE.md

### Code Organization
- Follow Go best practices and idioms
- Use meaningful variable and function names
- Keep functions focused and single-purpose
- Avoid deeply nested code - prefer early returns
- Use error wrapping with fmt.Errorf("context: %w", err)

## Development Workflow

### Before Committing
1. **Run tests**: `make test` or `go test ./...` must pass
2. **Check build**: `make build-all` must succeed
3. **Verify coverage**: `make test-coverage` and review results
4. **Run linters**: Code should be `go fmt` compliant
5. **Update documentation**: SPECIFICATION.md, ISSUES.md, README.md as needed

### Git Commits
- Make commits when code is in a functional, compilable state
- Use conventional commit format: `feat:`, `fix:`, `docs:`, `test:`, etc.
- Include descriptive commit messages
- Co-Author commits: `Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>`

### Build and Deploy
- **Do not start dev servers** - just build the application
- **Do not open web browsers** for testing
- Use Docker and docker-compose for local testing
- Verify builds compile before marking tasks complete

## Project-Specific Rules

### Language-Specific
- **Go**: This is primarily a Go project (Go 1.24+)
  - Use pgx/v5 for PostgreSQL interactions
  - Use NATS JetStream for messaging
  - Plugin system uses HashiCorp go-plugin
  - Follow micro-kernel architecture - business logic in plugins

- **TypeScript/Next.js**: Frontend in ui/ directory
  - Next.js 16 with React 19
  - Use Tailwind CSS 4 (no gradient colors)
  - shadcn/ui components
  - No direct DOM manipulation

- **Python**: Use virtual environments (`venv`) for any Python packages

### Architecture Principles
- **Micro-kernel**: Core kernel has minimal business logic
- **Plugin-based**: All major features (Auth, Storage, Encoder, Live, Publisher) are plugins
- **Separation of concerns**: Clear boundaries between layers
- **12-factor app**: Configuration via environment variables
- **Production-ready**: Treat all code as production code

### Error Handling and Logging
- **Proper logging is mandatory**: Use structured logging (zerolog)
- Log at appropriate levels (debug, info, warn, error)
- Include context in error messages
- Track errors in ISSUES.md

### Issue and Feature Tracking
- **ISSUES.md**: Document known bugs, technical debt, and workarounds
- **SPECIFICATION.md**: Single source of truth for features and architecture
- Update these files as you discover issues or implement features

## Publishing and Open Source

- Nearly every project will be published on GitHub
- Maintain a professional README.md with:
  - Clear project description
  - Quick start instructions
  - Architecture overview
  - Contributing guidelines
- Include proper LICENSE file (MIT)
- Keep sensitive data (API keys, secrets) out of the repository

## Testing Guidelines

### Unit Tests
- Test files: `*_test.go` alongside source files
- Use `testify` for assertions and mocks
- Mock external dependencies (DB, API calls, file I/O)
- Test edge cases and error conditions

### Integration Tests
- Mark integration tests with `t.Skip()` if dependencies unavailable
- Use Docker containers for test databases/services where possible
- Clean up resources after tests

### Coverage
- Aim for 100% coverage on all new code
- Focus on business logic and error paths
- Generated code (sqlc, protobuf) can have lower coverage
- Use `go test -coverprofile=coverage.out` to check coverage
- Review uncovered lines: `go tool cover -func=coverage.out`

## Specific Conventions

### Database
- Use SQLC for type-safe database queries
- Migrations in `migrations/` directory
- All database queries must be in `queries/` for SQLC generation
- Always use transactions for multi-statement operations

### API Design
- RESTful endpoints under `/api/v1/`
- Use proper HTTP status codes
- Consistent error response format
- OpenAPI/Swagger documentation in `docs/openapi.yaml`

### Plugin Development
- Follow the 5-pillar plugin system (Auth, Storage, Encoder, Live, Publisher)
- Use gRPC for plugin communication
- Health checks mandatory for all plugins
- Graceful shutdown required

## Performance Considerations
- Profile before optimizing
- Use connection pooling for databases
- Stream large files, don't buffer entirely in memory
- Monitor goroutine leaks
- Use context for cancellation and timeouts

## Security
- Never commit secrets, API keys, or credentials
- Use environment variables for sensitive configuration
- Validate all user input
- Sanitize file paths to prevent directory traversal
- Use prepared statements for SQL (SQLC does this automatically)

---

**Note**: These instructions are meant to guide Claude Code in maintaining consistency and quality. Human developers should also follow these guidelines when contributing to WebEncode.
