```text
|\\\|//
| |~~~| 
| |.:.;_     ____
| |:;:|  \  |  o | 
| |`:`|   |/ ___\|
 \|~~~;|_|_|/
```

## Changelog

### [v2.0.0] - 2025-03-27

While maintaining overall prose and structure, we had to make some changes
from the initial version of the book to come up with the current version
in a timely manner. We may add certain features (such as footnotes) in 
the upcoming versions by introducing further changes and automation scripts.

Here are notable changes since v1.0.0 of the book:

* Moved footnotes as "side notes", as they play better with markdown, and HTML
  formatting.
* Figure numbers has been renamed to start from index 1, in the former book
  each figure was linked to a chapter. We can change this with some Lua
  scripting later; but it turned out to be tedious to do it with an
  out-of-the-box Pandoc setup.
* Instead of a single PDF output, we now have four formats: html, markdown, 
  epub, and pdf. The **PDF** version is the most ready version for consumption
  while html and epub versions will need further work in the upcoming releases.
* The ebook generation is **fully-automatable** and scriptable, we already
  have generator scripts in the repo.
* Replaced the typography with Open Source fonts (Dejavu Sans family)
* The initial iteration was focused on keeping the content as intact as
  possible. A follow-up iteration can focus more on modernizing the content.
* Updated the "Secrets Management" chapter introducing products like
  SPIKE and VMware Secrets Manager---We plan to create an ebook soley on
  SPIFFE-based secrets management later.
