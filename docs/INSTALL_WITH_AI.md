# Install Garuda with an AI assistant

You can ask ChatGPT, Claude, Codex or Cursor to install Garuda for you. Copy one prompt, paste it, and the assistant reads our official instructions and walks you through.

What the assistant can do depends on the tool:

| Tool | What it can usually do | What you may still do yourself |
| :--- | :--- | :--- |
| **Cursor, Codex, Claude Code** (run commands on your machine) | Check your system, run the installer, check that it works | Approve each command it asks to run |
| **ChatGPT, Claude** (chat apps) | Find the right download for your system and give you a link and the exact commands | Run the commands in your own terminal |

Download links, file creation and running commands depend on the assistant's tools and your permissions. The [official GitHub release](https://github.com/myshra777-ai/garuda/releases/latest) is always the source of truth.

## The prompt (works in all four)

Copy everything in the box.

```text
Help me install Garuda, an open-source change-control tool for AI-written code.

Use ONLY these official sources, in this order:
1. https://github.com/myshra777-ai/garuda/blob/main/INSTALL.md
2. The latest release: https://github.com/myshra777-ai/garuda/releases/latest
3. The installers in that repository: install.sh (macOS, Linux, ChromeOS, Termux) and install.ps1 (Windows)

Steps:
1. Ask me which operating system and CPU I have, or detect them if you can run commands.
2. Give me the install command for my system from INSTALL.md. If you can, also give a direct link to the right release file for my system and tell me to check it against SHA256SUMS.txt.
3. If I only want to try it without a database, tell me: npx garuda-guard demo (needs Node 18 or newer).
4. After installing, have me run "garuda --version" and the health check from INSTALL.md, and tell me what a good result looks like.

Rules:
- Do not invent download links or commands. If something is not in the official sources, say so.
- Do not use unofficial builds or change Garuda's source code.
- If you can run commands, show me each command and ask before running anything that needs administrator rights or changes system settings.
- If you find a script that is piped into a shell, tell me I can read it first.
```

## Tool notes

- **ChatGPT.** Turn on web browsing or search so it can open the repository. If it offers a download link, check that the address starts with `https://github.com/myshra777-ai/garuda/releases/`.
- **Claude.** Same prompt. In Claude Code it can run the installer for you after you approve the command.
- **Codex and Cursor.** Open a terminal or an empty folder first. Add this line to the prompt: *"Run the commands for me, and ask before anything needs administrator rights."*

## Check what the assistant gave you

An assistant can still make a mistake. Before you run anything:

1. The download address must start with `https://github.com/myshra777-ai/garuda/`.
2. The commands must match [INSTALL.md](../INSTALL.md).
3. For a release file, compare its checksum with `SHA256SUMS.txt` on the same release page (`sha256sum <file>` on macOS and Linux, `Get-FileHash <file>` on Windows).

## One-line installs, without an assistant

```bash
npx garuda-guard demo                       # try it, no database, needs Node 18+
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash   # macOS, Linux, ChromeOS, Termux
```

```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex   # Windows
```

See [INSTALL.md](../INSTALL.md) for what each installer does and for fixes to common problems.
