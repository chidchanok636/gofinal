package task

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Customer struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

func GetCustomerByIdHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := findCustomerByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if customer == nil {
		c.JSON(http.StatusOK, gin.H{})
	} else {
		c.JSON(http.StatusOK, customer)
	}
}

func GetCustomersHandler(c *gin.Context) {
	customers, err := findAllCustomer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customers)
}

func UpdateCustomerByIdHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	editCustomer := Customer{}
	if err := c.ShouldBindJSON(&editCustomer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	editCustomer.ID = id
	rowEffected, err := updateCustomer(editCustomer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rowEffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No row effected"})
		return
	}
	c.JSON(http.StatusOK, editCustomer)
}

func DeleteCustomerHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowEffected, err := deleteCustomer(id)
	if rowEffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No row effected"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
}

func CreateCustomersHandler(c *gin.Context) {
	newCustomer := Customer{}
	if err := c.ShouldBindJSON(&newCustomer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := insertCustomer(&newCustomer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newCustomer)
}

func InitialCustomers() {
	db := getConnection()
	defer db.Close()

	createTable(db)
}

// ===================================  DB func
func createTable(db *sql.DB) {
	createTableSql := `
		CREATE TABLE IF NOT EXISTS customers (
			id SERIAL PRIMARY KEY,
			name TEXT,
			email TEXT,
			status TEXT
		);
		`
	_, err := db.Exec(createTableSql)
	if err != nil {
		log.Println("can't create table", err)
	} else {
		log.Println("create table success")
	}
}

func findCustomerByID(searchId int) (*Customer, error) {
	db := getConnection()
	defer db.Close()

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers WHERE id=$1")
	if err != nil {
		log.Println("can't prepare query one row statement", err)
		return nil, err
	}

	customer := &Customer{}
	row := stmt.QueryRow(searchId)
	err = row.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Status)
	switch err {
	case sql.ErrNoRows:
		log.Println("No rows were returned!")
		return nil, nil
	case nil:
		return customer, nil
	default:
		return nil, err
	}
}

func findAllCustomer() ([]Customer, error) {
	db := getConnection()
	defer db.Close()

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers")
	if err != nil {
		log.Println("can't prepare query all customers statement", err)
		return nil, err
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Println("can't query all customers", err)
		return nil, err
	}

	customers := []Customer{}
	for rows.Next() {
		customer := Customer{}
		err = rows.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Status)
		if err != nil {
			log.Println("can't Scan row into variable", err)
			return nil, err
		}
		customers = append(customers, customer)
	}

	return customers, nil
}

func insertCustomer(newCustomer *Customer) error {
	db := getConnection()
	defer db.Close()

	insertedRow := db.QueryRow("INSERT INTO customers (name, email, status) values ($1,$2,$3) RETURNING id", newCustomer.Name, newCustomer.Email, newCustomer.Status)
	err := insertedRow.Scan(&newCustomer.ID)
	if err != nil {
		log.Println("can't scan id", err)
		return err
	}
	return nil
}

func updateCustomer(customer Customer) (int64, error) {
	db := getConnection()
	defer db.Close()

	result, err := db.Exec("UPDATE customers SET name=$1, email=$2, status=$3 WHERE id=$4", customer.Name, customer.Email, customer.Status, customer.ID)
	if err != nil {
		log.Println("can't prepare statment update", err)
		return 0, err
	}
	rowEffected, err := result.RowsAffected()
	return rowEffected, err
}

func deleteCustomer(id int) (int64, error) {
	db := getConnection()
	defer db.Close()

	result, err := db.Exec("DELETE FROM customers WHERE id =$1", id)
	if err != nil {
		log.Println("can't prepare statment delete", err)
		return 0, err
	}
	rowEffected, err := result.RowsAffected()
	return rowEffected, err
}

func getConnection() *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("Connect to database error", err)
	}
	return db
}
