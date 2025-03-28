#!/usr/bin/env bash

# |\\\|//
# | |~~~|  Solving the Bottom Turtle
# | |.:.;_     ____
# | |:;:|  \  |  o |
# | |`:`|   |/ ___\|
#  \|~~~;|_|_|/

cp -r ./src/assets ./dist/html/
cp ./src/css/book.css ./dist/html/css

pandoc ./src/book.md -s -o dist/html/solving-the-bottom-turtle-v2.0.0.html \
  --css=./css/book.css \
  --toc \
  --metadata pagetitle="Solving the Bottom Turtle"