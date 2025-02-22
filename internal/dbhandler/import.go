package dbhandler

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// selector for importing
func Import(choice int) {
	switch choice {
	case 1:
		importCSV()
	case 2:
		importSQL()
	case 3:
		importTXT()
	default:
		log.Fatal("No such import exists!")
	}
}

func importCSV() {
	db, err := sql.Open("sqlite3", "games.db")
	if err != nil {
		log.Fatal("Error opening db:", err)
	}
	defer db.Close()

	file, err := os.Open("export.csv")
	if err != nil {
		log.Fatal("error opening CSV:", err)
	}
	defer file.Close()

	// read from csv and check formatting
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error reading CSV:", err)
	}
	if len(rows) < 1 {
		log.Fatal("CSV file is empty or improperly formatted")
	}

	// create the table
	var name string
	err = db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
		"games",
	).Scan(&name)
	if err != nil {
		createStmt := fmt.Sprintf(
			"CREATE TABLE games (name TEXT PRIMARY KEY, hltburl TEXT, completionatorurl TEXT, favorite INTEGER, main REAL, mainPlus REAL, comp REAL)",
		)
		_, err := db.Exec(createStmt)
		if err != nil {
			log.Fatal("Error creating table:", err)
		}
		fmt.Println("Table created")
	} else {
		log.Fatal("Error with query for table creation")
	}

	// setup transaction with dummy values
	cols := rows[0]
	temp := make([]string, len(cols))
	for i := range temp {
		temp[i] = "?"
	}
	insertStmt := fmt.Sprintf("INSERT INTO games (%s) VALUES (%s);",
		join(cols, ", "), join(temp, ", "))

	// start transaction and insert data
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Error starting transaction:", err)
	}
	for _, row := range rows[1:] {
		_, err := tx.Exec(insertStmt, convertRowToInterface(row)...)
		if err != nil {
			tx.Rollback()
			log.Fatal("Error inserting data:", err)
		}
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatal("Error committing transaction:", err)
	}

	fmt.Println("Import completed successfully")
}

func importSQL() {
	db, err := sql.Open("sqlite3", "games.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlDump, err := os.ReadFile("export.sql")
	if err != nil {
		log.Fatal(err)
	}

	// perform the import (dump)
	_, err = db.Exec(string(sqlDump))
	if err != nil {
		log.Fatal("Error importing sql database:", err)
	}

	fmt.Println("SQL database imported successfully")
}

func importTXT() {
	db, err := sql.Open("sqlite3", "games.db")
	if err != nil {
		log.Fatal("Error opening db:", err)
	}
	defer db.Close()

	file, err := os.Open("gamenames.txt")
	if err != nil {
		log.Fatal("error opening txt file:", err)
	}
	defer file.Close()

	// scan file and print line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
