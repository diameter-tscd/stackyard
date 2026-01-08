package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"test-go/config"
	"test-go/internal/monitoring"
	"test-go/internal/server"
	"test-go/pkg/logger"
	"test-go/pkg/tui"
	"test-go/pkg/utils"
	"time"
)

// Flag definitions - configure flags here
var flagDefinitions = []utils.FlagDefinition{
	{
		Name:         "c",
		DefaultValue: "",
		Description:  "URL to load configuration from (YAML format)",
		Validator: func(value interface{}) error {
			if str, ok := value.(string); ok && str != "" {
				if _, err := url.ParseRequestURI(str); err != nil {
					return fmt.Errorf("invalid config URL format: %w", err)
				}
			}
			return nil
		},
	},
	// Add new flags here easily:
	// {
	//     Name:         "port",
	//     DefaultValue: 8080,
	//     Description:  "Server port to listen on",
	//     Validator: func(value interface{}) error {
	//         if port, ok := value.(int); ok && (port < 1 || port > 65535) {
	//             return fmt.Errorf("port must be between 1 and 65535")
	//         }
	//         return nil
	//     },
	// },
	// {
	//     Name:         "debug",
	//     DefaultValue: false,
	//     Description: "Enable debug mode",
	//     Validator:   nil, // No validation needed for bool flags
	// },
}

func main() {
	// Clear the terminal screen for a fresh start
	utils.ClearScreen()

	// 1. Parse command line flags
	parsedFlags, err := utils.ParseFlags(flagDefinitions)
	if err != nil {
		fmt.Printf("Flag parsing error: %s\n", err.Error())
		utils.PrintUsage(flagDefinitions, "stackyard")
		os.Exit(1)
	}

	// 2. Load Config
	var cfg *config.Config
	if parsedFlags.ConfigURL != "" {
		// Load config from URL
		fmt.Printf("Loading config from URL: %s\n", parsedFlags.ConfigURL)
		if err := utils.LoadConfigFromURL(parsedFlags.ConfigURL); err != nil {
			fmt.Printf("Failed to load config from URL: %s\n", err.Error())
			os.Exit(1)
		}

		// Parse the loaded config
		cfg, err = config.LoadConfigWithURL(parsedFlags.ConfigURL)
		if err != nil {
			panic("Failed to parse config from URL: " + err.Error())
		}
	} else {
		// Load config from local file
		cfg, err = config.LoadConfig()
		if err != nil {
			panic("Failed to load config: " + err.Error())
		}
	}

	// Check if "web" folder exists, if not, disable web monitoring
	if _, err := os.Stat("web"); os.IsNotExist(err) {
		fmt.Println("\033[33m 'web' folder not found, disabling web monitoring\033[0m")
		cfg.Monitoring.Enabled = false
	}

	// 2. Load Banner
	var bannerText string
	if cfg.App.BannerPath != "" {
		banner, err := os.ReadFile(cfg.App.BannerPath)
		if err == nil {
			bannerText = string(banner)
		}
	}

	// 3. Check port availability
	if err := utils.CheckPortAvailability(cfg.Server.Port, cfg.Monitoring.Port, cfg.Monitoring.Enabled); err != nil {
		fmt.Printf("\033[31m Port Error: %s\033[0m\n", err.Error())
		fmt.Println("\033[33mPlease stop the conflicting service or change the port in config.yaml\033[0m")
		os.Exit(1)
	}

	// 4. Init Broadcaster for monitoring
	broadcaster := monitoring.NewLogBroadcaster()

	// Check if TUI mode is enabled
	if cfg.App.EnableTUI {
		// ===== TUI MODE =====
		runWithTUI(cfg, bannerText, broadcaster)
	} else {
		// ===== TRADITIONAL CONSOLE MODE =====
		runWithConsole(cfg, bannerText, broadcaster)
	}
}

// runWithTUI runs the application with fancy TUI interface
func runWithTUI(cfg *config.Config, bannerText string, broadcaster *monitoring.LogBroadcaster) {
	// Config conditions
	if !cfg.Monitoring.Enabled {
		cfg.Monitoring.Port = "disabled"
	}

	tuiConfig := tui.StartupConfig{
		AppName:     cfg.App.Name,
		AppVersion:  cfg.App.Version,
		Banner:      bannerText,
		Port:        cfg.Server.Port,
		MonitorPort: cfg.Monitoring.Port,
		Env:         cfg.App.Env,
		IdleSeconds: cfg.App.StartupDelay,
	}

	// Get service configurations
	serviceConfigs := getServiceConfigs(cfg)

	// Define boot sequence
	initQueue := []tui.ServiceInit{
		{Name: "Configuration", Enabled: true, InitFunc: nil},
	}

	// Add infrastructure services to boot queue
	for _, svc := range serviceConfigs {
		initQueue = append(initQueue, tui.ServiceInit{
			Name: svc.Name, Enabled: svc.Enabled, InitFunc: nil,
		})
	}

	initQueue = append(initQueue, tui.ServiceInit{Name: "Middleware", Enabled: true, InitFunc: nil})

	// Dynamically add services from config
	for name, enabled := range cfg.Services {
		initQueue = append(initQueue, tui.ServiceInit{Name: "Service: " + name, Enabled: enabled, InitFunc: nil})
	}

	// Add monitoring last
	initQueue = append(initQueue, tui.ServiceInit{Name: "Monitoring", Enabled: cfg.Monitoring.Enabled, InitFunc: nil})

	// Run the boot sequence TUI
	_, _ = tui.RunBootSequence(tuiConfig, initQueue)

	// Create Live TUI for continuous display
	liveTUI := tui.NewLiveTUI(tui.LiveConfig{
		AppName:     cfg.App.Name,
		AppVersion:  cfg.App.Version,
		Banner:      bannerText,
		Port:        cfg.Server.Port,
		MonitorPort: cfg.Monitoring.Port,
		Env:         cfg.App.Env,
		OnShutdown:  utils.TriggerShutdown, // Pass the shutdown callback
	})

	// Init Logger (quiet mode so logs go to TUI only)
	// We also broadcast to the monitoring system so the Web UI Live Logs work
	multiWriter := io.MultiWriter(liveTUI, broadcaster)
	l := logger.NewQuiet(cfg.App.Debug, multiWriter)

	// Start Live TUI in background
	liveTUI.Start()

	// Give TUI a moment to initialize
	time.Sleep(100 * time.Millisecond)

	// Add initial logs
	liveTUI.AddLog("info", "Server starting on port "+cfg.Server.Port)
	liveTUI.AddLog("info", "Environment: "+cfg.App.Env)

	// Start Server in background
	srv := server.New(cfg, l, broadcaster)
	go func() {
		liveTUI.AddLog("info", "HTTP server listening...")
		if err := srv.Start(); err != nil {
			liveTUI.AddLog("fatal", "Server error: "+err.Error())
		}
	}()

	// Give server a moment to start
	time.Sleep(500 * time.Millisecond)
	liveTUI.AddLog("info", "Server ready at http://localhost:"+cfg.Server.Port)
	if cfg.Monitoring.Enabled {
		liveTUI.AddLog("info", "Monitoring at http://localhost:"+cfg.Monitoring.Port)
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal or shutdown channel
	select {
	case <-sigChan:
		liveTUI.AddLog("warn", "Shutting down...")
		srv.Shutdown(context.Background(), l)
	case <-utils.ShutdownChan:
		liveTUI.AddLog("warn", "Shutting down...")
		srv.Shutdown(context.Background(), l)
	}

	liveTUI.Stop()

	// Give a moment for cleanup and then exit
	time.Sleep(100 * time.Millisecond)
	os.Exit(0)
}

// runWithConsole runs the application with traditional console logging
func runWithConsole(cfg *config.Config, bannerText string, broadcaster *monitoring.LogBroadcaster) {
	// Print banner to console
	if bannerText != "" {
		fmt.Print("\033[35m") // Purple color
		fmt.Println(bannerText)
		fmt.Print("\033[0m") // Reset color
	}

	// Init Logger (normal mode with console output)
	l := logger.New(cfg.App.Debug, broadcaster)

	// Log startup info
	l.Info("Starting Application", "name", cfg.App.Name, "env", cfg.App.Env)
	l.Info("TUI mode disabled, using traditional console logging")

	// Log enabled services
	l.Info("Initializing services...")

	// Log infrastructure services using unified config
	serviceConfigs := getServiceConfigs(cfg)
	for _, svc := range serviceConfigs {
		logServiceStatus(l, svc.Name, svc.Enabled)
	}

	// Dynamically log all services from config
	for name, enabled := range cfg.Services {
		logServiceStatus(l, "Service: "+name, enabled)
	}

	logServiceStatus(l, "Monitoring", cfg.Monitoring.Enabled)

	// Start Server
	srv := server.New(cfg, l, broadcaster)
	go func() {
		l.Info("HTTP server listening", "port", cfg.Server.Port)
		if err := srv.Start(); err != nil {
			l.Fatal("Server error", err)
		}
	}()

	// Give server a moment to start
	time.Sleep(500 * time.Millisecond)
	l.Info("Server ready", "url", "http://localhost:"+cfg.Server.Port)
	if cfg.Monitoring.Enabled {
		time.Sleep(500 * time.Millisecond)
		l.Info("Monitoring dashboard", "url", "http://localhost:"+cfg.Monitoring.Port)
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal
	<-sigChan

	l.Warn("Shutting down...")
	srv.Shutdown(context.Background(), l)

	// Give a moment for cleanup and then exit
	time.Sleep(100 * time.Millisecond)
	os.Exit(0)
}

// ServiceConfig represents a service with its name and enabled status
type ServiceConfig struct {
	Name    string
	Enabled bool
}

// getServiceConfigs returns a unified list of all service configurations
func getServiceConfigs(cfg *config.Config) []ServiceConfig {
	return []ServiceConfig{
		{Name: "Grafana", Enabled: cfg.Grafana.Enabled},
		{Name: "MinIO", Enabled: cfg.Monitoring.MinIO.Enabled},
		{Name: "Redis Cache", Enabled: cfg.Redis.Enabled},
		{Name: "Kafka Messaging", Enabled: cfg.Kafka.Enabled},
		{Name: "PostgreSQL", Enabled: cfg.Postgres.Enabled},
		{Name: "MongoDB", Enabled: cfg.Mongo.Enabled},
		{Name: "Cron Scheduler", Enabled: cfg.Cron.Enabled},
		{Name: "External Services", Enabled: (len(cfg.Monitoring.External.Services) > 0)},
	}
}

// logServiceStatus logs whether a service is enabled or skipped
func logServiceStatus(l *logger.Logger, name string, enabled bool) {
	if enabled {
		l.Info("Service initialized", "service", name, "status", "enabled")
	} else {
		l.Debug("Service skipped", "service", name, "status", "disabled")
	}
}
