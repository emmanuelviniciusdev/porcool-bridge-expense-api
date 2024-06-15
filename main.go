package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	_ "github.com/emmanuelviniciusdev/porcool-bridge-expense-api/docs"
	"github.com/gin-gonic/gin"
	swaggoFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           porcool-bridge-expense-api
// @description     HTTP Service created to make possible the creation of expenses in Porcool by other applications.

// @contact.name   Emmanuel Bergmann
// @contact.url    https://github.com/emmanuelviniciusdev
// @contact.email  emmanuelviniciusdev@gmail.com

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

// @BasePath /api

func main() {
	// TODO: REMOVE THESE "SET ENV" BEFORE GIT COMMIT.
	// TODO: REMOVE THESE "SET ENV" BEFORE GIT COMMIT.
	// TODO: REMOVE THESE "SET ENV" BEFORE GIT COMMIT.
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "./firebase-service-account.json")

	// ctxFirestore := context.Background()

	// clientFirestore, err := initFirestore(ctxFirestore)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer clientFirestore.Close()

	// randomDoc, err := clientFirestore.Collection("users").Doc("sm9p7ItxV3OFeXNUIzm0Zseithm2").Get(ctxFirestore)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(randomDoc.Data())

	r := gin.Default()

	rGroupAPI := r.Group("api")
	{
		rGroupAPI.GET("ping", getPingRoute)
		rGroupAPI.POST("expense", postExpenseRoute)
	}

	// Accessible by "swagger/index.html"
	r.GET("swagger/*any", ginSwagger.WrapHandler(swaggoFiles.Handler))

	r.Run()
}

func initFirestore(ctx context.Context) (*firestore.Client, error) {
	app, err := firebase.NewApp(ctx, nil)

	if err != nil {
		return nil, err
	}

	client, err := app.Firestore(ctx)

	if err != nil {
		return nil, err
	}

	return client, nil
}

type RequestDefaultOutput struct {
	Message string `json:"message"`
}

type RequestPostExpenseInput struct {
	UserEmail           string  `json:"userEmail"`
	ExpenseName         string  `json:"expenseName"`
	ExpenseStatus       string  `json:"expenseStatus"`
	ExpenseAmount       float32 `json:"expenseAmount"`
	ExpenseSpendingDate string  `json:"expenseSpendingDate"`
}

func (input *RequestPostExpenseInput) ExpenseSpendingDateForExpenseCreationStr() (*time.Time, error) {
	expenseSpendingDate, err := time.Parse("2006-01", input.ExpenseSpendingDate)

	if err != nil {
		return nil, err
	}

	parsedExpenseSpendingDate := time.Date(expenseSpendingDate.Year(), expenseSpendingDate.Month(), expenseSpendingDate.Day(), 0, 0, 0, 0, time.Local)

	return &parsedExpenseSpendingDate, nil
}

// @Tags Expense
// @Accept json
// @Produce json
// @Param RequestBody body RequestPostExpenseInput true "RequestBody"
// @Success 204
// @Router /api/expense [post]
func postExpenseRoute(ctx *gin.Context) {
	// TODO: Specify swagger annotations for failure HTTP statuses.
	// TODO: Specify swagger annotations for query params (incrementBalance, balanceName).

	ctxFirestore := context.Background()

	clientFirestore, err := initFirestore(ctxFirestore)

	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable, RequestDefaultOutput{Message: "Unable to communicate with Porcool"})
		return
	}

	defer clientFirestore.Close()

	var requestBodyInput RequestPostExpenseInput

	if err := ctx.BindJSON(&requestBodyInput); err != nil {
		ctx.JSON(http.StatusBadRequest, RequestDefaultOutput{Message: "Invalid request body"})
		return
	}

	incrementBalance, err := strconv.ParseBool(ctx.Query("incrementBalance"))

	if err != nil {
		incrementBalance = false
	}

	balanceName := ctx.DefaultQuery("balanceName", "AUTOMATIC_CREATION_BY_EXTERNAL_SERVICE")

	err = clientFirestore.RunTransaction(ctxFirestore, func(ctx context.Context, transactionFirestore *firestore.Transaction) error {
		// TODO: Retrive user ID by user e-mail (request body "userEmail").
		userID := ""

		err, errMessage := createNewExpense(clientFirestore, transactionFirestore, requestBodyInput, userID)

		if err != nil {
			return errors.New(errMessage)
		}

		if incrementBalance {
			err, errMessage = createOrIncrementExistingBalance(clientFirestore, transactionFirestore, balanceName, requestBodyInput.ExpenseAmount, userID)

			if err != nil {
				return errors.New(errMessage)
			}
		}

		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, RequestDefaultOutput{Message: err.Error()})
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func createOrIncrementExistingBalance(clientFirestore *firestore.Client, transactionFirestore *firestore.Transaction, balanceName string, amount float32, userID string) (error, string) {
	// TODO: Implement...
	// TODO: Implement...
	// TODO: Implement...

	return nil, ""
}

func createNewExpense(clientFirestore *firestore.Client, transactionFirestore *firestore.Transaction, input RequestPostExpenseInput, userID string) (error, string) {
	expenseSpendingDate, err := input.ExpenseSpendingDateForExpenseCreationStr()

	if err != nil {
		return err, "The request body value expenseSpendingDate is invalid. Expected format: YYYY-MM (example: 2024-12)"
	}

	expenseNewDocRef := clientFirestore.Collection("expenses").NewDoc()

	err = transactionFirestore.Create(expenseNewDocRef, map[string]interface{}{
		"expenseName":           input.ExpenseName,
		"amount":                input.ExpenseAmount,
		"status":                input.ExpenseStatus,
		"spendingDate":          expenseSpendingDate,
		"user":                  userID,
		"type":                  "expense",
		"alreadyPaidAmount":     0,
		"validity":              nil,
		"indeterminateValidity": false,
		"created":               time.Now(),
		"source":                "porcool-bridge-expense-api",
	})

	if err != nil {
		return err, "Something went wrong during the expense creation"
	}

	return nil, ""
}

// @Tags Test
// @Produce json
// @Success 200 {object} RequestDefaultOutput
// @Router /api/ping [get]
func getPingRoute(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, RequestDefaultOutput{Message: "pong"})
}
