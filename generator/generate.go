package generator

import (
	"github.com/Ventmere/sitemap/walker"
	"github.com/jabley/mustache"
	"log"
	"os"
	"regexp"
)

type Generator struct {
	OutputDir string
	Pattern   string
}

type lang struct {
	country  string
	language string
}

func (l lang) String() string {
	return l.language + "-" + l.country
}

func (g *Generator) Generate(r *walker.WalkerResult) error {
	exp := regexp.MustCompile(g.Pattern)
	lmap := map[string]*lang{}
	plmap := map[string]map[string]string{}
	for _, node := range r.Nodes {
		match := exp.FindStringSubmatch(node.Path)
		if match == nil {
			continue
		}
		m := map[string]string{}
		for i, name := range exp.SubexpNames() {
			m[name] = match[i]
		}
		if m["country"] != "" && m["language"] != "" {
			l := lang{
				country:  m["country"],
				language: m["language"],
			}
			lstr := l.String()
			if lmap[lstr] == nil {
				lmap[lstr] = &l
			}

			path := node.Path[len(lstr)+1:]
			if plmap[path] == nil {
				plmap[path] = map[string]string{}
			}
			plmap[path][lstr] = node.Path
		}
	}

	log.Println(g.OutputDir + "/sitemap.xml")

	tpl, err := mustache.ParseString(indexTemplate)
	if err != nil {
		log.Fatal(err)
	}

	fi, err := os.OpenFile(g.OutputDir+"/sitemap.xml", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer fi.Close()

	var ll []string
	for l := range lmap {
		ll = append(ll, r.Root+"/sitemap_"+l+".xml")
	}
	fi.WriteString(tpl.Render(ll))

	type sitemapEntryAlternate struct {
		HrefLang string
		Href     string
	}

	type sitemapEntry struct {
		Loc        string
		Alternates []sitemapEntryAlternate
	}

	tpl, err = mustache.ParseString(mapTemplate)
	if err != nil {
		log.Fatal(err)
	}

	for l := range lmap {
		fl, err := os.OpenFile(g.OutputDir+"/sitemap_"+l+".xml", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
		if err != nil {
			log.Fatal(err)
		}
		defer fl.Close()

		var entries []*sitemapEntry
		for path, lmap := range plmap {
			if lpath, ok := lmap[l]; ok {
				entry := &sitemapEntry{
					Loc: r.Root + lpath,
				}
				entries = append(entries, entry)
				for ll, lpath := range lmap {
					if ll != l {
						//FIXME hard coded for edfier.com
						if ll == "en-int" {
							ll = "en"
						}
						entry.Alternates = append(entry.Alternates, sitemapEntryAlternate{
							HrefLang: ll,
							Href:     r.Root + lpath,
						})
					}
				}
				log.Println(path)
			}
		}
		fl.WriteString(tpl.Render(entries))
	}

	return nil
}
