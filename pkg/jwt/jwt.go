package tokens

import (
	"log"
	"myproject/internal/models"
	"myproject/pkg/validations"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func GenerateJWT(user *models.User, duration int) (string, error) {
	return generateTokenByClaims(jwt.MapClaims{
		"id":             user.ID.Hex(),
		"personal_info":  user.PersonalInfo,
		"contact_info":   user.ContactInfo,
		"companies_info": user.CompaniesInfo,
		"created_at":     user.CreatedAt,
		"updated_at":     user.UpdatedAt,
		"deleted_at":     user.DeletedAt,
		"last_session":   user.LastSession,
		"iat":            time.Now().Unix(),
		"exp":            time.Now().Add(time.Hour * time.Duration(duration)).Unix(),
	})
}

func GenerateJWTEmail(email string, duration int) (string, error) {
	return generateTokenByClaims(jwt.MapClaims{
		"email": email,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour * time.Duration(duration)).Unix(),
	})
}

func generateTokenByClaims(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Firmar el token
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		log.Fatal(err)
	}
	return tokenString, nil
}

//-----------------------------------------\\

func GetClaims(tokenString string) (*jwt.MapClaims, error) {
	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, validations.ErrInvalidToken
	}
	return claims, nil
}

func GetTokenInHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return "", nil
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	return token, nil
}

func GetFieldOfHeaderToken(r *http.Request, field string) (string, error) {
	token, err := GetTokenInHeader(r)
	if err != nil {
		return "", err
	}

	value, err := GetFieldInToken(token, field)
	if err != nil {
		return "", err
	}
	return value, nil
}

func GetFieldInToken(token string, field string) (string, error) {
	claims, err := GetClaims(token)
	if err != nil {
		return "", err
	}

	value, ok := (*claims)[field].(string)
	if !ok {
		return "", validations.ErrInvalidToken
	}
	return value, nil
}

/*
func GetBdNameInToken(r *http.Request) (string, error) {
	id, err := GetFieldOfHeaderToken(r, "id")
	if err != nil {
		return "", err
	}
	return id + "_DB", nil
}
*/
