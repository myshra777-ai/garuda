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
exports.GarudaDiagnosticsProvider = void 0;
const vscode = __importStar(require("vscode"));
class GarudaDiagnosticsProvider {
    diagnosticCollection;
    mcpClient;
    constructor(mcpClient) {
        this.mcpClient = mcpClient;
        this.diagnosticCollection = vscode.languages.createDiagnosticCollection('garuda');
    }
    register(context) {
        context.subscriptions.push(this.diagnosticCollection);
        // Refresh on file save or document open
        context.subscriptions.push(vscode.workspace.onDidSaveTextDocument(() => this.refreshDiagnostics()), vscode.workspace.onDidOpenTextDocument(() => this.refreshDiagnostics()));
        // Initial scan
        this.refreshDiagnostics();
    }
    async refreshDiagnostics() {
        try {
            const contradictions = await this.mcpClient.getContradictions(100);
            this.diagnosticCollection.clear();
            const fileDiagnosticsMap = new Map();
            for (const contra of contradictions) {
                const parts = contra.location.split(':');
                const filePath = parts[0];
                let lineNum = 0;
                if (parts.length > 1) {
                    const parsedLine = parseInt(parts[1], 10);
                    lineNum = isNaN(parsedLine) || parsedLine <= 0 ? 0 : parsedLine - 1;
                }
                const range = new vscode.Range(new vscode.Position(lineNum, 0), new vscode.Position(lineNum, 120));
                const message = `[Garuda Quarantined Contradiction] Runtime violation in ${contra.caller_symbol}: unauthorized invocation target '${contra.unapproved_target}' (${contra.invocation_count}x detected).`;
                const diagnostic = new vscode.Diagnostic(range, message, vscode.DiagnosticSeverity.Error);
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
        }
        catch (err) {
            console.error('Garuda diagnostics refresh failed:', err);
        }
    }
}
exports.GarudaDiagnosticsProvider = GarudaDiagnosticsProvider;
//# sourceMappingURL=diagnosticsProvider.js.map