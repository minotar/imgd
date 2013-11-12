package main

////// THIS CODE IS RELEASED INTO THE PUBLIC DOMAIN. SEE "UNLICENSE" FOR MORE INFORMATION. http://unlicense.org //////

import (
	"github.com/lukegb/minotar"
	"log"
	"os"
)

const (
	TEST_USER = "lukegb"
)

func main() {
	log.Printf("Testing for '%s'\n", TEST_USER)
	skin, err := minotar.FetchSkinForUser(TEST_USER)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("OK", skin)

	f, err := os.Create("test.png")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	headImg, err := skin.Helm()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("OK", headImg)

	headImg = minotar.Resize(64, 0, headImg)

	err = minotar.WritePNG(f, headImg)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("OK")
}
