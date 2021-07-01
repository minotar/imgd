# Tiered Cache

## What

The key principle of the Tiered Cache is that we can use multiple cache.Cache providers (eg. LRU or BoltCache) as a single unified Cache.

The caches are ordered and the logic should be to use the quickest (and expensive ££) caches first, and subsequently slower (and cheaper) caches latterly. The idea being that the resources (or costs) are constrained for using the fastest cache, but there is still a benefit from using the slower cache.

## How

Any Inserts or Removals cascade down to each cache. Any retrievals syncronously hit each cache, and once they get a _hit_, then trigger a goroutine to update the earlier caches with the data.

If the TTL of key is less than a minute, they are not re-entered into the earlier caches (more likely the app should decided to re-fetch the data and re-cache).

## Why

Redis is great, but memory is expensive. The amount of data we want to store is multiple gigabytes and there is benefit from retaining a longer timeframe of data if it means we can keep serving successful requests. Entirely using slower/cheaper storage systems is not ideal for performance though. Hence, a hybrid solution should offer benefits.
