# skind

Pronounced "skinned".


## Endpoints

Exposes only basic endpoints for grabbing a raw Minecraft skin as a PNG

Supports either a a UUID or Username. 

```
/skin/<UUID>
/skin/<USERNAME>
/_302/skin/<USERNAME> (302 redirects to /skin/UUID)
```

### ESI

_No idea if this works, it's an idea..!_

To optimize Varnish cache usage, there is no point in caching both UUID and Username paths. So, any requests for this will instead trigger a Varnish request to the UUID version.

An ESI path will basically return the UUID wrapped in ESI markup
```
/_esi/skin/<USERNAME>
```

```
<esi:include src="/skin/<UUID>"/>
```

### Metrics

Prometheus metrics exposed:

```
/metrics
```

## Caching

Crucial.

Types of data we need to cache:

1. Username -> UUID mapping. 30+ days validity. Upstream rate limits are brutal
2. UUID -> User Data. 7 days?
3. Skin Path -> Texture. No ratelimits, just a speed thing. Could use a proxy here?
4. Error Cache. Transient. Just to avoid repeated broken calls (maybe not worth it)

Key requirements for 1 and 2 are reliability/durability between service failures. We can lose some/up to 20% of keys with no real concern - but this is super valuable data.

The key/value size for 1 is tiny, but million+ records. The data for 2 is larger.

Logically, filesystem is a good place as it allows a lot of data vs. the limitations of an in-memory store.

We can still see benefit from in-memory though - this is where the speed comes from.

Multi-tiered caching is the ideal route here for 1 & 2 - with a faster cache in-front of a slower cache.

Fast:
1. Local in-memory LRU/LFU type caches
    * Fine if single instance
    * Needs request hashing/sharding if multiple instances
    * Probably fastest (though probably not impactful)
    * Error prone memory handling can OOM
    * Complicates app
2. Redis shared among nodes
    * Decoupled from app - we can deploy app with 0 concern
    * Native state save/loading (reduced impact from restarts)
    * Ready made solution
    * Resilient deployment is complicated and memory hungry
3. Local write cache
    * Depending on the speed of the cache backends, we could have a list of writes
    * Before requesting from a slow cache, we query the write cache
    * Misses are expected, the cache length should stay short, guards against failures/latency


Slow:
1. boltdb
    * Memory usage for millions of files?
    * Perf issues from putting everything together?
    * Expiry?
2. Filesystem
    * Probably fairly unused / 20GB - 50GB free?
    * Would require care with directory sizes
    * Hash key, then split on the first 10 characters in pairs?
    * Probably would see a benefit of using a "modern" store for chunks of keys
        * eg. boltdb ? LevelDB, RocksDB ?
        * eg. last 50% of the hashed key ends up in a DB file.
    * Need to manually handle expiry... maybe another DB?
    * Native Linux file caching
3. Object storage (eg. DO Spaces or S3?)
    * Avoids a ratelimit, but adds an overhead for insert/retrieval
    * Would offer a builtin expiry management
    * Likely inefficient storage as minimum object billing would be much larger
    * Batching keys together would break expiry management
    * Maybe usable for an emergency? Or grace based cache


We would also want to avoid multiple in-flight requests for the same upstream item.

### Skin Path

Configure nginx or Caddy as a caching proxy to upstream. They can control their local file caches and are entirely transient. The data can be lost with 0 concern.

Assuming we are configuring this as a reverse proxy type thing, we'll want to ensure the URL we are proxying is the same one we got in the User Data (a sort of sanity check).

## Rate Limits

We need to control our flow to upstream. Once ratelimited, it's much more painful vs. floating under the radar.

Some sort of bucket system needed here to avoid hitting API limits.

Configurable limits (eg. requests per X seconds). What behaviour when we limit?

##

Configurable listeners? HTTP + GRPC?
