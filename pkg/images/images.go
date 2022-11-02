package images

import (
	"fmt"

	"github.com/loadtheaccumulator/wedgeledge/ec/pkg/api"
	"github.com/loadtheaccumulator/wedgeledge/ec/pkg/config"
)

// var cfg = config.Get()

func List(ecConfig *config.EdgeConfig) {
	url := api.JoinURL(ecConfig, api.EndpointImages)
	body := api.Call(ecConfig, "GET", url, nil)

	// Print the call result
	fmt.Printf("%s", body)
}
