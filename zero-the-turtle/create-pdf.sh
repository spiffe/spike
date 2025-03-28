#!/usr/bin/env bash

# |\\\|//
# | |~~~|  Solving the Bottom Turtle
# | |.:.;_     ____
# | |:;:|  \  |  o |
# | |`:`|   |/ ___\|
#  \|~~~;|_|_|/

pandoc ./src/book.md --pdf-engine=lualatex \
   -V 'mainfont:Dejavu Sans' \
   -V 'monofont:Dejavu Sans Mono' \
   -V geometry="left=1in,right=1in,top=1.5in,bottom=1.5in" \
   -V colorlinks=true \
   -o dist/pdf/solving-the-bottom-turtle-v2.0.0.pdf
