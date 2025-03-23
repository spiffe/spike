
This is the work-in-progress of converting the SPIFFE book into Markdown,
and extending it for the next release.

We will host it in this repository for a while until we determine a home for
it (*either a new repository; or a folder under `spiffe.io` GitHub repo)

Notable changes from the v1.0.0 of the book:

* Moved footnotes as "side notes", as they play better with markdown, and HTML
  formatting.
* Figure numbers has been renamed to start from index 1, in the former book
  each figure was linked to a chapter. We can change this with some Lua
  scripting later; but it turned out to be tedious to do it with an 
  out-of-the-box Pandoc setup.
* We now have three formats: html, markdown, and pdf. We can add further 
  formats like epub later.
* The ebook generation is **fully-automatable** and scriptable.
* Replace the typography with Open Source fonts (Dejavu Sans family)
* The initial iteration was focused on keeping the content as intact as
  possible. A follow-up iteration can focus more on modernizing the content.
* Updated the "Secrets Management" chapter introducing products like
  SPIKE and VMware Secrets Manager. --- We plan to create an ebook soley on
  SPIFFE-based secrets management later.
