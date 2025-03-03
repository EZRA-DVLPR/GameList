package ui

import (
	"fmt"
	"image/color"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/EZRA-DVLPR/GameList/internal/dbhandler"
	"github.com/EZRA-DVLPR/GameList/internal/scraper"
)

// window for popup that will be modified for the following functions
var w2 fyne.Window

// confirmation window for updating/deleting all db data
var w3 fyne.Window

func singleGameNameSearchPopup(
	a fyne.App,
	searchSource binding.String,
	sortCategory binding.String,
	sortOrder binding.Bool,
	searchText binding.String,
	dbData *MyDataBinding,
	selectedRow binding.Int,
) {
	// if w2 already exists then focus it and complete task
	if w2 != nil {
		w2.RequestFocus()
		return
	}

	// define w2 properties
	w2 = a.NewWindow("Single Game Name Search")
	w2.Resize(fyne.NewSize(400, 80))

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Game Name to Search")
	w2.SetContent(
		container.New(
			layout.NewVBoxLayout(),
			entry,
			widget.NewButton("Begin Search", func() {
				// if entry is non-empty then perform search
				if strings.TrimSpace(entry.Text) != "" {
					log.Println("Search for game data beginning!")
					ss, _ := searchSource.Get()
					// search game data then add to db
					dbhandler.SearchAddToDB(entry.Text, ss)

					// update dbData
					updateDBData(sortCategory, sortOrder, searchText, dbData)
					forceRenderDB(sortCategory, sortOrder, searchText, dbData, selectedRow)
				} else {
					log.Println("No game name given")
				}
				w2.Close()
			}),
		),
	)
	w2.SetOnClosed(func() {
		w2 = nil
	})
	w2.Show()
}

func manualEntryPopup(
	a fyne.App,
	sortCategory binding.String,
	sortOrder binding.Bool,
	searchText binding.String,
	dbData *MyDataBinding,
	selectedRow binding.Int,
) {
	if w2 != nil {
		w2.RequestFocus()
		return
	}

	w2 = a.NewWindow("Manual Game Data Entry")
	w2.Resize(fyne.NewSize(400, 100))

	gamename := widget.NewEntry()
	main := widget.NewEntry()
	mainplus := widget.NewEntry()
	comp := widget.NewEntry()
	hltbURL := widget.NewEntry()
	completionatorURL := widget.NewEntry()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{
				Text:   "Game Name",
				Widget: gamename,
			},
			{
				Text:   "Main (Hours)",
				Widget: main,
			},
			{
				Text:   "Main Plus Sides (Hours)",
				Widget: mainplus,
			},
			{
				Text:   "Completionist (Hours)",
				Widget: comp,
			},
			{
				Text:   "URL for HowLongToBeat",
				Widget: hltbURL,
			},
			{
				Text:   "URL for Completionator",
				Widget: completionatorURL,
			},
		},
		OnSubmit: func() {
			if strings.TrimSpace(gamename.Text) == "" ||
				strings.TrimSpace(main.Text) == "" ||
				strings.TrimSpace(mainplus.Text) == "" ||
				strings.TrimSpace(comp.Text) == "" {
				log.Println("Not enough game data given. Fill out top 4 fields")
			} else {
				if hltbURL.Text == "" {
					log.Println("No HLTB URL given for manual entry for game", strings.TrimSpace(gamename.Text))
				}
				if completionatorURL.Text == "" {
					log.Println("No Completionator URL given for manual entry for game", strings.TrimSpace(gamename.Text))
				}

				// check if main, mainplus, comp are valid floats
				mainfl, err := strconv.ParseFloat(main.Text, 64)
				if err != nil {
					log.Println("Improper value for Main Story. Make sure its a valid decimal.")
					w2.Close()
					return
				}
				mainplusfl, err := strconv.ParseFloat(mainplus.Text, 64)
				if err != nil {
					log.Println("Improper value for Main + Sides. Make sure its a valid decimal.")
					w2.Close()
					return
				}
				compfl, err := strconv.ParseFloat(comp.Text, 64)
				if err != nil {
					log.Println("Improper value for Completionist. Make sure its a valid decimal.")
					w2.Close()
					return
				}

				// insert the data into the db
				var newgame scraper.Game
				newgame.Name = strings.TrimSpace(gamename.Text)
				newgame.Main = float32(mainfl)
				newgame.MainPlus = float32(mainplusfl)
				newgame.Comp = float32(compfl)
				newgame.HLTBUrl = strings.TrimSpace(hltbURL.Text)
				newgame.CompletionatorUrl = strings.TrimSpace(completionatorURL.Text)
				newgame.Favorite = 0

				dbhandler.AddToDB(newgame)
				forceRenderDB(sortCategory, sortOrder, searchText, dbData, selectedRow)
			}
			w2.Close()
		},
		OnCancel: func() {
			w2.Close()
		},
	}
	w2.SetContent(
		form,
	)
	w2.SetOnClosed(func() {
		w2 = nil
	})
	w2.Show()
}

func settingsPopup(
	a fyne.App,
	w fyne.Window,
	searchSource binding.String,
	sortCategory binding.String,
	sortOrder binding.Bool,
	searchText binding.String,
	selectedRow binding.Int,
	dbData *MyDataBinding,
	textSize binding.Float,
	selectedTheme binding.String,
) {
	// if w2 already exists then focus it and complete task
	if w2 != nil {
		w2.RequestFocus()
		return
	}

	// define w2 properties
	w2 = a.NewWindow("Settings Window")
	w2.Resize(fyne.NewSize(400, 600))

	w2.SetContent(
		container.NewVScroll(
			container.New(
				layout.NewVBoxLayout(),
				searchSourceRadioWidget(searchSource),
				widget.NewSeparator(),
				themeSelector(selectedTheme, textSize, a),
				widget.NewSeparator(),
				textSlider(selectedTheme, textSize, a),
				widget.NewSeparator(),
				updateAllButton(a, sortCategory, sortOrder, searchText, dbData, selectedRow),
				widget.NewSeparator(),
				storageLocationSelector(w),
				widget.NewSeparator(),
				deleteAllButton(a, sortCategory, sortOrder, searchText, dbData, selectedRow),
			),
		),
	)
	w2.SetOnClosed(func() {
		w2 = nil
	})
	w2.Show()
}

// radio for selection of sources
func searchSourceRadioWidget(searchSource binding.String) *fyne.Container {
	label := widget.NewLabel("Search Source Selection")

	radio := widget.NewRadioGroup(
		[]string{
			"All",
			"HLTB",
			"Completionator",
		},
		func(value string) { searchSource.Set(value) },
	)

	// set default to search source saved
	ss, _ := searchSource.Get()
	radio.SetSelected(ss)

	return container.New(
		layout.NewVBoxLayout(),
		label,
		radio,
	)
}

// selector for the theme of the application
// TODO: Binding for themesDir location
func themeSelector(
	selectedTheme binding.String,
	textSize binding.Float,
	a fyne.App,
) *fyne.Container {
	st, _ := selectedTheme.Get()
	label := widget.NewLabel(fmt.Sprintf("Current Theme: %v", st))

	// TODO: Binding for themesDir location
	availableThemes, err := loadAllThemes("themes")
	if err != nil {
		log.Fatal("Error loading themes from themes folder:", err)
	}
	themeList := container.New(
		layout.NewVBoxLayout(),
	)

	// TODO: if themename is too long, want to abbreviate and append '...'
	for themeName, themeColors := range availableThemes {
		button := widget.NewButton(themeName, func(name string, colors ColorTheme) func() {
			return func() {
				label.SetText(fmt.Sprintf("Current Theme: %v", name))
				selectedTheme.Set(name)
				ts, _ := textSize.Get()
				a.Settings().SetTheme(
					&CustomTheme{
						Theme:    theme.DefaultTheme(),
						textSize: float32(ts),
						colors:   availableThemes[name],
					},
				)
			}
		}(themeName, themeColors))
		themeList.Add(button)

		colorPreviews := container.New(
			layout.NewGridLayout(6),
			fixedHeightRect(hexToColor(themeColors.Background)),
			fixedHeightRect(hexToColor(themeColors.Foreground)),
			fixedHeightRect(hexToColor(themeColors.Primary)),
			fixedHeightRect(hexToColor(themeColors.ButtonColor)),
			fixedHeightRect(hexToColor(themeColors.HoverColor)),
			fixedHeightRect(hexToColor(themeColors.InputBackgroundColor)),
		)
		themeList.Add(colorPreviews)
	}

	return container.New(
		layout.NewVBoxLayout(),
		label,
		themeList,
	)
}

func fixedHeightRect(color color.Color) *canvas.Rectangle {
	rect := canvas.NewRectangle(color)
	rect.SetMinSize(fyne.NewSize(0, 40))
	return rect
}

// TODO: Binding for themesDir location
func textSlider(
	selectedTheme binding.String,
	textSize binding.Float,
	a fyne.App,
) *fyne.Container {
	// TODO: Binding for themesDir location
	availableThemes, err := loadAllThemes("themes")
	if err != nil {
		log.Fatal("Error loading themes from themes folder:", err)
	}
	label := widget.NewLabel("Text Size Changer")
	ts, _ := textSize.Get()
	currSize := widget.NewLabel(fmt.Sprintf("Current size is: %v", ts))
	moveSize := widget.NewLabel("")
	moveSize.Hide()

	slider := widget.NewSliderWithData(12, 24, textSize)
	slider.OnChanged = func(res float64) {
		moveSize.Show()
		res32 := float32(res)
		moveSize.SetText(fmt.Sprintf("New Size will be: %v", res32))
	}
	slider.OnChangeEnded = func(res float64) {
		moveSize.Hide()
		res32 := float32(res)
		st, _ := selectedTheme.Get()
		currSize.SetText(fmt.Sprintf("Current size is: %v", res32))
		a.Settings().SetTheme(
			&CustomTheme{
				Theme:    theme.DefaultTheme(),
				textSize: res32,
				colors:   availableThemes[st],
			},
		)
	}
	return container.New(
		layout.NewVBoxLayout(),
		label,
		currSize,
		slider,
		moveSize,
	)
}

func updateAllButton(
	a fyne.App,
	sortCategory binding.String,
	sortOrder binding.Bool,
	searchText binding.String,
	dbData *MyDataBinding,
	selectedRow binding.Int,
) *fyne.Container {
	label := widget.NewLabel("Press this button to update all games within the database")

	updateAll := widget.NewButton("Update All", func() {
		if w3 != nil {
			w3.RequestFocus()
			return
		}
		w3 = a.NewWindow("Confirm Deletion of entire Database")
		w3.Resize(fyne.NewSize(400, 200))
		w3.SetContent(container.New(
			layout.NewGridLayout(2),
			widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
				w3.Close()
				w2.RequestFocus()
			}),
			widget.NewButtonWithIcon("Update all data", theme.ConfirmIcon(), func() {
				dbhandler.UpdateEntireDB()

				forceRenderDB(sortCategory, sortOrder, searchText, dbData, selectedRow)
				w3.Close()
				w2.Close()
			}),
		))
		w3.Show()
		w3.SetOnClosed(func() {
			w3 = nil
		})
		w2.SetOnClosed(func() {
			w2 = nil
		})
	})
	return container.New(
		layout.NewVBoxLayout(),
		label,
		updateAll,
	)
}

func storageLocationSelector(w fyne.Window) *fyne.Container {
	label := widget.NewLabel("Button to change location to store database and export files")

	storageDir := widget.NewButton("select dir", func() {
		w.RequestFocus()
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				log.Println("error:", err)
				w2.Close()
				return
			}
			if uri == nil {
				log.Println("no dir selected")
				w2.Close()
				return
			}
			log.Println("selected dir:", uri)
			w2.Close()
			w2.SetOnClosed(func() {
				w2 = nil
			})
		}, w)
	})

	return container.New(
		layout.NewVBoxLayout(),
		label,
		storageDir,
	)
}

func deleteAllButton(
	a fyne.App,
	sortCategory binding.String,
	sortOrder binding.Bool,
	searchText binding.String,
	dbData *MyDataBinding,
	selectedRow binding.Int,
) *fyne.Container {
	label := widget.NewLabel("Button to delete all games from data")

	deleteAll := widget.NewButton("Delete All", func() {
		if w3 != nil {
			w3.RequestFocus()
			return
		}
		w3 = a.NewWindow("Confirm Deletion of entire Database")
		w3.Resize(fyne.NewSize(400, 200))
		w3.SetContent(container.New(
			layout.NewGridLayout(2),
			widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
				w3.Close()
				w2.RequestFocus()
			}),
			widget.NewButtonWithIcon("Delete all data", theme.ConfirmIcon(), func() {
				dbhandler.DeleteAllDBData()

				forceRenderDB(sortCategory, sortOrder, searchText, dbData, selectedRow)
				w3.Close()
				w2.Close()
			}),
		))
		w3.Show()
		w3.SetOnClosed(func() {
			w3 = nil
		})
		w2.SetOnClosed(func() {
			w2 = nil
		})
	})
	return container.New(
		layout.NewVBoxLayout(),
		label,
		deleteAll,
	)
}
