package auth

import (
	"net/http"
)

type UserId = uint64

type Authenticator interface {
	Authenticate(req *http.Request) *UserId
}
