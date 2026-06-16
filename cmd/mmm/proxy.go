package mmm

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
	"github.com/yuukiyuuna/MangaMetaManager/internal/network"
)

func validProxyType(proxyType string) bool {
	switch strings.ToLower(proxyType) {
	case "http", "https", "socks5":
		return true
	default:
		return false
	}
}

func validateProxyInput(enabled bool, proxyType, host string, port int) error {
	if enabled {
		if !validProxyType(proxyType) {
			return fmt.Errorf("invalid proxy type: %s", proxyType)
		}
		if host == "" {
			return fmt.Errorf("host cannot be empty when proxy is enabled")
		}
		if port <= 0 || port > 65535 {
			return fmt.Errorf("invalid port number: %d", port)
		}
	}
	return nil
}

func validateProviderProxyStrategy(input models.ProviderProxyStrategy) error {
	switch input.Strategy {
	case "inherit", "disabled":
		return nil
	case "custom":
		return validateProxyInput(true, input.Type, input.Host, input.Port)
	default:
		return fmt.Errorf("invalid proxy strategy: %s", input.Strategy)
	}
}

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Manage global proxy settings",
}

var proxyShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current global proxy settings",
	Run: func(cmd *cobra.Command, args []string) {
		dbFlag, _ := cmd.Flags().GetString("db")
		dbPath := databasePath(dbFlag)
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
		dbFlag, _ := cmd.Flags().GetString("db")
		dbPath := databasePath(dbFlag)
		models.InitDB(dbPath)

		enabled, _ := cmd.Flags().GetBool("enabled")
		pType, _ := cmd.Flags().GetString("type")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		user, _ := cmd.Flags().GetString("username")
		pass, _ := cmd.Flags().GetString("password")
		noProxy, _ := cmd.Flags().GetString("noproxy")
		if err := validateProxyInput(enabled, pType, host, port); err != nil {
			log.Fatal(err)
		}

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
			if err := models.DB.Create(&newSettings).Error; err != nil {
				log.Fatalf("Failed to save proxy settings: %v", err)
			}
		} else {
			if pass == "" {
				newSettings.Password = settings.Password
			}
			if err := models.DB.Model(&settings).Updates(newSettings).Error; err != nil {
				log.Fatalf("Failed to update proxy settings: %v", err)
			}
		}
		fmt.Println("Global proxy settings updated.")
	},
}

var proxyTestCmd = &cobra.Command{
	Use:   "test [url]",
	Short: "Test proxy connectivity",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbFlag, _ := cmd.Flags().GetString("db")
		dbPath := databasePath(dbFlag)
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
		dbFlag, _ := cmd.Flags().GetString("db")
		dbPath := databasePath(dbFlag)
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
		if err := validateProviderProxyStrategy(input); err != nil {
			log.Fatal(err)
		}

		var existing models.ProviderProxyStrategy
		result := models.DB.Where("provider_id = ?", providerID).First(&existing)
		if result.Error != nil {
			if err := models.DB.Create(&input).Error; err != nil {
				log.Fatalf("Failed to save provider proxy strategy: %v", err)
			}
		} else {
			if input.Password == "" {
				input.Password = existing.Password
			}
			if err := models.DB.Model(&existing).Updates(input).Error; err != nil {
				log.Fatalf("Failed to update provider proxy strategy: %v", err)
			}
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
	proxyCmd.PersistentFlags().StringP("db", "d", "", "Path to SQLite database (default from config or mmm.db)")
	providerProxyCmd.PersistentFlags().StringP("db", "d", "", "Path to SQLite database (default from config or mmm.db)")
}
