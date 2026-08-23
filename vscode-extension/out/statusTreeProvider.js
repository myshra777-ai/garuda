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
exports.GarudaTreeItem = exports.GarudaContradictionsTreeProvider = exports.GarudaLedgerTreeProvider = void 0;
const vscode = __importStar(require("vscode"));
class GarudaLedgerTreeProvider {
    mcpClient;
    _onDidChangeTreeData = new vscode.EventEmitter();
    onDidChangeTreeData = this._onDidChangeTreeData.event;
    constructor(mcpClient) {
        this.mcpClient = mcpClient;
    }
    refresh() {
        this._onDidChangeTreeData.fire();
    }
    getTreeItem(element) {
        return element;
    }
    async getChildren(element) {
        if (element) {
            return [];
        }
        try {
            const state = await this.mcpClient.getRuntimeState();
            return [
                new GarudaTreeItem(`Block Height: #${state.block_height}`, 'shield', `Verified Block`),
                new GarudaTreeItem(`Status: ${state.status}`, 'check', 'Cryptographic Dual-Root State'),
                new GarudaTreeItem(`Entities: ${state.total_entities}`, 'symbol-class', 'AST Symbols'),
                new GarudaTreeItem(`Static Claims: ${state.total_static_claims}`, 'references', 'AST Edges'),
                new GarudaTreeItem(`Quarantined Drift: ${state.contradicted_claims}`, 'warning', 'Isolated Contradictions'),
                new GarudaTreeItem(`Static Root: ${state.static_root_hash.slice(0, 16)}...`, 'key', 'Static Tree Root'),
                new GarudaTreeItem(`Runtime Root: ${state.runtime_root_hash.slice(0, 16)}...`, 'pulse', 'Runtime Trace Root')
            ];
        }
        catch (err) {
            return [new GarudaTreeItem('Garuda Engine Offline', 'error', 'Check garuda dev or database connection')];
        }
    }
}
exports.GarudaLedgerTreeProvider = GarudaLedgerTreeProvider;
class GarudaContradictionsTreeProvider {
    mcpClient;
    _onDidChangeTreeData = new vscode.EventEmitter();
    onDidChangeTreeData = this._onDidChangeTreeData.event;
    constructor(mcpClient) {
        this.mcpClient = mcpClient;
    }
    refresh() {
        this._onDidChangeTreeData.fire();
    }
    getTreeItem(element) {
        return element;
    }
    async getChildren(element) {
        if (element) {
            return [];
        }
        try {
            const contradictions = await this.mcpClient.getContradictions(50);
            if (contradictions.length === 0) {
                return [new GarudaTreeItem('No Quarantined Contradictions', 'check', 'Zero active architectural drift')];
            }
            return contradictions.map((contra) => {
                const item = new GarudaTreeItem(`${contra.caller_symbol} → ${contra.unapproved_target}`, 'error', `Invocations: ${contra.invocation_count}x | ${contra.location}`);
                const parts = contra.location.split(':');
                if (parts.length > 0) {
                    const filePath = parts[0];
                    const line = parts.length > 1 ? parseInt(parts[1], 10) - 1 : 0;
                    item.command = {
                        command: 'vscode.open',
                        title: 'Open Contradiction File',
                        arguments: [
                            vscode.Uri.file(filePath),
                            { selection: new vscode.Range(line, 0, line, 0) }
                        ]
                    };
                }
                return item;
            });
        }
        catch (err) {
            return [new GarudaTreeItem('Failed to fetch contradictions', 'error', `${err}`)];
        }
    }
}
exports.GarudaContradictionsTreeProvider = GarudaContradictionsTreeProvider;
class GarudaTreeItem extends vscode.TreeItem {
    label;
    description;
    constructor(label, iconName, description) {
        super(label, vscode.TreeItemCollapsibleState.None);
        this.label = label;
        this.description = description;
        this.iconPath = new vscode.ThemeIcon(iconName);
    }
}
exports.GarudaTreeItem = GarudaTreeItem;
//# sourceMappingURL=statusTreeProvider.js.map