package main

import (
	"github.com/sean-/seed"

	memfs "github.com/zbiljic/memfs/cmd"
)

func init() {
	/* #nosec */
	seed.Init()
}

func main() {
	memfs.Main()
}
