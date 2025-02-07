package auth

import (
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var SECRET_KEY string = os.Getenv("SECRET_KEY")

func ValidateSession(c *gin.Context) bool {
	cookie, err := c.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired , please login again"})
			return false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while getting cookie"})
		return false
	}
	token, err := ValidateJWT(cookie)
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, signature invaid"})
			return false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while validating token"})
	}

	if !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized, invalid token"})
		return false
	}
	return true
}

func GenerateJWT(userid string) (string, error, time.Time) {
	//expiration time
	expirationTime := time.Now().Add(5 * time.Minute)
	//create jwt claims which includes the username and the exp time
	claims := &Claims{
		Username: userid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(SECRET_KEY))

	return tokenString, err, expirationTime
}

func ValidateJWT(token string) (jwt.Token, error) {
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	return *tkn, err
}

func RefreshToken(c *gin.Context) (bool, error, time.Time) {

	token, err := c.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			return true, nil, time.Time{}
		}
		return true, err, time.Time{}
	}

	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return true, nil, time.Time{}
		}
		return false, err, time.Time{}
	}
	if !tkn.Valid || time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) < 30*time.Second {
		return true, nil, time.Unix(claims.ExpiresAt, 0)
	}
	return false, nil, time.Unix(claims.ExpiresAt, 0)

}
