// crawler contains the main entry point, and is designed to orchestrate / trigger the scrape and crawl process.

package crawl

import (
	"log"
	"net/url"
)

func WalkTarget(target *url.URL) HtmlPage {
	// WalkTarget creates a page instance against the specified target, fires the appropriate scraper, then fires goroutines to recurse

	// allPages is a map of pointers to every page we have trawled
	// we do this to avoid loopbacks (i.e deep pages looping back to root and causing an infinite loop)

	// implementation note: I originally wrote this as just passing through a pointer to this variable to all goroutines
	// then I realised that caused massive race conditions when multiple workers are trying to write at the same time,
	// so I've switched to using read / write channels, which are passed down recursively through the tree scraper stack
	allPages := make(map[url.URL]*HtmlPage)

	getPage := make(chan *readPage)
	setPage := make(chan *writePage)

	// Start the Map Storage Provider on the getPage and setPage channels
	// see the comments inside that function for why it's necessary
	go mapStorageProvider(allPages, getPage, setPage)

	// Seed the root page
	root := HtmlPage{Url: target}

	// we DON'T add the current page to the allPages index here, because we need to handle redirects, normalisation, and other stuff
	// instead, fetchAndParse handles it for us
	allPages[*target] = &root

	// do fetchAndParse, check that's okay, THEN go into recurse
	// these are done separately and on the main thread because
	// we need the root element to complete (and be valid)
	// before we can start branching out
	err := root.fetchAndParse(getPage, setPage)

	if err != nil {
		// we were unable to scrape the initial page, so we can't continue!
		log.Fatalf("☠️ Unable to scrape root document; the following error occured: %s", err)
	}
	root.recurse(getPage, setPage)
	// root.recurse()

	log.Println("🙌 crawler finished!")

	return root
}
