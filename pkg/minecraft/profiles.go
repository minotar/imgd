package minecraft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type User struct {
	UUID     string `json:"id"`
	Username string `json:"name"`
}

type APIProfileResponse struct {
	User
	Legacy bool `json:"legacy"`
	Demo   bool `json:"demo"`
}

type SessionProfileResponse struct {
	User
	Properties []SessionProfileProperty `json:"properties"`
}

type SessionProfileProperty struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// GetAPIProfileCtx is the same as GetAPIProfile, but with Context on the Request
func (mc *Minecraft) GetAPIProfileCtx(ctx context.Context, username string) (APIProfileResponse, error) {
	apiBody, err := mc.apiRequestCtx(ctx, mc.Cfg.ProfileURL+username)
	if err != nil {
		return APIProfileResponse{}, fmt.Errorf("unable to GetAPIProfile: %w", err)
	}

	apiProfile := APIProfileResponse{}
	err = json.NewDecoder(apiBody).Decode(&apiProfile)
	if err != nil {
		return APIProfileResponse{}, fmt.Errorf("decoding GetAPIProfile failed: %w", err)
	}

	return apiProfile, nil
}

// GetAPIProfile returns the API profile for a given username primarily of use
// for getting the UUID, but can also correct the capitilzation of a username or
// possibly get the account status (legacy or demo) - only included when true
func (mc *Minecraft) GetAPIProfile(username string) (APIProfileResponse, error) {
	return mc.GetAPIProfileCtx(context.Background(), username)
}

// GetUUID returns the UUID for a given username (shorthand for GetAPIProfile)
func (mc *Minecraft) GetUUID(username string) (string, error) {
	apiProfile, err := mc.GetAPIProfile(username)
	return apiProfile.UUID, err
}

// NormalizePlayerForUUID takes either a Username or UUID and returns a UUID
// formatted without dashes, or an error (eg. no account or an API error)
func (mc *Minecraft) NormalizePlayerForUUID(player string) (string, error) {
	if RegexUsername.MatchString(player) {
		return mc.GetUUID(player)
	} else if RegexUUID.MatchString(player) {
		return strings.Replace(player, "-", "", 4), nil
	}

	// We shouldn't get this far as there should have been Regex checks already.
	return "", errors.New("unable to NormalizePlayerForUUID due to invalid Username/UUID")
}

// GetSessionProfileCtx is the same as GetSessionProfile, but with Context on the Request
func (mc *Minecraft) GetSessionProfileCtx(ctx context.Context, uuid string) (SessionProfileResponse, error) {
	apiBody, err := mc.apiRequestCtx(ctx, mc.Cfg.SessionServerURL+uuid)
	if err != nil {
		return SessionProfileResponse{}, fmt.Errorf("unable to GetSessionProfile: %w", err)
	}

	sessionProfile := SessionProfileResponse{}
	err = json.NewDecoder(apiBody).Decode(&sessionProfile)
	if err != nil {
		return SessionProfileResponse{}, fmt.Errorf("decoding GetSessionProfile failed: %w", err)
	}

	return sessionProfile, nil
}

// GetSessionProfile fetches the session profile of the UUID, this includes
// extra properties for the user (currently just a textures property)
// Rate limits if performing same request within 30 seconds
func (mc *Minecraft) GetSessionProfile(uuid string) (SessionProfileResponse, error) {
	return mc.GetSessionProfileCtx(context.Background(), uuid)
}
