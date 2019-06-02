package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	todoAnswer := `{
  "userId": %v,
  "id": %v,
  "title": "delectus aut autem",
  "completed": false
}`

	userAnswer := `{
  "id": %v,
  "name": "Leanne Graham",
  "username": "Bret",
  "email": "Sincere@april.biz",
  "address": {
    "street": "Kulas Light",
    "suite": "Apt. 556",
    "city": "Gwenborough",
    "zipcode": "92998-3874",
    "geo": {
      "lat": "-37.3159",
      "lng": "81.1496"
    }
  },
  "phone": "1-770-736-8031 x56442",
  "website": "hildegard.org",
  "company": {
	"id": %v,
    "name": "Romaguera-Crona",
    "catchPhrase": "Multi-layered client-server neural-net",
    "bs": "harness real-time e-markets"
  }
}`

	companyAnswer := `{
  "id": %v,
  "name": "Romaguera-Crona",
  "username": "Bret",
  "email": "Sincere@april.biz",
  "address": {
    "street": "Kulas Light",
    "suite": "Apt. 556",
    "city": "Gwenborough",
    "zipcode": "92998-3874",
    "geo": {
      "lat": "-37.3159",
      "lng": "81.1496"
    }
  },
  "ceoUserId": "%v",
  "website": "romaguera.com"
}`

	router := gin.Default()

	// This handler will match /user/john but will not match /user/ or /user
	router.GET("/todos/:todo_id", func(c *gin.Context) {

		todoID := c.Param("todo_id")
		userID := rand.Intn(65535)

		sleepTime := 2000 + rand.Intn(2000)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)

		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, todoAnswer, todoID, userID)
	})

	// This handler will match /user/john but will not match /user/ or /user
	router.GET("/users/:user_id", func(c *gin.Context) {

		userID := c.Param("user_id")
		companyID := rand.Intn(65535)

		sleepTime := 1000 + rand.Intn(1000)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)

		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, userAnswer, userID, companyID)
	})

	// This handler will match /user/john but will not match /user/ or /user
	router.GET("/companies/:company_id", func(c *gin.Context) {

		companyID := c.Param("company_id")
		ceoID := rand.Intn(65535)

		sleepTime := 500 + rand.Intn(500)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)

		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, companyAnswer, companyID, ceoID)
	})

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
