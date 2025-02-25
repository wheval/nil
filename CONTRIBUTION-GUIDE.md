# Contribution Guide

Thank you for your interest in contributing to this project! We welcome contributions from the community and appreciate your help in improving the project. We highly encourage meaningful pull requests (PRs) that include bug fixes, performance optimizations, or functional improvements. Additionally, opening issues to report critical bugs or suggest valuable improvements is just as important as submitting PRs.

_Because maintainers need to sign your commits, PRs with typo fixes are generally not favorable. Such PRs will be closed and replaced with an issue instead._

## Repository Structure

Understanding the project structure will help you navigate and contribute effectively. Below are the key directories and their purposes:

- **`nil/`** – Contains the core blockchain logic, including consensus mechanisms, networking protocols, cryptographic primitives, and state management functionalities

- **`niljs/`** – A TypeScript client library providing tools and abstractions for interacting with the nil blockchain

- **`clijs/`** – Contains the command-line interface tools for interacting with the nil blockchain, facilitating tasks such as node management, transaction submission, and network monitoring

- **`explorer_frontend/`** – Develops the user interface for the blockchain explorer, enabling users to visualize blockchain data, track transactions, and monitor network status. This folder also contains the playground which is an in-browser IDE to interact with the =nil; blockchain.

- **`wallet-extension`** – Contains the code to the extension that provides essential tools for interacting with the network or connecting with decentralized applications (dApps)

- **`create-nil-hardhat-project/`** – Provides a template or scaffolding for developers to quickly set up a Hardhat project tailored for the nil blockchain, streamlining the development and deployment of smart contracts

- **`docs/`** – Contains comprehensive documentation, including protocol specifications, developer guides, API references, and integration tutorials

- **`explorer_backend/`** – Manages the backend services for the blockchain explorer, including data indexing, API endpoints, and real-time data processing

- **`smart-contracts/`** – Hosts the [@nilfoundation/smart-contracts](https://www.npmjs.com/package/@nilfoundation/smart-contracts) package, comprising essential smart contracts for blockchain interactions

(_Note: Some directories like **`nix`** are omitted for brevity, but contributors are encouraged to explore the full repository if necessary._)

## How to Contribute

This section describes in some detail how changes can be made and proposed with pull requests.

### 1. Check Existing Issues

Before starting work on a feature or bug fix, check existing [issues](https://github.com/NilFoundation/nil/issues). If none exist, consider opening an issue first to discuss your approach. This ensures alignment with project maintainers and avoids duplicate work.

### 2. Fork the Repository

- Click the **Fork** button on the repository's page
- Clone your forked repository to your local machine:
  ```sh
  git clone https://github.com/your-username/nil.git
  ```
- Navigate to the project directory:
  ```sh
  cd nil
  ```

### 3. Create a Branch

- Create a new branch for your feature or bug fix:
  ```sh
  git checkout -b feature-branch-name
  ```

### 4. Make Changes

- Implement your changes following the project’s coding style
- Ensure that all functional changes include appropriate tests
- Test your changes thoroughly before submitting
- Commit and sign your changes with a descriptive message:
  ```sh
  git add .
  git commit -S -m "Add feature or fix description"
  ```

### 5. Push and Create a Pull Request

- Push your changes to your forked repository:
  ```sh
  git push origin feature-branch-name
  ```
- Open a pull request from your forked repository to the main repository
- Provide a clear description of your changes and the motivation behind them
- Once you open a PR, assign yourself to it and add the respective label

## Contribution Guidelines

### PR Review and Merging Process

- **PRs require approval and will be merged by a team member after signing, not directly by the author**. Maintainers will review and rebase them before merging
- PRs must be signed by maintainers before merging
- **The original authorship will be preserved**, but GPG signatures will be stripped, and the committer field will be changed
- Direct merges from `master` are not allowed. Instead, PRs should be rebased if conflicts arise to maintain a clean and linear commit history. This approach helps avoid unnecessary merge commits, makes debugging easier, improves change tracking for reviews, reduces potential merge conflicts, and keeps the repository history structured and easy to navigate

### Commit Message Guidelines

Commit messages should be concise and clear. Keep them as a single line summarizing the change.
Example:

```sh
Optimize block validation performance
```

If necessary, add more details in the PR description instead of the commit message.

### Squashing Commits

To keep the commit history clean, you should squash unnecessary commits before opening a PR. **Small fixes such as codestyle or typos should be squashed**, but having multiple well-structured commits in a PR is acceptable.
To squash commits interactively:

```sh
# Interactive rebase to squash commits
git rebase -i HEAD~n
```

Replace `n` with the number of commits you want to squash. Mark the first commit as `pick` and change the others to `squash` (`s`). Then, modify the commit message as needed.

Alternatively, squash all commits and push forcefully:

```sh
git reset HEAD~n --soft
git commit -S -m "Your final commit message"
git push -f
```

### Code Contribution Rules

- Follow the coding style and conventions used in this project
- Keep pull requests small and focused
- **No third-party dependencies should be introduced by external PRs.** Any such change must be discussed and approved before implementation
- Ensure that your changes do not break existing functionality
- Be respectful and considerate in discussions and code reviews
- Describe what the PR does in a clear and concise manner or mention the issue URL you are trying to solve in the PR

### Example PR Description:

```
### [project]: enhance block synchronization efficiency
This PR introduces an optimized peer-to-peer block synchronization mechanism,
reducing network overhead and improving sync speed by 25%.
Key Changes:
- Implemented a more efficient block fetching strategy to minimize redundant data transfer
- Improved peer prioritization logic to prefer lower-latency nodes
- Added a caching layer to streamline block validation and reduce reorg delays
```

_Note:_ The `[project]` in the PR title should match the directory modified. In this example, it corresponds to `nil` as it deals with the core protocol of block synchronization.

## Issue Reporting

- Before opening a new issue, check if it has already been reported
- Provide clear and concise information, including steps to reproduce the issue
- Append tags such as `nil`, `niljs`, `clijs`, `playground`, or `docs`under the labels section for the issue. This will help in categorizing the issues
- **Not sure if it's an issue? Ask the [team](https://t.me/+vtKXTAYAsx4zZGIx)!** We encourage discussion before creating an issue to ensure it's valid and actionable
- If you want to work on an issue, comment on it to avoid duplicate efforts

---

### Note on Contributions

We look forward to any contributions. That being said, we encourage contributions that are well thought out and tested. Thank you for your time and effort.
