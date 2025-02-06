package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Maheshkarri4444/todo-app/auth"
	"github.com/Maheshkarri4444/todo-app/database"
	"github.com/Maheshkarri4444/todo-app/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var SECRET_KEY string = os.Getenv("SECRET_KEY")
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func SignUp(c *gin.Context) {
	var user models.User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	emailCount, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	defer cancel()

	if err != nil {
		log.Panic(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while cheking for the email"})
	}

	password := HashPassword(*user.Password)
	user.Password = &password

	if emailCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user with this email already exists"})
		return
	}

	user.ID = primitive.NewObjectID()
	_, insertErr := userCollection.InsertOne(ctx, user)
	if insertErr != nil {
		msg := fmt.Sprintf("user item was not %s", "created")
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	defer cancel()

	userId := user.ID.Hex()
	username := *user.Name

	token, err, expirationTime := auth.GenerateJWT(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while generating token"})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  expirationTime,
		Path:     "/",
		Domain:   "", // Leave empty to default to current domain
		HttpOnly: true,
		Secure:   true, // Must be true when using SameSite=None
		SameSite: http.SameSiteNoneMode,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "userID",
		Value:    userId,
		Expires:  expirationTime,
		Path:     "/",
		Domain:   "", // Leave empty to default to current domain
		HttpOnly: true,
		Secure:   true, // Must be true when using SameSite=None
		SameSite: http.SameSiteNoneMode,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "username",
		Value:    username,
		Expires:  expirationTime,
		Path:     "/",
		Domain:   "", // Leave empty to default to current domain
		HttpOnly: true,
		Secure:   true, // Must be true when using SameSite=None
		SameSite: http.SameSiteNoneMode,
	})

	// c.JSON(http.StatusOK, resultInsertionNumber)
	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
		"user_id": userId,
	})

}

func Login(c *gin.Context) {

	var user models.User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var foundUser models.User
	err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "email is not valid"})
		return
	}

	passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)

	if !passwordIsValid {
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	if foundUser.Email == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	userId := foundUser.ID.Hex()
	username := *foundUser.Name

	shouldRefresh, err, expirationTime := auth.RefreshToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh token error"})
		return
	}

	if shouldRefresh {
		token, err, expirationTime := auth.GenerateJWT(userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while generating token"})
			return
		}
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "token",
			Value:    token,
			Expires:  expirationTime,
			Path:     "/",
			Domain:   "", // Leave empty to default to current domain
			HttpOnly: true,
			Secure:   true, // Must be true when using SameSite=None
			SameSite: http.SameSiteNoneMode,
		})

		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "userID",
			Value:    userId,
			Expires:  expirationTime,
			Path:     "/",
			Domain:   "", // Leave empty to default to current domain
			HttpOnly: true,
			Secure:   true, // Must be true when using SameSite=None
			SameSite: http.SameSiteNoneMode,
		})

		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "username",
			Value:    username,
			Expires:  expirationTime,
			Path:     "/",
			Domain:   "", // Leave empty to default to current domain
			HttpOnly: true,
			Secure:   true, // Must be true when using SameSite=None
			SameSite: http.SameSiteNoneMode,
		})
	} else {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "userID",
			Value:    userId,
			Expires:  expirationTime,
			Path:     "/",
			Domain:   "", // Leave empty to default to current domain
			HttpOnly: true,
			Secure:   true, // Must be true when using SameSite=None
			SameSite: http.SameSiteNoneMode,
		})

		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "username",
			Value:    username,
			Expires:  expirationTime,
			Path:     "/",
			Domain:   "", // Leave empty to default to current domain
			HttpOnly: true,
			Secure:   true, // Must be true when using SameSite=None
			SameSite: http.SameSiteNoneMode,
		})
	}
	c.JSON(http.StatusOK, gin.H{"msg": "login successfull"})

}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""
	if err != nil {
		msg = fmt.Sprintf("password is %s", "incorrect")
		check = false
	}
	return check, msg

}
