package mmm

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mmm",
	Short: "MangaMetaManager is a tool to manage manga metadata.",
	Long:  `MangaMetaManager is a tool to manage manga metadata, supporting CBZ/ZIP formats, ComicInfo.xml integration, and multiple metadata providers.`,
}

func Execute() error {
	return rootCmd.Execute()
}
