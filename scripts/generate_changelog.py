#!/usr/bin/env python3
"""
Garuda AI Changelog Generator
Generates CHANGELOG.md from git history using AI for summarization.
"""

import os
import sys
import subprocess
import json
import re
from datetime import datetime
import google.generativeai as genai

# Configuration
CHANGELOG_PATH = "CHANGELOG.md"
COMMITS_PER_RELEASE = 10

def get_commits_since_last_tag():
    """Get all commits since the last tag."""
    try:
        # Get the last tag
        tag_result = subprocess.run(
            ["git", "describe", "--tags", "--abbrev=0", "--always"],
            capture_output=True,
            text=True,
            check=True
        )
        last_tag = tag_result.stdout.strip()
        print(f"🔖 Last tag: {last_tag}")
    except subprocess.CalledProcessError:
        last_tag = None
        print("📌 No existing tags found. Starting from first commit.")

    # Get commits since last tag
    if last_tag:
        result = subprocess.run(
            ["git", "log", "--oneline", f"{last_tag}..HEAD"],
            capture_output=True,
            text=True,
            check=True
        )
    else:
        result = subprocess.run(
            ["git", "log", "--oneline", "--all"],
            capture_output=True,
            text=True,
            check=True
        )
    
    commits = result.stdout.strip().split("\n")
    commits = [c for c in commits if c]  # Remove empty lines
    return commits, last_tag

def categorize_commits(commits):
    """Categorize commits by type (feat, fix, docs, chore, etc.)."""
    categories = {
        "feat": [],
        "fix": [],
        "docs": [],
        "style": [],
        "refactor": [],
        "perf": [],
        "test": [],
        "chore": [],
        "other": []
    }
    
    for commit in commits:
        # Extract commit type from conventional commit format
        # e.g., "feat: add new endpoint" -> category: feat
        parts = commit.split(" ", 1)
        if len(parts) > 1 and ":" in parts[0]:
            msg = parts[1]
            type_part = parts[0].lower()
            if type_part.startswith("feat"):
                categories["feat"].append(commit)
            elif type_part.startswith("fix"):
                categories["fix"].append(commit)
            elif type_part.startswith("docs"):
                categories["docs"].append(commit)
            elif type_part.startswith("style"):
                categories["style"].append(commit)
            elif type_part.startswith("refactor"):
                categories["refactor"].append(commit)
            elif type_part.startswith("perf"):
                categories["perf"].append(commit)
            elif type_part.startswith("test"):
                categories["test"].append(commit)
            elif type_part.startswith("chore"):
                categories["chore"].append(commit)
            else:
                categories["other"].append(commit)
        else:
            categories["other"].append(commit)
    
    return categories

def generate_summary_with_ai(categories):
    """Use AI to generate a release summary."""
    genai.configure(api_key=os.environ.get("GEMINI_API_KEY"))
    model = genai.GenerativeModel('gemini-2.0-flash')
    
    # Build context
    summary_prompt = f"""
You are generating release notes for the Garuda project – an Organizational Intelligence Runtime for AI agents.

Here are the commits from this release:

Feat (new features):
{chr(10).join(categories["feat"])[:1500]}

Fixes:
{chr(10).join(categories["fix"])[:1500]}

Docs:
{chr(10).join(categories["docs"])[:500]}

Refactors:
{chr(10).join(categories["refactor"])[:500]}

Generate a concise, professional release summary (2-3 paragraphs) highlighting:
1. The major new features in this release
2. Key improvements or fixes
3. Any breaking changes (if present)

Make it suitable for CHANGELOG.md.
"""
    
    try:
        response = model.generate_content(summary_prompt)
        return response.text.strip()
    except Exception as e:
        print(f"❌ AI generation failed: {e}")
        return ""

def get_version_from_go_mod():
    """Extract version from go.mod or git tags."""
    try:
        result = subprocess.run(
            ["git", "describe", "--tags", "--abbrev=0"],
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout.strip()
    except:
        # Fallback: read go.mod for module path
        try:
            with open("go.mod", "r") as f:
                content = f.read()
                match = re.search(r'module\s+(\S+)', content)
                if match:
                    # Try to get latest version from git
                    return "v0.1.0"
        except:
            pass
        return "v0.1.0"

def update_changelog(version, summary, categories):
    """Append new release to CHANGELOG.md."""
    now = datetime.now().strftime("%Y-%m-%d")
    
    entry = f"""
## [{version}] - {now}

{summary}

### ✨ New Features
{chr(10).join(['- ' + c for c in categories["feat"][:5]]) if categories["feat"] else '- None'}

### 🐛 Bug Fixes
{chr(10).join(['- ' + c for c in categories["fix"][:5]]) if categories["fix"] else '- None'}

### 📚 Documentation
{chr(10).join(['- ' + c for c in categories["docs"][:5]]) if categories["docs"] else '- None'}

### 🔧 Refactors
{chr(10).join(['- ' + c for c in categories["refactor"][:5]]) if categories["refactor"] else '- None'}

---
"""
    
    # Check if CHANGELOG exists
    if os.path.exists(CHANGELOG_PATH):
        with open(CHANGELOG_PATH, "r", encoding="utf-8") as f:
            content = f.read()
        
        # Insert at the top (after the title)
        title_pattern = r"# Changelog\n\n"
        if re.search(title_pattern, content):
            content = re.sub(title_pattern, f"# Changelog\n\n{entry}", content)
        else:
            content = entry + content
    else:
        content = f"""# Changelog

All notable changes to Garuda will be documented in this file.

{entry}
"""
    
    with open(CHANGELOG_PATH, "w", encoding="utf-8") as f:
        f.write(content)

def main():
    print("📝 Garuda AI Changelog Generator")
    print("=" * 40)
    
    # Check API key
    if not os.environ.get("GEMINI_API_KEY"):
        print("❌ GEMINI_API_KEY environment variable not set")
        sys.exit(1)
    
    # Get commits
    commits, last_tag = get_commits_since_last_tag()
    if not commits:
        print("ℹ️ No new commits since last tag")
        sys.exit(0)
    
    print(f"📦 Found {len(commits)} commits since {last_tag or 'start'}")
    
    # Categorize commits
    categories = categorize_commits(commits)
    print(f"📊 Categories: {len(categories['feat'])} features, {len(categories['fix'])} fixes")
    
    # Generate summary
    print("🧠 Generating release summary with AI...")
    summary = generate_summary_with_ai(categories)
    
    if not summary:
        summary = "This release includes new features, bug fixes, and improvements."
    
    # Get version
    version = get_version_from_go_mod()
    new_version = increment_version(version)
    
    # Update changelog
    print(f"📝 Writing CHANGELOG.md for version {new_version}...")
    update_changelog(new_version, summary, categories)
    
    print(f"✅ CHANGELOG.md updated with version {new_version}!")
    print(f"📌 Run: git tag {new_version}")

def increment_version(version):
    """Increment patch version (simple) or use conventional commit type."""
    # Remove 'v' prefix
    if version.startswith('v'):
        version = version[1:]
    
    parts = version.split('.')
    if len(parts) == 3:
        # Increment patch
        parts[2] = str(int(parts[2]) + 1)
        return 'v' + '.'.join(parts)
    else:
        return 'v0.1.0'

if __name__ == "__main__":
    main()