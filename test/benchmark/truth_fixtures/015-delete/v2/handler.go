package routing

type V1Handler struct{}

// LegacyRoute deleted in v2 -> TOMBSTONED.
func (h *V1Handler) ActiveRoute() string {
	return "/api/v1/live"
}
