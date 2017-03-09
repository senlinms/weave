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

func do(discoveryEndpoint, token string, request interface{}, response interface{}) error {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(request)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", discoveryEndpoint+"/peer", body)
	if err != nil {
		return err
	}
	user.InjectIntoHTTPRequest(user.Inject(context.Background(), token), req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		rbody, _ := ioutil.ReadAll(resp.Body)
		return errors.New(resp.Status + ": " + string(rbody))
	}
	return json.NewDecoder(resp.Body).Decode(response)
}

func peerDiscovery(discoveryEndpoint, token, peername, nickname string, addresses []string) ([]string, error) {
	request := PeerUpdateRequest{
		Name:      peername,
		Nickname:  nickname,
		Addresses: addresses,
	}
	var updateResponse PeerUpdateResponse
	err := do(discoveryEndpoint, token, request, &updateResponse)
	return updateResponse.Addresses, err
}
