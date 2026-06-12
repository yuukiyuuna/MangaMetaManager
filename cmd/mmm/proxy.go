package mmm

import (
	"fmt"
	"log"

	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
	"github.com/yuukiyuuna/MangaMetaManager/internal/network"
	"github.com/spf13/cobra"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Manage global proxy settings",
}

var proxyShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current global proxy settings",
	Run: func(cmd *cobra.Command, args []string) {
		dbPath, _ := cmd.Flags().GetString("db")
		models.InitDB(dbPath)

		var settings models.ProxySettings
		if err := models.DB.First(&settings).Error; err != nil {
			fmt.Println("No global proxy settings found.")
			return
		}

		fmt.Printf("Enabled:  %v\n", settings.Enabled)
		fmt.Printf("Type:     %s\n", settings.Type)
		fmt.Printf("Host:     %s\n", settings.Host)
		fmt.Printf("Port:     %d\n", settings.Port)
		fmt.Printf("Username: %s\n", settings.Username)
		fmt.Printf("NoProxy:  %s\n", settings.NoProxy)
	},
}

var proxySetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set global proxy settings",
	Run: func(cmd *cobra.Command, args []string) {
		dbPath, _ := cmd.Flags().GetString("db")
		models.InitDB(dbPath)

		enabled, _ := cmd.Flags().GetBool("enabled")
		pType, _ := cmd.Flags().GetString("type")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		user, _ := cmd.Flags().GetString("username")
		pass, _ := cmd.Flags().GetString("password")
		noProxy, _ := cmd.Flags().GetString("noproxy")

		var settings models.ProxySettings
		result := models.DB.First(&settings)
		
		newSettings := models.ProxySettings{
			Enabled:  enabled,
			Type:     pType,
			Host:     host,
			Port:     port,
			Username: user,
			Password: pass,
			NoProxy:  noProxy,
		}

		if result.Error != nil {
			models.DB.Create(&newSettings)
		} else {
			if pass == "" {
				newSettings.Password = settings.Password
			}
			models.DB.Model(&settings).Updates(newSettings)
		}
		fmt.Println("Global proxy settings updated.")
	},
}

var proxyTestCmd = &cobra.Command{
	Use:   "test [url]",
	Short: "Test proxy connectivity",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbPath, _ := cmd.Flags().GetString("db")
		models.InitDB(dbPath)

		testURL := "https://www.google.com"
		if len(args) > 0 {
			testURL = args[0]
		}

		factory := network.NewHTTPClientFactory()
		client, err := factory.GetClient("")
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}

		fmt.Printf("Testing connection to %s...\n", testURL)
		resp, err := client.Get(testURL)
		if err != nil {
			fmt.Printf("Test failed: %v\n", err)
			return
		}
		defer resp.Body.Close()

		fmt.Printf("Success! Status Code: %d\n", resp.StatusCode)
	},
}

var providerProxyCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage provider specific proxy settings",
}

var providerProxySetCmd = &cobra.Command{
	Use:   "set [providerID]",
	Short: "Set proxy strategy for a provider",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbPath, _ := cmd.Flags().GetString("db")
		models.InitDB(dbPath)

		providerID := args[0]
		strategy, _ := cmd.Flags().GetString("strategy")
		
		var input models.ProviderProxyStrategy
		input.ProviderID = providerID
		input.Strategy = strategy

		if strategy == "custom" {
			input.Type, _ = cmd.Flags().GetString("type")
			input.Host, _ = cmd.Flags().GetString("host")
			input.Port, _ = cmd.Flags().GetInt("port")
			input.Username, _ = cmd.Flags().GetString("username")
			input.Password, _ = cmd.Flags().GetString("password")
		}

		var existing models.ProviderProxyStrategy
		result := models.DB.Where("provider_id = ?", providerID).First(&existing)
		if result.Error != nil {
			models.DB.Create(&input)
		} else {
			if input.Password == "" {
				input.Password = existing.Password
			}
			models.DB.Model(&existing).Updates(input)
		}
		fmt.Printf("Proxy strategy for provider '%s' updated to '%s'.\n", providerID, strategy)
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)
	proxyCmd.AddCommand(proxyShowCmd)
	proxyCmd.AddCommand(proxySetCmd)
	proxyCmd.AddCommand(proxyTestCmd)

	rootCmd.AddCommand(providerProxyCmd)
	providerProxyCmd.AddCommand(providerProxySetCmd)

	// Flags for proxy set
	proxySetCmd.Flags().Bool("enabled", true, "Enable proxy")
	proxySetCmd.Flags().String("type", "http", "Proxy type (http, https, socks5)")
	proxySetCmd.Flags().String("host", "", "Proxy host")
	proxySetCmd.Flags().Int("port", 0, "Proxy port")
	proxySetCmd.Flags().String("username", "", "Proxy username")
	proxySetCmd.Flags().String("password", "", "Proxy password")
	proxySetCmd.Flags().String("noproxy", "", "Comma separated list of hosts to bypass proxy")

	// Flags for provider proxy set
	providerProxySetCmd.Flags().String("strategy", "inherit", "Strategy (inherit, disabled, custom)")
	providerProxySetCmd.Flags().String("type", "http", "Custom proxy type")
	providerProxySetCmd.Flags().String("host", "", "Custom proxy host")
	providerProxySetCmd.Flags().Int("port", 0, "Custom proxy port")
	providerProxySetCmd.Flags().String("username", "", "Custom proxy username")
	providerProxySetCmd.Flags().String("password", "", "Custom proxy password")

	// Global DB flag for all commands
	proxyCmd.PersistentFlags().StringP("db", "d", "mmm.db", "Path to SQLite database")
	providerProxyCmd.PersistentFlags().StringP("db", "d", "mmm.db", "Path to SQLite database")
}
