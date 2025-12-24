package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)



func main() {
	// abbreviatePath shortens a file path to fit within a reasonable display width
	// Returns abbreviated path with ellipsis if path is too long
	abbreviatePath := func(path string, maxChars int) string {
		if len(path) <= maxChars {
			return path
		}
		
		// Try to show filename at least
		filename := filepath.Base(path)
		if len(filename) > maxChars-5 {
			// If even filename is too long, truncate with ellipsis
			return filename[:maxChars-3] + "..."
		}
		
		// Show ellipsis + filename
		return "..." + string(os.PathSeparator) + filename
	}





	myApp := app.NewWithID("us.hakubi.dsda-launch")
	myWindow := myApp.NewWindow("DSDA Doom Launcher")
	myWindow.SetTitle("DSDA Doom Launcher")
	
	// Load window size from preferences
	prefs := myApp.Preferences()
	width := float32(prefs.FloatWithFallback("windowWidth", 1000))
	height := float32(prefs.FloatWithFallback("windowHeight", 500))
	myWindow.Resize(fyne.NewSize(width, height))

	// Preference keys
	const dsdaPathKey = "dsdaDoomPath"
	const iwadPathKey = "iwadPath"

	// File path storage
	var dsdaDoomPath string
	var iwadFile string
	var complevel string
	var pwadPaths []string

	// Load paths from preferences
	dsdaDoomPath = prefs.StringWithFallback(dsdaPathKey, "")
	iwadFile = prefs.StringWithFallback("iwadFile", "doom2.wad")
	complevel = prefs.StringWithFallback("complevel", "9")
	// Load PWAD paths
	pwadJSON := prefs.StringWithFallback("pwadPaths", "[]")
	json.Unmarshal([]byte(pwadJSON), &pwadPaths)

	// DSDA Doom Executable Path
	dsdaLabel := widget.NewRichTextFromMarkdown("**DSDA Doom Executable:**")
	dsdaPathLabel := widget.NewLabel("No file selected")
	dsdaPathLabel.Alignment = fyne.TextAlignLeading
	if dsdaDoomPath != "" {
		dsdaPathLabel.SetText(abbreviatePath(dsdaDoomPath, 50))
	}

	// Helper function to update label text with abbreviated path
	updateDsdaLabel := func(path string) {
		if path == "" {
			dsdaPathLabel.SetText("No file selected")
		} else {
			dsdaPathLabel.SetText(abbreviatePath(path, 50))
		}
	}

	dsdaBrowseBtn := widget.NewButton("Browse...", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			if reader == nil {
				return
			}
			dsdaDoomPath = reader.URI().Path()
			prefs.SetString(dsdaPathKey, dsdaDoomPath)
			updateDsdaLabel(dsdaDoomPath)
			reader.Close()
		}, myWindow)
	})

	// Create a custom container with max width to keep label within window bounds
	dsdaPathLabelWrapper := container.NewMax(dsdaPathLabel)
	dsdaContainer := container.NewBorder(nil, nil, dsdaBrowseBtn, nil, dsdaPathLabelWrapper)



	// Compatibility Level
	complevelLabel := widget.NewRichTextFromMarkdown("**Compatibility Level:**")
	complevelOptions := []string{
		"0 - Doom 1.2",
		"1 - Doom 1.666",
		"2 - Doom 1.9",
		"3 - Ultimate Doom",
		"4 - Final Doom",
		"5 - DOSDoom",
		"6 - TASDoom",
		"7 - Boom",
		"8 - Boom v2.01",
		"9 - Boom v2.02",
		"10 - LxDoom",
		"11 - MBF",
		"12 - PrBoom v2.03beta",
		"13 - PrBoom v2.1.0",
		"14 - PrBoom v2.1.1-2.2.6",
		"15 - PrBoom v2.3.x",
		"16 - PrBoom v2.4.0",
		"17 - PrBoom v2.5+",
		"21 - MBF21",
	}
	complevelDropdown := widget.NewSelect(complevelOptions, func(value string) {})
	
	// Set the selected value based on stored complevel
	for _, option := range complevelOptions {
		if strings.HasPrefix(option, complevel+" -") {
			complevelDropdown.SetSelected(option)
			break
		}
	}
	if complevelDropdown.Selected == "" {
		complevelDropdown.SetSelected(complevelOptions[len(complevelOptions)-1]) // Default to MBF21
	}
	
	// Update complevel on selection
	complevelDropdown.OnChanged = func(value string) {
		// Extract the complevel number from the beginning of the string
		var levelCode string
		for i := 0; i < len(value); i++ {
			if value[i] == ' ' {
				levelCode = value[:i]
				break
			}
		}
		complevel = levelCode
		prefs.SetString("complevel", levelCode)
	}

	// IWAD File
	iwadLabel := widget.NewRichTextFromMarkdown("**IWAD File:**")
	iwadOptions := []string{"doom.wad", "doom2.wad", "tnt.wad", "plutonia.wad", "freedoom.wad", "freedoom2.wad"}
	iwadDropdown := widget.NewSelect(iwadOptions, func(value string) {})
	iwadDropdown.SetSelected(iwadFile)
	
	// Update IWAD on selection
	iwadDropdown.OnChanged = func(value string) {
		iwadFile = value
		prefs.SetString("iwadFile", value)
	}

	// PWAD Files
	pwadLabel := widget.NewRichTextFromMarkdown("**PWAD Files:**")
	var selectedPwadIdx int = -1
	var pwadList *widget.List
	
	// Helper function to save PWAD list to preferences
	savePwadList := func() {
		pwadJSON, _ := json.Marshal(pwadPaths)
		prefs.SetString("pwadPaths", string(pwadJSON))
	}
	
	// Helper function to refresh the PWAD list display
	var refreshPwadList func()
	
	// Create PWAD list widget
	pwadList = widget.NewList(
		func() int {
			return len(pwadPaths)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			lbl := obj.(*widget.Label)
			if i < len(pwadPaths) {
				lbl.SetText(pwadPaths[i])
			}
		},
	)

	
	refreshPwadList = func() {
		pwadList.Unselect(selectedPwadIdx)
		pwadList.Refresh()
	}
	
	// Remove button (disabled when nothing selected)
	removePwadBtn := widget.NewButton("－", func() {
		if selectedPwadIdx >= 0 && selectedPwadIdx < len(pwadPaths) {
			pwadPaths = append(pwadPaths[:selectedPwadIdx], pwadPaths[selectedPwadIdx+1:]...)
			if selectedPwadIdx >= len(pwadPaths) && selectedPwadIdx > 0 {
				selectedPwadIdx--
			} else if len(pwadPaths) == 0 {
				selectedPwadIdx = -1
			}
			savePwadList()
			refreshPwadList()
		}
	})
	removePwadBtn.Disable()
	
	// Add button
	addPwadBtn := widget.NewButton("＋", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			if reader == nil {
				return
			}
			pwadPath := reader.URI().Path()
			pwadPaths = append(pwadPaths, pwadPath)
			savePwadList()
			refreshPwadList()
			reader.Close()
		}, myWindow)
	})
	
	// Update remove button state when selection changes
	var originalOnSelected func(widget.ListItemID)
	originalOnSelected = func(id widget.ListItemID) {
		selectedPwadIdx = id
		if selectedPwadIdx >= 0 && selectedPwadIdx < len(pwadPaths) {
			removePwadBtn.Enable()
		} else {
			removePwadBtn.Disable()
		}
	}
	pwadList.OnSelected = originalOnSelected
	
	// Handle unselection
	pwadList.OnUnselected = func(id widget.ListItemID) {
		if id == selectedPwadIdx {
			selectedPwadIdx = -1
			removePwadBtn.Disable()
		}
	}
	
	// PWAD button container - vertically stacked on the left
	addPwadBtn.Alignment = widget.ButtonAlignCenter
	removePwadBtn.Alignment = widget.ButtonAlignCenter
	pwadButtonContainer := container.NewVBox(addPwadBtn, removePwadBtn)
	pwadContainer := container.NewBorder(nil, nil, pwadButtonContainer, nil, pwadList)


	// Launch button
	launchBtn := widget.NewButton("Launch Game", func() {
		if dsdaDoomPath == "" {
			dialog.ShowInformation("Missing Path", "Please select DSDA Doom executable path", myWindow)
			return
		}
		
		// Build command arguments
		args := []string{
			"-complevel", complevel,
			"-iwad", iwadFile,
		}
		
		// Add PWAD files if any
		if len(pwadPaths) > 0 {
			args = append(args, "-file")
			args = append(args, pwadPaths...)
		}
		
		// Execute in a goroutine to avoid blocking the UI
		go func() {
			cmd := exec.Command(dsdaDoomPath, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		// On Linux, remove AppImage-specific environment variables that cause library conflicts
		if os.Getenv("APPIMAGE") != "" {
			env := os.Environ()
			cleanEnv := []string{}
			
			// Variables to remove from AppImage that cause nested AppImage conflicts
			blacklist := map[string]bool{
				"LD_LIBRARY_PATH": true,  // Causes library version conflicts
				"LD_PRELOAD": true,       // Preloaded libraries from parent AppImage
				"APPDIR": true,           // Parent AppImage directory
				"APPIMAGE": true,         // Parent AppImage path
				"ARGV0": true,            // Parent AppImage argv
				"APPIMAGE_UUID": true,    // Parent AppImage UUID
				"OWD": true,              // Original working directory
			}
			
			for _, envVar := range env {
				key := strings.Split(envVar, "=")[0]
				if !blacklist[key] {
					cleanEnv = append(cleanEnv, envVar)
				}
			}
			
			cmd.Env = cleanEnv
		}
		
		// Print the executed command line
		fmt.Println("Executing:", cmd.String())
			
			err := cmd.Run()
			if err != nil {
				dialog.ShowError(err, myWindow)
			}
		}()
	})
	launchBtn.Importance = widget.HighImportance

	// Button container
	buttonContainer := container.NewHBox(
		launchBtn,
	)
	
	// Top section with fixed items
	topSection := container.NewVBox(
		dsdaLabel,
		dsdaContainer,
		complevelLabel,
		complevelDropdown,
		iwadLabel,
		iwadDropdown,
		pwadLabel,
	)
	
	// Main form with PWAD list filling remaining space
	form := container.NewBorder(
		topSection,      // top
		buttonContainer, // bottom
		nil,             // left
		nil,             // right
		pwadContainer, // center (expands to fill space)
	)

	// Add scroll if content is large
	scrollContainer := container.NewScroll(form)

	myWindow.SetContent(scrollContainer)
	
	// Save window size on close
	myWindow.SetOnClosed(func() {
		size := myWindow.Canvas().Size()
		prefs.SetFloat("windowWidth", float64(size.Width))
		prefs.SetFloat("windowHeight", float64(size.Height))
	})
	
	myWindow.ShowAndRun()
}
