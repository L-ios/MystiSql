//go:build e2e

package e2e

import (
	"fmt"
	"math/rand"
	"time"
)

type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	FullName     string
	Age          int
	Balance      float64
	IsActive     bool
}

type Product struct {
	ID          int64
	Name        string
	Description string
	Price       float64
	Stock       int
	Category    string
	IsAvailable bool
}

type Order struct {
	ID          int64
	UserID      int64
	OrderNumber string
	TotalAmount float64
	Status      string
	Notes       string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateTestUser() *User {
	id := rand.Int63n(1000000)
	return &User{
		Username:     fmt.Sprintf("testuser_%d", id),
		Email:        fmt.Sprintf("testuser_%d@example.com", id),
		PasswordHash: "$2a$10$testhashedpassword",
		FullName:     fmt.Sprintf("Test User %d", id),
		Age:          20 + rand.Intn(40),
		Balance:      float64(rand.Intn(10000)) + float64(rand.Intn(100))/100,
		IsActive:     true,
	}
}

func GenerateTestProduct() *Product {
	id := rand.Int63n(1000000)
	categories := []string{"Electronics", "Accessories", "Audio", "Furniture", "Storage"}
	return &Product{
		Name:        fmt.Sprintf("Test Product %d", id),
		Description: fmt.Sprintf("Description for test product %d", id),
		Price:       float64(rand.Intn(500)) + float64(rand.Intn(100))/100,
		Stock:       rand.Intn(200),
		Category:    categories[rand.Intn(len(categories))],
		IsAvailable: true,
	}
}

func GenerateTestOrder(userID int64) *Order {
	id := rand.Int63n(1000000)
	statuses := []string{"pending", "confirmed", "shipped", "delivered", "cancelled"}
	return &Order{
		UserID:      userID,
		OrderNumber: fmt.Sprintf("ORD-TEST-%d", id),
		TotalAmount: float64(rand.Intn(1000)) + float64(rand.Intn(100))/100,
		Status:      statuses[rand.Intn(len(statuses))],
		Notes:       "Test order",
	}
}

func GenerateTestUsers(count int) []*User {
	users := make([]*User, count)
	for i := 0; i < count; i++ {
		users[i] = GenerateTestUser()
	}
	return users
}

func GenerateTestProducts(count int) []*Product {
	products := make([]*Product, count)
	for i := 0; i < count; i++ {
		products[i] = GenerateTestProduct()
	}
	return products
}

func GenerateTestOrders(userIDs []int64, count int) []*Order {
	orders := make([]*Order, count)
	for i := 0; i < count; i++ {
		userID := userIDs[rand.Intn(len(userIDs))]
		orders[i] = GenerateTestOrder(userID)
	}
	return orders
}

var (
	DefaultTestUsers = []*User{
		{Username: "alice", Email: "alice@example.com", PasswordHash: "hash1", FullName: "Alice Johnson", Age: 28, Balance: 1500.50, IsActive: true},
		{Username: "bob", Email: "bob@example.com", PasswordHash: "hash2", FullName: "Bob Smith", Age: 35, Balance: 2350.75, IsActive: true},
		{Username: "charlie", Email: "charlie@example.com", PasswordHash: "hash3", FullName: "Charlie Brown", Age: 42, Balance: 890.00, IsActive: true},
	}

	DefaultTestProducts = []*Product{
		{Name: "Laptop Pro 15", Description: "High-performance laptop", Price: 1299.99, Stock: 50, Category: "Electronics", IsAvailable: true},
		{Name: "Wireless Mouse", Description: "Ergonomic mouse", Price: 49.99, Stock: 200, Category: "Accessories", IsAvailable: true},
		{Name: "USB-C Hub", Description: "Multi-port adapter", Price: 79.99, Stock: 150, Category: "Accessories", IsAvailable: true},
	}
)
