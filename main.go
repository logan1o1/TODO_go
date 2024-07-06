package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	Id        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Completed bool               `json:"completed"`
	Body      string             `json:"body"`
}

var collection *mongo.Collection

func main() {
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load(".env")

		if err != nil {
			log.Fatal("Error loading .env file", err)
		}
	}

	MONGO_URI := os.Getenv("MONGO_URI")
	clientOptions := options.Client().ApplyURI(MONGO_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to mongoDB")

	collection = client.Database("todo").Collection("todos")

	app := fiber.New()

	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins: "http://localhost:5173/",
	// 	AllowHeaders: "Origin, content-Type,Accept",
	// }))

	app.Post("/api/addTodos", createTodos)
	app.Get("/api/getTodods", getTodos)
	app.Put("/api/completeTodos/:id", completeTodos)
	app.Delete("/api/deleteTodos/:id", deleteTodos)

	if os.Getenv("ENV") == "production" {
		app.Static("/", "./todoClient/dist")
	}

	log.Fatal(app.Listen("127.0.0.1:4000"))
}

func createTodos(c *fiber.Ctx) error {
	todo := new(Todo)

	if err := c.BodyParser(todo); err != nil {
		return err
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Body doesn't exist"})
	}

	insertTodo, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		return err
	}

	if insertTodo == nil {
		return c.Status(400).JSON(fiber.Map{"error": "err in insertTodo"})
	}

	// todo.ID = insertTodo.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(todo)
}

func getTodos(c *fiber.Ctx) error {
	var todos []Todo

	cursor, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		return c.Status(404).JSON(err)
	}

	if cursor == nil {
		return c.Status(300).JSON(fiber.Map{"Message": "No todos set"})
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo Todo
		err := cursor.Decode(&todo)
		if err != nil {
			return c.Status(400).JSON(err)
		}
		todos = append(todos, todo)
	}

	return c.Status(200).JSON(todos)
}

func completeTodos(c *fiber.Ctx) error {

	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(err)
	}

	filter := bson.M{"_id": objectId}
	complete := bson.M{"$set": bson.M{"completed": true}}

	_, err = collection.UpdateOne(context.Background(), filter, complete)

	if err != nil {
		return c.Status(400).JSON(err)
	}

	return c.Status(200).JSON(fiber.Map{"message": "Todo succesfully completed"})
}

func deleteTodos(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(err)
	}

	filter := bson.M{"_id": objectId}

	_, err = collection.DeleteOne(context.Background(), filter)

	if err != nil {
		return c.Status(400).JSON(err)
	}

	return c.Status(200).JSON(fiber.Map{"message": "Todo was succesfully deleted"})
}
