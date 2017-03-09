package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/weaveworks/common/user"
)

// TODO: move these definitions somewhere more shareable
type PeerUpdateRequest struct {
	Name      string   `json:"peername"`
	Nickname  string   `json:"nickname"`  // optional
	Addresses []string `json:"addresses"` // can be empty
}

type PeerUpdateResponse struct {
	Addresses []string `json:"addresses"`
}

func peerDiscovery(discoveryEndpoint, token, peername, nickname string, addresses []string) ([]string, error) {
	request := PeerUpdateRequest{
		Name:      peername,
		Nickname:  nickname,
		Addresses: addresses,
	}

	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(request)
	if err != nil {
		return nil, err
	}
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
	if resp.StatusCode != http.StatusOK {
		rbody, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(resp.Status + ": " + string(rbody))
	}

	var updateResponse PeerUpdateResponse
	err = json.NewDecoder(resp.Body).Decode(&updateResponse)
	if err != nil {
		return nil, err
	}
	return updateResponse.Addresses, nil
}
