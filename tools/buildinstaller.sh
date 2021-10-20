#!/usr/bin/env bash
#
# This script assumes a linux environment

set -e
if [ -d "dist" ]; then
  rm -rf dist
fi

mkdir -p dist

cd src/server
go build -o fastforward
cd ../../
mv src/server/fastforward ./dist/

cp -r deployment/dep ./dist/
cp -r deployment/fastforwardSetup.sh ./dist/

cp -r config/cfg ./dist/
cd dist
tar -czvf fastforward.tar.gz --exclude fastforwardSetup.sh *


find . -type f ! \( -name "fastforward.tar.gz" -o -name "fastforwardSetup.sh" \) -delete
find . -type d -empty -delete
cat fastforward.tar.gz >> fastforwardSetup.sh
