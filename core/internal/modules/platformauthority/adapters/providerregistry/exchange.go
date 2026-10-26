package providerregistry

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

func (client *Client) exchange(
	ctx context.Context,
	method string,
	path string,
	body []byte,
) ([]byte, error) {
	request, err := http.NewRequestWithContext(
		ctx, method, client.baseURL+path, bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("build SaaS authority registry request")
	}
	request.Header.Set("Content-Type", "application/vnd.harpy.protected-authority+json")
	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("SaaS authority registry dependency failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("SaaS authority registry rejected the request")
	}
	result, err := io.ReadAll(io.LimitReader(response.Body, 256*1024+1))
	if err != nil || len(result) == 0 || len(result) > 256*1024 {
		return nil, fmt.Errorf("SaaS authority registry response is invalid")
	}
	return result, nil
}
