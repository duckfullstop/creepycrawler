# creepycrawler.displayTree 🌲

displayTree is a simple package designed to parse `crawl.HtmlPage` tree structs into more standard formats.

Right now it's just got one function, PrintPageTree(), which spews a representation of the tree into console (using `gotree`).

However, you could easily extend this package to allow for outputting in different formats (like HTML lists, XML, or JSON).

## Tests ✅

Test coverage is ~86.4%, though that's probably not doing it justice.
The only thing that isn't tested is StringPageTree(), but as that basically just stringifies pageTree() (which is thoroughly tested), 
it's not of extreme concern.