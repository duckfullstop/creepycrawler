// parser contains functions for parsing types of Page for links and other metadata.

package crawl

import (
	"golang.org/x/net/html"
	"io"
	"log"
)

func (p *HtmlPage) parseHTML(data *io.ReadCloser, getPage chan *readPage, setPage chan *writePage) (error error) {

	// parseHTML parses HTML from the provided ReadCloser, and writes information about it to its HtmlPage.
	// It uses the getPage and setPage channels to access the synchronous allpages data store, to check
	// for entry duplication.

	// This could easily be refactored into a generic parseHTML function, but for the purposes of this scraper,
	// it is bound against HtmlPage.
	doc, err := html.Parse(*data)

	if err != nil {
		return err
	}

	// f is a helper function that does the scraping for us (and can, as such, be called recursively)
	// it must be defined explicitly because otherwise it's not available inside the function during creation
	var f func(*html.Node)
	f = func(n *html.Node) {

		if n.Type == html.ElementNode && n.Data == "title" {
			// It's the page title! Let's set the HtmlPage title attribute.
			p.Title = n.FirstChild.Data
			log.Printf("ℹ️ (%s) title='%s'", p.Url.String(), p.Title)
		}
		if n.Type == html.ElementNode && n.Data == "a" {
			// It's a hyperlink! Let's add this to the discovered links array.

			// First, we need to get the thing it actually links to:
			var href *string
			for _, a := range n.Attr {
				if a.Key == "href" {
					href = &a.Val
					break
				}
			}

			// Next, check that it is part of our target
			targetUrl, err := p.createAbsoluteUrl(href)

			if err != nil {
				log.Printf("⚠️ (%s) recoverable error occured during parsing of a URL, ignoring: %s", p.Url.String(), err)
			// FIXME I'd probably replace this host==host lookup with more sensible hostIsValid and urlIsDuplicate calls
			// FIXME mainly because that would make it easier to ignore URIs like mailto:, plus gives a single authority
			} else if targetUrl.Host == p.Url.Host {
				// Finally, do we already have it in our stack?

				read := &readPage{key: *targetUrl, response: make(chan *HtmlPage)}

				getPage <- read

				response := <-read.response

				if response == nil {
					// Then we append a pointer to a new HtmlPage
					targetPage := &HtmlPage{Url: targetUrl}
					p.LinksTo = append(p.LinksTo, targetPage)
					// and append this link to the discovered list
					write := &writePage{key: *targetUrl, value: targetPage, response: make(chan bool)}
					setPage <- write
					<-write.response
					log.Printf("🌍 (%s) href=%s stored as %p", p.Url.String(), targetUrl.String(), targetPage)
				} else {
					// Append this page to the linksTo array, because it _is_ still a link to a different page
					// it won't be read again because fetchDataAndRecurse won't run on HtmlPages which have already been parsed
					// (this is guaranteed with a Mutex lock on Page)
					p.LinksTo = append(p.LinksTo, response)
					log.Printf("ℹ️ (%s) href=%s already discovered: is %p", p.Url.String(), targetUrl.String(), &response)
				}

			}
		}
		// We don't return from the above if because it's (theoretically) possible to have A's inside A's (even though it's stupid and totally against spec)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			// Let's recurse as far as possible!
			f(c)
		}
	}

	// Do the scrape!
	f(doc)

	return
}
