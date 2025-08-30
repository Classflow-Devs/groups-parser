package http

import (
	"encoding/json"
	"fmt"
	"github.com/corpix/uarand"
	"github.com/krispeckt/logx"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"net/url"
	"time"
)

var (
	logAlias = logx.New()
	client   = &fasthttp.Client{}
)

func FetchDataFromURL[T any](uri string, queryParams map[string]string) (*T, error) {
	start := time.Now()

	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("can't parse url: %w", err)
	}

	q := u.Query()
	for k, v := range queryParams {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(u.String())
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.Set("User-Agent", uarand.GetRandom())

	err = client.Do(req, resp)

	logAlias.WithFields(logrus.Fields{
		"url":  u.String(),
		"took": time.Since(start),
	}).Debug("request")

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), resp.Body())
	}

	var result T
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	return &result, nil
}
