package walker

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type WalkerResultNode struct {
	Path   string
	Status int
}

type WalkerResult struct {
	Root  string
	Nodes []*WalkerResultNode
}

type Walker struct {
	rootURL     *url.URL
	walkerCount int
}

func (w *Walker) parseTokenPath(parent *url.URL, token *html.Token) string {
	href := ""
	for _, attr := range token.Attr {
		if attr.Val != "" && strings.ToLower(attr.Key) == "href" {
			href = attr.Val
		}
	}

	if href == "" {
		return ""
	}

	u, err := url.Parse(href)
	if err != nil {
		return ""
	}

	ref := parent.ResolveReference(u)

	if ref.Host != w.rootURL.Host {
		return ""
	}

	return ref.Path
}

func (w *Walker) processPath(path *string) (int, []string, error) {
	u, _ := url.Parse(*path)
	absURL := w.rootURL.ResolveReference(u)

	//log.Println(absURL.Path)

	res, err := http.Get(absURL.String())
	if err != nil {
		return 0, nil, err
	}
	status, url := res.StatusCode, res.Request.URL
	defer res.Body.Close()

	if url.Host != w.rootURL.Host {
		return 0, nil, nil
	}

	*path = url.Path

	var children []string

	//log.Println(res.Header.Get("Content-Type"))
	if strings.Index(res.Header.Get("Content-Type"), "text/html") != -1 {
		z := html.NewTokenizer(res.Body)
		for {
			tt := z.Next()
			if tt == html.ErrorToken {
				break
			}

			token := z.Token()
			if token.DataAtom == atom.A {
				p := w.parseTokenPath(url, &token)
				if p != "" {
					children = append(children, p)
				}
			}
		}
	}

	return status, children, nil
}

func (w *Walker) Walk() (*WalkerResult, error) {
	r := &WalkerResult{Root: w.rootURL.String()}

	type result struct {
		err      error
		path     string
		status   int
		children []string
	}

	pendingChan := make(chan string)
	resultChan := make(chan result)
	wg := sync.WaitGroup{}

	walker := func() {
		for item := range pendingChan {
			status, children, err := w.processPath(&item)
			log.Println(item)
			resultChan <- result{path: item, status: status, children: children, err: err}
		}
	}

	for i := 0; i < w.walkerCount; i++ {
		go walker()
	}

	hit := map[string]bool{}
	done := map[string]int{}
	go func() {
		for result := range resultChan {
			if result.err != nil {
				log.Println(result)
			} else {
				if result.status != 0 {
					if _, ok := hit[result.path]; !ok {
						hit[result.path] = true
					}
					done[result.path] = result.status
					var addList []string
					for _, c := range result.children {
						if _, ok := hit[c]; !ok {
							addList = append(addList, c)
							hit[c] = true
						}
					}
					if len(addList) > 0 {
						wg.Add(len(addList))
						go func() {
							for _, add := range addList {
								pendingChan <- add
							}
						}()
					}
				}
			}
			wg.Done()
		}
	}()

	wg.Add(1)
	pendingChan <- "/"

	wg.Wait()
	close(pendingChan)
	close(resultChan)

	for path, status := range done {
		r.Nodes = append(r.Nodes, &WalkerResultNode{
			Path:   path,
			Status: status,
		})
	}

	return r, nil
}

func NewWalker(root string, walkerCount int) *Walker {
	u, err := url.Parse(root)
	if err != nil {
		log.Fatal(err)
	}
	return &Walker{rootURL: u, walkerCount: walkerCount}
}
