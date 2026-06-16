package mmm

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/yuukiyuuna/MangaMetaManager/internal/api"
	"github.com/yuukiyuuna/MangaMetaManager/internal/core"
	"github.com/yuukiyuuna/MangaMetaManager/internal/logger"
	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
	"github.com/yuukiyuuna/MangaMetaManager/internal/provider"
	"github.com/yuukiyuuna/MangaMetaManager/internal/scanner"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Web server",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize Logger
		logger.InitLogger()

		// Initialize Task Manager
		core.InitTaskManager()

		// Initialize Database
		dbFlag, _ := cmd.Flags().GetString("db")
		dbPath := databasePath(dbFlag)
		models.InitDB(dbPath)

		// Cleanup abandoned temp files on startup
		scanner.CleanupTempFiles()

		// Initialize Providers
		provider.InitProviders()

		r := gin.Default()

		apiGroup := r.Group("/api")
		{
			proxyHandler := api.NewProxyHandler()
			proxyHandler.RegisterRoutes(apiGroup)

			mangaHandler := api.NewMangaHandler()
			mangaHandler.RegisterRoutes(apiGroup)

			providerHandler := api.NewProviderHandler()
			providerHandler.RegisterRoutes(apiGroup)

			settingsHandler := api.NewSettingsHandler()
			settingsHandler.RegisterRoutes(apiGroup)
		}

		// Static Files
		r.Static("/assets", "web/dist/assets")
		r.StaticFile("/favicon.svg", "web/dist/favicon.svg")
		r.StaticFile("/icons.svg", "web/dist/icons.svg")

		r.NoRoute(func(c *gin.Context) {
			// Check if the request is for an API, if so, don't serve index.html
			if !strings.HasPrefix(c.Request.URL.Path, "/api") {
				c.File("web/dist/index.html")
			}
		})

		r.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetString("port")
		addr := serverAddress(host, port)

		log.Printf("Starting server on %s", addr)
		if err := r.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().String("host", "", "Host to listen on (default from config or 0.0.0.0)")
	serveCmd.Flags().StringP("port", "p", "", "Port to listen on (default from config or 8080)")
	serveCmd.Flags().StringP("db", "d", "", "Path to SQLite database (default from config or mmm.db)")
}
