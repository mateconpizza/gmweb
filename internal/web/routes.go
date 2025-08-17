package web

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/justinas/nosurf"
	"github.com/mateconpizza/gm/pkg/bookmark"

	"github.com/mateconpizza/gmweb/internal/files"
	"github.com/mateconpizza/gmweb/internal/forms"
	"github.com/mateconpizza/gmweb/internal/helpers"
	"github.com/mateconpizza/gmweb/internal/middleware"
	"github.com/mateconpizza/gmweb/internal/qr"
	"github.com/mateconpizza/gmweb/internal/responder"
	"github.com/mateconpizza/gmweb/internal/router"
	"github.com/mateconpizza/gmweb/ui"
)

func (h *Handler) Routes(mux *http.ServeMux) {
	requireIDAndDB := func(fn http.HandlerFunc) http.Handler {
		return middleware.RequireIDAndDBParam(fn)
	}
	requireDB := func(fn http.HandlerFunc) http.Handler {
		return middleware.RequireDBParam(fn)
	}

	r := h.router
	mux.HandleFunc("GET /not/implemented", h.notImplementedYet)
	mux.HandleFunc("GET "+r.Web.Index("{$}"), h.indexRedirect)
	mux.Handle("GET "+r.Web.All(), requireDB(h.index))
	mux.Handle("GET "+r.Web.New(), requireDB(h.newRecord))
	mux.Handle("GET "+r.Web.NewFrame(), requireDB(h.newRecordFrame))
	mux.Handle("GET "+r.Web.Detail("{id}"), requireIDAndDB(h.recordDetail))
	mux.Handle("GET "+r.Web.Edit("{id}"), requireIDAndDB(h.editRecord))
	mux.Handle("GET "+r.Web.QRCode("{id}"), requireIDAndDB(h.showQR))

	// user related
	mux.HandleFunc("GET "+r.Web.UserSignup(), h.userSignup)
	mux.HandleFunc("POST "+r.Web.UserSignup(), h.userSignupPost)
	mux.HandleFunc("GET "+r.Web.UserLogin(), h.userLogin)
	mux.HandleFunc("POST "+r.Web.UserLogin(), h.userLoginPost)
	mux.HandleFunc("POST "+r.Web.UserLogout(), h.userLogoutPost)

	// theme related
	mux.HandleFunc("/web/theme/change", h.changeTheme)

	// static and cache files
	staticFS, err := fs.Sub(h.files, "static")
	if err != nil {
		panic(err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	iconsPath := filepath.Join(h.cacheDir, "favicon")
	mux.Handle(ui.Cache, http.StripPrefix(ui.Cache, http.FileServer(http.Dir(iconsPath))))
}

func (h *Handler) indexRedirect(w http.ResponseWriter, r *http.Request) {
	h.router.SetRepo("main")
	http.Redirect(w, r, h.router.Web.All(), http.StatusSeeOther)
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	p := parseRequestParams(r)
	p.CurrentDB = r.PathValue("db")

	repo, err := h.repoLoader(p.CurrentDB)
	if err != nil {
		responder.ServerErr(w, r, err)
		return
	}

	records, err := repo.All(r.Context())
	if err != nil {
		responder.ServerErr(w, r, err)
		return
	}

	records = helpers.SortBy("newest", records)

	// Copy bookmarks
	currentBookmarks := make([]*bookmark.Bookmark, len(records))
	copy(currentBookmarks, records)

	// Filter + sort + paginate
	filtered := helpers.ApplyFiltersAndSorting(p.Tag, p.Query, p.Letter, p.FilterBy, currentBookmarks)
	pagination := calculatePagination(len(filtered), p.Page, h.itemsPerPage)
	paginated := filtered[pagination.StartIndex:pagination.EndIndex]

	faviconPath := filepath.Join(h.cacheDir, "favicon")
	if err := files.MkdirAll(faviconPath); err != nil {
		responder.ServerErr(w, r, err)
	}

	go func() {
		if err := loadFavicons(faviconPath, paginated); err != nil {
			h.logger.Error("background favicon loading", "error", err)
		}
	}()

	// Context
	ctx := &TemplateContext{
		App:        h.appCfg,
		Request:    r,
		Bookmarks:  paginated,
		Params:     p,
		Routes:     h.router,
		TagsFn:     helpers.GetTagsFn(p.Tag, p.Query, records, paginated),
		Pagination: pagination,
	}

	data := buildIndexTemplateData(ctx)
	data.Colorschemes = h.colorschemes
	data.CurrentColorscheme = getThemeFromCookie(r)
	data.CurrentTheme = getThemeModeFromCookie(r)

	h.renderPage(w, r, http.StatusOK, "pages/index", data)
}

func (h *Handler) showQR(w http.ResponseWriter, r *http.Request) {
	dbName := r.PathValue("db")
	repo, err := h.repoLoader(dbName)
	if err != nil {
		responder.ServerErr(w, r, err)
		return
	}

	bID, _ := strconv.Atoi(r.PathValue("id"))
	b, err := repo.ByID(r.Context(), bID)
	if err != nil {
		responder.ServerErr(w, r, err)
		return
	}
	h.logger.Info("show qr", "id", b.ID, "url", b.URL)

	q := qr.New(b.URL, h.qrImgSize)
	if err := q.Generate(); err != nil {
		responder.ServerErr(w, r, err)
		return
	}

	f, err := h.files.ReadFile("static/json/colorschemes.json")
	if err != nil {
		responder.ServerErr(w, r, err)
	}

	currentTheme := files.StripSuffixes(filepath.Base(getThemeFromCookie(r)))
	t, err := getCurrentTheme(f, currentTheme)
	if err != nil {
		responder.ServerErr(w, r, err)
	}

	bg := t.Light.Bg
	fg := t.Light.Fg
	mode := getThemeModeFromCookie(r)
	if mode == "dark" {
		bg = t.Dark.Bg
		fg = t.Dark.Fg
	}

	h.logger.Info("show QR", "current mode", mode)

	img, err := q.ImagePNG(bg, fg)
	if err != nil {
		responder.ServerErr(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(img.Bytes()); err != nil {
		responder.ServerErr(w, r, err)
		return
	}
}

func (h *Handler) editRecord(w http.ResponseWriter, r *http.Request) {
	p := parseRequestParams(r)
	p.CurrentDB = r.PathValue("db")

	repo, err := h.repoLoader(p.CurrentDB)
	if err != nil {
		responder.ServerErr(w, r, err)
		return
	}

	idStr := r.PathValue("id")
	bID, _ := strconv.Atoi(idStr)

	b, err := repo.ByID(r.Context(), bID)
	if err != nil {
		responder.ServerErr(w, r, err)
		return
	}

	data := &TemplateData{Params: p, Bookmark: b}
	data.Colorschemes = h.colorschemes
	data.CurrentColorscheme = getThemeFromCookie(r)
	data.PageTitle = "Edit Bookmark"
	data.CurrentURI = r.RequestURI
	data.CSRFToken = nosurf.Token(r)

	if err := h.template.ExecuteTemplate(w, "bookmark-edit", data); err != nil {
		responder.ServerErr(w, r, err)
	}
}

func (h *Handler) recordDetail(w http.ResponseWriter, r *http.Request) {
	p := parseRequestParams(r)
	p.CurrentDB = r.PathValue("db")
	repo, err := h.repoLoader(p.CurrentDB)
	if err != nil {
		responder.ServerErr(w, r, err)
	}

	idStr := r.PathValue("id")
	bID, _ := strconv.Atoi(idStr)
	b, err := repo.ByID(r.Context(), bID)
	if err != nil {
		responder.ServerErr(w, r, err)
	}

	faviconPath := filepath.Join(h.cacheDir, "favicon")
	ext := filepath.Ext(b.FaviconURL)
	hashDomain, _ := helpers.HashDomain(b.URL)
	if files.Exists(hashDomain + ext) {
		b.FaviconLocal = hashDomain + ext
	}

	if b.FaviconLocal == "" {
		go func() {
			if err := loadFavicons(faviconPath, []*bookmark.Bookmark{b}); err != nil {
				h.logger.Error("background favicon loading", "error", err)
			}
		}()
	}

	data := &TemplateData{Params: p, Bookmark: b}
	data.Colorschemes = h.colorschemes
	data.CurrentColorscheme = getThemeFromCookie(r)
	data.PageTitle = "Bookmark Detail"
	data.CurrentURI = r.RequestURI
	data.Routes = router.New(p.CurrentDB).Web
	data.CSRFToken = nosurf.Token(r)

	h.renderPage(w, r, http.StatusOK, "bookmark-detail", data)
}

func (h *Handler) newRecord(w http.ResponseWriter, r *http.Request) {
	// FIX:
	// maybe, redirect to `view` new record.
	p := parseRequestParams(r)
	p.CurrentDB = r.PathValue("db")

	data := &TemplateData{Params: p}
	data.Colorschemes = h.colorschemes
	data.CurrentColorscheme = getThemeFromCookie(r)
	data.PageTitle = "New Bookmark"
	data.CSRFToken = nosurf.Token(r)
	data.URL = buildURLs(p, r)

	if err := h.template.ExecuteTemplate(w, "bookmark-new", data); err != nil {
		responder.ServerErr(w, r, err)
	}
}

func (h *Handler) changeTheme(w http.ResponseWriter, r *http.Request) {
	themeName := r.FormValue("theme")
	isValid := slices.Contains(h.colorschemes, filepath.Base(themeName))

	if !isValid {
		http.Error(w, "Invalid theme: "+strings.Join(h.colorschemes, ", "), http.StatusBadRequest)
		return
	}

	setThemeCookie(w, themeName)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) newRecordFrame(w http.ResponseWriter, r *http.Request) {
	p := parseRequestParams(r)
	p.CurrentDB = r.PathValue("db")
	u := r.URL.Query().Get("url")

	repo, err := h.repoLoader(p.CurrentDB)
	if err != nil {
		responder.ServerErr(w, r, err)
		return
	}

	data := &TemplateData{Params: p}
	data.Colorschemes = h.colorschemes
	data.CurrentColorscheme = getThemeFromCookie(r)
	data.PageTitle = "New Bookmark"
	data.CSRFToken = nosurf.Token(r)
	data.URL = buildURLs(p, r)

	templateName := "bookmark-new-frame"
	if b, ok := repo.Has(r.Context(), u); ok {
		data.Bookmark = b
		data.PageTitle = "Edit Bookmark"
		templateName = "bookmark-edit-frame"
	}

	if err := h.template.ExecuteTemplate(w, templateName, data); err != nil {
		responder.ServerErr(w, r, err)
	}
}

func (h *Handler) userSignup(w http.ResponseWriter, r *http.Request) {
	p := parseRequestParams(r)
	p.CurrentDB = r.URL.Query().Get("db")

	data := newTemplateData(p.CurrentDB)
	data.Form = forms.UserSignUp{}
	data.Colorschemes = h.colorschemes
	data.CurrentColorscheme = getThemeFromCookie(r)
	data.PageTitle = "New User"
	data.Params = p
	data.URL = buildURLs(p, r)
	data.CSRFToken = nosurf.Token(r)

	h.renderPage(w, r, http.StatusOK, "signup", data)
}

func (h *Handler) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var f forms.UserSignUp
	err := forms.DecodePostForm(r, &f)
	if err != nil {
		h.logger.Error("signup", "error", err)
		responder.ServerCustomErr(w, r, err, http.StatusBadRequest)
		return
	}

	f.CheckField(forms.NotBlank(f.Name), "name", "Field 'name' cannot be blank")
	f.CheckField(forms.NotBlank(f.Password), "password", "Field 'password' cannot be blank")
	f.CheckField(forms.MinChars(f.Password, 8), "password", "Password must be at least 8 characters long")

	if !f.Valid() {
		dbName := r.URL.Query().Get("db")
		if dbName == "" {
			dbName = "main"
		}

		data := newTemplateData(dbName)
		data.Form = f
		data.FormHasErrors = true
		data.Colorschemes = h.colorschemes
		data.CurrentColorscheme = getThemeFromCookie(r)
		data.PageTitle = "New User"

		data.Params = parseRequestParams(r)
		data.URL = buildURLs(data.Params, r)
		data.Params.CurrentDB = dbName
		data.CSRFToken = nosurf.Token(r)

		h.renderPage(w, r, http.StatusUnprocessableEntity, "signup", data)
		return
	}

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (h *Handler) userLogin(w http.ResponseWriter, r *http.Request) {
	p := parseRequestParams(r)
	p.CurrentDB = r.URL.Query().Get("db")
	data := newTemplateData(p.CurrentDB)

	data.Form = forms.UserLogin{}
	data.Colorschemes = h.colorschemes
	data.CurrentColorscheme = getThemeFromCookie(r)
	data.PageTitle = "Login User"
	data.Params = p
	data.URL = buildURLs(data.Params, r)
	data.CSRFToken = nosurf.Token(r)

	h.renderPage(w, r, http.StatusOK, "login", data)
}

func (h *Handler) userLoginPost(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintln(w, "Authenticate and login the user...")
}

func (h *Handler) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintln(w, "Logout the user...")
}

func (h *Handler) notImplementedYet(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintln(w, "Not implemented yet! :D")
}

func (h *Handler) renderPage(w http.ResponseWriter, r *http.Request, n int, page string, d *TemplateData) {
	buf := new(bytes.Buffer)

	// Write the template to the buffer, instead of straight to the
	// http.ResponseWriter. If there's an error, call our responder.ServerErr() helper
	// and then return.
	err := h.template.ExecuteTemplate(buf, page, d)
	if err != nil {
		responder.ServerErr(w, r, err)
		return
	}

	// If the template is written to the buffer without any errors, we are safe
	// to go ahead and write the HTTP status code to http.ResponseWriter.
	w.WriteHeader(n)

	// Write the contents of the buffer to the http.ResponseWriter. Note: this
	// is another time where we pass our http.ResponseWriter to a function that
	// takes an io.Writer.
	if _, err := buf.WriteTo(w); err != nil {
		responder.ServerErr(w, r, err)
		return
	}
}
