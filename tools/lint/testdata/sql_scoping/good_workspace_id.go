// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package fixtures

// GoodWorkspaceID is properly scoped.
// Expected: 0 violations.
const GoodWorkspaceID = `
	SELECT id, name FROM entities
	WHERE tenant_id = $1 AND workspace_id = $2
`
