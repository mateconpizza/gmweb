// Package router provides URL path generation for web and API endpoints with
// database-specific routing.
package router

type Router struct {
	API  *API
	Web  *WebRouter
	User *User
}

func New(db string) *Router {
	if db == "" {
		panic("database name cannot be empty")
	}

	return &Router{
		Web:  NewWebRoutes(db),
		User: NewUserRoutes(),
		API:  NewAPIRoutes(db),
	}
}

// SetRepo updates the repository name for all routes.
func (r *Router) SetRepo(db string) *Router {
	r.API = NewAPIRoutes(db)
	r.Web = NewWebRoutes(db)

	return r
}

////////////////////

type BookmarkRoutes struct {
	db string
}

type RouteBuilder struct {
	db string
}

func NewRouteBuilder(db string) *RouteBuilder {
	return &RouteBuilder{db: db}
}

func (rb *RouteBuilder) Bookmarks() *BookmarkRoutes {
	return &BookmarkRoutes{db: rb.db}
}

func (rb *RouteBuilder) User() *User {
	return &User{}
}

func (rb *RouteBuilder) Static() *StaticRoutes {
	return &StaticRoutes{}
}

type StaticRoutes struct{}

func (sr *StaticRoutes) Favicon() string { return "/static/img/favicon.png" }
