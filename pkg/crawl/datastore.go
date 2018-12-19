package crawl

import "net/url"

type readPage struct {
	// readPage is a request to the allPages store for information about a page.
	// It's designed to be sent over a channel.
	key      url.URL
	response chan *HtmlPage
}

type writePage struct {
	// readPage is a request to the allPages store to write information about a page.
	// It's designed to be sent over a channel.
	key      url.URL
	value    *HtmlPage
	response chan bool
}

func mapStorageProvider(allPages map[url.URL]*HtmlPage, getPage chan *readPage, setPage chan *writePage) {
	// mapStorageProvider is a storage provider that provides access to the given allPages map.
	// It blocks on channel requests, ensuring that reads and writes are performed synchronously.
	// Without this, you can end up in a state where multiple parser threads declare a new page simultaneously,
	// and are unaware of another thread doing it, causing a race condition / duplication.
	for {
		select {
		case get := <-getPage:
			get.response <- allPages[get.key]
		case set := <-setPage:
			allPages[set.key] = set.value
			set.response <- true
		}
	}
}
