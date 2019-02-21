# cmd.creepycrawler ⌨️

The creepycrawler console application is built from here.

It imports appropriately from packages which are defined in `pkg/`.

(You should probably build this using the Makefile in `/`)

## Lack of Tests 🙾✅☹️

There aren't integration or unit tests here because I'm not completely sure how to test CLI commands in golang. If I wrote them, they would (pseudo):

  * test command help is displayed when an invalid argument set is provided
  * test that flags do what they are defined to do
  
(probably fairly similarly to `pkg/crawl/crawler_test.go`)