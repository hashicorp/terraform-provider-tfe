#!/bin/bash

while read dirname others; do
    mkdir "$dirname"
done < list.txt

# find . -type d -exec touch {}/import.sh \;
find . -type d -exec touch {}/data-source.tf \;