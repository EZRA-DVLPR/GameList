package mkdown

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func WriteToMarkdown() {
	db, err := sql.Open("sqlite3", "games.db")
	if err != nil {
		log.Fatal("Error accessing local dB: ", err)
	}
	defer db.Close()

	// select everything except the url to be grabbed
	rows, err := db.Query("SELECT name, favorite, main, mainPlus, comp FROM games")
	if err != nil {
		log.Fatal("Error retrieving games: ", err)
	}
	defer rows.Close()

	// open the markdown file we are going to be writing to
	mdfile, err := os.Create("GameList.md")
	if err != nil {
		log.Fatal("Error creating markdown file", err)
	}
	defer mdfile.Close()

	_, err = mdfile.WriteString("| No. | **Game Name** | **Main Story** | **Main + Sides** | **Completionist** | Favorite |\n")
	_, err = mdfile.WriteString("| :----: | :---- | ---- | ---- | ---- | ---- |\n")
	if err != nil {
		log.Fatal("Failed to begin writing to markdown file")
	}

	id := 1
	// for each row in games, add a line in the markdown file
	for rows.Next() {
		var name, main, mainPlus, comp string
		var favorite int
		if err := rows.Scan(&name, &favorite, &main, &mainPlus, &comp); err != nil {
			log.Fatal("Error scanning row: ", err)
		}

		//  | No. | name | Main Story | Main + Sides | Completionist |
		_, err = mdfile.WriteString(fmt.Sprintf("| %d. | %s | %s | %s | %s | %d |\n", id, name, main, mainPlus, comp, favorite))
		id += 1
	}
}
