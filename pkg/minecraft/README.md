# Minecraft Go API [![Build Status](https://travis-ci.org/minotar/minecraft.png)](https://travis-ci.org/minotar/minecraft)

~~~ go
package main

import "github.com/minotar/minecraft"

func main() {
  uuid, _ := minecraft.GetUUID("clone1018")

  skin, _ := minecraft.FetchSkinUUID(uuid)
}
~~~

Install the package (**go 1.1** and greater is required):
~~~
go get github.com/minotar/minecraft
~~~

## Features
* User fetching using the Minecraft API
* Skin fetching using AmazonS3
* Almost no actual functionality!


## License
This is free and unencumbered software released into the public domain.
