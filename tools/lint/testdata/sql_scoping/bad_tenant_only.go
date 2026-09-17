// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package fixtures

// BadTenantOnly is a scoped read that omits workspace_id.
// Expected: 1 sql-scoping violation against `entities`.
const BadTenantOnly = `
	SELECT id, name FROM entities
	WHERE tenant_id = $1
`
