# Security Policy

Garuda analyzes source code and stores software-structure, policy, runtime and governance data. Please report suspected vulnerabilities privately so they can be assessed before details become public.

## Reporting a vulnerability

**Do not open a public issue for an unpatched security vulnerability.**

Use GitHub's private vulnerability reporting feature for this repository if it is enabled. If it is unavailable, contact the repository maintainer privately through their GitHub profile and request a secure reporting channel. Do not include secrets, customer source code or sensitive production data in a public discussion.

Please include, when possible:
- A concise description and potential impact.
- A minimal reproduction or proof of concept.
- The affected version or commit.
- Relevant configuration and deployment context.
- Any proposed mitigation.

The project does not currently publish a guaranteed response-time SLA. The maintainer will acknowledge and triage reports as capacity allows. Please coordinate disclosure timing before publishing technical details.

## Scope

Potentially in-scope areas include:
- API authentication, authorization and tenant/workspace isolation.
- CLI, MCP server and agent-integration boundaries.
- Policy evaluation and change-control gates.
- Database access, migrations and persisted governance decisions.
- Merkle proof generation and verification.
- Installation scripts, generated hooks and deployment configuration.
- Sensitive information exposure through logs, API responses or dashboard views.

Third-party dependency vulnerabilities should include the affected dependency and a reproduction or advisory reference.

## Security design and limits

Garuda includes security-related mechanisms such as Ed25519 JWT support, scoped persistence, policy evaluation and Merkle-based integrity verification. Their presence is not a blanket guarantee that every deployment is secure. Configuration, route-level authorization, tenant scoping, secret management, network exposure and installation practices must be reviewed in context.

A Merkle inclusion proof can verify consistency between a record, proof path and supplied root. Unless an independent trusted anchor is used, it does not by itself prove an external timestamp or protect against every threat to the host or root source.

Agent instructions and MCP tools are not, by themselves, security boundaries. Teams relying on Talon or Git/IDE hooks should test bypass behavior, fail-open/fail-closed behavior and CI enforcement in their own configuration.

## Deployment guidance

- Do not use development credentials or sample secrets in a shared or production environment.
- Keep JWT signing keys, database credentials and control-plane tokens out of source control and logs.
- Bind network services only to intended interfaces and apply network-level access controls.
- Review `setup.sh` before running it on a machine with existing services or Docker containers.
- Run the available tests and inspect the current API/route authorization before exposing the service to a network.
- Keep dependencies, database backups and operating-system packages maintained.

## Supported versions

No formal long-term support matrix or security backport SLA is published in this document. For security-sensitive use, prefer a current reviewed commit and check the repository's release notes and open advisories. Do not assume older versions receive backported fixes unless a release explicitly says so.
