package main

import (
	"fmt"
	"os"

	"github.com/yuukiyuuna/MangaMetaManager/cmd/mmm"
)

func main() {
	if err := mmm.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
