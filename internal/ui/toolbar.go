package ui

import (
	_ "embed"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/EZRA-DVLPR/GameList/internal/dbhandler"
)

// used to grab the custom SVG for the heart
//
//go:embed assets/heart.svg
var heartSVG []byte

// creates the toolbar with the options that will be displayed to manage the rendered DB
func createMainWindowToolbar(
	sortCategory binding.String,
	sortOrder binding.Bool,
	searchText binding.String,
	selectedRow binding.Int,
	dbData *MyDataBinding,
	searchSource binding.String,
	textSize binding.Float,
	selectedTheme binding.String,
	availableThemes map[string]ColorTheme,
	a fyne.App,
	w fyne.Window,
) (toolbar *fyne.Container) {
	return container.New(
		layout.NewHBoxLayout(),
		createSortButton(sortOrder),
		layout.NewSpacer(),
		createAddButton(a, sortCategory, sortOrder, searchText, dbData, selectedRow, searchSource, selectedTheme, w),
		layout.NewSpacer(),
		createUpdateButton(sortCategory, sortOrder, searchText, selectedRow, dbData),
		layout.NewSpacer(),
		createRemoveButton(selectedRow, sortCategory, sortOrder, searchText, dbData),
		layout.NewSpacer(),
		createRandomButton(selectedRow, dbData),
		layout.NewSpacer(),
		createFaveButton(selectedRow, sortCategory, sortOrder, searchText, dbData),
		layout.NewSpacer(),
		createExportButton(selectedTheme, w),
		layout.NewSpacer(),
		createHelpButton(selectedTheme, w),
		layout.NewSpacer(),
		createSettingsButton(a, searchSource, sortCategory, sortOrder, searchText, selectedRow, dbData, textSize, selectedTheme, availableThemes),
	)

	// PERF: change size of each button depending on the size of the given window
	// 1. make toolbar use gridwraplayout
	// container.New(
	// 	layout.NewGridWrapLayout(fyne.NewSize(200, 50)),
	// 	createSortButton(sortOrder),
	// 	createAddButton(a, sortCategory, sortOrder, searchText, dbData, selectedRow, searchSource, toolbarCanvas),
	// 	createUpdateButton(sortCategory, sortOrder, searchText, selectedRow, dbData),
	// 	createRemoveButton(selectedRow, sortCategory, sortOrder, searchText, dbData),
	// 	createRandomButton(selectedRow, dbData),
	// 	createFaveButton(selectedRow, sortCategory, sortOrder, searchText, dbData),
	// 	createExportButton(toolbarCanvas),
	// 	createHelpButton(toolbarCanvas),
	// 	createSettingsButton(a, w, searchSource, sortCategory, sortOrder, searchText, selectedRow, dbData, textSize, selectedTheme),
	// )
	// 2. remove text next to buttons
}

// toggles sort Order (ASC->DESC->ASC)
func createSortButton(sortOrder binding.Bool) (sortButton *widget.Button) {
	// create the button with empty label
	sortButton = widget.NewButtonWithIcon("", theme.MenuDropUpIcon(), func() {
		// whatever curr value of sortOrder is, we want opposite when clicked
		val, _ := sortOrder.Get()
		log.Println("Sort Order changed to:", !val)
		sortOrder.Set(!val)
	})

	// listen for changes, and update text+icon
	sortOrder.AddListener(binding.NewDataListener(func() {
		val, _ := sortOrder.Get()
		if val {
			sortButton.SetText("Sort ASC")
			sortButton.SetIcon(theme.MenuDropUpIcon())
		} else {
			sortButton.SetText("Sort DESC")
			sortButton.SetIcon(theme.MenuDropDownIcon())
		}
	}))
	return sortButton
}

// export data from db
func createExportButton(
	selectedTheme binding.String,
	w fyne.Window,
) (exportButton *widget.Button) {
	menuItems := []*fyne.MenuItem{
		fyne.NewMenuItem("Export to CSV", func() {
			// idea is to have user pick folder, then entry for filename
			dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
				if err != nil {
					log.Println("Error writing to CSV file:", err)
					return
				}
				if uri == nil {
					log.Println("No file Selected to export to CSV")
					return
				}
				defer uri.Close() // close uri when dialog closes

				// delete the partially created file as i will create it in export function
				os.Remove(uri.URI().Path())

				//export
				dbhandler.Export(1, uri.URI().Path())
			}, w)
		}),
		fyne.NewMenuItem("Export to SQL", func() {
			dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
				if err != nil {
					log.Println("Error writing to SQL file:", err)
					return
				}
				if uri == nil {
					log.Println("No file Selected to export to SQL")
					return
				}
				defer uri.Close() // close uri when dialog closes
				os.Remove(uri.URI().Path())
				dbhandler.Export(2, uri.URI().Path())
			}, w)
		}),
		fyne.NewMenuItem("Export to Markdown", func() {
			dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
				if err != nil {
					log.Println("Error writing to MD file:", err)
					return
				}
				if uri == nil {
					log.Println("No file Selected to export to MD")
					return
				}
				defer uri.Close() // close uri when dialog closes
				os.Remove(uri.URI().Path())
				dbhandler.Export(3, uri.URI().Path())
			}, w)
		}),
	}

	// define the popup
	menuPopup := NewThemeAwareMenu(menuItems, w.Canvas())

	exportButton = widget.NewButtonWithIcon("", theme.MailSendIcon(), func() {
		menuPopup.Show(exportButton.Position().Add(fyne.NewPos(0, exportButton.Size().Height)))
	})

	// refresh menupopup when the theme changes
	selectedTheme.AddListener(binding.NewDataListener(func() {
		menuPopup.Refresh()
	}))

	return exportButton
}

// PERF:
//
//	find way to implement menu without opening new window for window tiling managers
func createSettingsButton(
	a fyne.App,
	searchSource binding.String,
	sortCategory binding.String,
	sortOrder binding.Bool,
	searchText binding.String,
	selectedRow binding.Int,
	dbData *MyDataBinding,
	textSize binding.Float,
	selectedTheme binding.String,
	availableThemes map[string]ColorTheme,
) (settingsButton *widget.Button) {
	settingsButton = widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		settingsPopup(
			a,
			searchSource,
			sortCategory,
			sortOrder,
			searchText,
			selectedRow,
			dbData,
			textSize,
			selectedTheme,
			availableThemes,
		)
	})

	return settingsButton
}

// get data and add it to the DB
func createAddButton(
	a fyne.App,
	sortCategory binding.String,
	sortOrder binding.Bool,
	searchText binding.String,
	dbData *MyDataBinding,
	selectedRow binding.Int,
	searchSource binding.String,
	selectedTheme binding.String,
	w fyne.Window,
) (addButton *widget.Button) {
	ss, _ := searchSource.Get()
	menuItems := []*fyne.MenuItem{
		fyne.NewMenuItem("Single Game Search", func() {
			singleGameNameSearchPopup(
				a,
				searchSource,
				sortCategory,
				sortOrder,
				searchText,
				dbData,
				selectedRow,
			)
		}),
		fyne.NewMenuItem("Manual Entry", func() {
			manualEntryPopup(
				a,
				sortCategory,
				sortOrder,
				searchText,
				dbData,
				selectedRow,
			)
		}),
		fyne.NewMenuItem("From CSV", func() {
			fileDialog := dialog.NewFileOpen(func(uri fyne.URIReadCloser, err error) {
				if err != nil {
					log.Println("Error opening CSV file:", err)
					return
				}
				if uri == nil {
					log.Println("No file Selected for importing from CSV")
					return
				}
				defer uri.Close() // close uri when dialog closes
				dbhandler.Import(1, ss, uri.URI().Path())
				forceRenderDB(sortCategory, sortOrder, searchText, dbData, selectedRow)
			}, w)
			// set file extension to only allow csv files
			fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
			fileDialog.Show()
		}),
		// INFO: will drop the existing table and replace with imported SQL file
		fyne.NewMenuItem("From SQL", func() {
			fileDialog := dialog.NewFileOpen(func(uri fyne.URIReadCloser, err error) {
				if err != nil {
					log.Println("Error opening SQL file:", err)
					return
				}
				if uri == nil {
					log.Println("No file Selected for importing from SQL")
					return
				}
				defer uri.Close()
				dbhandler.Import(2, ss, uri.URI().Path())
				forceRenderDB(sortCategory, sortOrder, searchText, dbData, selectedRow)
			}, w)
			fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".sql"}))
			fileDialog.Show()
		}),
		// INFO: game names must be separated by new lines with 1 game per line
		fyne.NewMenuItem("From TXT", func() {
			fileDialog := dialog.NewFileOpen(func(uri fyne.URIReadCloser, err error) {
				if err != nil {
					log.Println("Error opening TXT file:", err)
					return
				}
				if uri == nil {
					log.Println("No file Selected for importing from TXT")
					return
				}
				defer uri.Close()
				dbhandler.Import(3, ss, uri.URI().Path())
				forceRenderDB(sortCategory, sortOrder, searchText, dbData, selectedRow)
			}, w)
			fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt"}))
			fileDialog.Show()
		}),
		// INFO: These require user input to be completed
		fyne.NewMenuItem("From GOG", func() {
			integrationImport(searchSource, "gog", w)
		}),
		fyne.NewMenuItem("From psn", func() {
			integrationImport(searchSource, "psn", w)
		}),
		fyne.NewMenuItem("From steam", func() {
			integrationImport(searchSource, "steam", w)
		}),
		fyne.NewMenuItem("From Epic", func() {
			integrationImport(searchSource, "epic", w)
		}),
	}

	// define the popup
	menuPopup := NewThemeAwareMenu(menuItems, w.Canvas())

	addButton = widget.NewButtonWithIcon("Add Game", theme.ContentAddIcon(), func() {
		menuPopup.Show(addButton.Position().Add(fyne.NewPos(0, addButton.Size().Height)))
	})

	// refresh menupopup when the theme changes
	selectedTheme.AddListener(binding.NewDataListener(func() {
		menuPopup.Refresh()
	}))

	return addButton
}

// finds selected row game name, and deletes it from DB
func createRemoveButton(
	selectedRow binding.Int,
	sortCategory binding.String,
	sortOrder binding.Bool,
	searchText binding.String,
	dbData *MyDataBinding,
) (removeButton *widget.Button) {
	removeButton = widget.NewButtonWithIcon("Remove Game", theme.ContentRemoveIcon(), func() {
		selrow, _ := selectedRow.Get()
		if selrow >= 0 {
			// get the game name and send query for deletion
			dbdata, _ := dbData.Get()
			log.Println("Removing Game:", dbdata[selrow][0])
			dbhandler.DeleteFromDB(dbdata[selrow][0])

			forceRenderDB(sortCategory, sortOrder, searchText, dbData, selectedRow)
		}
	})

	return removeButton
}

// lists help options such as tutorial, manual, support, etc.
func createHelpButton(
	selectedTheme binding.String,
	w fyne.Window,
) (helpButton *widget.Button) {
	menuItems := []*fyne.MenuItem{
		fyne.NewMenuItem("Video Tutorial Series", func() {
			println("Youtube video series on EZRA-DVLPR channel where I explain how to use the program")
			goToWebsite("https://example.com")
		}),
		fyne.NewMenuItem("Online Manual", func() {
			println("Link to my website which has a manual describing what each thing does in a pdf format")
			goToWebsite("https://example.com")
		}),
		fyne.NewMenuItem("Bug/Feature Tracker", func() {
			println("Link to github repo with pre-made tags for feature requests and bugs and stuff")
			goToWebsite("https://example.com")
		}),
		fyne.NewMenuItem("Blog Post", func() {
			println("Link to my website with a blogpost")
			goToWebsite("https://example.com")
		}),
		fyne.NewMenuItem("Support Me <3", func() {
			println("Is a link such that, when clicked will take you to Ko-Fi, Paypal, etc. on my website")
			goToWebsite("https://example.com")
		}),
	}

	menuPopup := NewThemeAwareMenu(menuItems, w.Canvas())

	helpButton = widget.NewButtonWithIcon("", theme.QuestionIcon(), func() {
		menuPopup.Show(helpButton.Position().Add(fyne.NewPos(0, helpButton.Size().Height)))
	})

	selectedTheme.AddListener(binding.NewDataListener(func() {
		menuPopup.Refresh()
	}))

	return helpButton
}

// randomly selects a row to highlight
func createRandomButton(
	selectedRow binding.Int,
	dbData *MyDataBinding,
) (removeButton *widget.Button) {
	removeButton = widget.NewButtonWithIcon("Random Row", theme.SearchReplaceIcon(), func() {
		dbdata, _ := dbData.Get()
		selectedRow.Set(rand.Intn(len(dbdata)))
	})

	return removeButton
}

// toggle favorite for game defined by selectedRow
func createFaveButton(
	selectedRow binding.Int,
	sortCategory binding.String,
	sortOrder binding.Bool,
	searchText binding.String,
	dbData *MyDataBinding,
) (faveButton *widget.Button) {
	heartIcon := fyne.NewStaticResource("heart.svg", heartSVG)
	faveButton = widget.NewButtonWithIcon("(Un)Favorite", theme.NewThemedResource(heartIcon), func() {
		selrow, _ := selectedRow.Get()
		if selrow >= 0 {
			// get the game name and send query for toggling favorite
			dbdata, _ := dbData.Get()
			dbhandler.ToggleFavorite(dbdata[selrow][0])

			forceRenderDB(sortCategory, sortOrder, searchText, dbData, selectedRow)
		}
	})

	return faveButton
}

// update the selected game defined by selectedRow
func createUpdateButton(
	sortCategory binding.String,
	sortOrder binding.Bool,
	searchText binding.String,
	selectedRow binding.Int,
	dbData *MyDataBinding,
) (updateButton *widget.Button) {
	updateButton = widget.NewButtonWithIcon("Update", theme.MediaReplayIcon(), func() {
		selrow, _ := selectedRow.Get()
		if selrow >= 0 {
			log.Println("Updating highlighted entry")

			dbdata, _ := dbData.Get()
			dbhandler.UpdateGame(dbdata[selrow][0])

			forceRenderDB(sortCategory, sortOrder, searchText, dbData, selectedRow)
		}
	})

	return updateButton
}

func forceRenderDB(
	sortCategory binding.String,
	sortOrder binding.Bool,
	searchText binding.String,
	dbData *MyDataBinding,
	selectedRow binding.Int,
) {
	// update dbData and selectedRow to render changes
	updateDBData(sortCategory, sortOrder, searchText, dbData)
	selectedRow.Set(-1)
}

func goToWebsite(link string) {
	var cmd *exec.Cmd

	// change cmd based on which OS is being used
	switch runtime.GOOS {
	case "darwin": // mac = darwin
		cmd = exec.Command("open", link)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", link)
	default: // linux, bsd, etc.
		cmd = exec.Command("xdg-open", link)
	}

	err := cmd.Start()
	if err != nil {
		// Handle error
		log.Println("Error opening link:", link)
		log.Fatal(err)
	}
}
