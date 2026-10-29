Name:           garuda
Version:        1.9.1
Release:        1%{?dist}
Summary:        Local-first AI change-control gate, Garuda Talon harness, financial audit & 35-tool MCP engine
License:        Apache-2.0
URL:            https://github.com/myshra777-ai/garuda
Source0:        https://github.com/myshra777-ai/garuda/releases/download/v%{version}/garuda-v%{version}-linux-amd64.tar.gz
BuildArch:      x86_64 aarch64

%description
Garuda v1.9.1 combines compiler-grade AST analysis, active AI agent guardrails (Garuda Talon),
human-in-the-loop 0.0-1.0 formula weighting, 3-tier surface criticality, continuous financial
discrepancy auditing (48 Cr vs 4.8 Cr), and RFC 6962 cryptographic Merkle proofs.

%prep
%setup -q -n garuda-v%{version}-linux-amd64

%install
mkdir -p %{buildroot}%{_bindir}
install -m 0755 bin/garuda %{buildroot}%{_bindir}/garuda
install -m 0755 bin/garuda-api %{buildroot}%{_bindir}/garuda-api
install -m 0755 bin/garuda-mcp %{buildroot}%{_bindir}/garuda-mcp
install -m 0755 bin/garuda-talon %{buildroot}%{_bindir}/garuda-talon

%files
%license LICENSE
%doc README.md INSTALL.md PLAYBOOK.md CHANGELOG.md RELEASE_NOTES.md
%{_bindir}/garuda
%{_bindir}/garuda-api
%{_bindir}/garuda-mcp
%{_bindir}/garuda-talon
