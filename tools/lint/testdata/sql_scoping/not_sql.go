// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package fixtures

// NotSQL is a log message that happens to contain SQL keywords. The
// verb-prefix guard in the check prevents it from firing.
// Expected: 0 violations.
const NotSQL = "rows loaded from entities filtered by tenant_id"
