#!/usr/bin/env python3
import os
import json
import argparse
import subprocess
import google.genai as genai

def get_commits_since_last_tag():
    try:
        tag = subprocess.check_output(['git', 'describe', '--tags', '--abbrev=0'], text=True).strip()
        since = tag
    except:
        since = 'HEAD~50'
    log_cmd = ['git', 'log', f'{since}..HEAD', '--pretty=format:%h|%s|%an|%ad', '--date=short']
    output = subprocess.check_output(log_cmd, text=True)
    commits = []
    for line in output.strip().split('\n'):
        if not line:
            continue
        sha, msg, author, date = line.split('|', 3)
        commits.append({'sha': sha, 'message': msg, 'author': author, 'date': date})
    return commits, since

def load_context(path):
    with open(path) as f:
        return json.load(f)

def generate_changelog(context):
    commits, since = get_commits_since_last_tag()
    prompt = f"""You are generating the CHANGELOG for Garuda.

The last release was based on commit/tag: {since}.
The following commits are new since then:

Commits (sha, message, author, date):
{json.dumps(commits, indent=2)}

The current product context is:
{json.dumps(context, indent=2)}

Generate a CHANGELOG in the typical format:

## [Unreleased]

### Added
- ...

### Changed
- ...

### Fixed
- ...

### Removed
- ...

### Semantic Changes
- ... (if any, derived from the commits)

Do not invent capabilities that are not mentioned. Use the commits as the primary source.
"""
    client = genai.Client(api_key=os.environ['GEMINI_API_KEY'])
    response = client.models.generate_content(
        model='gemini-3.6-flash',
        contents=prompt
    )
    return response.text.strip()

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--context", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()
    context = load_context(args.context)
    changelog = generate_changelog(context)
    with open(args.output, 'w') as f:
        f.write(changelog)

if __name__ == "__main__":
    main()