package auth

import (
	"github.com/gin-gonic/gin/binding"
	"testing"
)

func TestValidation(t *testing.T) {
	// set up user and validator for testing
	var (
		user User
		err  error
	)
	bValidator := binding.Validator

	// missing required field case
	user = User{
		Email:    "missingpassword@gmail.com",
		Password: "",
	}

	err = bValidator.ValidateStruct(user)
	if err == nil {
		t.Fatal("expected err nil")
	}

	// correct validation
	user = User{
		Email:    "notmissing@gmail.com",
		Password: "passwordishereja",
	}

	err = bValidator.ValidateStruct(user)
	if err != nil {
		t.Fatal("expected err non-nil")
	}
}
