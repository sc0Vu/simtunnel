package simtunnel

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type TunnelTest struct {
	SrcPort     int
	ForwardPort int
	ForwardHost string
	secure      bool
}

var (
	tunnelTests = []TunnelTest{
		{
			SrcPort:     9999,
			ForwardPort: 443,
			ForwardHost: "www.google.com",
			secure:      true,
		},
		{
			SrcPort:     8888,
			ForwardPort: 80,
			ForwardHost: "www.google.com",
			secure:      false,
		},
	}
)

func getHTTP(addr string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	client := &http.Client{
		Transport: tr,
	}
	resp, err := client.Get(addr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)

}

func TestTunneling(t *testing.T) {
	for _, v := range tunnelTests {
		tunnel := NewTunnel()
		go func() {
			if err := tunnel.ListenAndServe(v.SrcPort, v.ForwardHost, v.ForwardPort); err != nil && err != ErrClosedListener {
				t.Errorf("failed to listen and serve the tunnel: %s\n", err.Error())
			}
		}()
		protocol := "http"
		if v.secure {
			protocol = fmt.Sprintf("%ss", protocol)
		}
		tunAddr := fmt.Sprintf("%s://localhost:%d", protocol, v.SrcPort)
		resFromTun, err := getHTTP(tunAddr)
		if err != nil {
			t.Errorf("failed to GET from tunnel: %s\n", err.Error())
		}
		if len(resFromTun) <= 0 {
			t.Errorf("failed to fetch from tunnel\n")
		}
		// forwardAddr := fmt.Sprintf("%s://%s:%d", protocol, v.ForwardHost, v.ForwardPort)
		// resFromForward, err := getHTTP(forwardAddr)
		// if err != nil {
		// 	t.Errorf("failed to GET from forward connection: %s\n", err.Error())
		// }
		tunnel.Close()
	}
}
