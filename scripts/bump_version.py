#!/usr/bin/env python3
"""
Garuda AI Version Bumper
Automatically bumps version based on commit messages using semantic versioning.
"""

import os
import sys
import subprocess
import re
from datetime import datetime
import google.generativeai as genai

VERSION_FILE = "version.txt"
GO_MOD_PATH = "go.mod"

def get_current_version():
    """Get current version from version.txt or go.mod."""
    # Try version.txt first
    if os.path.exists(VERSION_FILE):
        with open(VERSION_FILE, "r") as f:
            return f.read().strip()
    
    # Try git tag
    try:
        result = subprocess.run(
            ["git", "describe", "--tags", "--abbrev=0"],
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout.strip()
    except:
        pass
    
    return "v0.1.0"

def analyze_commits_with_ai(commits):
    """Use AI to determine version bump type."""
    genai.configure(api_key=os.environ.get("GEMINI_API_KEY"))
    model = genai.GenerativeModel('gemini-2.0-flash')
    
    prompt = f"""
Analyze these git commits for Garuda.

Commits:
{commits[:500]}

Determine what type of version bump is needed based on semantic versioning:
- MAJOR (x.0.0): Breaking changes
- MINOR (0.x.0): New features
- PATCH (0.0.x): Bug fixes

Return only one of: major, minor, patch
"""
    
    try:
        response = model.generate_content(prompt)
        result = response.text.strip().lower()
        if "major" in result:
            return "major"
        elif "minor" in result:
            return "minor"
        else:
            return "patch"
    except:
        return "patch"

def get_commits_since_last_tag():
    """Get commits since last tag."""
    try:
        result = subprocess.run(
            ["git", "log", "--oneline", "--no-merges"],
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout.strip()
    except:
        return ""

def bump_version(current, bump_type):
    """Increment version based on bump_type."""
    # Remove 'v' prefix
    if current.startswith('v'):
        current = current[1:]
    
    parts = current.split('.')
    if len(parts) != 3:
        return "v0.1.0"
    
    major, minor, patch = parts
    
    if bump_type == "major":
        major = str(int(major) + 1)
        minor = "0"
        patch = "0"
    elif bump_type == "minor":
        minor = str(int(minor) + 1)
        patch = "0"
    else:  # patch
        patch = str(int(patch) + 1)
    
    return f"v{major}.{minor}.{patch}"

def update_version_file(version):
    """Update version.txt."""
    with open(VERSION_FILE, "w") as f:
        f.write(version)
    print(f"📝 Updated {VERSION_FILE} to {version}")

def update_go_mod(version):
    """Update version in go.mod."""
    if not os.path.exists(GO_MOD_PATH):
        return
    
    with open(GO_MOD_PATH, "r") as f:
        content = f.read()
    
    # Update version line if it exists
    pattern = r'(module\s+\S+)\s+//\s+v[\d.]+'
    if re.search(pattern, content):
        content = re.sub(pattern, r'\1 // ' + version, content)
    else:
        # Add version comment if not exists
        content = content.replace('module github.com/myshra777-ai/garuda', f'module github.com/myshra777-ai/garuda // {version}')
    
    with open(GO_MOD_PATH, "w") as f:
        f.write(content)
    print(f"📝 Updated {GO_MOD_PATH} with version {version}")

def main():
    print("🔢 Garuda AI Version Bumper")
    print("=" * 40)
    
    # Check API key
    if not os.environ.get("GEMINI_API_KEY"):
        print("⚠️ GEMINI_API_KEY not set. Using heuristic for version bump...")
        bump_type = "patch"
    else:
        # Analyze commits
        commits = get_commits_since_last_tag()
        if commits:
            print("🧠 Analyzing commits with AI...")
            bump_type = analyze_commits_with_ai(commits)
        else:
            bump_type = "patch"
    
    print(f"📊 Bump type: {bump_type}")
    
    # Get current version
    current_version = get_current_version()
    print(f"📌 Current version: {current_version}")
    
    # Bump version
    new_version = bump_version(current_version, bump_type)
    print(f"🚀 New version: {new_version}")
    
    # Update files
    update_version_file(new_version)
    update_go_mod(new_version)
    
    # Create git tag
    subprocess.run(["git", "tag", new_version], capture_output=True)
    print(f"✅ Git tag created: {new_version}")
    
    print(f"\n📌 Run: git push origin {new_version}")

if __name__ == "__main__":
    main()