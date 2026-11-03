// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Downloads the Garuda release archive for this machine (once), checks its
// SHA-256 against the release's SHA256SUMS.txt, unpacks it under
// ~/.garuda/npm/<version>/, and runs the requested binary. No dependencies.
'use strict';

const fs = require('fs');
const os = require('os');
const path = require('path');
const crypto = require('crypto');
const cp = require('child_process');

const pkg = require('../package.json');
const TAG = 'v' + String(pkg.version).replace(/^v/, '');
const BASE = (process.env.GARUDA_RELEASE_BASE ||
  `https://github.com/myshra777-ai/garuda/releases/download/${TAG}`).replace(/\/+$/, '');

function fail(msg) {
  process.stderr.write(`garuda: ${msg}\n`);
  process.exit(1);
}

// Maps this machine to the archive name that scripts/build_release_matrix.sh publishes.
function archiveName(platform = process.platform, arch = process.arch) {
  const cpu = { x64: 'amd64', arm64: 'arm64' }[arch];
  if (!cpu) return null;
  if (platform === 'android') {
    return cpu === 'arm64' ? `garuda-${TAG}-android-arm64-termux.tar.gz` : null;
  }
  const osName = { linux: 'linux', darwin: 'darwin', win32: 'windows', freebsd: 'freebsd' }[platform];
  if (!osName) return null;
  if (osName === 'freebsd' && cpu !== 'amd64') return null;
  return `garuda-${TAG}-${osName}-${cpu}` + (osName === 'windows' ? '.zip' : '.tar.gz');
}

async function download(url) {
  let res;
  try {
    res = await fetch(url, { redirect: 'follow' });
  } catch (e) {
    fail(`could not reach ${url}: ${e.message}\n  Check your network or proxy, or install from https://github.com/myshra777-ai/garuda#readme`);
  }
  if (!res.ok) {
    fail(`download failed (${res.status}) for ${url}\n  This release may not be published yet: https://github.com/myshra777-ai/garuda/releases`);
  }
  return Buffer.from(await res.arrayBuffer());
}

function expectedHash(sums, name) {
  for (const line of sums.split(/\r?\n/)) {
    const m = line.trim().match(/^([0-9a-fA-F]{64})\s+\*?(.+)$/);
    if (m && path.basename(m[2]) === name) return m[1].toLowerCase();
  }
  return null;
}

async function ensureInstalled() {
  const name = archiveName();
  if (!name) fail(`no prebuilt Garuda binary for ${process.platform}/${process.arch}. See https://github.com/myshra777-ai/garuda/blob/main/INSTALL.md`);

  const root = path.join(os.homedir(), '.garuda', 'npm', TAG);
  const stem = name.replace(/\.(tar\.gz|zip)$/, '');
  const binDir = path.join(root, stem, 'bin');
  const ext = process.platform === 'win32' ? '.exe' : '';
  if (fs.existsSync(path.join(binDir, 'garuda' + ext))) return binDir;

  process.stderr.write(`garuda: first run, downloading ${name} (one time)...\n`);
  const archive = await download(`${BASE}/${name}`);
  const sums = (await download(`${BASE}/SHA256SUMS.txt`)).toString('utf8');
  const want = expectedHash(sums, name);
  if (!want) fail(`SHA256SUMS.txt has no entry for ${name}; refusing to run an unverified download.`);
  const got = crypto.createHash('sha256').update(archive).digest('hex');
  if (got !== want) fail(`checksum mismatch for ${name} (expected ${want}, got ${got}); refusing to run it.`);

  fs.mkdirSync(root, { recursive: true });
  const tmp = path.join(root, `${name}.part`);
  fs.writeFileSync(tmp, archive);
  const r = cp.spawnSync('tar', ['-xf', tmp, '-C', root], { stdio: 'inherit' });
  fs.rmSync(tmp, { force: true });
  if (r.error || r.status !== 0) fail('could not unpack the download: the "tar" command is required (included with Windows 10+, macOS and Linux).');
  if (!fs.existsSync(path.join(binDir, 'garuda' + ext))) fail('the archive did not contain bin/garuda; please report this at https://github.com/myshra777-ai/garuda/issues');
  return binDir;
}

async function main(binary) {
  const binDir = await ensureInstalled();
  const ext = process.platform === 'win32' ? '.exe' : '';
  const exe = path.join(binDir, binary + ext);
  if (process.platform !== 'win32') { try { fs.chmodSync(exe, 0o755); } catch (_) { /* best effort */ } }
  const r = cp.spawnSync(exe, process.argv.slice(2), { stdio: 'inherit' });
  if (r.error) fail(`could not start ${exe}: ${r.error.message}`);
  process.exit(r.status === null ? 1 : r.status);
}

module.exports = { main: (b) => main(b).catch((e) => fail(e.message)), archiveName, expectedHash };
