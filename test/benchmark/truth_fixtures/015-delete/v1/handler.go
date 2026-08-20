package routing

type V1Handler struct{}

func (h *V1Handler) LegacyRoute() string {
	return "/api/v1/old"
}

func (h *V1Handler) ActiveRoute() string {
	return "/api/v1/live"
}
