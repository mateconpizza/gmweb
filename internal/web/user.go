//go:build ignore

package web

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/mateconpizza/gmweb/internal/forms"
)

func (h *Handler) userSignup(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData()
	data.Form = forms.UserSignUp{}
	data.Colorschemes = h.colorschemes
	data.CurrentColorscheme = getThemeFromCookie(r)
	data.PageTitle = "New User"

	p := parseRequestParams(r)
	p.CurrentDB = r.URL.Query().Get("db")
	data.Params = p
	data.Routes = &webRoutes{}

	h.renderPage(w, r, http.StatusOK, "signup", data)
}

func (h *Handler) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var f forms.UserSignUp
	err := forms.DecodePostForm(r, &f)
	if err != nil {
		h.logger.Error("signup", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	f.CheckField(forms.NotBlank(f.Name), "name", "Field 'name' cannot be blank")
	f.CheckField(forms.NotBlank(f.Password), "password", "Field 'password' cannot be blank")
	f.CheckField(forms.MinChars(f.Password, 8), "password", "Password must be at least 8 characters long")

	if !f.Valid() {
		data := newTemplateData()
		data.Form = f
		data.FormHasErrors = true
		data.Colorschemes = h.colorschemes
		data.CurrentColorscheme = getThemeFromCookie(r)
		data.PageTitle = "New User"

		data.Params = parseRequestParams(r)
		data.Params.CurrentDB = "main"
		h.renderPage(w, r, http.StatusUnprocessableEntity, "signup", data)
		return
	}

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (h *Handler) userLogin(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData()
	data.Form = forms.UserLogin{}
	data.Colorschemes = h.colorschemes
	data.CurrentColorscheme = getThemeFromCookie(r)
	data.PageTitle = "Login User"
	p := parseRequestParams(r)
	data.Params = p
	data.Params.CurrentDB = "main"

	fmt.Printf("data: %v\n", data)

	h.renderPage(w, r, http.StatusOK, "login", data)
}

func (h *Handler) userLoginPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Authenticate and login the user...")
}

func (h *Handler) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Logout the user...")
}
