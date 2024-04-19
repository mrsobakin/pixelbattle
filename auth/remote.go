package auth

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type authServerResponse struct {
	UserId *UserId `json:"session_id"`
}

type RemoteAuthenticator struct {
	Endpoint url.URL
	Client   *http.Client
}

func NewRemoteAuthenticator(endpoint url.URL) *RemoteAuthenticator {
	return &RemoteAuthenticator{
		Endpoint: endpoint,
		Client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (auth *RemoteAuthenticator) Authenticate(req *http.Request) *UserId {
	cookie, err := req.Cookie("session")
	if err != nil {
		log.Println(err)
		return nil
	}

	req, err = http.NewRequest("GET", auth.Endpoint.String(), nil)
	if err != nil {
		log.Println(err)
		return nil
	}

	q := req.URL.Query()
	q.Add("cookie", cookie.Value)
	req.URL.RawQuery = q.Encode()

	resp, err := auth.Client.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Println(err)
		return nil
	}

	var result authServerResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Println(err)
		return nil
	}

	return result.UserId
}
