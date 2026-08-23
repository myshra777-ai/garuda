// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

import * as vscode from 'vscode';
import { GarudaMCPClient, Contradiction } from './mcpClient';

export class GarudaDiagnosticsProvider {
    private diagnosticCollection: vscode.DiagnosticCollection;
    private mcpClient: GarudaMCPClient;

    constructor(mcpClient: GarudaMCPClient) {
        this.mcpClient = mcpClient;
        this.diagnosticCollection = vscode.languages.createDiagnosticCollection('garuda');
    }

    public register(context: vscode.ExtensionContext): void {
        context.subscriptions.push(this.diagnosticCollection);

        // Refresh on file save or document open
        context.subscriptions.push(
            vscode.workspace.onDidSaveTextDocument(() => this.refreshDiagnostics()),
            vscode.workspace.onDidOpenTextDocument(() => this.refreshDiagnostics())
        );

        // Initial scan
        this.refreshDiagnostics();
    }

    public async refreshDiagnostics(): Promise<void> {
        try {
            const contradictions = await this.mcpClient.getContradictions(100);
            this.diagnosticCollection.clear();

            const fileDiagnosticsMap = new Map<string, vscode.Diagnostic[]>();

            for (const contra of contradictions) {
                const parts = contra.location.split(':');
                const filePath = parts[0];
                let lineNum = 0;
                if (parts.length > 1) {
                    const parsedLine = parseInt(parts[1], 10);
                    lineNum = isNaN(parsedLine) || parsedLine <= 0 ? 0 : parsedLine - 1;
                }

                const range = new vscode.Range(
                    new vscode.Position(lineNum, 0),
                    new vscode.Position(lineNum, 120)
                );

                const message = `[Garuda Quarantined Contradiction] Runtime violation in ${contra.caller_symbol}: unauthorized invocation target '${contra.unapproved_target}' (${contra.invocation_count}x detected).`;

                const diagnostic = new vscode.Diagnostic(
                    range,
                    message,
                    vscode.DiagnosticSeverity.Error
                );

                diagnostic.source = 'Garuda';
                diagnostic.code = {
                    value: 'ARCH_DRIFT_001',
                    target: vscode.Uri.parse('http://localhost:8080/api/v1/graph')
                };

                const existing = fileDiagnosticsMap.get(filePath) || [];
                existing.push(diagnostic);
                fileDiagnosticsMap.set(filePath, existing);
            }

            for (const [filePath, diagnostics] of fileDiagnosticsMap.entries()) {
                const uri = vscode.Uri.file(filePath);
                this.diagnosticCollection.set(uri, diagnostics);
            }
        } catch (err) {
            console.error('Garuda diagnostics refresh failed:', err);
        }
    }
}
