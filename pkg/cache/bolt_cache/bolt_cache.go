package bolt_cache

import (
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/minotar/imgd/pkg/cache"
	"github.com/minotar/imgd/pkg/storage"
	"github.com/minotar/imgd/pkg/storage/bolt_store"
)

const (
	// Minimum Duration between full bucket scans looking for expired keys
	COMPACTION_SCAN_INTERVAL = 5 * time.Second
	// Max number of keys to scan in a single DB Transaction
	COMPACTION_MAX_SCAN = 50000
)

var ErrCompactionInterupted = errors.New("Compaction was interupted")
var ErrCompactionFinished = errors.New("Compaction has finished")

type clock interface {
	Now() time.Time
}

type realClock struct{}

func (r realClock) Now() time.Time { return time.Now() }

type BoltCache struct {
	*bolt_store.BoltStore
	nameExpiry               string
	clock                    clock
	closer                   chan bool
	running                  bool
	compaction_scan_interval time.Duration
}

// ensure that the storage.Storage interface is implemented
var _ cache.Cache = new(BoltCache)

func NewBoltCache(path, name string) (*BoltCache, error) {
	nameExpiry := fmt.Sprintf("%s_expiry", name)

	bs, err := bolt_store.NewBoltStore(path, name)
	if err != nil {
		return nil, err
	}

	err = bs.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(nameExpiry))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to create bucket \"%s\": %s", nameExpiry, err)
	}

	bc := &BoltCache{
		BoltStore:                bs,
		nameExpiry:               nameExpiry,
		clock:                    realClock{},
		closer:                   make(chan bool),
		compaction_scan_interval: COMPACTION_SCAN_INTERVAL,
	}

	return bc, nil
}

func (bc *BoltCache) Insert(key string, value []byte) error {
	return bc.InsertTTL(key, value, 0)
}

func (bc *BoltCache) InsertTTL(key string, value []byte, ttl time.Duration) error {
	bce := bc.NewBoltCacheEntry(key, value, ttl)
	keyBytes, valueBytes := bce.Encode()
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		//e := tx.Bucket([]byte(bc.nameExpiry))
		b := tx.Bucket([]byte(bc.Name))
		return b.Put(keyBytes, valueBytes)
	})
	if err != nil {
		return fmt.Errorf("Inserting \"%s\" into \"%s\": %s", key, bc.Name, err)
	}
	return nil
}

func (bc *BoltCache) InsertBatch(key string, value []byte) error {
	return fmt.Errorf("Not implemented")
}

func (bc *BoltCache) retrieveBCE(key string) (*BoltCacheEntry, error) {
	var bce *BoltCacheEntry
	keyBytes := []byte(key)
	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bc.Name))
		v := b.Get(keyBytes)
		if v == nil {
			return storage.ErrNotFound
		}
		// Set byte slice length for copy
		data := make([]byte, len(v))
		copy(data, v)
		bce = DecodeBoltCacheEntry(keyBytes, data)
		return nil
	})
	if err == storage.ErrNotFound {
		return nil, storage.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("Retrieving \"%s\" from \"%s\": %s", key, bc.Name, err)
	}

	// We could at this stage Remove Expired entries - but returning expired is better
	return bce, nil
}

func (bc *BoltCache) Retrieve(key string) ([]byte, error) {
	bce, err := bc.retrieveBCE(key)
	if err != nil {
		return nil, err
	}

	// Optionally we could further check expiry here
	// (though general preference for us is to return stale)
	return bce.Value, nil
}

// Must check error return - 0 expiry is either "No Expiry" or error
// Returned TTL can be used on Insert
func (bc *BoltCache) TTL(key string) (time.Duration, error) {
	bce, err := bc.retrieveBCE(key)
	if err != nil {
		return 0, err
	}

	if bce.HasExpiry() {
		ttl := bce.Expiry().Sub(bc.clock.Now())
		if ttl == time.Duration(0) {
			// Technically, we could get back a 0 Duration - but that is ambiguous
			ttl = time.Duration(1)
		}
		return ttl, nil
	}
	// No expiry is a 0 TTL
	return 0, nil
}

func (bc *BoltCache) Remove(key string) error {
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bc.Name))
		return b.Delete([]byte(key))
	})
	if err != nil {
		return fmt.Errorf("Removing \"%s\" from \"%s\": %s", key, bc.Name, err)
	}
	return nil
}

func firstOrSeek(c *bolt.Cursor, keyMarker string) (k, v []byte) {
	if keyMarker == "" {
		return c.First()
	} else {
		return c.Seek([]byte(keyMarker))
	}
}

func (bc *BoltCache) expiryScan(reviewTime time.Time, chunkSize int) {
	var scannedCount, expiredCount int
	var keyMarker string
	var scanErr error
	dbLength := int(bc.Len())

	// Keep scanning until it's interupted or finishes (via a return)
	// Crucially, this loop ensures that we aren't holding a Write lock for an extended time
	for {
		if keyMarker != "" {
			fmt.Printf("kM starts as %s\n", keyMarker)
		}
		// Start a transaction for processing a chunk of keys
		// Technically, keys might be added/removed between transactions
		_ = bc.DB.Update(func(tx *bolt.Tx) error {
			c := tx.Bucket([]byte(bc.Name)).Cursor()

			var i int
			var k, v []byte
			// Loop through each key/value until we've reached the max chunk scan size
			// Either start at beginning, or use the marker if it's set
			for k, v = firstOrSeek(c, keyMarker); i < chunkSize; k, v = c.Next() {
				// Check on every key that the cache is still running/not stopping
				if !bc.running {
					// Todo: debug logging
					fmt.Printf("Caching is no longer running, exiting ExpiryScan\n")
					scanErr = ErrCompactionInterupted
					return nil
				}

				if k == nil {
					// Todo: debug logging
					fmt.Printf("Nil Key means that all keys have been scanned\n")
					scanErr = ErrCompactionFinished
					return nil
				}
				// Increment counters for stats and chunk max
				scannedCount++
				i++
				fmt.Printf("Scanned %s\n", k)

				if HasExpired(v[:4], reviewTime) {
					err := c.Delete()
					// Todo: It seems this advances the cursor????
					fmt.Printf("Deleted: %s\n", k)
					if err != nil {
						// Todo: logging
						fmt.Printf("Error Removing Expired \"%s\" from \"%s\": %s\n", string(k), bc.Name, err)
					}
					expiredCount++
				}
			}

			// More keys to be scanned
			// The for loop above would have assigned k to the next key, but not scanned it
			// We now use that assignment to set the Marker for the next iteration
			keyMarker = string(k)
			fmt.Printf("kM ends as %s\n", keyMarker)
			if k == nil {
				// Todo: debug logging
				fmt.Printf("Nil Key means that all keys have been scanned\n")
				scanErr = ErrCompactionFinished
				return nil
			}
			return nil
			// end of DB transaction
		})
		// We need to check scanErr for globally set errors/state

		// Todo: Merge all this??
		if scanErr == ErrCompactionInterupted {
			// Todo: info logging
			fmt.Printf("Caching is %+v, exiting ExpiryScan. Len: %d, Scanned: %d, Expired: %d\n", bc.running, dbLength, scannedCount, expiredCount)
			return
		} else if scanErr == ErrCompactionFinished {
			// Todo: info logging
			fmt.Printf("All keys have been scanned. Len: %d, Scanned: %d, Expired: %d\n", dbLength, scannedCount, expiredCount)
			return
		}
	}

}

func (bc *BoltCache) ExpiryScan() {
	bc.expiryScan(bc.clock.Now(), COMPACTION_MAX_SCAN)
}

// Todo: Incomplete. Could be used when we want to maitain a DB length/size
func (bc *BoltCache) randomEvict(evictScan int) error {
	// The level of randomness isn't actually crucial - though getting a sample from a wide range is beneficial
	// Randomly grabbing a value from the DB seems difficult as there is no context of index based on ID.
	// A full scan would be hugely inefficient as well...
	// At a byte level, we know the first key is the "lowest" and the last is highest
	// We could try and `seek` to a random point between the lowest and highest.
	// That would probably be straightforward were it not for the issue of uneven key lengths
	// A way we can overcome that (while reducing the randomness...) is to perform a scan of some keys
	// at each length of slice between the First/Last key. Then generate a random byte slice
	// and try and seek to it and read the key.
	// Our random byte slices can probably prefer ascii characters
	// (Seeking an unknown key returns the next key after where it would be positioned)
	// Based on the returned values/expiry from the Seek, chooses the oldest key to Delete

	if dbLength := int(bc.Len()); dbLength < evictScan {
		evictScan = dbLength
	}

	err := bc.DB.Update(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bc.Name)).Cursor()
		firstKey, _ := c.First()
		lastKey, _ := c.Last()
		fmt.Printf("First: %+v\nLast: %+v\n", firstKey, lastKey)
		_ = evictScan
		return nil
	})
	if err != nil {
		return fmt.Errorf("Evicting a random key", err)
	}
	return nil
}

func (bc *BoltCache) Start() {
	if !bc.running {
		bc.running = true
		go bc.runCompactor()
	}
}

func (bc *BoltCache) Stop() {
	if bc.running {
		// Close() waits on a sync.Mutex for all transactions to finish
		bc.closer <- true
		// Setting to false so an in-progress compaction can stop
		bc.running = false
	}
}

func (bc *BoltCache) Close() {
	bc.Stop()
	bc.BoltStore.Close()
}

// runCompactor is in its own goroutine and thus needs the closer to stop
func (bc *BoltCache) runCompactor() {
	// Run immediately
	bc.ExpiryScan()
	ticker := time.NewTicker(bc.compaction_scan_interval)

COMPACT:
	for {
		select {
		case <-bc.closer:
			break COMPACT
		case <-ticker.C:
			bc.ExpiryScan()
		}
	}

	ticker.Stop()
}