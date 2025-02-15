package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	// "fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func createSearchBar(showLabel bool) (searchBar *fyne.Container) {
	searchButton := createSearchButton(showLabel)
	searchText := createSearchTextBox()

	searchBar = container.NewBorder(nil, nil, searchButton, nil, searchText)
	return searchBar
}

func createSearchButton(showLabel bool) (searchButton *widget.Button) {
	startText := ""
	if showLabel {
		startText = "Search"
	}

	searchButton = widget.NewButtonWithIcon(startText, theme.SearchIcon(), func() {
		log.Println("show search bar when typing into this after clicking or pressing hotkey")
	})

	return searchButton
}

func createSearchTextBox() (searchTextBox *widget.Entry) {
	searchTextBox = widget.NewEntry()
	searchTextBox.SetPlaceHolder("Search Game Names Here!")
	return
}
