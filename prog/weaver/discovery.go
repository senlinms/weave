package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/weaveworks/common/user"
	"github.com/weaveworks/weave/ipam"
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

var httpClient = &http.Client{Timeout: 30 * time.Second}

func do(verb string, discoveryEndpoint, token string, request interface{}, response interface{}) error {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(request)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(verb, discoveryEndpoint+"/peer", body)
	if err != nil {
		return err
	}
	user.InjectIntoHTTPRequest(user.Inject(context.Background(), token), req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		rbody, _ := ioutil.ReadAll(resp.Body)
		return errors.New(resp.Status + ": " + string(rbody))
	}
	if response == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(response)
}

func peerDiscoveryUpdate(discoveryEndpoint, token, peername, nickname string, addresses []string) ([]string, error) {
	request := PeerUpdateRequest{
		Name:      peername,
		Nickname:  nickname,
		Addresses: addresses,
	}
	var updateResponse PeerUpdateResponse
	err := do("POST", discoveryEndpoint, token, request, &updateResponse)
	return updateResponse.Addresses, err
}

func HandleHTTPPeer(router *mux.Router, alloc *ipam.Allocator) {
	router.Methods("DELETE").Path("/peer").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if alloc != nil {
			alloc.Shutdown()
		}
		w.WriteHeader(204)
	})

	router.Methods("DELETE").Path("/peer/{id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ident := mux.Vars(r)["id"]
		if alloc != nil {
			transferred := alloc.AdminTakeoverRanges(ident)
			fmt.Fprintf(w, "%d IPs taken over from %s\n", transferred, ident)
		}
	})
}
