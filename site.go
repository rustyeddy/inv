package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Site struct {
	URL    string
	IPAddr string
	Health bool
	Pagemap
}

type Sitemap map[string]*Site

var (
	sites Sitemap
)

func GetSites() Sitemap {
	if sites == nil {
		st := getStorage()
		if _, err := st.FetchObject("sites", &sites); err != nil {
			log.Errorf(" failed get stored object 'sites' %v", err)
			return nil
		}
	}
	return sites
}

// SiteFromURL will create and index the site based on
// hostname:port.
func SiteFromURL(ustr string) (s *Site) {
	s = &Site{
		URL:     ustr,
		Pagemap: make(Pagemap),
	}
	return s
}

func (s Sitemap) Find(url string) (site *Site, ex bool) {
	site, ex = s[url]
	return site, ex
}

func (s Sitemap) Get(url string) (site *Site) {
	if site, ex := s.Find(url); ex {
		return site
	}
	return nil
}

func (s Sitemap) Exists(url string) bool {
	_, ex := s.Find(url)
	return ex
}

func (s Sitemap) Delete(url string) {
	if _, ex := s[url]; ex {
		delete(s, url)
	}
}

func (s Sitemap) Store() {
	st := getStorage()
	if _, err := st.StoreObject("sites", s); err != nil {
		log.Errorf("failed saving sites %v", err)
	}
}

func (s *Sitemap) Fetch() {
	st := getStorage()
	if _, err := st.FetchObject("sites", s); err != nil {
		log.Errorf("failed to fetch sites %v", err)
	}
}

func SiteListHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, sites)
}

func SiteIdHandler(w http.ResponseWriter, r *http.Request) {
	url := urlFromRequest(r)

	switch r.Method {
	case "GET":
		site := sites.Get(url)
		if site != nil {
			writeJSON(w, site)
		} else {
			JSONError(w, ErrorNotFound(url+" site not found "))
		}

	case "PUT", "POST":
		sites[url] = SiteFromURL(url)
		sites.Store()

	case "DELETE":
		sites.Delete(url)

	default:
		JSONError(w, ErrorNotSupported("unspported method "+r.Method))
	}
	writeJSON(w, "OK")
}
