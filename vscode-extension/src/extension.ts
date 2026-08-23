// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

import * as vscode from 'vscode';
import * as cp from 'child_process';
import { GarudaMCPClient } from './mcpClient';
import { GarudaDiagnosticsProvider } from './diagnosticsProvider';
import { GarudaHoverProvider } from './hoverProvider';
import { GarudaLedgerTreeProvider, GarudaContradictionsTreeProvider } from './statusTreeProvider';

let statusBarItem: vscode.StatusBarItem;

export function activate(context: vscode.ExtensionContext) {
    console.log('🦅 Activating Garuda Epistemic Architecture Shield...');

    const mcpClient = new GarudaMCPClient();

    // 1. Diagnostics Provider (Inline Squiggles)
    const diagnosticsProvider = new GarudaDiagnosticsProvider(mcpClient);
    diagnosticsProvider.register(context);

    // 2. Hover Provider (Blast Radius Context)
    const hoverProvider = new GarudaHoverProvider(mcpClient);
    context.subscriptions.push(
        vscode.languages.registerHoverProvider({ language: 'go', scheme: 'file' }, hoverProvider)
    );

    // 3. Activity Bar Tree Views
    const ledgerTree = new GarudaLedgerTreeProvider(mcpClient);
    const contraTree = new GarudaContradictionsTreeProvider(mcpClient);

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
    context.subscriptions.push(
        vscode.commands.registerCommand('garuda.refreshState', async () => {
            await diagnosticsProvider.refreshDiagnostics();
            ledgerTree.refresh();
            contraTree.refresh();
            await updateStatusBar(mcpClient);
            vscode.window.showInformationMessage('Garuda state and contradictions refreshed.');
        }),

        vscode.commands.registerCommand('garuda.openVisualizer', () => {
            const daemonUrl = vscode.workspace.getConfiguration('garuda').get<string>('daemonUrl') || 'http://localhost:8080';
            vscode.env.openExternal(vscode.Uri.parse(`${daemonUrl}/api/v1/graph`));
        }),

        vscode.commands.registerCommand('garuda.reanalyzeWorkspace', () => {
            const config = vscode.workspace.getConfiguration('garuda');
            const execPath = config.get<string>('executablePath') || 'garuda';
            const dbUrl = config.get<string>('databaseUrl') || 'postgres://test:test@localhost:5433/garuda_test?sslmode=disable';

            vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: "Garuda: Re-indexing workspace AST...",
                cancellable: false
            }, async () => {
                return new Promise<void>((resolve) => {
                    cp.exec(`${execPath} analyze . --workspace uuid-ws -s`, {
                        cwd: vscode.workspace.workspaceFolders?.[0]?.uri.fsPath,
                        env: { ...process.env, DATABASE_URL: dbUrl }
                    }, async (err, stdout, stderr) => {
                        if (err) {
                            vscode.window.showErrorMessage(`Re-analysis failed: ${stderr || err.message}`);
                        } else {
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
        })
    );

    console.log('✓ Garuda extension initialized successfully.');
}

async function updateStatusBar(mcpClient: GarudaMCPClient): Promise<void> {
    try {
        const state = await mcpClient.getRuntimeState();
        if (state.contradicted_claims > 0) {
            statusBarItem.text = `$(shield) Garuda: #${state.block_height} ($(warning) ${state.contradicted_claims} Violations)`;
            statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.warningBackground');
            statusBarItem.tooltip = `Dual-Root Merkle Block #${state.block_height} | ${state.contradicted_claims} Quarantined Architectural Contradictions`;
        } else {
            statusBarItem.text = `$(shield) Garuda: #${state.block_height} (✓ Verified)`;
            statusBarItem.backgroundColor = undefined;
            statusBarItem.tooltip = `Dual-Root Merkle Block #${state.block_height} | Cryptographically Verified`;
        }
        statusBarItem.show();
    } catch (err) {
        statusBarItem.text = `$(shield) Garuda: Offline`;
        statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.errorBackground');
        statusBarItem.tooltip = `Could not connect to Garuda Engine via MCP or PostgreSQL`;
        statusBarItem.show();
    }
}

export function deactivate() {}
