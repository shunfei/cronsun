#!/bin/sh
BASEDIR=$(dirname "$0")
cd $BASEDIR/ui
npm run build
cd ..
go-bindata -pkg "web" -prefix "ui/dist/" -o static_assets.go ./ui/dist/
