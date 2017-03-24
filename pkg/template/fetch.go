package template

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Fetch fetchs content from the given URL string.  Supported schemes are http:// https:// file:// unix://
func Fetch(s string, opt Options, customize func(*http.Request)) ([]byte, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "file":
		return ioutil.ReadFile(u.Path)

	case "http", "https":
		return doHTTPGet(u, customize, &http.Client{})

	case "unix":
		// unix: will look for a socket that matches the host name at a
		// directory path set by environment variable.
		c, err := socketClient(u, opt.SocketDir)
		if err != nil {
			return nil, err
		}
		u.Scheme = "http"
		return doHTTPGet(u, customize, c)
	}

	return nil, fmt.Errorf("unsupported url:%s", s)
}

func doHTTPGet(u *url.URL, customize func(*http.Request), client *http.Client) ([]byte, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	if customize != nil {
		customize(req)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func socketClient(u *url.URL, socketDir string) (*http.Client, error) {
	socketPath := filepath.Join(socketDir, u.Host)
	if f, err := os.Stat(socketPath); err != nil {
		return nil, err
	} else if f.Mode()&os.ModeSocket == 0 {
		return nil, fmt.Errorf("not-a-socket:%v", socketPath)
	}
	return &http.Client{
		Transport: &http.Transport{
			Dial: func(proto, addr string) (conn net.Conn, err error) {
				return net.Dial("unix", socketPath)
			},
		},
	}, nil
}
