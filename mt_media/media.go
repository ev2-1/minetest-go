package mt_media

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"crypto/sha1"
	"encoding/base64"

	"fmt"
	"log"
	"os"
)

func init() {
	os.Mkdir(minetest.Path("model_media"), 0777)
	files, err := os.ReadDir(minetest.Path("model_media"))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[mt_media] reading...")
	l := len(files)

	for i, file := range files {
		fmt.Fprint(os.Stdout, "\r"+fmt.Sprintf("(%d/%d)", i+1, l))

		data, err := os.ReadFile(minetest.Path("model_media/" + file.Name()))
		if err != nil {
			log.Fatal(err)
		}

		binhash := sha1.Sum(data)

		minetest.AddMedia(struct{ Name, Base64SHA1 string }{
			Name:       file.Name(),
			Base64SHA1: base64.StdEncoding.EncodeToString(binhash[:]), //"yfKIyplFgzDsgbO4AX+MEmLtnVM=",
		})
	}

	fmt.Printf("\n")

	minetest.RegisterPktProcessor(func(c *minetest.Client, pkt *mt.Pkt) {
		switch cmd := pkt.Cmd.(type) {
		case *mt.ToSrvReqMedia:
			// respond:
			go func() {
				res := mt.ToCltMedia{
					N: 1,
					I: 1,

					Files: []struct {
						Name string
						Data []byte
					}{},
				}

				for _, file := range cmd.Filenames {
					data, err := os.ReadFile(minetest.Path("model_media/" + file))
					if err != nil {
						log.Println(err)
						continue
					}

					res.Files = append(res.Files, struct {
						Name string
						Data []byte
					}{
						Name: file,
						Data: data,
					})
				}

				c.SendCmd(&res)
			}()
		}
	})
}
