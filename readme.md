# Airline Seat Assignment System

This Go program connects to a MySQL database and simulates an airline check-in system. It assigns available seats to users from the `users` table, ensuring that no two users are assigned the same seat, even in a concurrent environment. The system leverages Go's concurrency (goroutines) and SQL transactions to safely assign seats to 120 users in parallel.

## Features
- **Concurrency**: The program creates 120 threads (goroutines) for 120 users.
- **Database Transactions**: Each thread starts a transaction to safely assign a seat, ensuring no two threads assign the same seat.
- **SQL Locking**: It uses the SQL `FOR UPDATE SKIP LOCKED` clause to prevent concurrency issues, allowing threads to skip over locked rows and find the next available seat without waiting.
- **Logging**: The program logs seat assignments and final seat allocations for transparency.

## Prerequisites
- Go 1.16+
- MySQL installed and running on `localhost:3306`
- MySQL user credentials (`username`, `password`)
- Pre-populated MySQL database (`airlineCheckin`) with the following structure:
    - `users` table with 120 sample users (`id`, `name`)
    - `seats` table with 120 seats (`seat_no`, `userID`), all initially having `NULL` for `userID`

### Database Structure

#### Users Table
| id  | name   |
| --- | ------ |
| 1   | User1  |
| 2   | User2  |
| ... | ...    |
| 120 | User120|

#### Seats Table
| seat_no | userID |
| ------- | ------ |
| 1A      | NULL   |
| 1B      | NULL   |
| ...     | ...    |
| 20F     | NULL   |

## Setup Instructions

### 1. Install Go and MySQL
Ensure that Go and MySQL are installed on your machine.

### 2. Install MySQL Driver for Go
Install the Go MySQL driver:

```bash
go get -u github.com/go-sql-driver/mysql
```

### 3. Set Up MySQL Database
1. Create the database:

    ```sql
    CREATE DATABASE airlineCheckin;
    ```

2. Create the `users` and `seats` tables and populate them:

    ```sql
    USE airlineCheckin;

    CREATE TABLE users (
        id INT PRIMARY KEY AUTO_INCREMENT,
        name VARCHAR(255) NOT NULL
    );

    -- Insert 120 users
    INSERT INTO users (name)
    VALUES ('User1'), ('User2'), ..., ('User120');

    CREATE TABLE seats (
        seat_no VARCHAR(3) PRIMARY KEY,
        userID INT NULL,
        FOREIGN KEY (userID) REFERENCES users(id)
    );

    -- Insert seat numbers (1A to 20F)
    INSERT INTO seats (seat_no, userID)
    VALUES ('1A', NULL), ('1B', NULL), ..., ('20F', NULL);
    ```

### 4. Run the Go Program

1. Clone the repository or create a new Go project:
    ```bash
    go mod init airline-seat-assignment
    ```

2. Add the `go-sql-driver/mysql` dependency:
    ```bash
    go get -u github.com/go-sql-driver/mysql
    ```

3. Run the Go program:
    ```bash
    go run main.go
    ```

### 5. Output
Once you run the program, it will:
- Fetch all 120 users from the database.
- Create 120 goroutines (one for each user).
- Each goroutine will attempt to assign the first available seat to its user using a transaction.
- Finally, the program will log and print all seat assignments.

Example output:
```
Assigned seat 1A to userID 1
Assigned seat 1B to userID 2
...
Query executed in 50ms

Seat Assignments:
Seat: 1A => UserID: 1
Seat: 1B => UserID: 2
...
```

## Key Concepts

### 1. Goroutines
The program creates 120 goroutines, one for each user, to concurrently assign seats.

### 2. SQL Transactions
Each seat assignment is wrapped inside a transaction to ensure data integrity.

### 3. SQL Locking with `FOR UPDATE SKIP LOCKED`
The `FOR UPDATE SKIP LOCKED` clause allows us to lock rows during seat assignment so that no two goroutines can assign the same seat at the same time.

### 4. Logging
The program logs the seat assignment for each user and prints the final seat allocation at the end of execution.


### My Notes

- Approach - 1 : Without any locking, non-deterministic, many will get the same seat, last one will be mapped
- Approach 2: With unique locking in the where clause, all will get seats.
    - First will get locked, all others will wait
    - Once others get the signal the first lock is released since that seat is filled
    - The select statement will send the updated data like 2
    - Now others wait for 2 seat locks to complete and they will get updated data like 3
    - This is how it’ll work, correct solution but a bit slow (430 ms) logged
- Approach - 3: With unique locking along with SKIP LOCKED
    - This time it’ll be much faster since the other processes are not waiting for lock to be released
    - Other processes will get different data simply, and it’ll be faster (87 ms) logged

- Use Cases: Airline Check-in, Flash Sales, Movie Ticket Booking