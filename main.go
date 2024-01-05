package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/abdullah/go-fiber-postgres/models"
	"github.com/abdullah/go-fiber-postgres/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Book struct {
	Author		string		`json:"author"`
	Title		string		`json:"title"`
	Publisher	string		`json:"publisher"`
}

type Repository struct {
	DB *gorm.DB  // `DB` is a pointer to `*gorm.DB`
}

// Define a method `SetupRoutes` for the type or struct `Repository`
func (r *Repository) SetupRoutes(app *fiber.App){
	api := app.Group("/api")
	api.Post("/create_books", r.CreateBooks)
	api.Delete("delete_book/:id", r.DeleteBook)
	api.Get("/get_books/:id", r.GetBookByID)
	api.Get("/books", r.GetBooks)
}


// function CreateBooks
func (r *Repository) CreateBooks(context *fiber.Ctx) error{
	book := Book{}

	//Using `context.BodyParser` to convert the `json` into book format
	err := context.BodyParser(&book)
	if err != nil{
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	// *************************************
	err = r.DB.Create(&book).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not create book"})
		return err
	}

	// If everything goes well without `errors` we then send a status `200`
	context.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "book has been created"})
	return nil
}


// function DeleteBook
func (r *Repository) DeleteBook(context *fiber.Ctx) error {
	bookModel := models.Books{}
	id := context.Params("id")

	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}
	

	err := r.DB.Delete(bookModel, id)
	if err.Error != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not delete book",
		})
		return err.Error
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book deleted successfully",
	})
	return nil
}

// function GetBooks
func (r *Repository) GetBooks(context *fiber.Ctx) error {
	// `bookModels` is a pointer to an empty slice of `models.Book`.
	bookModels := &[]models.Books{}

	err := r.DB.Find(bookModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not find book"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book has been found",
		"data": bookModels,
	})
	return nil
}

// function GetBookByID
func (r *Repository) GetBookByID(context *fiber.Ctx) error {
	bookModel := &models.Books{}
	id := context.Params("id")

	if id == ""{
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}

	fmt.Println("the id is", id)

	err := r.DB.Where("id = ?", id).First(bookModel).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "cannot get id",	
		})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book id has been found",
		"data": bookModel,	
	})
	return nil

}

// main function
func main(){
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host: 		os.Getenv("DB_HOST"),
		Port: 		os.Getenv("DB_PORT"),
		Password: 	os.Getenv("DB_PASSWORD"),
		User: 		os.Getenv("DB_USER"),
		DBName: 	os.Getenv("DB_NAME"),
		SSLMode: 	os.Getenv("DB_SSLMODE"),
	}

	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("could not load the database")
	}

	r := Repository{
		DB: db,
	}

	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("could not migrate db")
	}

	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":8080")
}