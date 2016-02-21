package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// TODO:
// - Support AuthZ
// - Support Digest.

// AuthN is an authentication provider.
//
// It is responsible for taking raw HTTP Authorization data and determining
// whether the request is authenticated.
type AuthN interface {
	Authenticate(string) (bool, error)
}

// UserPasswordLookup provides services for lookup up a username/password combo.
type UserPasswordLookup interface {
	// IsValid looks up a username and password, and returns true if the auth is successful.
	//
	// An error is only returned if the underlying mechanism is failing. A false
	// with a nil error means the authentication simply failed.
	IsValid(username, password string) (bool, error)
}

// BasicAuth is an authentication provider for type HTTP Basic Auth.
type BasicAuth struct {
	Users UserPasswordLookup
}

// Authenticate performs an authentication step on raw HTTP Authorization data.
func (a *BasicAuth) Authenticate(data string) (bool, error) {
	user, pass, err := parseBasicString(data)
	if err != nil {
		return false, fmt.Errorf("Basic authentication parsing failed: %s", err)
	}

	return a.Users.IsValid(user, pass)
}

// Create a new HTTPAuth object with HTTP Basic support.
//
// This requires a UserPasswordLookup service.
func NewBasicHTTPAuth(pwdb UserPasswordLookup) *HTTPAuth {
	return &HTTPAuth{
		Realm: "secret",
		auths: map[string]AuthN{"basic": &BasicAuth{Users: pwdb}},
	}
}

// HTTPAuth provides HTTP authentication services.
type HTTPAuth struct {
	Realm string
	auths map[string]AuthN
}

// This will attempt to authenticate, and return an HTTP error if auth fails.
//
// If this returns `false`, a 403 Unauthorized response has already been sent.
func (h *HTTPAuth) Authenticate(res http.ResponseWriter, req *http.Request) bool {
	authz := strings.TrimSpace(req.Header.Get("Authorization"))

	// FIXME: This should extract the authn type and look it up in the auths map.
	if len(authz) == 0 || !strings.Contains(authz, "Basic ") {
		sendUnauthorized(h.Realm, res)
		return false
	}
	authn, ok := h.auths["basic"]
	if !ok {
		sendUnauthorized(h.Realm, res)
		return false
	}
	// END fixme

	if ok, _ := authn.Authenticate(authz); !ok {
		sendUnauthorized(h.Realm, res)
		return false
	}

	return ok
}

func parseBasicString(header string) (user, pass string, err error) {
	parts := strings.Split(header, " ")
	user = ""
	pass = ""
	if len(parts) < 2 {
		err = errors.New("No auth string found.")
		return
	}

	full, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return
	}

	parts = strings.SplitN(string(full), ":", 2)
	user = parts[0]
	if len(parts) > 0 {
		pass = parts[1]
	}
	return
}

func sendUnauthorized(realm string, res http.ResponseWriter) {
	// Send a 403
	res.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", realm))
	http.Error(res, "Authentication Required", http.StatusUnauthorized)
}
