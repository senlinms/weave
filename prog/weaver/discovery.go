package main

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/weaveworks/common/user"
)

func peerDiscovery(discoveryEndpoint, token, peername, nickname string, addresses []string) ([]string, error) {
	values := url.Values{}
	values.Add("peername", peername)
	values.Add("nickname", nickname)
	values.Add("addresses", strings.Join(addresses, ","))

	body := strings.NewReader(values.Encode())
	req, err := http.NewRequest("POST", discoveryEndpoint+"/peer", body)
	if err != nil {
		return nil, err
	}
	user.InjectIntoHTTPRequest(user.Inject(context.Background(), token), req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, errors.New(resp.Status + ": " + string(rbody))
	}

	if len(rbody) == 0 {
		return nil, nil
	}
	return strings.Split(string(rbody), ","), nil
}
