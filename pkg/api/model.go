package api

import "io"

const (
	// Edge API Endpoint paths
	EndpointImages = "api/edge/v1/images"
)

type API struct {
	Method string
	URL    string
	FH     io.Reader
}
