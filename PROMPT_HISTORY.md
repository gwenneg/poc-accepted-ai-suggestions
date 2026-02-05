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

### Prompt 6
Which other AI-assisted review tools integrated into GitHub are you aware of?

### Prompt 7
We'll target all AI review bots in this tool. Ignore the static analysis tools for now.

### Prompt 8
The repo will need a README.md file with a focus on how to run the CLI. Do not include details about the architecture. Keep it simple.

### Prompt 9
We might eventually have an implementation for all 8 AI-assisted review tools listed in the plan, but I want a focus only on CodeRabbit and Sourcery.ai at first. Do not implement something for the other 6 until instructed otherwise. Put the implementation for each review tool into a separate Go module within the project.

### Prompt 10
The acceptance criteria will likely vary depending on each AI review tool. Start searching details about that for CodeRabbit and Sourcery now. Update the plan with your findings.

### Prompt 11
Let's start with PRs from the past month only. It's fine to hardcode that for now. We might make the time window configurable later.

### Prompt 12
When the Sourcery bot reviews a PR, it will be marked as a reviewer in the PR. That's probably the best way to filter out PRs from the analysis without analyzing comments to determine whether an AI review tool was involved. Check if that also works with CodeRabbit.

### Prompt 13
Although unlikely, a single PR could technically be analyzed by two or more AI review tools. To make it possible in this tool, the output should be a JSON array. Each JSON item from that array should be dedicated to the analysis of a particular AI review tool. Do not try to unify the data structure between all AI review tools. For now, let's reflect the exact data that is available for each one of them. We might unify that into a single JSON data structure later.

### Prompt 14
The plan can be cleaned up a bit. Review it.

### Prompt 15
The thread resolution doesn't necessarily mean the AI comment was applied. Developers might use it to discard a comment they considered irrelevant. Let's only count the threads that were resolved by the AI tool itself, which should mean the AI analyzed a newer commit and determined the initial issue was fixed.

### Prompt 16
Analyzing thumbs up and down might be useful for Coderabbit. Include that in the output as well.

### Prompt 17
It seems CodeRabbit offers an API - probably not available with a free plan - which helps gather statistics about the reviews it did and the suggestions acceptance. Research this. Determine if and how we could use it in this project.

### Prompt 18
Let's keep that in the plan but consider it an option for later. This should not be implemented immediately.

### Prompt 19
Does Sourcery offer a similar API?

### Prompt 20
Review the plan again. Clean it up if needed. Ask me any questions that you consider important before you start implementing it.

### Prompt 21
"ai-review-analyzer" is fine to the CLI tool name. Analyze all PRs regardless of their status. In case of error, log it and skip the PR. Report partial results. Add simple and reasonable logs to report progress.

### Prompt 22
Add an .env.example file with all environment variables needed to run the project. Each one should have a short comment explaining why it's needed.

### Prompt 23
You were not supposed to create any new file until I told you to start implementing. It's fine though, we're about to finalize the plan.

### Prompt 24
Review the plan again. Are we ready to start implementing?

### Prompt 25
Start implementing the plan. Raise issues early. Ask for guidance in case of doubt. Test the code with this repo: https://github.com/RedHatInsights/notifications-backend. I added a GitHub token in the .env file which you can use to perform real requests to the GitHub API. Do not write Go tests yet. We'll take care of that later.

### Prompt 26
We'll address that later. Make this optional in the plan, mention that we'll use GraphQL and why.

### Prompt 27
Create the README.md and split the code into modules

### Prompt 28
It seems all Sourcery comments come with one thumb up and one thumb down by default. Research this and confirm it. If confirmed, do not count these defaults reactions. Also look for something similar with Coderabbit.

### Prompt 29
Explain why github.com/google/go-querystring is needed

### Prompt 30
Did you implement something to deal with paginated comments from a PR? The GitHub API limits how many comments can be fetched with a single API call. We might need to iterate over pages for bigger PRs.

### Prompt 31
Let's introduce a new feature in the plan: submitting for analysis to an LLM all comments from a GitHub discussion that was initiated by an AI-assisted review tool suggestion. We'll ask the LLM to analyze these comments and return a JSON data structure with two fields: a score (from 0 to 100) that should reflect how useful the AI suggestion was, and a summary which should summarize the user discussions. The goal is to reflect the users opinion and intent in the summary. We'll do that with Claude. Looks at how https://github.com/RedHatInsights/release-confidence-score/tree/main/internal/llm is implemented for inspiration. In this project, we'll implement the new feature in the simplest possible way. Do not over-engineer this. We will only use Claude. Do not plan for extra LLMS for now.

### Prompt 32
If a thread doesn't have any user comments, it should not be submitted for analysis

### Prompt 33
Make sure we only include useful information into the user prompt sent to the LLM. For instance, there's no need to include usernames. For the initial implementation, only the message from each comment should be enough. This feature should keep the cost as low as possible, so consumed tokens have to remain low.

### Prompt 34
You will need to add the new required env vars to .env.example (with comments).

### Prompt 35
Start implementing this.

### Prompt 36
I don't want to send requests to the Anthropic API. Look at the repo I shared for inspiration. I want the same approach here: a configurable model API URL and user key. The model ID can be hardcoded.

### Prompt 37
Is the new code similar to https://github.com/RedHatInsights/release-confidence-score/blob/main/internal/llm/providers/claude.go?

### Prompt 38
I want separate system and user prompt. Create a markdown file in this project for the system prompt. The user prompt can be created from the code.

### Prompt 39
Let's update the JSON data structure returned while executing this project. I want details about the LLM analysis included. Keep returning the current values which aggregate all metrics fetched or built from multiple PRs. Also include a new "details" field at the top level. It should be an array with details about each PR (same metrics than aggregated) and also contain the LLM analysis result.

### Prompt 40
Did you include an average LLM score to the aggregated metrics section? If not, do it.

### Prompt 41
Update README.me with the latest changes. Remember it should only be focused on running the CLI.
