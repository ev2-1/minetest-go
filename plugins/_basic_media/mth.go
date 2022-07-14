package main

import (
	"github.com/ev2-1/minetest-go"

	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
)

func generateIndex() {
	path := minetest.Path("media/")

	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	os.Mkdir(minetest.Path("hashfiles/"), 0777)
	indexFile, err := os.Create(minetest.Path("hashfiles/index.mth"))
	if err != nil {
		log.Fatal(err)
	}

	indexFile.Write([]byte("MTHS\x00\x01")) // header
	defer indexFile.Close()

	l := len(files)

	log.Printf("[media_mth] generating...")

	// generate media
	for i, file := range files {
		fmt.Fprint(os.Stdout, "\r"+fmt.Sprintf("(%d/%d)", i+1, l))

		data, err := os.ReadFile(minetest.Path("media/" + file.Name()))
		if err != nil {
			log.Fatal(err)
		}

		// generate hashes
		binhash := sha1.Sum(data)

		strhash := hex.EncodeToString(binhash[:])

		// copy file with hex hash as name
		copyFile(minetest.Path("media/"+file.Name()), "hashfiles/"+strhash)

		// add file to index
		indexFile.Write(binhash[:])

		// register file
		minetest.AddMedia(struct{ Name, Base64SHA1 string }{
			Name:       file.Name(),
			Base64SHA1: base64.StdEncoding.EncodeToString(binhash[:]), //"yfKIyplFgzDsgbO4AX+MEmLtnVM=",
		})
	}

	data, _ := indexFile.Stat()
	fmt.Printf("\n")
	log.Printf("[media_mth] done generated %d files; %d bytes mth", l+1, data.Size())
}

func copyFile(src, dst string) {
	srcf, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}

	defer srcf.Close()

	dstf, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}

	defer dstf.Close()

	_, err = io.Copy(dstf, srcf)
	if err != nil {
		log.Fatal(err)
	}
}
