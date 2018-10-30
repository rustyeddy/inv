package moni

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

// Crawler represents the crawling
type CrawlDispatcher struct {
	UrlQ   chan string // public used externally to submit urls
	crawlQ chan *Page
	saveQ  chan *Page
	errQ   chan error

	*AccessList
	qsize int

	*log.Entry
}

// CrawlJob is created periodically to manage a crawl request
type CrawlJob struct {
	crawlId int64
	url     string  // site url for this job
	pages   []*Page // pages for this crawl
	depth   int     // crawl depth (see conni)
	time.Time
}

var (
	CrawlDepth int = 2
	Crawler    *CrawlDispatcher
)

func init() {
	Crawler = NewDispatcher()
}

// NewCrawler will handle scheduling all call requests
func NewDispatcher() (crawler *CrawlDispatcher) {
	cr := &CrawlDispatcher{
		AccessList: NewACL(),
		qsize:      2,
	}
	cr.UrlQ = make(chan string, cr.qsize)
	cr.crawlQ = make(chan *Page, cr.qsize)
	cr.saveQ = make(chan *Page, cr.qsize)
	cr.errQ = make(chan error, cr.qsize)

	flds := log.Fields{
		"Name": "Dispatcher",
	}
	cr.Entry = log.WithFields(flds)
	Crawler = cr
	return Crawler
}

func (cr *CrawlDispatcher) WatchChannels() {
	for {
		log.Infoln("URLQ Channel Watcher waiting for URL ... ")
		ts := time.Now()

		select {
		case url := <-cr.UrlQ:
			log.Infof("urlChan recieved %s ~ %v ", url, time.Since(ts))

			// normalize the URL
			urlstr, err := NormalizeURL(url)
			if err != nil {
				cr.errQ <- fmt.Errorf("url normaization failed %v", err)
				continue
			}

			page := FetchPage(urlstr)
			IfNilFatal(page, "get page "+urlstr)

			if !cr.IsAllowed(page.URL) {
				continue
			}

			if page.CrawlReady {
				cr.crawlQ <- page
			}

		case page := <-cr.crawlQ:
			cr.Crawl(page)

		case page := <-cr.saveQ:
			StorePage(page)

		case err := <-cr.errQ:
			log.Error(err)
		}
	}
}

// Crawl will visit the given URL, and depending on configuration
// options potentially walk internal links.
//
// Order of the callbacks http://go-colly.org/docs/introduction/start/
func (cr *CrawlDispatcher) Crawl(pg *Page) {

	// Create the collector and go get shit! (preserve?)
	c := colly.NewCollector(
		colly.MaxDepth(4),
	)

	c.OnRequest(func(r *colly.Request) {
		ustr := r.URL.String()

		cr.UrlQ <- ustr
	})

	// OnHTML will be called when we encounter a page reference
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if link := e.Request.AbsoluteURL(e.Attr("href")); link != "" {
			// Just send the link to the URL Q for processing
			cr.UrlQ <- link
		}
	})

	c.OnResponse(func(r *colly.Response) {
		link := r.Request.URL.String()
		log.Debugln("  response from", link, "status", r.StatusCode)
		pg.StatusCode = r.StatusCode
		pg.Finish = time.Now()
		pg.CrawlState = CrawlResponseRecieved
		pages[link] = pg
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Infoln("error:", r.StatusCode, err)
		pg.Err = err
		pg.StatusCode = r.StatusCode
		pg.Finish = time.Now()
		pg.CrawlState = CrawlErrored
		link := r.Request.URL.String()
		pages[link] = pg
	})

	pg.Start = time.Now()
	c.Visit(pg.URL)

	log.Infoln("Crawl Finished ", pg.URL)
}

// CrawlOrNot will determine if the provided url is allowed to be crawled,
// and if enough time has passed before the url can be scanned again
func (cr *CrawlDispatcher) CrawlOrNot(urlstr string) (pi *Page) {
	cr.Infoln("crawl or not ", urlstr)
	if !cr.IsAllowed(urlstr) {
		cr.Infof("  not allowed %s add reason ..", urlstr)
		return nil
	}

	if pi = PageFromURL(urlstr); pi == nil {
		cr.Errorf("page not found url %s", urlstr)
		return nil
	}

	if pi.CrawlState != CrawlReady {
		cr.Infof("  %s not ready to crawl ~ crawl bit off ", urlstr)
		return nil
	}
	return pi
}

func NameFromURL(urlstr string) (name string) {
	u, err := url.Parse(urlstr)
	if err != nil {
		log.Errorln(err)
		return
	}

	name = u.Hostname()
	name = "crawl-" + TimeStamp() + "-" + strings.Replace(name, ".", "-", -1)
	return name
}

func FindCrawls(pattern string) (crawl []string) {
	panic("TodO implement")
	return crawl
}

// GetCrawls
func GetCrawls() (crawl []string) {
	panic("todo write this function")
	return crawl
}

// GetTimeStamp returns a timestamp in a modified RFC3339
// format, basically remove all colons ':' from filename, since
// they have a specific use with Unix pathnames, hence must be
// escaped when used in a filename.
func TimeStamp() string {
	ts := time.Now().UTC().Format(time.RFC3339)
	return strings.Replace(ts, ":", "", -1) // get rid of offesnive colons
}
