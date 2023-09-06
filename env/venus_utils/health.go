package venusutils

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/ipfs-force-community/brightbird/types"
)

func VenusHealthCheck(ctx context.Context, endpoint types.Endpoint) error {
	req, err := retryablehttp.NewRequest("GET", fmt.Sprintf("http://%s/healthcheck", endpoint), nil)
	if err != nil {
		return err
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5

	resp, err := retryClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	log.Debugf("track status %s %d", resp.Status, resp.StatusCode)
	return fmt.Errorf("receive health %s", resp.Status)
}
