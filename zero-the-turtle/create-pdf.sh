#!/usr/bin/env bash

# pandoc book.md -o turtle-book.pdf --number-sections --number-offset=0

# pandoc book.md -o turtle-book.pdf # --variable=numbersections:false

#pandoc book.md --pdf-engine=lualatex \
pandoc book.md --pdf-engine=lualatex \
   -V 'mainfont:Dejavu Sans' \
   -V 'monofont:Dejavu Sans Mono' \
   -V geometry="left=1in,right=1in,top=1.5in,bottom=1.5in" \
   -V colorlinks=true \
   -o turtle-book.pdf



#   --toc --toc-depth=3 \
#   -V 'sansfont:Ubuntu Mono' \
#   -V 'monofont:Ubuntu Mono' \
#   -V 'mathfont:Ubuntu Mono' \
#   -V 'mainfontoptions:Extension=.ttf, UprightFont=*, BoldFont=*-Bold, ItalicFont=*-Italic, BoldItalicFont=*-BoldItalic' \


# --variable mainfont="Monolisa" \
# --variable sansfont="Monolisa" --variable monofont="Monolisa"