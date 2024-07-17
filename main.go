package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type PersonInfo struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	City        string `json:"city"`
	State       string `json:"state"`
	Street1     string `json:"street1"`
	Street2     string `json:"street2"`
	ZipCode     string `json:"zip_code"`
}

func initDB() {
	var err error
	db, err = sql.Open("mysql", "username:password@tcp(localhost:3306)/database_name")
	if err != nil {
		log.Fatal(err)
	}
}

func getPersonInfo(c *gin.Context) {
	personID, err := strconv.Atoi(c.Param("person_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid person ID"})
		return
	}

	var personInfo PersonInfo

	// Query to fetch person's info, phone number, and address using JOINs
	query := `
        SELECT p.name, ph.number AS phone_number, a.city, a.state, a.street1, a.street2, a.zip_code
        FROM person p
        JOIN phone ph ON p.id = ph.person_id
        JOIN address_join aj ON p.id = aj.person_id
        JOIN address a ON aj.address_id = a.id
        WHERE p.id = ?
    `
	err = db.QueryRow(query, personID).Scan(
		&personInfo.Name,
		&personInfo.PhoneNumber,
		&personInfo.City,
		&personInfo.State,
		&personInfo.Street1,
		&personInfo.Street2,
		&personInfo.ZipCode,
	)
	if err != nil {
		log.Println("Error fetching person info:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch person info"})
		return
	}

	c.JSON(http.StatusOK, personInfo)
}

func main() {
	initDB()
	defer db.Close()

	r := gin.Default()

	r.GET("/person/:person_id/info", getPersonInfo)
	r.POST("/person/create", createPerson)

	port := "8080" // or any other port you prefer
	log.Printf("Server running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

type Person struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	City        string `json:"city"`
	State       string `json:"state"`
	Street1     string `json:"street1"`
	Street2     string `json:"street2"`
	ZipCode     string `json:"zip_code"`
}

func createPerson(c *gin.Context) {
	var person Person
	if err := c.ShouldBindJSON(&person); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Insert into database
	stmt, err := db.Prepare(`
        INSERT INTO person(name) VALUES (?)
    `)
	if err != nil {
		log.Println("Error preparing statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create person"})
		return
	}
	defer stmt.Close()

	// Execute the SQL statement
	_, err = stmt.Exec(person.Name)
	if err != nil {
		log.Println("Error executing statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create person"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Person created successfully"})
}
