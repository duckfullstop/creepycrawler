package crawl

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestDatastore(t *testing.T) {
	// TestDatastore simply ensures that the mapStorageProvider can consistently produce a single result.
	testUrl, err := url.Parse("http://testsite.test/")

	if err != nil {
		t.Error("Failed to parse test URL (fault in url library???)")
	}

	allPages := make(map[url.URL]*HtmlPage)
	getPage := make(chan *readPage)
	setPage := make(chan *writePage)

	// Start the Map Storage Provider on the getPage and setPage channels
	// see the comments inside that function for why it's necessary
	go mapStorageProvider(allPages, getPage, setPage)

	testPage := &HtmlPage{Url: testUrl}

	write := &writePage{key: *testUrl, value: testPage, response: make(chan bool)}

	setPage <- write

	if !<-write.response {
		t.Error("datastore writer returned failure on first storage event")
	}

	read := &readPage{key: *testUrl, response: make(chan *HtmlPage)}

	getPage <- read

	readResponse := <-read.response
	if readResponse != testPage {
		t.Errorf("datastore reader returned something that wasn't a pointer to our HtmlPage instance: expected %p, got %p", testPage, readResponse)
	}
}

func TestDatastoreSynchro(t *testing.T) {
	// TestDatastoreSynchro simply ensures that the mapStorageProvider can consistently produce results,
	// even when lots of requests are being flung at it.
	// _LOTS_ of requests (because trees could go very deep, and we could have lots of workers trying to read and write).
	allPages := make(map[url.URL]*HtmlPage)
	getPage := make(chan *readPage)
	setPage := make(chan *writePage)

	chErrors := make(chan error)
	chComplete := make(chan bool)

	defer close(chComplete)
	defer close(chErrors)

	// Start the Map Storage Provider on the getPage and setPage channels
	// see the comments inside that function for why it's necessary
	go mapStorageProvider(allPages, getPage, setPage)

	for totalWorkers := 0; totalWorkers < 300; totalWorkers++ {
		go testDatastoreWorker(getPage, setPage, chErrors, chComplete)
	}

	// Hold a lock until all channels complete
	var breakout bool
	var errorset []error
	for waiting := 0; waiting < 300 || breakout; {
		select {
		case <-chComplete:
			// each time a goroutine finishes (and sends us the message through chComplete), increment the finished counter
			waiting++
		case returnedError := <-chErrors:
			// an error occurred! (oh noes!)
			errorset = append(errorset, returnedError)
		case <-time.After(5 * time.Second):
			// print a warning after 5 seconds if all goroutines haven't returned yet
			log.Printf("5s timeout! (waiting for %d of %d goroutine(s))\n", 300-waiting, 300)
			breakout = true
		}
	}

	if len(errorset) > 0 {
		for _, v := range errorset {
			log.Printf("worker error: %s", v)
		}
		t.Error("one or more worker errors were reported!")
	}

}

func testDatastoreWorker(getPage chan *readPage, setPage chan *writePage, errorchannel chan error, completed chan bool) {
	defer func() { completed <- true }()
	// randomHostname is just a random 63-bit integer cast to string.
	// it's a little silly, but this is probably the fastest way to produce random hostnames
	// (without iterating randomly through, say, a glyph array)

	// this is potentially dangerous if we somehow end up with a collision, but we only have 300 instances
	// so the odds are pretty small
	randomHostname := strconv.Itoa(rand.Int())
	randomUrl, err := url.Parse("http://" + randomHostname + ".test/")

	if err != nil {
		errorchannel <- err
	}

	testPage := &HtmlPage{Url: randomUrl}

	write := &writePage{key: *randomUrl, value: testPage, response: make(chan bool)}

	setPage <- write

	if !<-write.response {
		errorchannel <- errors.New("datastore writer returned failure on first storage event")
	}

	read := &readPage{key: *randomUrl, response: make(chan *HtmlPage)}

	getPage <- read

	readResponse := <-read.response
	if readResponse != testPage {
		errorchannel <- errors.New(fmt.Sprintf("datastore reader returned something that wasn't a pointer to a worker's HtmlPage instance: expected %p, got %p", testPage, readResponse))
	}

}
