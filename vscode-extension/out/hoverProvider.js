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
exports.GarudaHoverProvider = void 0;
const vscode = __importStar(require("vscode"));
class GarudaHoverProvider {
    mcpClient;
    constructor(mcpClient) {
        this.mcpClient = mcpClient;
    }
    async provideHover(document, position, _token) {
        const config = vscode.workspace.getConfiguration('garuda');
        if (!config.get('enableHoverBlastRadius', true)) {
            return null;
        }
        const range = document.getWordRangeAtPosition(position);
        if (!range) {
            return null;
        }
        const symbol = document.getText(range);
        if (!symbol || symbol.length < 3 || /^(func|type|struct|interface|package|import|var|const)$/.test(symbol)) {
            return null;
        }
        try {
            const blast = await this.mcpClient.getBlastRadius(symbol, 2);
            if (blast.upstream_callers_count === 0 && blast.downstream_deps_count === 0) {
                return null;
            }
            const md = new vscode.MarkdownString();
            md.isTrusted = true;
            md.appendMarkdown(`### 🦅 Garuda Architectural Context: \`${symbol}\`\n\n`);
            md.appendMarkdown(`**Blast Radius:** \`${blast.upstream_callers_count}\` Upstream Callers | \`${blast.downstream_deps_count}\` Downstream Dependencies\n\n`);
            if (blast.upstream_callers && blast.upstream_callers.length > 0) {
                md.appendMarkdown(`**Direct Callers:**\n`);
                for (const caller of blast.upstream_callers.slice(0, 5)) {
                    md.appendMarkdown(`- \`${caller.name}\` (${caller.kind}) · *${caller.package}*\n`);
                }
                if (blast.upstream_callers.length > 5) {
                    md.appendMarkdown(`- *+${blast.upstream_callers.length - 5} more callers*\n`);
                }
                md.appendMarkdown(`\n`);
            }
            if (blast.downstream_deps && blast.downstream_deps.length > 0) {
                md.appendMarkdown(`**Dependencies / Implements:**\n`);
                for (const dep of blast.downstream_deps.slice(0, 5)) {
                    md.appendMarkdown(`- \`${dep.name}\` (${dep.kind}) · *${dep.package}*\n`);
                }
                if (blast.downstream_deps.length > 5) {
                    md.appendMarkdown(`- *+${blast.downstream_deps.length - 5} more deps*\n`);
                }
            }
            md.appendMarkdown(`\n---\n[Open in Visualizer](http://localhost:8080/api/v1/graph)`);
            return new vscode.Hover(md, range);
        }
        catch (err) {
            return null;
        }
    }
}
exports.GarudaHoverProvider = GarudaHoverProvider;
//# sourceMappingURL=hoverProvider.js.map