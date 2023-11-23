package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Todo struct {
	gorm.Model
	Description string `json:"description"`
	Done bool	`json:"done"`
}

func MakeTodo(desc string, done bool) Todo {
	return Todo{Description: desc, Done: done}
}

func SmokeTest(c *fiber.Ctx) error {
	return c.JSON(map[string]string{"message": "API is running"})
}

func AddTodo(db *gorm.DB, todo Todo) Todo {
	db.Create(&todo)
	return todo
}

func DeleteTodo(db *gorm.DB, todoId int) {
	db.Delete(&Todo{}, todoId)
}

func QueryTodos(db *gorm.DB) []Todo{
	var todos []Todo
	db.Find(&todos)
	return todos
}

func ToggleTodo(db *gorm.DB, todoId int) Todo {
	var todo Todo
	db.First(&todo, todoId)
	todo.Done = !todo.Done
	db.Save(&todo)
	return todo
}

func main(){
    // Initialize standard Go html template engine
    engine := html.New("./views", ".html")
    // If you want other engine, just replace with following
    // Create a new engine with django
    // engine := django.New("./views", ".django")

    app := fiber.New(fiber.Config{
        Views: engine,
    })
	app.Use(logger.New())
	app.Static("/", "./public")
	db, err := gorm.Open(sqlite.Open("config.db"), &gorm.Config{})
	if err != nil {
	  panic("failed to connect database")
	}
	db.AutoMigrate(&Todo{})

	
	
	app.Get("/", SmokeTest)
	app.Get("/todo-app", func(c *fiber.Ctx) error {
		// todos := QueryTodos(db)
		return c.Render("index", fiber.Map{
			"Name": "Joe",
		})
	})

	app.Get("/todo-table", func(c *fiber.Ctx) error {
		todos := QueryTodos(db)
		return c.Render("partials/todo-table", fiber.Map{"Todos": todos})
	})

	app.Get("/todo", func(c *fiber.Ctx) error {
		return c.JSON(QueryTodos(db))
	})
	app.Post("/todo", func(c *fiber.Ctx) error {
		todo := new(Todo)
		if err := c.BodyParser(todo); err != nil {
			return err
		}
		if todo.Description == "" {
			return fmt.Errorf("Description cannot be empty")
		}
		 result := AddTodo(db, *todo)
		 return c.Render("partials/todo-row", fiber.Map{
			"ID": result.ID,
			"Done": result.Done,
			"Description": result.Description,
		 })
	})

	app.Put("/todo/:todoId/toggle", func(c *fiber.Ctx) error {
		todoId, err := c.ParamsInt("todoId")
		if err != nil {
			return c.JSON(map[string]string{"message": "todo ID must be an int!"})
		} 
		todo := ToggleTodo(db, todoId)
		return c.Render("partials/todo-row", fiber.Map{
			"ID": todo.ID,
			"Done": todo.Done,
			"Description": todo.Description,
		})
	})

	app.Delete("/todo/:todoId", func(c *fiber.Ctx) error {
		todoId, err := c.ParamsInt("todoId")
		if err != nil {
			return c.JSON(map[string]string{"message": "todo ID must be an int!"})
		}
		DeleteTodo(db, todoId)
		return c.SendString("")
	})
	log.Fatal(app.Listen(":1313"))
}