// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

// Package migrations embeds the SQL migrations into every Garuda binary, so
// a released binary can set up its database wherever it is installed
// (Homebrew, a CI runner, a container) without a migrations folder beside it.
package migrations

import "embed"

// FS holds the top-level *.sql files of this directory. Files under
// archive/ are historical and are not applied.
//
//go:embed *.sql
var FS embed.FS
