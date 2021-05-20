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
	COMPACTION_MAX_SCAN = 500000
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

func (bc *BoltCache) Retrieve(key string) ([]byte, error) {
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
		// We could at this stage Remove Expired entries - but returning expired is better
		return nil
	})
	if err == storage.ErrNotFound {
		return nil, storage.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("Retrieving \"%s\" from \"%s\": %s", key, bc.Name, err)
	}

	return bce.Value, nil
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

func expirySeek(c *bolt.Cursor, keyMarker string, i int) (k, v []byte) {
	if i != 0 {
		// Midway through a scan
		k, v = c.Next()
	} else if i == 0 {
		// First run on an Update transaction
		// Need to set Cursor - keyMarker may or may not be set
		if keyMarker == "" {
			// Very first iter / first run
			k, v = c.First()
		} else {
			// X iter - use given keymarker
			k, v = c.Seek([]byte(keyMarker))
		}
	}
	return k, v
}

func (bc *BoltCache) ExpiryScan() {
	var scannedCount, expiredCount int
	var iterations int = 1
	scanSize := COMPACTION_MAX_SCAN

	// If the DB is larger than max scan, we'll chunk it into even sized
	dbLength := int(bc.Len())
	if dbLength > scanSize {
		iterations = (dbLength / scanSize) + 1
		scanSize = (dbLength / iterations) + 1
	}

	for iter := 0; iter < iterations; iter++ {
		var keyMarker string
		var err error
		now := bc.clock.Now()
		// Start a transaction for processing a chunk of keys
		// Technically, keys might be added/removed between transactions
		_ = bc.DB.Update(func(tx *bolt.Tx) error {
			c := tx.Bucket([]byte(bc.Name)).Cursor()

			for i := 0; i < scanSize; i++ {
				// If it's signalled to be stopping, stop an in-progress scan
				if !bc.running {
					// Todo: logging
					fmt.Printf("Caching is no longer running, exiting ExpiryScan")
					err = ErrCompactionInterupted
					return nil
				}

				k, v := expirySeek(c, keyMarker, i)
				if k == nil {
					fmt.Printf("Nil Key means that all keys have been scanned")
					err = ErrCompactionFinished
					return nil
				}
				scannedCount++

				// Todo: debug
				//currentTimeStr := now.String()
				//fmt.Printf("Current time is: %s\n", currentTimeStr)

				// Todo: debug
				//expiryTime := getExpiry(getExpirySeconds(v))
				//expiryTimeStr := expiryTime.String()
				//fmt.Printf("Expiry time is: %s\n", expiryTimeStr)

				if HasExpired(v[:4], now) {
					err := c.Delete()
					if err != nil {
						// Todo: logging
						fmt.Printf("Error Removing Expired \"%s\" from \"%s\": %s\n", string(k), bc.Name, err)
					}
					expiredCount++
				}
			}

			if iter == iterations {
				// Last run, make sure all keys are done
				for k, v := c.Next(); k != nil; k, v = c.Next() {
					scannedCount++
					if HasExpired(v[:4], now) {
						err := c.Delete()
						if err != nil {
							// Todo: logging
							fmt.Printf("Error Removing Expired \"%s\" from \"%s\": %s\n", string(k), bc.Name, err)
						}
						expiredCount++
					}
				}
			} else {
				// Set next key for next iteration
				k, _ := c.Next()
				keyMarker = string(k)
			}
			// Todo: more logic here for errors?
			return nil
		})

		if err == ErrCompactionInterupted {
			// Todo: logging
			fmt.Printf("Caching is %+v, exiting ExpiryScan. Len: %d, Scanned: %d, Expired: %d\n", bc.running, dbLength, scannedCount, expiredCount)
			return
		} else if err == ErrCompactionFinished {
			// Todo: logging
			fmt.Printf("All keys have been scanned. Len: %d, Scanned: %d, Expired: %d\n", dbLength, scannedCount, expiredCount)
			return
		}
	}
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
