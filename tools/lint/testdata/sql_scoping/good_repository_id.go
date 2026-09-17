// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package fixtures

// GoodRepositoryID is scoped by repository_id, which the audit's scope
// rule treats as equivalent to workspace_id.
// Expected: 0 violations.
const GoodRepositoryID = `
	SELECT id, name FROM entities
	WHERE tenant_id = $1 AND repository_id = $2
`
