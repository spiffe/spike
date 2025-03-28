```text
|\\\|//
| |~~~|  Solving the Bottom Turtle 
| |.:.;_     ____
| |:;:|  \  |  o | 
| |`:`|   |/ ___\|
 \|~~~;|_|_|/
```

## About

This is the source code repository of "[Solving the Bottom Turtle][turtle-book]", 
a comprehensive guide to the SPIFFE standard for service identity and SPIRE, 
its reference implementation.

[turtle-book]: https://spiffe.io/book/ "Solving the Bottom Turtle"

## About the Book

This book presents the SPIFFE standard for service identity and SPIRE, the 
reference implementation for SPIFFE. These projects provide a uniform identity 
control plane across modern, heterogeneous infrastructure. Both projects are open 
source and are part of the Cloud Native Computing Foundation.

As organizations grow their application architectures to make the most of new 
infrastructure technologies, their security models must also evolve. In this 
new infrastructure world, SPIFFE and SPIRE help keep systems secure by providing 
a deep understanding of the identity problem and how to solve it. With these 
projects, developers and operators can build software using new infrastructure 
technologies while allowing security teams to step back from expensive and 
time-consuming manual security processes.

## Repository Purpose

The initial version of the book was created in a format that was relatively 
harder to contribute to. This repository is a "Markdown" transformation of the 
book to allow contributions from the community.

This edition has been converted from the original PDF version of the book into 
Markdown format. While Markdown is less visually expressive compared to the 
original format, and certain decorative elements have been omitted during the 
conversion process, every effort has been made to preserve the integrity and 
intent of the original content.

The maintainers, with help from the community, have enriched the book by 
expanding the content, introducing new chapters, and updating the content by 
addressing suggestions from the community.

You can get the older version of the book [here][turtle-v1].

[turtle-v1]: dist/pdf/solving-the-bottom-turtle-v1.0.0.pdf

## Building From Source

To build the book locally make sure you have the fonts *Dejavu Sans* and 
*Dejavu Sans Mono* installed on your system.

### Prerequisites

Make sure you have installed `pandoc`

For Debian-based systems, you can use the package manager to install Pandoc:

```bash
sudo apt install pandoc
```

Also, make sure you install the necessary Pandoc plugins:

```bash
sudo apt install texlive-full
```

### Ebook Creation

The markdown version can be found at `./src/book.md`

You can run the following scripts to generate the format you need based on the 
markdown source (`./src/book.md`):

```bash
# Execute scripts from this folder:
cd ./zero-the-turtle

# Create PDF ebook at ./dist/pdf/$book.pdf
./create-pdf.sh

# Create HTML ebook at ./dist/html/$book.html
./create-html.sh

# Create EPUB ebook at ./dist/epub/$book.epub
./create-epub.sh
```

## Contributing

We welcome contributions from the community to help improve this book. To 
contribute:

1. Fork the repository
2. Make your changes to the Markdown source in `./src/book.md`
3. Test your changes by building the book locally
4. Submit a pull request with a clear description of your changes

Please ensure your contributions maintain the technical accuracy and writing 
style of the existing content.
