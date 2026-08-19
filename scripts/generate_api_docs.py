#!/usr/bin/env python3
import os
import json
import argparse
import google.genai as genai

def load_context(path):
    with open(path) as f:
        return json.load(f)

def generate_api_docs(context):
    commands = context.get('cli', {}).get('commands', [])
    capabilities = context.get('capabilities', {})
    stable_caps = capabilities.get('stable', [])
    planned_caps = capabilities.get('planned', [])

    prompt = f"""You are generating API documentation for Garuda CLI.

The following commands are available according to the documentation context:

{json.dumps(commands, indent=2)}

The stable capabilities include: {', '.join(stable_caps)}
Planned capabilities: {', '.join(planned_caps)}

Produce a Markdown document with a section for each command. For each command, include:
- Name and description
- Usage example
- Flags (if any, you can infer some common flags like --output, --save, etc.)
- Example output

Do not invent commands that are not listed. If a command is planned but not yet implemented, mark it as "(planned)".
The output should be suitable for docs/API.md.
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
    api_docs = generate_api_docs(context)
    with open(args.output, 'w') as f:
        f.write(api_docs)

if __name__ == "__main__":
    main()