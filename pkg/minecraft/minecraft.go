// Package minecraft is a library for interacting with the profiles and skins of Minecraft players
package minecraft

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/minotar/imgd/pkg/minecraft/util/log"
)

const (
	// ValidUsernameRegex is proper Minecraft username regex
	ValidUsernameRegex = `[a-zA-Z0-9_]{1,16}`

	// ValidUUIDPlainRegex is proper Minecraft UUID regex (no dashes)
	ValidUUIDPlainRegex = `[0-9a-f]{32}`

	// ValidUUIDDashRegex is proper Minecraft UUID regex (with dashes 👎)
	ValidUUIDDashRegex = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`

	// ValidUUIDRegex is proper Minecraft UUID regex
	ValidUUIDRegex = ValidUUIDPlainRegex + "|" + ValidUUIDDashRegex

	// ValidUsernameOrUUIDRegex is proper Minecraft Username OR UUID regex
	ValidUsernameOrUUIDRegex = ValidUUIDRegex + "|" + ValidUsernameRegex

	// ValidUsernameOrPlainUUIDRegex is proper Minecraft Username OR Plain-UUID regex
	ValidUsernameOrPlainUUIDRegex = ValidUUIDPlainRegex + "|" + ValidUsernameRegex
)

var (
	// RegexUsername is our compiled once Username matching Regex
	RegexUsername = regexp.MustCompile("^" + ValidUsernameRegex + "$")

	// RegexUUIDPlain is our compiled once PLAIN UUID matching Regex
	RegexUUIDPlain = regexp.MustCompile("^" + ValidUUIDPlainRegex + "$")

	// RegexUUID is our compiled once DASHED UUID matching Regex
	RegexUUIDDash = regexp.MustCompile("^" + ValidUUIDDashRegex + "$")

	// RegexUUID is our compiled once UUID matching Regex
	RegexUUID = regexp.MustCompile("^(" + ValidUUIDRegex + ")$")

	// RegexUsernameOrUUID is our compiled once Username OR UUID matching Regex
	RegexUsernameOrUUID = regexp.MustCompile("^(" + ValidUsernameOrUUIDRegex + ")$")

	// RegexUsernameOrPlainUUID is our compiled once Username OR Plain -UID matching Regex
	RegexUsernameOrPlainUUID = regexp.MustCompile("^(" + ValidUsernameOrPlainUUIDRegex + ")$")
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrRateLimit    = errors.New("rate limited")
)

// UUIDAPIConfig is the current method for performing Mojang API requests using UUIDs
type UUIDAPIConfig struct {
	// SessionServerURL is the address where we can append a UUID and get back a SessionProfileResponse (UUID, Username and Properties/Textures)
	SessionServerURL string
	// ProfileURL is the address where we can append a Username and get back a APIProfileResponse (UUID and Username)
	ProfileURL string
}

// UsernameAPIConfig allows manually choosing the texture lookup location with a username
// This is very limited reasons to use this - realistically, just use UUIDs!
type UsernameAPIConfig struct {
	// SkinURL will have "username.png" appeneded (eg. use "http://minotar.net/skins/" to make use of our API - bypassing rate-limits to Mojang)
	SkinURL string
	// CapeURL will have "username.png" appended
	CapeURL string
}

type Config struct {
	Logger log.Logger
	UUIDAPIConfig
	UsernameAPIConfig
	UserAgent      string
	RequestTimeout time.Duration
}

var (
	// DefaultConfig is used by NewDefaultMinecraft and also serves as the defaults when using the Config.RegisterFlags
	DefaultConfig Config = Config{
		Logger:         log.NewStdLogger(),
		UserAgent:      "minotar/imgd/pkg/minecraft (https://github.com/minotar/imgd) - default",
		RequestTimeout: 10 * time.Second,
		UUIDAPIConfig: UUIDAPIConfig{
			SessionServerURL: "https://sessionserver.mojang.com/session/minecraft/profile/",
			ProfileURL:       "https://api.mojang.com/users/profiles/minecraft/",
		},
	}
)

// Optionally can be used for registering flags when creating parent Config objects
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.UserAgent, "minecraft.useragent", DefaultConfig.UserAgent, "UserAgent for Minecraft API Client")
	f.DurationVar(&c.RequestTimeout, "minecraft.request-timeout", DefaultConfig.RequestTimeout, "Timeout for Minecraft API Client")
	f.StringVar(&c.SessionServerURL, "minecraft.sessionserver-url", DefaultConfig.SessionServerURL, "API for UUID -> Texture Properties")
	f.StringVar(&c.ProfileURL, "minecraft.profile-url", DefaultConfig.ProfileURL, "API for Username -> UUID lookups")
}

// Minecraft is our structure for keeping of the required URLs
type Minecraft struct {
	// Client allows the supply of a custom RoundTripper (among other things)
	Client *http.Client
	Cfg    Config
}

// NewMinecraft returns a Minecraft structure with default values (HTTP Timeout of 10 seconds and UsernameAPI will be nil)
func NewDefaultMinecraft() *Minecraft {
	cfg := DefaultConfig
	return NewMinecraft(cfg)
}

func NewMinecraft(cfg Config) *Minecraft {
	return &Minecraft{
		Client: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
		Cfg: cfg,
	}
}

func (mc *Minecraft) get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}
	req.Header.Set("User-Agent", mc.Cfg.UserAgent)

	resp, err := mc.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to GET URL: %w", err)
	}

	return resp, nil
}

func processGetReq(r *http.Response, err error) (io.ReadCloser, error) {
	if err != nil {
		return nil, err
	}

	switch r.StatusCode {

	case http.StatusOK:
		return r.Body, nil

	case http.StatusNoContent:
		r.Body.Close()
		return nil, ErrUserNotFound

	case http.StatusTooManyRequests:
		r.Body.Close()
		return nil, ErrRateLimit

	default:
		r.Body.Close()
		return nil, errors.New("minecraft HTTP GET got unexpected: " + r.Status)
	}
}

// Mojang APIs have fairly standard responses and this makes those requests and
// catches the errors. Remember to close the response if there is no error present!
func (mc *Minecraft) apiRequestCtx(ctx context.Context, url string) (io.ReadCloser, error) {
	return processGetReq(mc.get(ctx, url))
}

func (mc *Minecraft) apiRequest(url string) (io.ReadCloser, error) {
	return mc.apiRequestCtx(context.Background(), url)
}
