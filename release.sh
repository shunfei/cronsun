#!/bin/sh

version=v0.1
if [[ $# -gt 0 ]]; then
	version="$1"
fi


declare -a goos=(
	linux
	darwin
)

for os in "${goos[@]}"; do
	export GOOS=$os GOARCH=amd64
	echo building $GOOS-$GOARCH
	sh build.sh
	mv dist cronsun-$version
	7z a cronsun-$version-$GOOS-$GOARCH.zip cronsun-$version
	rm -rf cronsun-$version
	echo
done
