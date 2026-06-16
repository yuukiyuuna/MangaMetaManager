package mmm

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
	"github.com/yuukiyuuna/MangaMetaManager/internal/scanner"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan library folders for manga",
	Run: func(cmd *cobra.Command, args []string) {
		dbFlag, _ := cmd.Flags().GetString("db")
		dbPath := databasePath(dbFlag)
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
		dbFlag, _ := cmd.Flags().GetString("db")
		dbPath := databasePath(dbFlag)
		models.InitDB(dbPath)

		path, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatalf("Invalid folder path: %v", err)
		}
		info, err := os.Stat(path)
		if err != nil {
			log.Fatalf("Invalid folder path: %v", err)
		}
		if !info.IsDir() {
			log.Fatalf("Path is not a directory: %s", path)
		}

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
		dbFlag, _ := cmd.Flags().GetString("db")
		dbPath := databasePath(dbFlag)
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
	scanCmd.Flags().StringP("db", "d", "", "Path to SQLite database (default from config or mmm.db)")

	rootCmd.AddCommand(libraryCmd)
	libraryCmd.AddCommand(libraryAddCmd)
	libraryCmd.AddCommand(libraryCleanCmd)
	libraryCmd.PersistentFlags().StringP("db", "d", "", "Path to SQLite database (default from config or mmm.db)")
}
