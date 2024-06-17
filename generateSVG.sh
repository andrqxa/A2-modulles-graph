#!/bin/bash

# Check if the file argument is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <dot-file>"
  exit 1
fi

DOT_FILE=$1
SVG_FILE="${DOT_FILE%.*}.svg"

# Generate SVG from the dot file
dot -Tsvg "$DOT_FILE" -o "$SVG_FILE"

# Output the result
echo "SVG file generated: $SVG_FILE"
