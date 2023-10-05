#!/bin/sh

VERSION=$(echo "$GITHUB_REF" | rev | cut -d/ -f1 | rev)

echo ${VERSION:=$(date +%s)}

exit 0