# Skind

## Cache Storage

We already hardcode the Textures URL as `http://textures.minecraft.net`, we can futher include the `/texture/` path in this and add a `bool` to determine where this is valid (which will help futureproof the optimization).

Encoding with Protobuf seems to save ~50% bytes per McUser (Username + UUID + SkinPath). This is because we aren't encoding the full struct specification with each object, instead just the actual values. 

This means we can cache double the amount of data. This also provides a built-in method for versioning and adding new fields later on with 0 effort.

This is similar to the "ExpiryRecord" encoding used in the `pkg/cache/util/expiry`, though because the timestamp is a fixed length, we pack that more effieciently than if we used Protobufs.

Further to the packing, we should further compress the resulting Protobuf (which is basically just the combined values as bytes). This can save another ~20% per User. I did further look at using a predefined dictionary, but this is only really suitable for the Textures URL and we can already optimize that out.
