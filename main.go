// nolint
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

// nolint: funlen
func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}

	err = wails.Run(&options.App{
		Title:             "Multibase",
		Width:             1024,
		Height:            800,
		MinWidth:          1024,
		MinHeight:         300,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		AlwaysOnTop:       false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Logger:        nil,
		LogLevel:      logger.DEBUG,
		OnStartup:     app.startup,
		OnDomReady:    app.domReady,
		OnShutdown:    app.shutdown,
		OnBeforeClose: app.beforeClose,
		Bind: []interface{}{
			app,
			app.ProjectModule,
			app.GRPCModule,
			app.ThriftModule,
			app.KafkaModule,
		},
		WindowStartState:                 options.Normal,
		EnableFraudulentWebsiteDetection: false,
		// Windows platform specific options
		Windows: &windows.Options{
			// WebviewIsTransparent: false,
			// WindowIsTranslucent:  false,
			DisableWindowIcon: false,
			// DisableFramelessWindowDecorations: false,
			WebviewUserDataPath: "",
		},
		Mac: &mac.Options{
			// TitleBar: &mac.TitleBar{
			// 	TitlebarAppearsTransparent: true,
			// 	HideTitle:                  false,
			// 	HideTitleBar:               false,
			// 	FullSizeContent:            false,
			// 	UseToolbar:                 false,
			// 	HideToolbarSeparator:       true,
			// },
			// Appearance:           mac.NSAppearanceNameDarkAqua,
			// WebviewIsTransparent: true,
			// WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "Multibase",
				Message: "Fast and lightweight GUI client for interacting with gRPC and Thrift servers.",
				Icon:    icon,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
