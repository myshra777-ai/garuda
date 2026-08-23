// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

import * as vscode from 'vscode';
import { GarudaMCPClient, RuntimeState, Contradiction } from './mcpClient';

export class GarudaLedgerTreeProvider implements vscode.TreeDataProvider<GarudaTreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<GarudaTreeItem | undefined | void> = new vscode.EventEmitter<GarudaTreeItem | undefined | void>();
    readonly onDidChangeTreeData: vscode.Event<GarudaTreeItem | undefined | void> = this._onDidChangeTreeData.event;

    constructor(private mcpClient: GarudaMCPClient) {}

    public refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    public getTreeItem(element: GarudaTreeItem): vscode.TreeItem {
        return element;
    }

    public async getChildren(element?: GarudaTreeItem): Promise<GarudaTreeItem[]> {
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
        } catch (err) {
            return [new GarudaTreeItem('Garuda Engine Offline', 'error', 'Check garuda dev or database connection')];
        }
    }
}

export class GarudaContradictionsTreeProvider implements vscode.TreeDataProvider<GarudaTreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<GarudaTreeItem | undefined | void> = new vscode.EventEmitter<GarudaTreeItem | undefined | void>();
    readonly onDidChangeTreeData: vscode.Event<GarudaTreeItem | undefined | void> = this._onDidChangeTreeData.event;

    constructor(private mcpClient: GarudaMCPClient) {}

    public refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    public getTreeItem(element: GarudaTreeItem): vscode.TreeItem {
        return element;
    }

    public async getChildren(element?: GarudaTreeItem): Promise<GarudaTreeItem[]> {
        if (element) {
            return [];
        }

        try {
            const contradictions = await this.mcpClient.getContradictions(50);
            if (contradictions.length === 0) {
                return [new GarudaTreeItem('No Quarantined Contradictions', 'check', 'Zero active architectural drift')];
            }

            return contradictions.map((contra) => {
                const item = new GarudaTreeItem(
                    `${contra.caller_symbol} → ${contra.unapproved_target}`,
                    'error',
                    `Invocations: ${contra.invocation_count}x | ${contra.location}`
                );

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
        } catch (err) {
            return [new GarudaTreeItem('Failed to fetch contradictions', 'error', `${err}`)];
        }
    }
}

export class GarudaTreeItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        iconName: string,
        public readonly description?: string
    ) {
        super(label, vscode.TreeItemCollapsibleState.None);
        this.iconPath = new vscode.ThemeIcon(iconName);
    }
}
