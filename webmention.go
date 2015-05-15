// webmention project webmention.go
package webmention

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/andyleap/microformats"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type WebMention struct {
	mentionQueue chan *mention
	timer        *time.Timer
	Mention      func(source, target *url.URL)
}

func New() *WebMention {
	wm := &WebMention{
		mentionQueue: make(chan *mention, 100),
	}
	wm.timer = time.NewTimer(5 * time.Second)
	go func() {
		for _ = range wm.timer.C {
			wm.process()
		}
	}()
	return wm
}

type mention struct {
	source *url.URL
	target *url.URL
}

func (wm *WebMention) GetTargetEndpoint(target url.URL) (*url.URL, error) {
	resp, err := http.Get(target.String())
	if err != nil {
		return nil, err
	}

	links := GetHeaderLinks(resp.Header["Link"])
	for _, link := range links {
		for _, rel := range link.Params["rel"] {
			if rel == "webmention" || rel == "http://webmention.org" {
				return link.URL, nil
			}
		}
	}

	parser := microformats.New()

	mf2data := parser.Parse(resp.Body)

	resp.Body.Close()

	for _, link := range mf2data.Rels["webmention"] {
		wmurl, err := url.Parse(link)
		if err != nil {
			fmt.Println(err)
			continue
		}
		return wmurl, nil
	}

	return nil, nil
}

func (wm *WebMention) SendNotification(target url.URL, source url.URL) {
	endpoint, err := wm.GetTargetEndpoint(target)
	if err != nil {
		fmt.Println(err)
		return
	}
	var values url.Values
	values.Set("source", source.String())
	values.Set("target", target.String())
	http.PostForm(endpoint.String(), values)
}

func (wm *WebMention) WebMentionEndpoint(rw http.ResponseWriter, req *http.Request) {
	source := req.FormValue("source")
	target := req.FormValue("target")
	if source != "" && target != "" {
		sourceurl, _ := url.Parse(source)
		targeturl, _ := url.Parse(target)
		wm.mentionQueue <- &mention{
			sourceurl,
			targeturl,
		}
	}
	rw.WriteHeader(http.StatusAccepted)
}

func (wm *WebMention) process() {
	mention := <-wm.mentionQueue

	resp, err := http.Get(mention.source.String())
	if err != nil {
		log.Printf("Error getting source %s: %s", mention.source, err)
		return
	}

	body, err := html.Parse(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Printf("Error parsing source %s: %s", mention.source, err)
		return
	}

	found := searchLinks(body, mention.target)
	if found {
		wm.Mention(mention.source, mention.target)
	}
}

func searchLinks(node *html.Node, link *url.URL) bool {
	if node.Type == html.ElementNode {
		if node.DataAtom == atom.A {
			if href := getAttr(node, "href"); href != "" {
				target, err := url.Parse(href)
				if err == nil {
					if target.String() == link.String() {
						return true
					}
				}
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		found := searchLinks(c, link)
		if found {
			return found
		}
	}
	return false
}

func getAttr(node *html.Node, name string) string {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return attr.Val
		}
	}
	return ""
}
