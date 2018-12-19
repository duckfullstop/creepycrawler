// page contains Page struct instances, as well as functions for working directly against Pages.

package crawl

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type HtmlPage struct {
	// A 'HtmlPage' is a representation of a HTML format webpage, with absolute URL, title, and a list of subpages.

	// Only 'Url' is mandatory.

	// All of these attributes are public.

	// ParseLock defines whether this HtmlPage is currently locked for parsing.
	// ParseLock should be checked before performing any action that modifies the state of this HtmlPage.
	ParseLock sync.Mutex

	// Url MUST be absolute, or an error will be thrown during parse.
	Url *url.URL

	// Title is the HTML title of this page (where available)
	Title string

	// I'm undecided as to whether there should also be a Data element, that stores the entire body of the page
	// this seems like a waste of memory though, as we're really only interested in mapping page relations

	LinksTo []*HtmlPage

	// IsParsed should be flipped to True if this page has been parsed for content (even if none was found).
	// (this saves having to scrape a page twice)
	IsParsed bool

	// CrawlError is set if there was an issue crawling this page (for example, it threw a HTTP error, or scraping failed)
	CrawlError error
}

type Page interface {
	// A 'page' is an interface to all types of webpage
	// right now that's only HTML, but this allows for easily extending into other schemas, like XML

	// This Page interface is currently unused, because, I will admit, I'm not 100% sure how to use it.
	// I'd like to accept generic Page instances during main execution (i.e fetchDataAndRecurse)

	fetchAndParse() bool
	recurse() bool

	createAbsoluteUrl()
	getQueryUrl() string
}

func (p *HtmlPage) fetchAndParse(getPage chan *readPage, setPage chan *writePage) (error error) {
	// we use a new HTTP client for each request because they're occurring async
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.getQueryUrl(), nil)

	if err != nil {
		return err
	}

	log.Printf("request: (%s)", p.getQueryUrl())

	// set a User-Agent so that we properly identify as a crawler
	// could make this configurable, but arguably out of scope
	req.Header.Set("User-Agent", "Go_CreepyCrawler/1.0")

	// do the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// connection must now be closed once we're done with it, so we add a deferred close function
	defer resp.Body.Close()

	// now parse the body
	err = p.parseHTML(&resp.Body, getPage, setPage)
	if err != nil {
		return err
	}

	// the HtmlPage is now populated; return
	p.IsParsed = true
	return
}

func (p *HtmlPage) fetchDataAndRecurse(getPage chan *readPage, setPage chan *writePage, complete chan<- bool) {
	// fetchDataAndRecurse simply runs fetchData() and then recurse() (to go even deeper) in order.
	// It's here to be run as a goroutine (and returns on <-complete when finished)

	defer func() { complete <- true }()

	p.ParseLock.Lock()
	err := p.fetchAndParse(getPage, setPage)
	if err != nil {
		p.CrawlError = err
		log.Printf("Failed to parse a page: %s", err)
	}
	log.Printf("Unlocking %s", p.Url.String())
	p.ParseLock.Unlock()
	p.recurse(getPage, setPage)

	return
}

func (p *HtmlPage) recurse(getPage chan *readPage, setPage chan *writePage) (error error) {
	chComplete := make(chan bool)

	// totalCalled tracks the number of worker routines we're firing, so that we can keep track of things
	// I originally had the IsParsed check inside fetchDataAndRecurse(), but that caused some weird goroutine timeout issues
	var totalCalled int

	for _, value := range p.LinksTo {
		log.Printf("Locking %s", value.Url.String())
		value.ParseLock.Lock()
		if !value.IsParsed {
			// we unlock again after testing IsParsed because the goroutine could (theoretically) take a bit to launch
			// plus this feels saner because fetchDataAndRecurse sets and releases its own lock instead of relying on this function's
			value.ParseLock.Unlock()
			log.Printf("%s not yet parsed, parsing", value.Url.String())
			go value.fetchDataAndRecurse(getPage, setPage, chComplete)
			totalCalled++
		} else {
			log.Printf("%s already parsed, ignoring", value.Url.String())
			value.ParseLock.Unlock()
		}

	}

	// Hold a lock until all channels complete
	// I could also use a WaitGroup here instead of a completed channel, but this syntax seems clearer
	for waiting := 0; waiting < totalCalled; {
		select {
		case <-chComplete:
			// each time a goroutine finishes (and sends us the message through chComplete), increment the finished counter
			waiting++
		case <-time.After(5 * time.Second):
			// print a warning after 5 seconds if all goroutines haven't returned yet
			log.Printf("%s: 5s timeout! (waiting for %d of %d goroutine(s))\n", p.Url.String(), len(p.LinksTo)-waiting, len(p.LinksTo))
			break
		}
	}

	// finally, close up
	close(chComplete)

	return
}

func normaliseUrl(target *url.URL) *url.URL {
	// this is handled slightly funky because we have to normalise URLs
	// right now that just involves removing any fragment, but you could
	// normalise on other things (say, for example, protocol)
	// by just extending this function
	target.Fragment = ""
	return target
}

func (p *HtmlPage) createAbsoluteUrl(target *string) (targetUrl *url.URL, err error) {
	// This function takes any relative or absolute URL, and converts it to absolute
	// based on the information available from the present HtmlPage.

	// It also normalises it by removing things like anchor tags.

	// Improved deduplication could be added here.

	if target == nil || *target == "" {
		// a nil pointer causes a nil pointer dereference so we check that first before checking for an empty string
		return nil, errors.New("target URL is nil or empty")
	}

	urlTarget, err := url.Parse(*target)

	if err != nil {
		return nil, err
	}

	normaliseUrl(urlTarget)

	if urlTarget.IsAbs() {
		// Absolute URL with scheme; just pass it through
		return urlTarget, nil
	}

	// discovering url.ResolveReference is beautiful and makes me want to move in with Go full time
	return p.Url.ResolveReference(urlTarget), nil
}

func (p *HtmlPage) getQueryUrl() string {
	// getQueryUrl() normalises a given HtmlPage's Url field into a valid HTTP URL.
	// (basically, it adds a scheme if one doesn't exist)

	if p.Url.Scheme == "" {
		// use the default
		// maybe this should gracefully degrade? https is the norm for the web now though
		p.Url.Scheme = "https"
	}
	return p.Url.String()
}
