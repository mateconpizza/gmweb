package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/mateconpizza/gm/pkg/bookio"
	"github.com/mateconpizza/gm/pkg/bookmark"
	"github.com/mateconpizza/gm/pkg/files"

	"github.com/mateconpizza/gmweb/internal/forms"
	"github.com/mateconpizza/gmweb/internal/helpers"
	"github.com/mateconpizza/gmweb/internal/middleware"
	"github.com/mateconpizza/gmweb/internal/qr"
	"github.com/mateconpizza/gmweb/internal/responder"
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

	// Bookmarks
	mux.Handle("GET "+r.Web.All(), requireDB(h.index))
	mux.Handle("GET "+r.Web.New(), requireDB(h.recordNew))
	mux.Handle("GET "+r.Web.NewFrame(), requireDB(h.recordNewFrame))
	mux.Handle("GET "+r.Web.Edit("{id}"), requireIDAndDB(h.recordEdit))
	mux.Handle("GET "+r.Web.QRCode("{id}"), requireIDAndDB(h.recordQR))
	mux.Handle("GET "+r.Web.Export(), requireDB(h.recordExport))
	mux.HandleFunc("POST "+r.Web.Settings(), h.settings)

	// User related
	mux.HandleFunc("GET "+r.User.Signup, h.userSignup)
	mux.HandleFunc("POST "+r.User.Signup, h.userSignupPost)
	mux.HandleFunc("GET "+r.User.Login, h.userLogin)
	mux.HandleFunc("POST "+r.User.Login, h.userLoginPost)
	mux.HandleFunc("POST "+r.User.Logout, h.userLogoutPost)

	// static and cache files
	staticFS, err := fs.Sub(h.files, "static")
	if err != nil {
		panic(err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	iconsPath := filepath.Join(h.cacheDir, "favicon")
	mux.Handle(ui.CacheFavicon, http.StripPrefix(ui.CacheFavicon, http.FileServer(http.Dir(iconsPath))))
}

func (h *Handler) indexRedirect(w http.ResponseWriter, r *http.Request) {
	dbName := cookie.get(r, cookie.jar.defaultRepoName, h.appCfg.MainDB)
	h.router.SetRepo(dbName)
	http.Redirect(w, r, h.router.Web.All(), http.StatusSeeOther)
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	p := parseRequestParams(r)
	p.CurrentDB = r.PathValue("db")

	h.itemsPerPage = cookie.getInt(r, cookie.jar.itemsPerPage, h.itemsPerPage)

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

	favicons := NewFaviconProcessor(repo, faviconPath, ui.CacheFavicon)
	go favicons.Process(paginated)

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
	h.renderPage(w, r, http.StatusOK, "index", data)
}

func (h *Handler) recordQR(w http.ResponseWriter, r *http.Request) {
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

	f, err := h.files.ReadFile(ui.ColorSchemesJSON)
	if err != nil {
		responder.ServerErr(w, r, err)
	}

	currentTheme := files.StripSuffixes(
		filepath.Base(cookie.get(r, cookie.jar.themeCurrent, ui.DefaultColorsCSS)),
	)
	t, err := getCurrentTheme(f, currentTheme)
	if err != nil {
		responder.ServerErr(w, r, err)
	}

	bg := t.Light.Bg
	fg := t.Light.Fg
	mode := cookie.get(r, cookie.jar.themeMode, "light")
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

func (h *Handler) recordEdit(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	repo, err := h.repoLoader(data.Params.CurrentDB)
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

	data.Bookmark = b
	data.Colorschemes = h.colorschemes
	data.PageTitle = "Edit Bookmark"

	if err := h.template.ExecuteTemplate(w, "bookmark-edit", data); err != nil {
		responder.ServerErr(w, r, err)
	}
}

func (h *Handler) recordNew(w http.ResponseWriter, r *http.Request) {
	// FIX: maybe, redirect to `view` new record.
	data := newTemplateData(r)
	data.Colorschemes = h.colorschemes
	data.PageTitle = "New Bookmark"
	data.URL = buildURLs(data.Params, r)

	if err := h.template.ExecuteTemplate(w, "bookmark-new", data); err != nil {
		responder.ServerErr(w, r, err)
	}
}

func (h *Handler) recordNewFrame(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	u := r.URL.Query().Get("url")

	repo, err := h.repoLoader(data.Params.CurrentDB)
	if err != nil {
		responder.ServerErr(w, r, err)
		return
	}

	data.Colorschemes = h.colorschemes
	data.PageTitle = "New Bookmark"
	data.URL = buildURLs(data.Params, r)

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
	// FIX: this is broken, new router constructor.
	data := newTemplateData(r)
	data.Params.CurrentDB = r.URL.Query().Get("db")
	data.Form = forms.UserSignUp{}
	data.Colorschemes = h.colorschemes
	data.PageTitle = "New User"
	data.URL = buildURLs(data.Params, r)

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

		data := newTemplateData(r)
		data.Form = f
		data.Params.CurrentDB = dbName
		data.FormHasErrors = true
		data.Colorschemes = h.colorschemes
		data.PageTitle = "New User"
		data.URL = buildURLs(data.Params, r)

		h.renderPage(w, r, http.StatusUnprocessableEntity, "signup", data)
		return
	}

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (h *Handler) userLogin(w http.ResponseWriter, r *http.Request) {
	td := newTemplateData(r)
	td.Params.CurrentDB = r.URL.Query().Get("db")
	td.Form = forms.UserLogin{}
	td.Colorschemes = h.colorschemes
	td.PageTitle = "Login User"
	td.URL = buildURLs(td.Params, r)

	h.renderPage(w, r, http.StatusOK, "login", td)
}

func (h *Handler) recordExport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="bookmarks.html"`)
	p := parseRequestParams(r)

	dbName := r.PathValue("db")
	repo, err := h.repoLoader(dbName)
	if err != nil {
		h.logger.Error("repo info", "error", err, "repo", dbName)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	records, err := repo.All(r.Context())
	if err != nil {
		http.Error(w, "failed to get all the bookmarks", http.StatusInternalServerError)
		return
	}

	filtered := helpers.ApplyFiltersAndSorting(p.Tag, p.Query, p.Letter, p.FilterBy, records)
	if err := bookio.ExportToNetscapeHTML(filtered, w); err != nil {
		http.Error(w, "failed to export bookmarks", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) settings(w http.ResponseWriter, r *http.Request) {
	var f forms.AppSettings
	err := forms.DecodePostForm(r, &f)
	if err != nil {
		h.logger.Error("settings", "error", err)
		responder.ServerCustomErr(w, r, err, http.StatusBadRequest)
		return
	}

	themeMode := "light"
	if f.DarkMode {
		themeMode = "dark"
	}

	cookie.set(w, cookie.jar.compactMode, strconv.FormatBool(f.CompactMode))
	cookie.set(w, cookie.jar.itemsPerPage, strconv.Itoa(f.ItemsPerPage))
	cookie.set(w, cookie.jar.themeCurrent, f.ThemeName)
	cookie.set(w, cookie.jar.themeMode, themeMode)
	cookie.set(w, cookie.jar.vimMode, strconv.FormatBool(f.VimMode))

	http.Redirect(w, r, h.router.Web.All(), http.StatusSeeOther)
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
	var tmpl *template.Template
	var err error

	if devMode {
		// always reload from disk in dev
		tmpl, err = template.New("pages/base").Funcs(templateFuncs).ParseGlob("ui/templates/**/*.gohtml")
	} else {
		tmpl = h.template
	}

	if err != nil {
		responder.ServerErr(w, r, err)
		return
	}

	buf := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(buf, page, d); err != nil {
		responder.ServerErr(w, r, err)
		return
	}

	w.WriteHeader(n)
	if _, err := buf.WriteTo(w); err != nil {
		responder.ServerErr(w, r, err)
		return
	}
}
