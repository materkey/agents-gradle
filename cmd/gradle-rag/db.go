package main

import "embed"

//go:embed db/*.db
var embeddedDBs embed.FS
