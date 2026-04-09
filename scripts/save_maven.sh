#!/bin/bash
# Quick script to add Maven output from clipboard
# Usage: ./save_maven_output.sh

DIR="testdata/collector/full"

if [ ! -d "$DIR" ]; then
    mkdir -p "$DIR"
fi

TIMESTAMP=$(date +%s%N)
FILENAME="$DIR/maven_${TIMESTAMP}.txt"

# Try clipboard
if command -v xclip &> /dev/null; then
    xclip -selection clipboard -o > "$FILENAME"
elif command -v pbpaste &> /dev/null; then
    pbpaste > "$FILENAME"
else
    echo "No clipboard tool found. Save output to $FILENAME manually"
    exit 1
fi

echo "Saved to $FILENAME"

# Create meta file
echo "manual:clipboard" > "$FILENAME.meta"
echo "Created $FILENAME.meta"