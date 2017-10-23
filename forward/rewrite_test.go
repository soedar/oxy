package forward

import (
	"fmt"
	"net/http"
	"testing"
)

func TestRewriter(t *testing.T) {
	testCases := []struct {
		desc               string
		url                string
		remoteAddr         string
		host               string
		hostName           string
		trustForwardHeader bool
		reqHeaders         map[string]string
		expectedHeaders    map[string]string
	}{
		{
			desc:               "don't trust X Headers",
			url:                "http://foo.bar",
			remoteAddr:         "fii.bir:800",
			hostName:           "fuu.bur",
			trustForwardHeader: false,
			reqHeaders:         dumbHeaders(XHeaders),
			expectedHeaders: map[string]string{
				XForwardedProto:  "http",
				XForwardedFor:    "fii.bir",
				XForwardedHost:   "foo.bar",
				XForwardedPort:   "80",
				XForwardedServer: "fuu.bur",
				XRealIp:          "fii.bir",
			},
		},
		{
			desc:               "trust X Headers",
			url:                "http://foo.bar",
			remoteAddr:         "fii.bir:800",
			trustForwardHeader: true,
			reqHeaders:         dumbHeaders(XHeaders),
			expectedHeaders: map[string]string{
				XForwardedProto:  "fake",
				XForwardedFor:    "fake, fii.bir",
				XForwardedHost:   "fake",
				XForwardedPort:   "fake",
				XForwardedServer: "fake",
				XRealIp:          "fake",
			},
		},
		{
			desc:               "no X Headers",
			url:                "http://foo.bar",
			remoteAddr:         "fii.bir:800",
			hostName:           "fuu.bur",
			trustForwardHeader: true,
			reqHeaders:         make(map[string]string),
			expectedHeaders: map[string]string{
				XForwardedProto:  "http",
				XForwardedFor:    "fii.bir",
				XForwardedHost:   "foo.bar",
				XForwardedPort:   "80",
				XForwardedServer: "fuu.bur",
				XRealIp:          "fii.bir",
			},
		},
		{
			desc:               "request host",
			url:                "http://127.0.0.1:8000/",
			remoteAddr:         "fii.bir:800",
			host:               "fyy.byr",
			hostName:           "fuu.bur",
			trustForwardHeader: false,
			expectedHeaders: map[string]string{
				XForwardedProto:  "http",
				XForwardedFor:    "fii.bir",
				XForwardedHost:   "fyy.byr",
				XForwardedPort:   "80",
				XForwardedServer: "fuu.bur",
				XRealIp:          "fii.bir",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			hr := HeaderRewriter{
				TrustForwardHeader: test.trustForwardHeader,
				Hostname:           test.hostName,
			}

			req, err := http.NewRequest(http.MethodGet, test.url, nil)
			if err != nil {
				t.Fatal(err)
			}
			if test.host != "" {
				req.Host = test.host
			}
			if test.remoteAddr != "" {
				req.RemoteAddr = test.remoteAddr
			}

			for key, value := range test.reqHeaders {
				req.Header.Add(key, value)
			}

			hr.Rewrite(req)

			for key, expectedValue := range test.expectedHeaders {
				currentValue := req.Header.Get(key)
				if currentValue != expectedValue {
					t.Errorf("key: %s, currentValue: %s, expectedValue: %s", key, currentValue, expectedValue)
				}
			}
			if t.Failed() {
				for key, currentValue := range req.Header {
					fmt.Println(key, currentValue)
				}
			}
		})
	}

}

func TestRewriterCleanHopHeaders(t *testing.T) {
	hr := HeaderRewriter{}

	req, err := http.NewRequest(http.MethodGet, "http://foo.bar", nil)

	for key, value := range dumbHeaders(HopHeaders) {
		req.Header.Add(key, value)
	}

	if err != nil {
		t.Fatal(err)
	}

	hr.Rewrite(req)

	for _, hop := range HopHeaders {
		if req.Header.Get(hop) != "" {
			t.Errorf("error %s", hop)
		}
	}
}

func dumbHeaders(selectedHeaders []string) map[string]string {
	headers := make(map[string]string)
	for _, header := range selectedHeaders {
		headers[header] = "fake"
	}
	return headers
}
