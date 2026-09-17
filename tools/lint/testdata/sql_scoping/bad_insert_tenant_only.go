// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package fixtures

// BadInsertTenantOnly is an INSERT that sets tenant_id but not
// workspace_id on a workspace-scoped table.
// Expected: 1 sql-scoping violation against `entities`.
const BadInsertTenantOnly = `
	INSERT INTO entities (id, tenant_id, name)
	VALUES ($1, $2, $3)
`
