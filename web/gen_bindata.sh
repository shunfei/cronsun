#!/bin/sh

cd ui
npm run build
cd ..
go-bindata -pkg "web" -prefix "ui/dist/" -o static_assets.go ./ui/dist/
