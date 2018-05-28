package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/docker/infrakit/pkg/callable/backend"
	"github.com/docker/infrakit/pkg/run/scope"
)

func init() {
	backend.Register("http", HTTP, func(params backend.Parameters) {
		params.StringSlice("http-header", []string{}, "Header")
		params.String("http-method", "", "HTTP Method")
		params.String("http-url", "", "Target URL")
		params.String("target", "", "Alias to --http-url")
	})
}

// HTTP takes a method parameter (string) and a URL (string) and then
// performs the http operation with the rendered data
func HTTP(scope scope.Scope, test bool, opt ...interface{}) (backend.ExecFunc, error) {

	method := "POST"
	url := ""
	headers := map[string]string{}

	if len(opt) > 2 {
		m, is := opt[0].(string)
		if !is {
			return nil, fmt.Errorf("method must be string")
		}
		method = m

		u, is := opt[1].(string)
		if !is {
			return nil, fmt.Errorf("url must be string")
		}
		url = u

		// remaining are headers
		for i := 2; i < len(opt); i++ {
			h, is := opt[i].(string)
			if !is {
				return nil, fmt.Errorf("header spec must be a string %v", opt[i])
			}
			parts := strings.SplitN(h, "=", 2)
			if len(parts) == 2 {
				headers[parts[0]] = parts[1]
			}
		}
	}

	return func(ctx context.Context, script string, parameters backend.Parameters, args []string) error {

		// Override from Parameters
		if tt, err := parameters.GetString("target"); err == nil {
			url = tt
		} else if t, err := parameters.GetString("http-url"); err == nil {
			url = t
		} else if url == "" {
			return fmt.Errorf("no url")
		}

		m, err := parameters.GetString("http-method")
		if err == nil {
			method = m
		} else if method == "" {
			return fmt.Errorf("no http method")
		}

		if _, has := map[string]bool{
			"META":   true,
			"HEAD":   true,
			"GET":    true,
			"PUT":    true,
			"POST":   true,
			"DELETE": true,
		}[method]; !has {
			return fmt.Errorf("not a valid method: %v", method)
		}

		h, err := parameters.GetStringSlice("http-header")
		if err == nil {
			for _, v := range h {
				kv := strings.SplitN(v, "=", 2)
				if len(kv) > 1 {
					headers[kv[0]] = kv[1]
				}
			}
		}

		body := bytes.NewBufferString(script)
		client := &http.Client{}

		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return err
		}

		req.Header.Set("User-Agent", "infrakit-http/0.6")
		for k, v := range headers {
			req.Header.Add(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("error %s", resp.Status)
		}

		defer resp.Body.Close()
		_, err = io.Copy(backend.GetWriter(ctx), resp.Body)
		return err
	}, nil
}
