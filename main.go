package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	dbUser     = "root"
	dbPassword = "Root@123"
	dbName     = "airlineCheckin"
)

func main() {
	// Connecting to MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", dbUser, dbPassword, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening DB: %v\n", err)
	}
	defer db.Close()

	// Fetch all users
	rows, err := db.Query("SELECT id FROM users")
	if err != nil {
		log.Fatalf("Error fetching users: %v\n", err)
	}
	defer rows.Close()

	// Store all users in a slice
	var userIDs []int
	for rows.Next() {
		var userID int
		rows.Scan(&userID)
		userIDs = append(userIDs, userID)
	}

	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Create 120 threads (goroutines) for 120 users
	startTime := time.Now()
	for _, userID := range userIDs {
		wg.Add(1)
		go assignSeat(db, userID, &wg)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	endTime := time.Now()

	fmt.Printf("Query executed in %v\n", endTime.Sub(startTime))

	// Print the seats table to show assigned users
	printSeats(db)
}

// Function to assign seat to a user
func assignSeat(db *sql.DB, userID int, wg *sync.WaitGroup) {
	defer wg.Done()

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Error starting transaction: %v\n", err)
	}
	defer tx.Rollback()

	// Find the first available seat (lock it for the current transaction)
	var seatNo string

	err = tx.QueryRow("SELECT seat_no FROM seats WHERE userID IS NULL LIMIT 1 FOR UPDATE SKIP LOCKED").Scan(&seatNo)
	//FOR UPDATE adds unique locking, this will fix the concurrency issue, but other processes will wait for each lock to release
	// and then they move forward to next one, and same follows until last seat is left for last thread

	//If we use SKIP LOCKED, this will make the query much faster, since in this case, the threads are not waiting for the lock to be
	//released, they'll directly move to find other NULL row
	if err != nil {
		log.Printf("Error finding seat for user %d: %v\n", userID, err)
		return
	}

	// Assign the seat to the user
	_, err = tx.Exec("UPDATE seats SET userID = ? WHERE seat_no = ?", userID, seatNo)
	if err != nil {
		log.Printf("Error assigning seat for user %d: %v\n", userID, err)
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Fatalf("Error committing transaction: %v\n", err)
	}

	// Log the seat assignment
	log.Printf("Assigned seat %s to userID %d\n", seatNo, userID)
}

// Print the seats table to show seat assignments
func printSeats(db *sql.DB) {
	rows, err := db.Query("SELECT seat_no, userID FROM seats ORDER BY seat_no")
	if err != nil {
		log.Fatalf("Error printing seats: %v\n", err)
	}
	defer rows.Close()

	log.Println("Seat Assignments:")
	for rows.Next() {
		var seatNo string
		var userID sql.NullInt64
		rows.Scan(&seatNo, &userID)
		if userID.Valid {
			log.Printf("Seat: %s => UserID: %d\n", seatNo, userID.Int64)
		} else {
			log.Printf("Seat: %s => UserID: None\n", seatNo)
		}
	}
}
