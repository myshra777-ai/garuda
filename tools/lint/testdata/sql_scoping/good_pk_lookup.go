// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package fixtures

// GoodPKLookup is scoped by the primary key, which is a globally
// unique UUID. The audit's scope rule treats this as safe.
// Expected: 0 violations.
const GoodPKLookup = `
	SELECT id, name FROM entities
	WHERE tenant_id = $1 AND id = $2
`
