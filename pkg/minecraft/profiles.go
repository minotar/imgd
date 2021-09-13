package minecraft

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
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

// GetAPIProfile returns the API profile for a given username primarily of use
// for getting the UUID, but can also correct the capitilzation of a username or
// possibly get the account status (legacy or demo) - only included when true
func (mc *Minecraft) GetAPIProfile(username string) (APIProfileResponse, error) {
	url := mc.UUIDAPI.ProfileURL
	url += username

	apiBody, err := mc.apiRequest(url)
	if apiBody != nil {
		defer apiBody.Close()
	}
	if err != nil {
		return APIProfileResponse{}, errors.Wrap(err, "unable to GetAPIProfile")
	}

	apiProfile := APIProfileResponse{}
	err = json.NewDecoder(apiBody).Decode(&apiProfile)
	if err != nil {
		return APIProfileResponse{}, errors.Wrap(err, "decoding GetAPIProfile failed")
	}

	return apiProfile, nil
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

// GetSessionProfile fetches the session profile of the UUID, this includes
// extra properties for the user (currently just a textures property)
// Rate limits if performing same request within 30 seconds
func (mc *Minecraft) GetSessionProfile(uuid string) (SessionProfileResponse, error) {
	url := mc.UUIDAPI.SessionServerURL
	url += uuid

	apiBody, err := mc.apiRequest(url)
	if apiBody != nil {
		defer apiBody.Close()
	}
	if err != nil {
		return SessionProfileResponse{}, errors.Wrap(err, "unable to GetSessionProfile")
	}

	sessionProfile := SessionProfileResponse{}
	err = json.NewDecoder(apiBody).Decode(&sessionProfile)
	if err != nil {
		return SessionProfileResponse{}, errors.Wrap(err, "decoding GetSessionProfile failed")
	}

	return sessionProfile, nil
}
