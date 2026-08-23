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
exports.activate = activate;
exports.deactivate = deactivate;
const vscode = __importStar(require("vscode"));
const cp = __importStar(require("child_process"));
const mcpClient_1 = require("./mcpClient");
const diagnosticsProvider_1 = require("./diagnosticsProvider");
const hoverProvider_1 = require("./hoverProvider");
const statusTreeProvider_1 = require("./statusTreeProvider");
let statusBarItem;
function activate(context) {
    console.log('🦅 Activating Garuda Epistemic Architecture Shield...');
    const mcpClient = new mcpClient_1.GarudaMCPClient();
    // 1. Diagnostics Provider (Inline Squiggles)
    const diagnosticsProvider = new diagnosticsProvider_1.GarudaDiagnosticsProvider(mcpClient);
    diagnosticsProvider.register(context);
    // 2. Hover Provider (Blast Radius Context)
    const hoverProvider = new hoverProvider_1.GarudaHoverProvider(mcpClient);
    context.subscriptions.push(vscode.languages.registerHoverProvider({ language: 'go', scheme: 'file' }, hoverProvider));
    // 3. Activity Bar Tree Views
    const ledgerTree = new statusTreeProvider_1.GarudaLedgerTreeProvider(mcpClient);
    const contraTree = new statusTreeProvider_1.GarudaContradictionsTreeProvider(mcpClient);
    vscode.window.registerTreeDataProvider('garuda-ledger-view', ledgerTree);
    vscode.window.registerTreeDataProvider('garuda-contradictions-view', contraTree);
    // 4. Status Bar Item
    statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    statusBarItem.command = 'garuda.refreshState';
    context.subscriptions.push(statusBarItem);
    updateStatusBar(mcpClient);
    // Periodic status bar & tree refresh (every 15s)
    const timer = setInterval(() => {
        updateStatusBar(mcpClient);
        ledgerTree.refresh();
        contraTree.refresh();
    }, 15000);
    context.subscriptions.push({ dispose: () => clearInterval(timer) });
    // 5. Commands
    context.subscriptions.push(vscode.commands.registerCommand('garuda.refreshState', async () => {
        await diagnosticsProvider.refreshDiagnostics();
        ledgerTree.refresh();
        contraTree.refresh();
        await updateStatusBar(mcpClient);
        vscode.window.showInformationMessage('Garuda state and contradictions refreshed.');
    }), vscode.commands.registerCommand('garuda.openVisualizer', () => {
        const daemonUrl = vscode.workspace.getConfiguration('garuda').get('daemonUrl') || 'http://localhost:8080';
        vscode.env.openExternal(vscode.Uri.parse(`${daemonUrl}/api/v1/graph`));
    }), vscode.commands.registerCommand('garuda.reanalyzeWorkspace', () => {
        const config = vscode.workspace.getConfiguration('garuda');
        const execPath = config.get('executablePath') || 'garuda';
        const dbUrl = config.get('databaseUrl') || 'postgres://test:test@localhost:5433/garuda_test?sslmode=disable';
        vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "Garuda: Re-indexing workspace AST...",
            cancellable: false
        }, async () => {
            return new Promise((resolve) => {
                cp.exec(`${execPath} analyze . --workspace uuid-ws -s`, {
                    cwd: vscode.workspace.workspaceFolders?.[0]?.uri.fsPath,
                    env: { ...process.env, DATABASE_URL: dbUrl }
                }, async (err, stdout, stderr) => {
                    if (err) {
                        vscode.window.showErrorMessage(`Re-analysis failed: ${stderr || err.message}`);
                    }
                    else {
                        vscode.window.showInformationMessage('AST Analysis complete and anchored to Merkle ledger.');
                        await diagnosticsProvider.refreshDiagnostics();
                        ledgerTree.refresh();
                        contraTree.refresh();
                        await updateStatusBar(mcpClient);
                    }
                    resolve();
                });
            });
        });
    }));
    console.log('✓ Garuda extension initialized successfully.');
}
async function updateStatusBar(mcpClient) {
    try {
        const state = await mcpClient.getRuntimeState();
        if (state.contradicted_claims > 0) {
            statusBarItem.text = `$(shield) Garuda: #${state.block_height} ($(warning) ${state.contradicted_claims} Violations)`;
            statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.warningBackground');
            statusBarItem.tooltip = `Dual-Root Merkle Block #${state.block_height} | ${state.contradicted_claims} Quarantined Architectural Contradictions`;
        }
        else {
            statusBarItem.text = `$(shield) Garuda: #${state.block_height} (✓ Verified)`;
            statusBarItem.backgroundColor = undefined;
            statusBarItem.tooltip = `Dual-Root Merkle Block #${state.block_height} | Cryptographically Verified`;
        }
        statusBarItem.show();
    }
    catch (err) {
        statusBarItem.text = `$(shield) Garuda: Offline`;
        statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.errorBackground');
        statusBarItem.tooltip = `Could not connect to Garuda Engine via MCP or PostgreSQL`;
        statusBarItem.show();
    }
}
function deactivate() { }
//# sourceMappingURL=extension.js.map