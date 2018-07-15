package main

import (
	"bytes"
	"errors"
	"image/png"
	"net/url"
	"strings"
	"time"

	"github.com/minotar/imgd/storage"
	"github.com/minotar/minecraft"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/singleflight"
)

// Todo: Config for these options?
const (
	day = 24 * time.Hour

	usernameTTL          = 60 * day
	usernameUnknownTTL   = 14 * day
	usernameRateLimitTTL = 2 * time.Hour
	usernameErrorTTL     = 1 * time.Hour

	uuidTTL          = 60 * day
	uuidFreshTTL     = 2 * time.Hour
	uuidUnknownTTL   = 7 * day
	uuidRateLimitTTL = 1 * time.Hour
	uuidErrorTTL     = 30 * time.Minute

	skinTTL      = 30 * day
	skinErrorTTL = 15 * time.Minute

	metaUnknownCode   = "204"
	metaRateLimitCode = "429"
	metaErrorCode     = "0"
	textureURL        = "http://textures.minecraft.net"
)

var (
	sfUsername singleflight.Group
	sfUUID     singleflight.Group
	sfSkinPath singleflight.Group
)

type textures struct {
	SkinPath string
	//CapePath string
}

type mcUser struct {
	minecraft.User
	Textures  textures
	Timestamp time.Time
}

func getUUID(username string) (string, error) {
	// Check the Username cache for whether we have a matching Username->UUID
	retrieveTimer := prometheus.NewTimer(cacheDuration.WithLabelValues("cacheUUID", "retrieve"))
	uuid, err := storage.RetrieveKV(cache["cacheUUID"], username)
	retrieveTimer.ObserveDuration()
	if err == nil && uuid != "" {
		// UUID would be set to one of the meta Codes if an API error occured
		log.Debugf("Found Username in cacheUUID (%s): %s", username, uuid)
		stats.CacheUUID("hit")
	} else if err != nil && err != storage.ErrNotFound {
		// An error with the cache is considered fatal
		log.Errorf("Failed Retrieve from cacheUUID (%s): %s", username, err.Error())
		stats.CacheUUID("error")
		return "", err
	} else {
		// If the UUID is still empty/not in cache, then we better request from the API :(
		log.Debugf("Did not find Username in cacheUUID (%s)", username)
		stats.CacheUUID("miss")
		ttl := usernameTTL
		stats.APIRequested("GetAPIProfile")
		apiTimer := prometheus.NewTimer(getDuration.WithLabelValues("GetAPIProfile"))
		apiProfile, err := mcClient.GetAPIProfile(username)
		apiTimer.ObserveDuration()
		uuid = apiProfile.UUID
		if err != nil {
			var code string
			switch errMsg := err.Error(); errMsg {

			// Based on returned error, we save a "meta" UUID to cache["cacheUUID"] for differing TTLs
			case "unable to GetAPIProfile: user not found":
				code = metaUnknownCode
				ttl = usernameUnknownTTL
				log.Infof("Failed UUID lookup (%s): %s", username, errMsg)
				// Previously named "UnknownUsername"
				stats.Errored("APIProfileUnknown")
			case "unable to GetAPIProfile: rate limited":
				code = metaRateLimitCode
				ttl = usernameRateLimitTTL
				log.Noticef("Failed UUID lookup (%s): %s", username, errMsg)
				// Previously named "LookupUUIDRateLimit"
				stats.Errored("APIProfileRateLimit")
			default:
				code = metaErrorCode
				ttl = usernameErrorTTL
				log.Warningf("Failed UUID lookup: (%s): %s", username, errMsg)
				// Previously named "LookupUUID"
				stats.Errored("APIProfileGeneric")
			}
			uuid = code
		}

		// If we make an API Request, we need to cache the response (pass or fail)
		insertTimer := prometheus.NewTimer(cacheDuration.WithLabelValues("cacheUUID", "insert"))
		err = storage.InsertKV(cache["cacheUUID"], username, uuid, ttl)
		insertTimer.ObserveDuration()
		if err != nil {
			stats.CacheUUID("error")
			log.Errorf("Failed Insert to cacheUUID (%s:%s): %s", username, uuid, err.Error())
		}
	}

	// Before returning we strip the "meta" UUID (possibly from cache["cacheUUID"])
	switch uuid {
	case metaUnknownCode:
		return "", errors.New("user not found")
	case metaRateLimitCode:
		return "", errors.New("rate limited")
	case metaErrorCode:
		return "", errors.New("UUID lookup failed")
	case "":
		// We shouldn't trigger here
		return "", errors.New("empty UUID")
	default:
		return uuid, nil
	}
}

// mcUser.Minecraft.UUID must be set
func (u *mcUser) pullSessionProfile() error {
	if u.UUID == "" {
		return errors.New("pullSessionProfile needs a UUID")
	}
	stats.APIRequested("GetSessionProfile")
	spTimer := prometheus.NewTimer(getDuration.WithLabelValues("GetSessionProfile"))
	sessionProfile, err := mcClient.GetSessionProfile(u.UUID)
	spTimer.ObserveDuration()
	if err != nil {
		stats.Errored("SessionProfileGeneric")
		return err
	}

	username := strings.ToLower(sessionProfile.Username)
	if username == "" {
		return errors.New("pullSessionProfile unable to grab Username")
	}
	u.Username = username

	profileTextureProperty, err := minecraft.DecodeTextureProperty(sessionProfile)
	if err != nil {
		return err
	}

	skinURL, err := url.Parse(profileTextureProperty.Textures.Skin.URL)
	if err != nil {
		return errors.New("pullSessionProfile URL parsing failed: " + err.Error())
	}

	u.Textures.SkinPath = skinURL.Path
	return nil
}

// When we have no useful data to return, we should error
func getUser(uuid string) (*mcUser, error) {
	now := time.Now()
	user := &mcUser{
		User: minecraft.User{UUID: uuid},
	}
	retrieveTimer := prometheus.NewTimer(cacheDuration.WithLabelValues("cacheUserData", "retrieve"))
	err := storage.RetrieveGob(cache["cacheUserData"], uuid, user)
	retrieveTimer.ObserveDuration()
	if err == nil && user.Textures.SkinPath != "" {
		// SkinPath can be used to store meta Codes
		// We must have now (at least partially) populated our User struct
		// We will manually verify the time is Fresh, otherwise re-request
		log.Debugf("Found UUID in cacheUserData (%s): %+v", uuid, user)
		stats.CacheUserData("hit")
		staleAt := user.Timestamp.Add(uuidFreshTTL)
		if now.Before(staleAt) {
			// If the Timestamp was "Zero", then this would never evaulate (and we only set Timestamp for good results)
			// Todo: Potentially, we could perform a goroutine to execute the pull for a level of grace?
			// Todo: Log time it's fresh from
			log.Debugf("Data from cacheUserData was fresh (%s)", uuid)
			stats.CacheUserData("fresh")
			return user, nil
		}
		log.Debugf("Data from cacheUserData was stale, will try and freshen (%s)", uuid)
		stats.CacheUserData("stale")
	} else if err != nil && err != storage.ErrNotFound {
		// An error with the cache is considered fatal
		log.Errorf("Failed Retrieve from cacheUserData (%s): %s", uuid, err.Error())
		stats.CacheUserData("error")
		return &mcUser{}, err
	} else {
		// In a separate block to ensure we don't double count stale
		log.Debugf("Did not find UUID in cacheUserData (%s)", uuid)
		stats.CacheUserData("miss")
	}

	// EITHER: cache["cacheUserData"] hit but has old/stale data, it was actually a meta Code, or it was a miss
	// Detect stale/fallback data (we only set Timestamp when valid)
	canFallback := !user.Timestamp.IsZero()
	if user.Textures.SkinPath == "" /* no entry in cache */ || canFallback {
		ttl := uuidTTL
		// Stats are counted within pullSessionProfile()
		err = user.pullSessionProfile()
		if err != nil /* API failed */ && canFallback {
			log.Warningf("Stale SessionProfile used (%s): %s", uuid, err.Error())
			stats.Errored("SessionProfileStale")
			return user, nil
		} else if err != nil {
			var code string
			switch errMsg := err.Error(); errMsg {

			// Based on returned error, we save a "meta" SkinPath to cache["cacheUserData"] for differing TTLs
			case "unable to GetSessionProfile: user not found":
				code = metaUnknownCode
				ttl = uuidUnknownTTL
				log.Infof("Failed SessionProfile lookup (%s): %s", uuid, errMsg)
				// Previously named "UnknownUsername"
				stats.Errored("SessionProfileUnknown")
			case "unable to GetSessionProfile: rate limited":
				code = metaRateLimitCode
				ttl = uuidRateLimitTTL
				log.Noticef("Failed SessionProfile lookup (%s): %s", uuid, errMsg)
				// Previously named "LookupUUIDRateLimit"
				stats.Errored("SessionProfileRateLimit")
			default:
				code = metaErrorCode
				ttl = uuidErrorTTL
				log.Warningf("Failed SessionProfile lookup: (%s): %s", uuid, errMsg)
				// Previously named "LookupUUID"
				stats.Errored("SessionProfileGeneric")
			}
			user = &mcUser{
				User:     minecraft.User{UUID: uuid},
				Textures: textures{SkinPath: code},
			}

		} else if user.Username != "" {
			// We only set the Timestamp (used by the fresh check) if data is good.
			user.Timestamp = now
			// If we got (good) fresh data, we should update the cache["cacheUUID"] as well
			// We do not need to block on this as it is not linked to the FlightGroup
			go func() {
				insertTimer := prometheus.NewTimer(cacheDuration.WithLabelValues("cacheUUID", "insert"))
				err = storage.InsertKV(cache["cacheUUID"], user.Username, uuid, usernameTTL)
				insertTimer.ObserveDuration()
				if err != nil {
					stats.CacheUUID("error")
					log.Errorf("Failed Insert to cacheUUID (%s:%s)", user.Username, uuid)
				}
			}()
		}
		insertTimer := prometheus.NewTimer(cacheDuration.WithLabelValues("cacheUserData", "insert"))
		err = storage.InsertGob(cache["cacheUserData"], uuid, user, ttl)
		insertTimer.ObserveDuration()
		if err != nil {
			stats.CacheUserData("error")
			log.Errorf("Failed Insert to cacheUserData (%s:%v)", uuid, user)
		}
	}

	switch user.Textures.SkinPath {
	case metaUnknownCode:
		user.Textures.SkinPath = ""
		return user, errors.New("user not found")
	case metaRateLimitCode:
		user.Textures.SkinPath = ""
		return user, errors.New("rate limited")
	case metaErrorCode:
		user.Textures.SkinPath = ""
		return user, errors.New("SessionProfile lookup failed")
	case "":
		// We shouldn't trigger here
		return user, errors.New("no SkinPath")
	default:
		return user, nil
	}
}

func getSkin(skinPath string) (minecraft.Skin, error) {
	retrieveTimer := prometheus.NewTimer(cacheDuration.WithLabelValues("cacheSkin", "retrieve"))
	skin, err := RetrieveSkin(cache["cacheSkin"], skinPath)
	retrieveTimer.ObserveDuration()
	if err == nil {
		log.Debugf("Found SkinPath in cacheSkin (%s): %s", skinPath, skin.Hash)
		stats.CacheSkin("hit")
		return skin, nil
	} else if err != nil && err != storage.ErrNotFound {
		// An error with the cache is considered fatal
		log.Errorf("Failed Retrieve from cacheSkin (%s): %s", skinPath, err.Error())
		stats.CacheSkin("error")
		return skin, err
	}
	log.Debugf("Did not find SkinPath in cacheSkin (%s)", skinPath)
	stats.CacheSkin("miss")

	// The cache["cacheSkin"] is not optimized for expiry based updates (eg. groupcache)
	// We use cache["cacheSkinTransient"] for temporary results (eg. errors)
	retrieveTimer = prometheus.NewTimer(cacheDuration.WithLabelValues("cacheSkinTransient", "retrieve"))
	state, err := storage.RetrieveKV(cache["cacheSkinTransient"], skinPath)
	retrieveTimer.ObserveDuration()
	if err == nil {
		// A hit means the Skin errored on last fetch
		log.Warningf("Found SkinPath in cacheSkinTransient: " + state)
		stats.CacheSkinTransient("hit")
		return skin, errors.New("getSkin previously failed: " + state)
	}
	stats.CacheSkinTransient("miss")

	skin.Mc = mcClient
	skin.URL = textureURL + skinPath

	stats.APIRequested("TextureFetch")
	texTimer := prometheus.NewTimer(getDuration.WithLabelValues("TextureFetch"))
	err = skin.Fetch()
	texTimer.ObserveDuration()
	if err != nil {
		errMsg := err.Error()
		log.Errorf("Failed TextureFetch (%s): %s", skinPath, errMsg)
		stats.Errored("TextureFetch")

		insertTimer := prometheus.NewTimer(cacheDuration.WithLabelValues("cacheSkinTransient", "insert"))
		err = storage.InsertKV(cache["cacheSkinTransient"], skinPath, errMsg, skinErrorTTL)
		insertTimer.ObserveDuration()
		if err != nil {
			stats.CacheSkinTransient("error")
			log.Errorf("Failed Insert to cacheSkinTransient (%s)", skinPath)
		}
		return skin, errors.New("getSkin failed: " + errMsg)
	}

	insertTimer := prometheus.NewTimer(cacheDuration.WithLabelValues("cacheSkin", "insert"))
	err = InsertSkin(cache["cacheSkin"], skinPath, skin)
	insertTimer.ObserveDuration()
	if err != nil {
		stats.CacheSkin("error")
		log.Errorf("Failed Insert to cacheSkin (%s)", skinPath)
	}

	return skin, nil
}

func InsertSkin(cache storage.Storage, key string, skin minecraft.Skin) error {
	skinBuf := new(bytes.Buffer)
	_ = png.Encode(skinBuf, skin.Image)

	return cache.Insert(key, skinBuf.Bytes(), skinTTL)
}

func RetrieveSkin(cache storage.Storage, key string) (minecraft.Skin, error) {
	skin := &minecraft.Skin{}
	respBytes, err := cache.Retrieve(key)
	if err != nil {
		return minecraft.Skin{}, err
	}

	imgBuf := bytes.NewReader(respBytes)

	skinErr := skin.Decode(imgBuf)
	if skinErr != nil {
		return minecraft.Skin{}, skinErr
	}

	return *skin, nil
}

func wrapUUIDLookup(uuid string) (*mcUser, error) {
	// Singleflight ensures that lookups for same uuid will block
	// crucial as the lookup could come from multiple places (not otherwise blocking)
	user, err, _ := sfUUID.Do(uuid, func() (interface{}, error) {
		// Todo: Could stat here and above to track usage
		return getUser(uuid)
	})
	return user.(*mcUser), err
}

func wrapUsernameLookup(username string) (*mcUser, error) {
	// Singleflight ensures that lookups for same username will block
	// and then all return same data on completion

	user, err, _ := sfUsername.Do(username, func() (interface{}, error) {
		// Todo: Could stat here and above to track usage
		user := &mcUser{}
		var err error
		// If the Username doesn't match, we will run through this twice (but at max twice)
		for i := 0; i < 2; i++ {
			uuid, err := getUUID(username)
			if err != nil {
				return user, err
			}

			// calls getUser(uuid) in a Singleflight
			user, err = wrapUUIDLookup(uuid)
			// Check that all errors returned are fatal?
			if err != nil {
				return user, err
			}

			// the cache["cacheUUID"] might return old data where Username pointed to the wrong UUID
			// We then realise after looking up the UUID and this detects that inconsistency
			if username != user.Username && i == 0 /* first iteration */ {
				log.Debugf("Cached Username did not match UUID (%s:%s:%s)", username, uuid, user.Username)
				stats.Errored("APIProfileStale")
				// Re-setting the cached value to "" and re-running this function seems like the best method
				// We could otherwise try to manually GetAPIProfile, but more complications
				insertTimer := prometheus.NewTimer(cacheDuration.WithLabelValues("cacheUUID", "insert"))
				err = storage.InsertKV(cache["cacheUUID"], username, "", uuidErrorTTL)
				insertTimer.ObserveDuration()
				if err != nil {
					stats.CacheUUID("error")
					log.Errorf("Failed Insert to cacheUUID (%s:%s): %s", username, uuid, err.Error())
				}
				// Once the wrong data is cleared from the cache, we can re-run this function
				continue
			}

			if user.Textures.SkinPath == "" {
				return user, errors.New("no Texture Path")
			}

			// Unless we got mismatched Usernames, we do not need to re-run
			break
		}
		return user, err
	})
	return user.(*mcUser), err
}

func fetchUsernameSkin(username string) *mcSkin {
	username = strings.ToLower(username)
	if username == "char" || username == "mhf_steve" /* MHF_Steve */ {
		skin, _ := minecraft.FetchSkinForSteve()
		return &mcSkin{Skin: skin}
	}

	// We wrap the cache and API lookups in Singleflight to reduce outbound
	// requests, reduce rate-limit chance and maybe speed up some requests

	user, err := wrapUsernameLookup(username)
	if err != nil {
		skin, _ := minecraft.FetchSkinForSteve()
		return &mcSkin{Skin: skin}
	}

	skinPath := user.Textures.SkinPath

	skin, err, _ := sfSkinPath.Do(skinPath, func() (interface{}, error) {
		return getSkin(skinPath)
	})
	if err != nil {
		skin, _ := minecraft.FetchSkinForSteve()
		return &mcSkin{Skin: skin}
	}

	return &mcSkin{Processed: nil, Skin: skin.(minecraft.Skin)}

}
