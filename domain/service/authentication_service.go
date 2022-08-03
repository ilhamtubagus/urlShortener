package service

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ilhamtubagus/urlShortener/domain/constant"
	"github.com/ilhamtubagus/urlShortener/domain/entity"
	"github.com/ilhamtubagus/urlShortener/interface/dto"
	"github.com/ilhamtubagus/urlShortener/utils"
	"github.com/labstack/echo/v4"
)

type AuthenticationService interface {
	SignIn(user *entity.User) (*entity.Token, error)
	GoogleSignIn(credential string) (*entity.Token, error)
}
type authenticationService struct {
	userService         UserService
	oauth2GoogleService Oauth2GoogleService
	hash                utils.Hash
}

func (a authenticationService) SignIn(user *entity.User) (*entity.Token, error) {
	searchedUser, err := a.userService.FindUserByEmail(user.Email)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, dto.NewDefaultResponse("unexpected database error", http.StatusInternalServerError))
	}
	const EMPTY_STRING = ""
	if searchedUser == nil || searchedUser.Password == EMPTY_STRING {
		return nil, echo.NewHTTPError(http.StatusNotFound, dto.NewDefaultResponse("user was not found", http.StatusNotFound))
	}
	if err := a.hash.CompareHash(user.Password, searchedUser.Password); err != nil {
		fmt.Println(err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, dto.NewDefaultResponse("password does not match", http.StatusBadRequest))
	}
	hour, _ := strconv.Atoi(os.Getenv("TOKEN_EXP"))
	claims := entity.Claims{
		Role:   searchedUser.Role,
		Email:  searchedUser.Email,
		Status: searchedUser.Status,
		StandardClaims: jwt.StandardClaims{
			//token expires within x hours
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(hour)).Unix(),
			Subject:   searchedUser.ID.String(),
		}}
	token, err := claims.GenerateJwt()
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, dto.NewDefaultResponse("unexpected server error", http.StatusInternalServerError))
	}
	return token, nil
}

func (a authenticationService) GoogleSignIn(credential string) (*entity.Token, error) {
	googleClaims, err := a.oauth2GoogleService.VerifyToken(credential)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	user, err := a.userService.FindUserByEmail(googleClaims.Email)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, dto.NewDefaultResponse("unexpected database error", http.StatusInternalServerError))
	}
	if user == nil {
		//insert new user into database
		user = &entity.User{Name: googleClaims.Name, Email: googleClaims.Email, Sub: googleClaims.Sub, Status: constant.ACTIVE, Role: constant.MEMBER}
		err := a.userService.SaveUser(user)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusInternalServerError, dto.NewDefaultResponse("unexpected database error", http.StatusInternalServerError))
		}
	}
	//create our own jwt and send back to client
	hour, _ := strconv.Atoi(os.Getenv("TOKEN_EXP"))
	claims := entity.Claims{
		Role:   user.Role,
		Email:  user.Email,
		Status: user.Status,
		StandardClaims: jwt.StandardClaims{
			//token expires within x hours
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(hour)).Unix(),
			Subject:   user.ID.String(),
		}}
	token, err := claims.GenerateJwt()
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, dto.NewDefaultResponse("unexpected server error", http.StatusInternalServerError))
	}
	return token, nil
}

func NewAuthenticationService(userService UserService, hash utils.Hash, oauth2GoogleService Oauth2GoogleService) AuthenticationService {
	return authenticationService{userService: userService, hash: hash, oauth2GoogleService: oauth2GoogleService}
}