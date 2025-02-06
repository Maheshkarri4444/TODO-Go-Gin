package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Maheshkarri4444/todo-app/auth"
	"github.com/Maheshkarri4444/todo-app/database"
	"github.com/Maheshkarri4444/todo-app/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var todoCollection *mongo.Collection = database.OpenCollection(database.Client, "todos")

func GetTodo(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	id := c.Param("id")
	objId, _ := primitive.ObjectIDFromHex(id)

	var todo models.Todo
	err := todoCollection.FindOne(ctx, bson.M{"_id": objId}).Decode(&todo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, todo)
}

func ClearAll(c *gin.Context) {
	session := auth.ValidateSession(c)
	if !session {
		return
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	userid := c.Param("userid")
	_, err := todoCollection.DeleteMany(ctx, bson.M{"userid": userid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "all todos are deleted"})
}

func GetTodos(c *gin.Context) {
	session := auth.ValidateSession(c)
	if !session {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	userid := c.Param("userid")
	findResult, err := todoCollection.Find(ctx, bson.M{"userid": userid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"FindError": err.Error()})
		return
	}

	var todos []models.Todo
	for findResult.Next(ctx) {
		var todo models.Todo
		err := findResult.Decode(&todo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Decode Error": err.Error()})
			return
		}
		todos = append(todos, todo)
	}

	c.JSON(http.StatusOK, todos)
}

func DeleteTodo(c *gin.Context) {
	session := auth.ValidateSession(c)
	if !session {
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	id := c.Param("id")
	userid := c.Param("userid")
	objId, _ := primitive.ObjectIDFromHex(id)
	deleteResult, err := todoCollection.DeleteOne(ctx, bson.M{"_id": objId, "userid": userid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if deleteResult.DeletedCount == 0 {
		msg := fmt.Sprintf("No todo with id: %v was found", id)
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	msg := fmt.Sprintf("todo with id: %v was deleted successfully ", id)
	c.JSON(http.StatusOK, gin.H{"success": msg})

}

func UpdateTodo(c *gin.Context) {
	session := auth.ValidateSession(c)
	if !session {
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var newTodo models.Todo
	if err := c.BindJSON(&newTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := todoCollection.UpdateOne(ctx, bson.M{"_id": newTodo.ID, "userid": newTodo.UserID}, bson.M{"$set": newTodo})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		fmt.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, newTodo)
}

func AddTodo(c *gin.Context) {
	session := auth.ValidateSession(c)
	if !session {
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var todo *models.Todo
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	todo.ID = primitive.NewObjectID()
	todo.UserID = c.Param("userid")

	_, err := todoCollection.InsertOne(ctx, todo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"insertedId": todo.ID})
}
