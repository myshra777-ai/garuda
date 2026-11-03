# Plans and pricing (draft, nothing here is for sale yet)

This page records how Garuda is meant to be offered. **There is no billing system and no paid plan today.** Everything that ships in this repository is free under Apache-2.0. Items marked *planned* are not built. Prices are not set; they will be decided with early customers.

## The principle

The protection is free. The coordination around it, across a team or a company, is what will be paid.

| | Community (today) | Team (planned) | Enterprise (planned) |
| :--- | :--- | :--- | :--- |
| Who | Individual developers, open-source projects, small self-run teams | AI-heavy engineering teams | Large or regulated organizations |
| Price | Free, Apache-2.0 | To be decided with design partners | Custom contract |
| Code analysis, dependency graph, change impact (Go, Python, TypeScript) | Yes | Yes | Yes |
| Pre-edit briefing, post-edit checks, Talon, local policies and templates | Yes | Yes | Yes |
| MCP server, skills, plugins, CLI, GitHub Action, local dashboard | Yes | Yes | Yes |
| Tamper-evident evidence and local verification | Yes | Yes | Yes |
| Central policy management and signed policy rollout | No | Planned | Planned |
| Approval and exception workflow | Local hooks only | Planned | Planned |
| Organization dashboard and reporting | No | Planned | Planned |
| Managed evidence retention and compliance exports | Local records, `garuda.audit_export` | Planned | Planned |
| Managed hosting | No | Planned | Planned |
| SSO / SAML / OIDC, SCIM, advanced roles | No | No | Planned |
| Dedicated deployment, SLA, onboarding | No | Priority support | Planned |

## Commitments

- Basic policy enforcement, local Talon, the analyzer, the agent integrations and local evidence verification **stay free**.
- An expired or missing subscription will never switch off local checks or delete local evidence.
- Source code is not uploaded to any Garuda service by default.
- Paid services will live in a separate, private repository that uses released versions of this engine. The engine stays open.

## Available now

Help with a first rulebook, an architecture scan of your repository, and setup support. [Open an issue](https://github.com/myshra777-ai/garuda/issues/new/choose) to ask.
