package crawl

import (
	"net/url"
	"testing"
)

func TestCrawlOfHttpSite(t *testing.T) {
	// TestCrawlOfHttpSite tests a crawl of a live HTTP site.
	// Right now it's set up to crawl projectflower.eu, which is my pet project
	// (though you'd obviously ideally switch this out to something more controlled)

	// It does, however, require the system which this test is being run against to have an internet connection.

	// This test, as a result, covers most of the crawl package on its own.

	rootUrl, err := url.Parse("projectflower.eu")

	if err != nil {
		t.Error("Failed to parse test URL (fault in url library???)")
	}

	// return type is already asserted inside WalkTarget
	var result = WalkTarget(rootUrl)

	// ensure that the parser parsed things correctly:
	if result.Title != "Home | Flower" {
		t.Errorf("Parsing of HTML <title/> returned an unexpected result: expected \"%s\", got \"%s\"", "Home | Flower", result.Title)
	}

	if len(result.LinksTo) < 0 {
		t.Errorf("Parsing of HTML <a/> tags returned an unexpected number of results: expected at least some, got %d", len(result.LinksTo))
	}

	// I can't really flesh this out as much as I'd like to, because projectflower's sublinks are fairly changeable / isn't controlled.

	// Treat this test as more of a loose example.

	// See parser_test for an example of how I'd iterate through to test that links are correctly identified.

}
