package mmm

import (
	"fmt"
	"log"

	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
	"github.com/yuukiyuuna/MangaMetaManager/internal/scanner"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan library folders for manga",
	Run: func(cmd *cobra.Command, args []string) {
		dbPath, _ := cmd.Flags().GetString("db")
		models.InitDB(dbPath)

		fmt.Println("Scanning library folders...")
		if err := scanner.ScanLibrary(nil); err != nil {
			log.Fatalf("Scan failed: %v", err)
		}
		fmt.Println("Scan completed.")
	},
}

var libraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Manage library folders",
}

var libraryAddCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Add a library folder",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbPath, _ := cmd.Flags().GetString("db")
		models.InitDB(dbPath)

		path := args[0]
		folder := models.LibraryFolder{Path: path}
		if err := models.DB.Create(&folder).Error; err != nil {
			log.Fatalf("Failed to add folder: %v", err)
		}
		fmt.Printf("Added library folder: %s\n", path)
	},
}

var libraryCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean missing files from database",
	Run: func(cmd *cobra.Command, args []string) {
		dbPath, _ := cmd.Flags().GetString("db")
		models.InitDB(dbPath)

		fmt.Println("Cleaning library database...")
		if err := scanner.CleanLibrary(nil); err != nil {
			log.Fatalf("Clean failed: %v", err)
		}
		fmt.Println("Clean completed.")
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().StringP("db", "d", "mmm.db", "Path to SQLite database")

	rootCmd.AddCommand(libraryCmd)
	libraryCmd.AddCommand(libraryAddCmd)
	libraryCmd.AddCommand(libraryCleanCmd)
	libraryCmd.PersistentFlags().StringP("db", "d", "mmm.db", "Path to SQLite database")
}
