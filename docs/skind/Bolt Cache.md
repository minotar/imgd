# Boltdb Cache


## Expiry Scan

Every $interval, we scan through every key in the Boltdb store and look for those with expired TTLs.

We do this in transaction chunks - the logic being we don't want to hold open an Update (write blocking) Tx for too long. We can then use the current key as the seek marker for next iteration.

Using consistent chunk sizes (vs. scaling chunk size to be even for each iteration) means that the max chunk size is used for most iterations and thus any performance bottlebneck is much more visible.


## Space Usage

Thought/investigation is needed for deletion/space recovery. It seems the file can only grow, and deletions free up space that can be later used.

Likely needs some sort of max file size monitor. When adding a new key, we can check the available allocations, then evict a key as required? Either randomly evict, or can we keep track of soon to expire keys? Or a random sample for eviction similar to Redis - eg. check 5 keys and choose the oldest.

Maybe a filesystem check to not use more than X free?
