"use strict";
// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.GarudaMCPClient = void 0;
const cp = __importStar(require("child_process"));
const fs = __importStar(require("fs"));
const path = __importStar(require("path"));
const vscode = __importStar(require("vscode"));
class GarudaMCPClient {
    getExecutablePath() {
        const config = vscode.workspace.getConfiguration('garuda');
        const configuredPath = config.get('executablePath');
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
    getDatabaseUrl() {
        const config = vscode.workspace.getConfiguration('garuda');
        return config.get('databaseUrl') || process.env.DATABASE_URL || 'postgres://test:test@localhost:5433/garuda_test?sslmode=disable';
    }
    executeRpc(toolName, args) {
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
                    resolve(parsedData);
                }
                catch (err) {
                    reject(new Error(`Failed to parse MCP JSON-RPC output: ${err} (Raw: ${stdout})`));
                }
            });
            proc.stdin.write(JSON.stringify(rpcRequest) + '\n');
            proc.stdin.end();
        });
    }
    async getContradictions(limit = 50) {
        return this.executeRpc('get_contradictions', { limit });
    }
    async getBlastRadius(symbol, maxDepth = 3) {
        return this.executeRpc('get_blast_radius', { symbol, max_depth: maxDepth });
    }
    async getRuntimeState() {
        return this.executeRpc('get_runtime_state', {});
    }
}
exports.GarudaMCPClient = GarudaMCPClient;
//# sourceMappingURL=mcpClient.js.map