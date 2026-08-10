#!/usr/bin/env python3
"""
Garuda AI API Docs Generator
Generates API documentation from OpenAPI spec and code comments.
"""

import os
import sys
import json
import yaml
import subprocess
import google.generativeai as genai

OPENAPI_PATH = "openapi.yaml"
DOCS_PATH = "docs/API.md"

def load_openapi():
    """Load OpenAPI specification."""
    if os.path.exists(OPENAPI_PATH):
        with open(OPENAPI_PATH, "r") as f:
            if OPENAPI_PATH.endswith(".yaml") or OPENAPI_PATH.endswith(".yml"):
                return yaml.safe_load(f)
            else:
                return json.load(f)
    return None

def generate_docs_with_ai(spec):
    """Use AI to generate API documentation."""
    genai.configure(api_key=os.environ.get("GEMINI_API_KEY"))
    model = genai.GenerativeModel('gemini-2.0-flash')
    
    # Extract key endpoints and schemas
    endpoints = []
    if "paths" in spec:
        for path, methods in spec["paths"].items():
            for method, details in methods.items():
                if method in ["get", "post", "put", "delete"]:
                    endpoints.append({
                        "path": path,
                        "method": method.upper(),
                        "summary": details.get("summary", ""),
                        "description": details.get("description", "")
                    })
    
    prompt = f"""
Generate API documentation for Garuda.

Endpoints:
{json.dumps(endpoints[:10], indent=2)}

Generate a clean, readable API documentation in Markdown format with:
1. Overview of the API
2. Authentication
3. Endpoint table (method, path, description)
4. Request/response examples

Make it developer-friendly and concise.
"""
    
    try:
        response = model.generate_content(prompt)
        return response.text.strip()
    except Exception as e:
        print(f"❌ AI generation failed: {e}")
        return ""

def update_api_docs(content):
    """Write API documentation to docs/API.md."""
    os.makedirs("docs", exist_ok=True)
    
    with open(DOCS_PATH, "w") as f:
        f.write(f"""# Garuda API Reference

{content}

---
_Generated from OpenAPI spec and code comments on {__import__('datetime').datetime.now().strftime('%Y-%m-%d %H:%M:%S')}_
""")
    print(f"✅ API docs written to {DOCS_PATH}")

def main():
    print("📚 Garuda AI API Docs Generator")
    print("=" * 40)
    
    # Load OpenAPI spec
    spec = load_openapi()
    if not spec:
        print("❌ OpenAPI spec not found")
        sys.exit(1)
    
    print(f"📖 Loaded OpenAPI spec from {OPENAPI_PATH}")
    
    # Generate docs
    if os.environ.get("GEMINI_API_KEY"):
        print("🧠 Generating API docs with AI...")
        content = generate_docs_with_ai(spec)
    else:
        print("⚠️ GEMINI_API_KEY not set. Using basic template.")
        content = "## API Overview\n\nSee OpenAPI spec for details."
    
    # Write docs
    update_api_docs(content)
    print("✅ API documentation generated!")

if __name__ == "__main__":
    main()