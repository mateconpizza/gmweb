package router

type User struct {
	Signup string
	Login  string
	Logout string
}

func NewUserRoutes() *User {
	return &User{
		Signup: "/user/signup",
		Login:  "/user/login",
		Logout: "/user/logout",
	}
}
