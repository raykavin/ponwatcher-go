package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

func NewRequest(
	ctx context.Context,
	url, method string,
	payload any,
	headers map[string]string,
	requestTimeout ...time.Duration,
) (
	[]byte,
	int,
	error,
) {
	client := new(http.Client)
	if len(requestTimeout) > 0 {
		client.Timeout = requestTimeout[0]
	}

	var p io.Reader
	switch v := payload.(type) {
	case []byte:
		p = bytes.NewReader(v)
	case *bytes.Buffer:
		p = v
	default:
		p = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, p)

	if err != nil {
		return nil, 0, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer res.Body.Close()

	statusCode := res.StatusCode

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, statusCode, err
	}

	return respBody, statusCode, nil
}
