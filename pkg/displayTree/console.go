package displayTree

import (
	"github.com/disiqueira/gotree"
	"github.com/luaduck/creepycrawler/pkg/crawl"
	"net/url"
	"fmt"
)

type stackElement struct {
	htmlPage *crawl.HtmlPage
	tree *gotree.Tree
}

func pageTree(page *crawl.HtmlPage, displayBackrefs *bool) gotree.Tree {
	// This isn't the most performant thing on the planet (it's not async, for one), but it's only here
	// because we need some way to dump the data in a human readable format.

	// It could easily be made way more performant by doing async with data storage similar to how I do it in crawl.

	// you get the idea

	tree := gotree.New(page.Url.String()+" ("+page.Title+")")

	// a map is absolutely overkill here, but it saves having to write a custom array search function
	// hell, you could probably do this as map[*url.URL]bool

	// map key is a pointer because it saves memory, plus we never deal with creating new url.URL entries here
	// and stuff is deduplicated for us beforehand (hence it's safe to do lookups)
	// if you wanted to be super safe you could make this a resolved lookup with duplication testing
	// but that feels overkill
	allPages := map[*url.URL]*crawl.HtmlPage{
		page.Url: page,
	}

	var f func(r *crawl.HtmlPage, src gotree.Tree)
	f = func(r *crawl.HtmlPage, src gotree.Tree) {
		// iterateStack is an array of recursion targets
		// this will be run after we have finished dealing with everything on the current layer
		var iterateStack []*stackElement
		for _, elem := range r.LinksTo {
			_, ok := allPages[elem.Url]
			if ok {
				// we have already found this index, so do not recurse
				if *displayBackrefs {
					src.Add(elem.Url.String()+" ("+elem.Title+") (🔙 lower or already parsed page)")
				}
			} else {
				// we haven't yet found this index, add it to the stack and keep recursing
				// (but only if we were able to parse it)
				if elem.IsParsed {
					subpage := src.Add(elem.Url.String()+" ("+elem.Title+")")
					allPages[elem.Url] = elem
					iterateStack = append(iterateStack, &stackElement{htmlPage: elem, tree: &subpage})
				} else {
					src.Add(fmt.Sprintf("%s (parse error: %s)", elem.Url.String(), elem.CrawlError))
				}

			}
		}

		for _, elem := range iterateStack {
			f(elem.htmlPage, *elem.tree)
		}
	}
	f(page, tree)

	return tree
}

func StringPageTree(page *crawl.HtmlPage, displayBackrefs *bool) *string {
	// This function is public, and calls pageTree in order to get a gotree tree representation
	// of the given HtmlPage.

	// This is a separate function so we can easily call into pageTree during our tests.

	treeObject := pageTree(page, displayBackrefs)

	treeString := treeObject.Print()

	return &treeString
}