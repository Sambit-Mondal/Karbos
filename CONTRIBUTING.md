# Contributing to Karbos

Thank you for your interest in contributing to Karbos! We welcome contributions from the community.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Be respectful and constructive.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/Sambit-Mondal/Karbos/issues)
2. If not, create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Docker version, etc.)
   - Screenshots if applicable

### Suggesting Features

1. Check [Discussions](https://github.com/Sambit-Mondal/Karbos/discussions) for similar ideas
2. Create a new discussion or issue with:
   - Clear description of the feature
   - Use case and benefits
   - Potential implementation approach

### Pull Requests

1. **Fork the repository**
2. **Create a branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**:
   - Follow the existing code style
   - Add tests for new functionality
   - Update documentation as needed

4. **Commit your changes** using [Conventional Commits](https://www.conventionalcommits.org/):
   ```bash
   git commit -m "Feat: Added carbon intensity caching"
   git commit -m "Fix: Resolved Redis connection timeout"
   git commit -m "Docs: Updated API endpoint documentation"
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Open a Pull Request**:
   - Provide a clear title and description
   - Reference related issues
   - Add screenshots for UI changes

## Development Guidelines

### Code Style

**Go (Backend)**:
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Run `gofmt` before committing
- Use meaningful variable names
- Add comments for complex logic

**TypeScript (Frontend)**:
- Follow the existing ESLint configuration
- Use TypeScript strict mode
- Prefer functional components with hooks
- Add JSDoc comments for complex functions

### Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `style:` Code style changes (formatting, semicolons, etc.)
- `refactor:` Code refactoring
- `perf:` Performance improvements
- `test:` Adding or updating tests
- `chore:` Maintenance tasks

Examples:
```
feat: add support for multiple carbon API providers
fix: resolve database connection pool exhaustion
docs: update deployment guide for Kubernetes
perf: optimize job queue polling interval
```

### Testing

- Write tests for new features
- Ensure all tests pass before submitting PR
- Aim for >80% code coverage

**Run tests**:
```bash
# Backend
cd server && go test -v ./...

# Frontend
cd client && npm test
```

### Documentation

- Update README.md for user-facing changes
- Add comments for complex algorithms
- Create/update diagrams for architectural changes

## Project Structure

```
Karbos/
├── client/          # Frontend (Next.js)
├── server/          # Backend (Go)
├── docs/            # Documentation
├── docker-compose.yml
└── README.md
```

## Getting Help

- **LinkedIn**: [Connect with me!](https://linkedin.com/in/sambitm02)
- **Email**: [Mail me!](sambitmondal2005@gmail.com)


## License

By contributing, you agree that your contributions will be licensed under the Apache-2.0 License.