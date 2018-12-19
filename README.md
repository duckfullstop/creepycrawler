# CREEPYCRAWLER 🕸

_creepycrawler_ is a simple, concurrent web scraper and mapper, written in Go.

It is not special or significant in any way.

It's also the second thing I've ever written in Go, so a little slack would be appreciated! 🙏

Considerations which have not been implemented, but are arguably out of scope for this assignment:

 * robots.txt testing (don't scrape things which we aren't allowed to)
 * Display of a proper "web" of relations
   * this is stored and can be easily calculated, but without writing a load of boilerplate code for making fancy webs, the tree will have to do
   (it prints a note next to back references)
 * Better logging / the ability to turn logging off
   * Right now it's super verbose so you can see what's happening, but I would probably switch to using something like `github.com/sirupsen/logrus` to add definable loglevels.

Package versions are pinned using `go dep`.

## Stack Design

I've written the stack like this:

1. Command comes in through `cmd/creepycrawler/main.go`
2. `cmd` calls `crawl.WalkTarget` with the target URL
3. `crawl.WalkTarget` pulls the target URL into a `HtmlPage` struct, parses it, and fills it out
  * Any links on the page are checked against a central truth store (to avoid branch duplication)
  * This truth store is a simple `map` accessible only via a read and write channel (controlled by a single worker goroutine)
4. `crawl.WalkTarget` then calls `HtmlPage.recurse()`
5. For every link in `HtmlPage`, `recurse()` starts a new goroutine
  * This goroutine runs the page scraper, then calls `recurse()` of its own
  * The idea being that each subpage gets its own goroutine worker, then each subpage of each subpage gets its own goroutine worker, etcetera
  * To ensure workers don't try to start recursing a page already belonging to another worker, a `mutex` is used on `HtmlPage`
6. Once each worker has finished, it returns a Completed boolean message on a channel back to its owning worker, then exits
  * A `recurse()` func only returns once all of its launched subworkers have returned Completed messages
7. `crawl.WalkTarget` returns a completed HtmlPage containing a web of pointers to other HtmlPages.
  * Loops are possible if you just keep recursing down `HtmlPage.LinksTo`
8. `cmd` then calls the `displayTree` package in order to get a human-readable tree, which is calculated with the help of `github.com/disiqueira/gotree`
  * Some funky deduplication / recursion checking goes on inside `displayTree` to avoid infinite loops / make it clear which sublinks are links to other, already found pages
9. `cmd` prints the result of `displayTree` to console

## Usage Instructions 🤔

A makefile is included to make this as simple as possible. To run:

1. Move everything here to `$GOROOT/src/github.com/luaduck/creepycrawler`
2. Install dependencies: `make dep`
  * You need `dep` installed to do this. If it's not available for whatever reason, `go get -u github.com/golang/dep/cmd/dep`
3. (optionally) Run tests: `make test`
4. Build the app: `make build`
  * This drops a binary at `./creepycrawler`
5. Run the crawler: `./creepycrawler`
  * The syntax is as follows: `./creepycrawler [OPTIONS] domain`
  * The option `-show-backrefs` will show entries in the tree which refer to a page whose tree has already been displayed, such as those further back down the stack
    (e.g a subpage referring back to root):
    turning this off will purely show a list of all domains, along with the depth at which they were first discovered
  * `domain` should be defined in full RFC1738 format. If a scheme isn't provided, it will default to `https`.

## License ⚖️

creepycrawler is licensed under the Unlicense. Please see [LICENSE](LICENSE) for more information.