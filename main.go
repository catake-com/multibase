// nolint
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
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
		Title:     "Multibase",
		Width:     1024,
		Height:    768,
		MinWidth:  1024,
		MinHeight: 300,
		// MaxWidth:          1280,
		// MaxHeight:         800,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		RGBA:              &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		Assets:            assets,
		Menu:              nil,
		Logger:            nil,
		LogLevel:          logger.DEBUG,
		OnStartup:         app.startup,
		OnDomReady:        app.domReady,
		OnBeforeClose:     app.beforeClose,
		OnShutdown:        app.shutdown,
		WindowStartState:  options.Normal,
		Bind: []interface{}{
			app,
			app.ProjectModule,
			app.GRPCModule,
		},
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
				Title:   "Vanilla Template",
				Message: "Part of the Wails projects",
				Icon:    icon,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
