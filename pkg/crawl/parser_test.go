package crawl

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"testing"
)

// Define some test data sets.
var testDataSingleAnchor = "<html><head><title>TestData</title></head><body><a href=\"http://testsite.test/1\">Test Site</a></body></html>"
var testDataDuplicateAnchor = "<html><head></head><body><a href=\"http://testsite.test/1\">Test Site</a><a href=\"http://testsite.test/1\">Duplicate Test Site</a></body></html>"
var testDataInvalidAnchorMarkup = "<html><body><a>This is screwed up.</a></body></html>"

func TestHtmlParsing(t *testing.T) {

	rootUrl, err := url.Parse("http://testsite.test/")

	if err != nil {
		t.Error("Failed to parse test URL (fault in url library???)")
	}

	testPage := HtmlPage{Url: rootUrl}

	testReader := ioutil.NopCloser(bytes.NewBufferString(testDataSingleAnchor))

	allPages := make(map[url.URL]*HtmlPage)
	getPage := make(chan *readPage)
	setPage := make(chan *writePage)

	// Start the Map Storage Provider on the getPage and setPage channels
	// see the comments inside that function for why it's necessary
	go mapStorageProvider(allPages, getPage, setPage)

	err = testPage.parseHTML(&testReader, getPage, setPage)

	if err != nil {
		t.Errorf("HtmlPage.parseHTML() returned an error during parsing: %s", err)
	}

	// ensure that the parser parsed things correctly:
	if testPage.Title != "TestData" {
		t.Errorf("Parsing of HTML <title/> returned an unexpected result: expected \"%s\", got \"%s\"", "TestData", testPage.Title)
	}

	if len(testPage.LinksTo) != 1 {
		t.Errorf("Parsing of HTML <a/> tags returned an unexpected number of results: expected 1, got %d", len(testPage.LinksTo))
	}

	for _, element := range testPage.LinksTo {
		// iteration is unfortunately a little mandatory here because we're dealing with a list,
		// even though that list only has one element - I guess it's open in future though
		if element.Url.String() != "http://testsite.test/1" {
			t.Errorf("Parsing of HTML <a/> tags returned an unexpected HREF: expected %s, got %s", "http://testsite.test/1", element.Url.String())
		}
	}

}

func TestHtmlParseDeduplication(t *testing.T) {
	rootUrl, err := url.Parse("http://testsite.test/")

	if err != nil {
		t.Error("Failed to parse test URL (fault in url library???)")
	}

	testPage := HtmlPage{Url: rootUrl}

	testReader := ioutil.NopCloser(bytes.NewBufferString(testDataDuplicateAnchor))

	allPages := make(map[url.URL]*HtmlPage)
	getPage := make(chan *readPage)
	setPage := make(chan *writePage)

	// Start the Map Storage Provider on the getPage and setPage channels
	// see the comments inside that function for why it's necessary
	go mapStorageProvider(allPages, getPage, setPage)

	err = testPage.parseHTML(&testReader, getPage, setPage)

	if err != nil {
		t.Errorf("HtmlPage.parseHTML() returned an error during parsing: %s", err)
	}

	if len(testPage.LinksTo) != 2 {
		t.Errorf("Parsing of HTML <a/> tags returned an unexpected number of results: expected 2, got %d", len(testPage.LinksTo))
	}

	origElement := testPage.LinksTo[0]

	for _, element := range testPage.LinksTo {
		// iteration is unfortunately a little mandatory here because we're dealing with a list,
		// even though that list only has one element - I guess it's open in future though
		if element.Url.String() != "http://testsite.test/1" {
			t.Errorf("Parsing of HTML <a/> tags returned an unexpected HREF: expected %s, got %s", "http://testsite.test/1", element.Url.String())
		}
		// store the first element
		if origElement != element {
			// deduping failed
			t.Errorf("Parsing of duplicated HTML <a/> tags returned pointer to different page (duplication!): expected %p, got %p", origElement, element)
		}
	}
}

func TestHtmlInvalidMarkup(t *testing.T) {
	rootUrl, err := url.Parse("http://testsite.test/")

	if err != nil {
		t.Error("Failed to parse test URL (fault in url library???)")
	}

	testPage := HtmlPage{Url: rootUrl}

	testReader := ioutil.NopCloser(bytes.NewBufferString(testDataInvalidAnchorMarkup))

	allPages := make(map[url.URL]*HtmlPage)
	getPage := make(chan *readPage)
	setPage := make(chan *writePage)

	// Start the Map Storage Provider on the getPage and setPage channels
	// see the comments inside that function for why it's necessary
	go mapStorageProvider(allPages, getPage, setPage)

	err = testPage.parseHTML(&testReader, getPage, setPage)

	if err != nil {
		t.Errorf("HtmlPage.parseHTML() returned an error during parsing: %s", err)
	}

	if len(testPage.LinksTo) > 0 {
		t.Errorf("Parsing of HTML <a/> tags returned an unexpected number of results: expected 0, got %d", len(testPage.LinksTo))
	}
}
