// Package minecraft is a library for interacting with the profiles and skins of Minecraft players
package minecraft

import (
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/pkg/errors"
)

const (
	// ValidUsernameRegex is proper Minecraft username regex
	ValidUsernameRegex = `[a-zA-Z0-9_]{1,16}`

	// ValidUUIDPlainRegex is proper Minecraft UUID regex
	ValidUUIDPlainRegex = `[0-9a-f]{32}`

	// ValidUUIDDashRegex is proper Minecraft UUID regex
	ValidUUIDDashRegex = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`

	// ValidUUIDRegex is proper Minecraft UUID regex
	ValidUUIDRegex = "(" + ValidUUIDPlainRegex + "|" + ValidUUIDDashRegex + ")"

	// ValidUsernameOrUUIDRegex is proper Minecraft Username-or-UUID regex
	ValidUsernameOrUUIDRegex = "(" + ValidUUIDRegex + "|" + ValidUsernameRegex + ")"
)

var (
	// RegexUsername is our compiled once Username matching Regex
	RegexUsername = regexp.MustCompile("^" + ValidUsernameRegex + "$")

	// RegexUUIDPlain is our compiled once PLAIN UUID matching Regex
	RegexUUIDPlain = regexp.MustCompile("^" + ValidUUIDPlainRegex + "$")

	// RegexUUID is our compiled once DASHED UUID matching Regex
	RegexUUIDDash = regexp.MustCompile("^" + ValidUUIDDashRegex + "$")

	// RegexUUID is our compiled once UUID matching Regex
	RegexUUID = regexp.MustCompile("^" + ValidUUIDRegex + "$")

	// RegexUsernameOrUUID is our compiled once Username OR UUID matching Regex
	RegexUsernameOrUUID = regexp.MustCompile("^" + ValidUsernameOrUUIDRegex + "$")
)

// UUIDAPI is the "recent" method for performing Mojang API requests using UUIDs
type UUIDAPI struct {
	// SessionServerURL is the address where we can append a UUID and get back a SessionProfileResponse (UUID, Username and Properties/Textures)
	SessionServerURL string
	// ProfileURL is the address where we can append a Username and get back a APIProfileResponse (UUID and Username)
	ProfileURL string
}

// UsernameAPI allows manually choosing the texture lookup location with a username
type UsernameAPI struct {
	// SkinURL will have "username.png" appeneded (eg. use "http://minotar.net/skins/" to make use of our API - bypassing rate-limits to Mojang)
	SkinURL string
	// CapeURL will have "username.png" appended
	CapeURL string
}

// Minecraft is our structure for keeping of the required URLs
type Minecraft struct {
	// Client allows the supply of a custom RoundTripper (among other things)
	Client    *http.Client
	UserAgent string
	UUIDAPI
	UsernameAPI
}

// NewHTTPClient is a lazy function for returning an HTTP Client with a 10 second Timeout
func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 10,
	}
}

// NewMinecraft returns a Minecraft structure with default values (HTTP Timeout of 10 seconds and UsernameAPI will be nil)
func NewMinecraft() *Minecraft {
	return &Minecraft{
		Client:    NewHTTPClient(),
		UserAgent: "minotar/minecraft (https://github.com/minotar/minecraft)",
		UUIDAPI: UUIDAPI{
			SessionServerURL: "https://sessionserver.mojang.com/session/minecraft/profile/",
			ProfileURL:       "https://api.mojang.com/users/profiles/minecraft/",
		},
	}
}

// Mojang APIs have fairly standard responses and this makes those requests and
// catches the errors. Remember to close the response!
func (mc *Minecraft) apiRequest(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create request")
	}

	req.Header.Set("User-Agent", mc.UserAgent)

	resp, err := mc.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "unable to GET URL")
	}

	switch resp.StatusCode {

	case http.StatusOK:
		return resp.Body, nil

	case http.StatusNoContent:
		return resp.Body, errors.New("user not found")

	case http.StatusTooManyRequests:
		return resp.Body, errors.New("rate limited")

	default:
		return resp.Body, errors.Errorf("apiRequest HTTP %s", resp.Status)
	}
}
