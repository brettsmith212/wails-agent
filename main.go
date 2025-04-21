package main

import (
	"context"
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:            "AI Agent",
		Width:            1200,
		Height:           800,
		MinWidth:         800,
		MinHeight:        600,
		MaxWidth:         1920,
		MaxHeight:        1080,
		DisableResize:    false,
		Fullscreen:       false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       func(_ context.Context) { os.Exit(0) },
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
			Theme:                             windows.SystemDefault,
		},
		Mac: &mac.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			TitleBar:             mac.TitleBarDefault(),
			Appearance:           mac.DefaultAppearance,
			About: &mac.AboutInfo{
				Title:   "AI Agent",
				Message: "© 2025 Your Name",
			},
		},
		Linux: &linux.Options{
			WebviewGpuPolicy: linux.WebviewGpuPolicyAlways,
			WindowIsTranslucent: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// package main

// import (
// 	"bufio"
// 	"context"
// 	"fmt"
// 	"log"
// 	"os"
// 	"path/filepath"

// 	"github.com/anthropics/anthropic-sdk-go"

// 	"wails-agent/agent"
// 	"wails-agent/logger"
// 	"wails-agent/tools"
// )

// func main() {
// 	// Initialize logger
// 	logDir := filepath.Join(".", "logs")
// 	if err := logger.Initialize(logDir); err != nil {
// 		log.Fatal("Failed to initialize logger:", err)
// 	}
// 	defer logger.Close()

// 	client := anthropic.NewClient()

// 	scanner := bufio.NewScanner(os.Stdin)
// 	getUserMessage := func() (string, bool) {
// 		if !scanner.Scan() {
// 			return "", false
// 		}
// 		return scanner.Text(), true
// 	}

// 	toolDefs := []tools.ToolDefinition{
// 		tools.ReadFileDefinition,
// 		tools.ListFilesDefinition,
// 		tools.EditFileDefinition,
// 		tools.RunShellCommandDefinition,
// 	}
// 	myAgent := agent.NewAgent(&client, getUserMessage, toolDefs)
// 	err := myAgent.Run(context.TODO())
// 	if err != nil {
// 		fmt.Printf("Error: %s\n", err.Error())
// 	}
// }
