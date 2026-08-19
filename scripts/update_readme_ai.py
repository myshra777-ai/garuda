#!/usr/bin/env python3
import os
import json
import argparse
import google.genai as genai

def load_context(path):
    with open(path) as f:
        return json.load(f)

def generate_readme(context):
    header = """# 🦅 Garuda

> Evidence-backed Company Brain for AI-native engineering

[![CI](https://github.com/myshra777-ai/garuda/actions/workflows/garuda-ci.yml/badge.svg)](https://github.com/myshra777-ai/garuda/actions/workflows/garuda-ci.yml)
[![Go Reference](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
"""

    prompt = f"""You are generating the README for Garuda.

IMPORTANT RULES:
1. The supplied documentation context is authoritative.
2. Never invent capabilities, commands, APIs, benchmarks, supported languages, or implementation status.
3. Never describe a planned capability as implemented.
4. Never infer that something is unsupported merely because it is absent from the context. Use "not currently documented" or "unknown" where appropriate.
5. Preserve the product positioning and thesis.
6. Use precise technical language.
7. Prefer evidence and measurable facts over marketing language.
8. Do not claim SOC 2, HIPAA, enterprise compliance, production readiness, or benchmark performance unless explicitly present in the supplied context.
9. The roadmap is a roadmap, not a list of current capabilities.

Generate the README using this exact section structure:

1. Why Garuda
2. What Garuda does
3. 60-second example
4. Core capabilities (list from capabilities.yaml)
5. Semantic model
6. Architecture
7. CLI (list commands)
8. Installation
9. Quick start
10. Example output
11. Evidence and trust model
12. Current status
13. Roadmap
14. Benchmarks
15. Security
16. Documentation
17. Development
18. Contributing
19. License

Documentation context (JSON):
{json.dumps(context, indent=2)}

Generate the README content in Markdown. Do not include the header section (the one with the title, badges, etc.) because that will be added separately.
"""

    client = genai.Client(api_key=os.environ['GEMINI_API_KEY'])
    response = client.models.generate_content(
        model='gemini-3.6-flash',
        contents=prompt
    )
    body = response.text.strip()
    return header + "\n" + body

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--context", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()
    context = load_context(args.context)
    readme = generate_readme(context)
    with open(args.output, 'w') as f:
        f.write(readme)

if __name__ == "__main__":
    main()