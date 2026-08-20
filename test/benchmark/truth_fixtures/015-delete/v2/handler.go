// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package routing

type V1Handler struct{}

// LegacyRoute deleted in v2 -> TOMBSTONED.
func (h *V1Handler) ActiveRoute() string {
	return "/api/v1/live"
}
