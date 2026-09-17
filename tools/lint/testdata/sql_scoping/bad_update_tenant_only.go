// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package fixtures

// BadUpdateTenantOnly is a scoped UPDATE that omits workspace_id.
// Expected: 1 sql-scoping violation against `entities`.
const BadUpdateTenantOnly = `
	UPDATE entities
	SET name = $1
	WHERE tenant_id = $2
`
