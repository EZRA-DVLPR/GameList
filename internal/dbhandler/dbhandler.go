package dbhandler

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/EZRA-DVLPR/GameList/internal/scraper"
	_ "github.com/mattn/go-sqlite3"
)

// INFO: STRUCTURE OF THE DB
// games (table) {
// 		name					string		(text)			PRIMARY KEY
// 		hltburl					string		(text)
// 		completionatorurl		string		(text)
//		favorite				int			(integer)
//		main					float		(real)
//		mainPlus				float		(real)
//		comp					float		(real)
//		}

// creates the DB with table
func CreateDB() {
	fmt.Println("Creating the DB")

	db, err := sql.Open("sqlite3", "games.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS games (
		name TEXT PRIMARY KEY,
		hltburl TEXT,
		completionatorurl TEXT,
		favorite INTEGER,
		main REAL,
		mainPlus REAL,
		comp REAL
	);
	`)
	if err != nil {
		log.Fatal("Error creating games table:", err)
	}

	log.Println("Created the local DBs successfully")
}

func CheckDBExists() bool {
	db, err := sql.Open("sqlite3", "games.db")
	if err != nil {
		log.Fatal("Error opening db:", err)
	}
	defer db.Close()

	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name = ?", "games").Scan(&name)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		log.Fatal("Error checking table with no data:", err)
	}
	return true
}

func DeleteAllDBData() {
	db, err := sql.Open("sqlite3", "games.db")
	if err != nil {
		log.Fatal("Error opening db:", err)
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM games")
	if err != nil {
		log.Fatal("Error deleting entire DB")
	}

	log.Println("Deleted all db data")
}

// given a game struct, will search DB for the name of the game to delete it
func DeleteFromDB(gameName string) {
	db, err := sql.Open("sqlite3", "games.db")
	if err != nil {
		log.Fatal("Failed to access db")
	}

	res, err := db.Exec("DELETE FROM games WHERE name = ?", gameName)
	if err != nil {
		log.Fatal("Error deleting game from games table: ", err)
	}

	if rowsAffected(res, gameName) {
		log.Println("Game deleted: ", gameName)
	}
}

// if the given game is not empty and not already existent in DB, then add to the DB
func AddToDB(game scraper.Game) {
	if (game.HLTBUrl == "") &&
		(game.CompletionatorUrl == "") &&
		(game.Main == -1) &&
		(game.MainPlus == -1) &&
		(game.Comp == -1) {
		log.Println("No game data received for associate game.")
		return
	}

	db, err := sql.Open("sqlite3", "games.db")
	if err != nil {
		log.Fatal("Failed to access db")
	}
	defer db.Close()

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM games WHERE name = ?)", game.Name).Scan(&exists)
	if err != nil {
		log.Fatal("Error checking game existence", err)
	}
	if exists {
		log.Println("Game already exists in local DB! Skipping insertion")
		return
	}

	log.Println("Adding the game data to the local DB for game:", game.Name)

	_, err = db.Exec(
		"INSERT OR IGNORE INTO games (name, hltburl, completionatorurl, favorite, main, mainPlus, comp) VALUES (?,?,?,?,?,?,?)",
		game.Name,
		game.HLTBUrl,
		game.CompletionatorUrl,
		game.Favorite,
		game.Main,
		game.MainPlus,
		game.Comp,
	)
	if err != nil {
		log.Fatal("Error inserting game: ", err)
	}

	log.Println("Finished adding the game data to the local DB for game", game.Name)
}

// given the name of a game, search from data sites, then add struct to DB
func SearchAddToDB(gameName string, searchSource string) {
	// get the data from scraper using sources
	var newgame scraper.Game

	switch searchSource {
	case "All":
		log.Println("Searching from all sources for game data")

		// search both and obtain game structs from each source
		hltbSearch := scraper.SearchGameHLTB(gameName)
		completionatorSearch := scraper.SearchGameCompletionator(gameName)

		newgame = compareGetGameData(hltbSearch, completionatorSearch)

		newgame.Name = gameName

	case "HLTB":
		newgame = scraper.SearchGameHLTB(gameName)

	case "Completionator":
		newgame = scraper.SearchGameCompletionator(gameName)

	default:
		log.Println("No such search style")
		return
	}

	// with the data retrieved, add it to DB
	AddToDB(newgame)
}

// if the given game is not empty, then toggle favorite
func ToggleFavorite(gameName string) {
	db, err := sql.Open("sqlite3", "games.db")
	if err != nil {
		log.Fatal("Failed to access db")
	}
	defer db.Close()

	// get value of favorite for given game
	var favorite bool
	err = db.QueryRow("SELECT favorite FROM games WHERE name = ?", gameName).Scan(&favorite)
	if err != nil {
		log.Fatal("Error obtaining favorite value from game", err)
	}

	// update game favorite value to the opposite value
	res, err := db.Exec("UPDATE games SET favorite = ? WHERE name = ?", !favorite, gameName)
	if err != nil {
		log.Fatal("Error updating game to be favorite", err)
	}

	if rowsAffected(res, gameName) {
		log.Println("Toggled Favorite for given game:", gameName)
	}
}

// given a game name, will update its contents with newer information
func UpdateGame(gameName string) {
	db, err := sql.Open("sqlite3", "games.db")
	if err != nil {
		log.Fatal("Failed to access db")
	}
	defer db.Close()

	// get urls for given game
	var hltbURL, completionatorURL string
	err = db.QueryRow("SELECT hltburl, completionatorurl FROM games WHERE name = ?", gameName).Scan(&hltbURL, &completionatorURL)
	if err != nil {
		log.Fatal("Error obtaining URLs for given game")
	}

	// if no URL from source, then try to scrape data from beginning for the link.
	// o/w directly scrape from the saved page
	var hltbSearch, completionatorSearch scraper.Game
	if hltbURL == "" {
		log.Println("No URL found to obtain information from HLTB. Attempting to get link.")
		hltbSearch = scraper.SearchGameHLTB(gameName)
	} else {
		hltbSearch = scraper.FetchHLTB(hltbURL)
	}
	if completionatorURL == "" {
		log.Println("No URL found to obtain information from Completionator. Attempting to get link.")
		completionatorSearch = scraper.SearchGameCompletionator(gameName)
	} else {
		completionatorSearch = scraper.FetchCompletionator(completionatorURL)
	}

	newgamedata := compareGetGameData(hltbSearch, completionatorSearch)

	// overwrite the old data with the new Data
	rows, err := db.Exec(
		"UPDATE games SET hltburl = ?, completionatorurl = ?, main = ?, mainPlus = ?, comp = ? WHERE name = ?",
		newgamedata.HLTBUrl,
		newgamedata.CompletionatorUrl,
		newgamedata.Main,
		newgamedata.MainPlus,
		newgamedata.Comp,
		gameName,
	)
	if err != nil {
		log.Println("Error updating value for game in table", gameName)
	}
	if rowsAffected(rows, gameName) {
		log.Println("Successfully updated values for", gameName)
	}
}

// returns query from db as [][]string given cat, ord, and query
func SortDB(sortCategory string, sortOrder bool, queryName string) (dbOutput [][]string) {
	db, err := sql.Open("sqlite3", "games.db")
	if err != nil {
		log.Fatal("Error accessing local dB: ", err)
	}
	defer db.Close()

	// if sortOrder is true => ASC. false => DESC
	so := ""
	if sortOrder {
		so = "ASC"
	} else {
		so = "DESC"
	}

	// if queryName is empty, sort DB without searching for similar game names
	var rows *sql.Rows
	if queryName == "" {
		rows, err = db.Query(
			fmt.Sprintf(`
				SELECT name, main, mainPlus, comp 
				FROM games 
				ORDER BY favorite DESC, 
				CASE 
					WHEN typeof(%[1]s) = 'integer' OR %[1]s GLOB '[0-9]*' THEN CAST(%[1]s AS INTEGER) 
					ELSE %[1]s 
				END %[2]s;`,
				sortCategory,
				so,
			),
		)
	} else {
		rows, err = db.Query(
			fmt.Sprintf(`
				SELECT name, main, mainPlus, comp 
				FROM games 
				WHERE name LIKE ?
				ORDER BY favorite DESC, 
				CASE 
					WHEN typeof(%[1]s) = 'integer' OR %[1]s GLOB '[0-9]*' THEN CAST(%[1]s AS INTEGER) 
					ELSE %[1]s 
				END %[2]s;`,
				sortCategory,
				so,
			), "%"+queryName+"%")
	}
	if err != nil {
		log.Fatal("Error sorting games from games table: ", err)
	}

	// format data for return
	for rows.Next() {
		var name string
		var main, mainPlus, comp float64
		if err := rows.Scan(&name, &main, &mainPlus, &comp); err != nil {
			log.Fatal("Error scanning row: ", err)
		}
		dbOutput = append(dbOutput, []string{
			name,
			strconv.FormatFloat(main, 'f', -1, 64),
			strconv.FormatFloat(mainPlus, 'f', -1, 64),
			strconv.FormatFloat(comp, 'f', -1, 64),
		})
	}
	return dbOutput
}

func convertRowToInterface(row []string) []interface{} {
	result := make([]interface{}, len(row))
	for i, v := range row {
		result[i] = v
	}
	return result
}

func join(elements []string, sep string) string {
	if len(elements) == 0 {
		return ""
	}
	result := elements[0]
	for _, element := range elements[1:] {
		result += sep + element
	}
	return result
}

// if given rows were affected then returns true. o/w false
func rowsAffected(res sql.Result, name string) (wereAffected bool) {
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Fatal("Error checking affected rows: ", err)
	}
	if rowsAffected == 0 {
		log.Printf("Game `%s` not found in local database\n", name)
		return false
	}
	return true
}

// WARN: if there is another source, consider making a single parameter of[]scraper.Game and
// go by ref to that to do comparisons more effectively
func compareGetGameData(
	firstGame scraper.Game,
	secondGame scraper.Game,
) (resultGame scraper.Game) {
	// if both are empty, then dont update anything (as no new data was found)
	if firstGame.Name == "" &&
		firstGame.Main == 0 &&
		firstGame.MainPlus == 0 &&
		firstGame.Comp == 0 &&
		secondGame.Name == "" &&
		secondGame.Main == 0 &&
		secondGame.MainPlus == 0 &&
		secondGame.Comp == 0 {
		log.Println("No Game Data for game Found!")
		return
	}

	// save each url
	resultGame.HLTBUrl = firstGame.HLTBUrl
	resultGame.CompletionatorUrl = secondGame.CompletionatorUrl

	// compare the values of each game and take the higher of both from each
	if firstGame.Main < secondGame.Main {
		resultGame.Main = secondGame.Main
	} else {
		resultGame.Main = firstGame.Main
	}
	if firstGame.MainPlus < secondGame.MainPlus {
		resultGame.MainPlus = secondGame.MainPlus
	} else {
		resultGame.MainPlus = firstGame.MainPlus
	}
	if firstGame.Comp < secondGame.Comp {
		resultGame.Comp = secondGame.Comp
	} else {
		resultGame.Comp = firstGame.Comp
	}

	return
}
