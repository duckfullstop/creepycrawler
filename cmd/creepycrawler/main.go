package main

import (
	"flag"
	"fmt"
	"os"
	"net/url"
	"log"
	"github.com/luaduck/creepycrawler/pkg/crawl"
	"github.com/luaduck/creepycrawler/pkg/displayTree"
)

func cmdUsage() {
	// cmdUsage simply prints how the command should be used.
	fmt.Printf("Usage: %s [OPTIONS] domain\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	displayBackrefs := flag.Bool("show-backrefs", false, "Show references to previously parsed / lower pages in the map tree.")

	// specify that the flag package should use our custom help handler for usage information
	// not sure if this is strictly necessary?
	flag.Usage = cmdUsage

	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// fire the main scraper code

	// FIXME whoops, I overrode url here accidentially, would probably declare this with a different name
	url, err := url.Parse(flag.Args()[0])

	if err != nil {
		// make this fail more gracefully
		log.Fatalln(err)
	}

	scrapedPage := crawl.WalkTarget(url)

	fmt.Println(*displayTree.StringPageTree(&scrapedPage, displayBackrefs))
}