// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

import * as cp from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import * as vscode from 'vscode';

export interface Contradiction {
    id: string;
    caller_symbol: string;
    caller_package: string;
    location: string;
    unapproved_target: string;
    invocation_count: number;
    last_evaluated_at: string;
}

export interface BlastRadiusResult {
    target_symbol: string;
    upstream_callers: Array<{ id: string; name: string; kind: string; package: string; repo: string }> | null;
    upstream_callers_count: number;
    downstream_deps: Array<{ id: string; name: string; kind: string; package: string; repo: string }> | null;
    downstream_deps_count: number;
}

export interface RuntimeState {
    block_height: number;
    contradicted_claims: number;
    runtime_root_hash: string;
    snapshot_hash: string;
    static_root_hash: string;
    status: string;
    total_entities: number;
    total_static_claims: number;
    verified_claims: number;
}

export class GarudaMCPClient {
    private getExecutablePath(): string {
        const config = vscode.workspace.getConfiguration('garuda');
        const configuredPath = config.get<string>('executablePath');
        if (configuredPath && configuredPath !== 'garuda') {
            return configuredPath;
        }

        // Auto-detect ./bin/garuda in open workspace folder
        const workspaceFolders = vscode.workspace.workspaceFolders;
        if (workspaceFolders && workspaceFolders.length > 0) {
            const localBin = path.join(workspaceFolders[0].uri.fsPath, 'bin', 'garuda');
            if (fs.existsSync(localBin)) {
                return localBin;
            }
        }

        // Fallback to /usr/local/bin/garuda or system PATH
        if (fs.existsSync('/usr/local/bin/garuda')) {
            return '/usr/local/bin/garuda';
        }

        return 'garuda';
    }

    private getDatabaseUrl(): string {
        const config = vscode.workspace.getConfiguration('garuda');
        return config.get<string>('databaseUrl') || process.env.DATABASE_URL || 'postgres://test:test@localhost:5433/garuda_test?sslmode=disable';
    }

    private executeRpc<T>(toolName: string, args: Record<string, any>): Promise<T> {
        return new Promise((resolve, reject) => {
            const execPath = this.getExecutablePath();
            const dbUrl = this.getDatabaseUrl();

            const rpcRequest = {
                jsonrpc: '2.0',
                id: Date.now(),
                method: 'tools/call',
                params: {
                    name: toolName,
                    arguments: args
                }
            };

            const proc = cp.spawn(execPath, ['mcp'], {
                env: {
                    ...process.env,
                    DATABASE_URL: dbUrl
                }
            });

            let stdout = '';
            let stderr = '';

            proc.stdout.on('data', (data) => {
                stdout += data.toString();
            });

            proc.stderr.on('data', (data) => {
                stderr += data.toString();
            });

            proc.on('error', (err) => {
                reject(new Error(`Failed to spawn Garuda binary at '${execPath}': ${err.message}`));
            });

            proc.on('close', (code) => {
                if (code !== 0) {
                    return reject(new Error(`MCP process exited with code ${code}: ${stderr || stdout}`));
                }

                try {
                    const response = JSON.parse(stdout.trim());
                    if (response.error) {
                        return reject(new Error(response.error.message || 'RPC Error'));
                    }
                    const textContent = response.result?.content?.[0]?.text;
                    if (!textContent) {
                        return reject(new Error('Empty RPC response payload'));
                    }
                    const parsedData = JSON.parse(textContent);
                    resolve(parsedData as T);
                } catch (err) {
                    reject(new Error(`Failed to parse MCP JSON-RPC output: ${err} (Raw: ${stdout})`));
                }
            });

            proc.stdin.write(JSON.stringify(rpcRequest) + '\n');
            proc.stdin.end();
        });
    }

    public async getContradictions(limit: number = 50): Promise<Contradiction[]> {
        return this.executeRpc<Contradiction[]>('get_contradictions', { limit });
    }

    public async getBlastRadius(symbol: string, maxDepth: number = 3): Promise<BlastRadiusResult> {
        return this.executeRpc<BlastRadiusResult>('get_blast_radius', { symbol, max_depth: maxDepth });
    }

    public async getRuntimeState(): Promise<RuntimeState> {
        return this.executeRpc<RuntimeState>('get_runtime_state', {});
    }
}
