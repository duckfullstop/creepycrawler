package displayTree

import (
	"testing"
	"github.com/luaduck/creepycrawler/pkg/crawl"
	"net/url"
	"errors"
)

func genTestTree() *crawl.HtmlPage {
	// there's almost definitely a cleaner way of doing this but whatever
	testRoot := &crawl.HtmlPage{
		Url: &url.URL{Scheme: "https", Host: "testsite.test", Path: "/"},
		Title: "TestRoot",
		IsParsed: true,
	}

	testElement1 := &crawl.HtmlPage{
		Url: &url.URL{Scheme: "https", Host: "testsite.test", Path: "/1"},
		Title: "TestElem1",
		IsParsed: true,
	}

	testElement2 := &crawl.HtmlPage{
		Url: &url.URL{Scheme: "https", Host: "testsite.test", Path: "/2"},
		Title: "TestElem2",
		IsParsed: true,
	}

	testElement3 := &crawl.HtmlPage{
		Url: &url.URL{Scheme: "https", Host: "testsite.test", Path: "/3"},
		Title: "TestElem3",
		IsParsed: false,
		CrawlError: errors.New("test"),
	}

	testRoot.LinksTo = append(testRoot.LinksTo, testElement1, testElement2)

	// element1 links to element3
	testElement1.LinksTo = append(testElement1.LinksTo, testElement3)

	// element2 also links back to root
	testElement2.LinksTo = append(testElement2.LinksTo, testRoot)

	return testRoot
}

func TestTreeData(t *testing.T) {
	displayBackrefs := true
	targetTree := pageTree(genTestTree(), &displayBackrefs)

	if targetTree.Text() != "https://testsite.test/ (TestRoot)" {
		t.Errorf("formatting on root node changed: expected %s, got %s", "https://testsite.test/ (TestRoot)", targetTree.Text())
	}

	if len(targetTree.Items()) != 2 {
		t.Errorf("TestElem1 has the wrong number of sub entries: expected %d, got %d", 2, len(targetTree.Items()))
	}

	// there is almost definitely a better way to test this rather than using statically declared if loops
	// would probably at least use helper utility functions if I had a sane place to declare them
	// e.g testUtilEnsureCount(item, count), testUtilEnsureMatch(item.Text(), "target")
	for _, item := range targetTree.Items() {
		// if we're hitting this point we can safely assume that formatting has not changed
		// so we bind against it
		if item.Text() == "https://testsite.test/1 (TestElem1)" {
			if len(item.Items()) != 1 {
				t.Errorf("TestElem1 has the wrong number of sub entries: expected %d, got %d", 1, len(item.Items()))
			}
			for _, subitem := range item.Items() {
				if subitem.Text() != "https://testsite.test/3 (parse error: test)" {
					t.Errorf("TestElem2->TestElem1 format incorrect, potential backref detection guru mediation? expected %s, got %s", "https://testsite.test/3 (parse error: test)", subitem.Text())
				}
			}
		} else if item.Text() == "https://testsite.test/2 (TestElem2)" {
			if len(item.Items()) != 1 {
				t.Errorf("TestElem2 has the wrong number of sub entries: expected %d, got %d", 1, len(item.Items()))
			}
			for _, subitem := range item.Items() {
				if subitem.Text() != "https://testsite.test/ (TestRoot) (🔙 lower or already parsed page)" {
					t.Errorf("TestElem2->TestElem1 format incorrect, potential backref detection guru mediation? expected %s, got %s", "https://testsite.test/ (TestRoot) (🔙 lower or already parsed page)", subitem.Text())
				}
			}
		} else {
			t.Errorf("encountered an unexpected element with text %s", item.Text())
		}
	}
}