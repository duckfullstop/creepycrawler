# creepycrawler.crawl 🚼

crawl is a package which recursively crawls a given HTML page for hyperlinks, and returns a tree of relations between them.

You probably want one function, and one function alone; `crawl.WalkTarget(string)`

It returns instances of `crawl.HtmlPage`, which implements the interface `crawl.Page`.
I've implemented it this way to allow for extensibility in future (for example, mapping things that aren't HTML).


## Tests ✅

Test coverage is ~91.8%. Most everything is tested, excluding some dire emergency error catch clauses.

This obviously isn't ideal, but it should be possible to get this to 100% with a little extra work.

`page.go` is tested by other testsuites (including `crawler_test.go`).