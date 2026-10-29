# typed: false
# frozen_string_literal: true

class Garuda < Formula
  desc "Local-first AI code change-control gate, active agent harness (Talon), financial audit & 35-tool MCP engine"
  homepage "https://github.com/myshra777-ai/garuda"
  version "1.9.1"
  license "Apache-2.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda-v1.9.1-darwin-arm64.tar.gz"
    else
      url "https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda-v1.9.1-darwin-amd64.tar.gz"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda-v1.9.1-linux-arm64.tar.gz"
    else
      url "https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda-v1.9.1-linux-amd64.tar.gz"
    end
  end

  def install
    bin.install "bin/garuda"
    bin.install "bin/garuda-api"
    bin.install "bin/garuda-mcp"
    bin.install "bin/garuda-talon"
    pkgshare.install "migrations", "policies", "setup.sh", "install.sh"
    doc.install "README.md", "INSTALL.md", "PLAYBOOK.md", "CHANGELOG.md", "RELEASE_NOTES.md"
  end

  test do
    assert_match "1.9.1", shell_output("#{bin}/garuda --version")
  end
end
