  ## Code Reviews

  Industry Standard Approach

  1. Pre-commit Hooks (Fast feedback)
  - Linting (Ruff)
  - Formatting (Ruff format)
  - Basic security checks
  - ❌ NOT full type checking or tests (too slow)

  2. CI/CD Pipeline (Comprehensive)
  - All quality checks (lint, format, type-check, tests)
  - Static analysis tools
  - Security scanning
  - This is where most teams do thorough automated review

  3. PR Review Tools (Optional but common)
  - SonarQube/SonarCloud
  - CodeClimate
  - Codacy
  - GitHub Code Scanning (CodeQL)
  - AI-powered reviewers (Sourcery, CodeRabbit)

  Your Current Setup

  You already have:
  - ✅ Pre-commit hooks (via task hooks:run)
  - ✅ CI pipeline (.github/workflows/ci.yml)

  What's Missing for Automated Code Review

  Looking at your ci.yml, you could add:

  1. SonarCloud (most popular free option for open source)
  2. CodeQL (GitHub's security scanner)
  3. Dependency vulnerability scanning (already have Trivy for containers)
  4. Code coverage reporting (Codecov/Coveralls)

  Recommendation

  Best practice: CI pipeline is primary gate. Add to your ci.yml:

  1. CodeQL for security analysis
  2. SonarCloud for code quality/technical debt
  3. Coverage badge via Codecov

  Want me to add any of these to your CI workflow?
