package main

import (
	"encoding/json"
	"flag"
	"github.com/Ventmere/sitemap/generator"
	"github.com/Ventmere/sitemap/walker"
	"log"
	"net/url"
	"os"
	"runtime"
)

var root = flag.String("root", "http://www.edifier.com", "root url")
var wc = flag.Int("worker", 8, "worker count")
var cached = flag.Bool("cached", false, "use cache")

const cachePath = "./out/cache.json"

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(8)

	useCache := *cached
	var r *walker.WalkerResult
	var err error

	u, err := url.Parse(*root)
	if err != nil {
		log.Fatal(err)
	}

	log.SetPrefix("[WALK] ")

	if useCache {
		log.Println("skipped: use cache")
		fcache, err := os.Open(cachePath)
		if err != nil {
			log.Fatal(err)
		}
		defer fcache.Close()

		var ri walker.WalkerResult
		if err := json.NewDecoder(fcache).Decode(&ri); err != nil {
			log.Fatal(err)
		}
		r = &ri
	} else {
		w := walker.NewWalker(*root, *wc)
		r, err = w.Walk()
		if err != nil {
			log.Fatal(err)
		}

		fcache, err := os.OpenFile(cachePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			log.Fatal(err)
		}
		defer fcache.Close()

		encoder := json.NewEncoder(fcache)
		if err := encoder.Encode(r); err != nil {
			log.Fatal(err)
		}
	}

	log.SetPrefix("[GENERATE] ")

	out := "./out/" + u.Host
	os.RemoveAll(out)
	os.Mkdir(out, 0777)
	g := &generator.Generator{OutputDir: out, Pattern: `^/(?P<country>\w{2,3})/(?P<language>\w{2})`}
	if err := g.Generate(r); err != nil {
		log.Fatal(err)
	}
}
