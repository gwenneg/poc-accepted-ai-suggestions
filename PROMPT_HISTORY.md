# Prompt History

## Session 1

### Prompt 1
This is an empty project that will be used to build a PoC. Create a PLAN.md file. We'll build a plan together in that file before you start implementing the project. Add an instruction in CLAUDE.md so that you read PLAN.md at the beginning of each of our sessions. I will tell you when we're ready to start implementing code.

### Prompt 2
The PoC will be a CLI tool written in GoLang. Its main goal will be to fetch data from the GitHub REST API about a specific repository. From that data, we want to detect comments that were added by AI-assisted review tools such as Sourcery.ai or CodeRabbit. For each of these comments, we want to determine whether they were accepted or rejected by the users who pushed commits to the PR or reviewed it. The CLI should accept the URL from a GitHub repo as the input. The initial output should be a simple JSON structure with details about which AI-assisted review tools commented a PR, how many they did it, and how many of these comments were accepted. The biggest challenge will likely be about how we determine a suggestion from AI was accepted.

### Prompt 3
Keep the codebase as simple as possible. Add an instruction in CLAUDE.md about that. I don't want over-engineered code.

### Prompt 4
We'll use a GITHUB_TOKEN env var for authentication with the GitHub API.

### Prompt 5
Use go version 1.24.0. Use github.com/google/go-github/v80 to interact with the GitHub API.
