#!/usr/bin/env bash

# |\\\|//
# | |~~~|  Solving the Bottom Turtle
# | |.:.;_     ____
# | |:;:|  \  |  o |
# | |`:`|   |/ ___\|
#  \|~~~;|_|_|/

cp -r src/assets .

pandoc ./src/book.md -o dist/epub/solving-the-bottom-turtle-v2.0.0.epub \
  --epub-cover-image=src/assets/Image_001.jpg \
  --toc \
  --metadata title="Solving the Bottom Turtle" \
  --metadata author="Daniel Feldman, Emily Fox, Evan Gilman, Ian Haken, \
Frederick Kautz, Umair Khan, Max Lambrecht, Brandon Lum, Agustín Martínez Fayó, \
Eli Nesterov, Andres Vega, Michael Wardrop, Volkan Özçelik"

rm -rf ./assets