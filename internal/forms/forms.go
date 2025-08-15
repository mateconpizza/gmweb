// Package forms provides a way to decode form data from HTTP requests into Go
// structs.
package forms

import (
	"errors"
	"net/http"
	"time"

	form "github.com/go-playground/form/v4"
)

type BookmarkCreate struct {
	URL         string
	Tags        []string
	Title       string
	Description string
	FaviconURL  string `db:"favicon_url"`
	CreatedAt   time.Time
	Validator   `form:"-"`
}

type UserSignUp struct {
	Name      string `form:"name"`
	Email     string `form:"email"`
	Password  string `form:"password"`
	Validator `form:"-"`
}

type UserLogin struct {
	Name      string `form:"name"`
	Password  string `form:"password"`
	Validator `form:"-"`
}

func DecodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	// Call Decode() on our decoder instance, passing the target destination as
	// the first parameter.
	decoder := form.NewDecoder()
	err = decoder.Decode(dst, r.PostForm)
	if err != nil {
		// If we try to use an invalid target destination, the Decode() method
		// will return an error with the type *form.InvalidDecoderError.We use
		// errors.As() to check for this and raise a panic rather than returning
		// the error.
		var invalidDecoderError *form.InvalidDecoderError
		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		// For all other errors, we return them as normal.
		return err
	}

	return nil
}
