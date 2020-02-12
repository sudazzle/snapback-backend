// install postgres
// install go

// https://www.digitalocean.com/community/tutorials/how-to-install-and-use-postgresql-on-ubuntu-18-04
// https://medium.com/@adigunhammedolalekan/build-and-deploy-a-secure-rest-api-with-go-postgresql-jwt-and-gorm-6fadf3da505b
// https://medium.com/coding-blocks/creating-user-database-and-adding-access-on-postgresql-8bfcd2f4a91e

package main

import (
	"log"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"

	// "github.com/gorilla/handlers"
	// "github.com/gorilla/csrf"
	// "snapback/app"

	handlers "snapback/handlers"
	"snapback/models"
	u "snapback/utils"

	// "path/filepath"
	"fmt"
	"net/http"
	"os"
)

// AppInfo for csrf token
type AppInfo struct {
	Version string
	Name    string
}

func main() {
	router := mux.NewRouter()
	// csrfMiddleware := csrf.Protect([]byte(os.Getenv("csrf_token_key")), csrf.Secure(false), csrf.CookieName("snapback_csrf"), csrf.MaxAge(0), csrf.Path("/"))

	// This one is for when the backend and the frontend are in the different servers
	//  csrfMiddleware := csrf.Protect([]byte("32-byte-long-auth-key"), csrf.TrustedOrigins([]string{"http://localhost:8080"}))

	// this is for when backend and frontend are in different server

	// router.HandleFunc("/api/users/new", controllers.CreateUser).Methods("POST")
	// router.HandleFunc("/api/users/login", controllers.Authenticate).Methods("POST")

	// router.HandleFunc("/api/user/changePassword")
	// router.HandleFunc("/api/user/resetPassword")

	// router.HandleFunc("/api/sessions/new", controllers.CreateSession).Methods("POST")

	// router.HandleFunc("/api/sessions/{id}", controllers.CreateSession).Methods("DELETE")
	// router.HandleFunc("/api/sessions/{id}", controllers.CreateSession).Methods("PATCH")

	// router.HandleFunc("/api/sessions/signup", controllers.DoSignup).Methods("POST")
	// router.HandleFunc("/api/sessions/cancel", controllers.CancelSignup).Methods("PATCH")

	var apis = router.PathPrefix("/api").Subrouter()

	fmt.Println(u.RandomString(9))

	apis.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("api endpoint not found")
		w.WriteHeader(http.StatusNotFound)
	})

	apis.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println(r.RequestURI)
			next.ServeHTTP(w, r)
		})
	})

	apis.Use(handlers.JwtAuthentication)
	// apis.Use(csrfMiddleware)

	apis.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-CSRF-Token", csrf.Token(r))
		u.Respond(w, r, &AppInfo{"1.1.0", "Snapback"}, "", "info", http.StatusOK)
	}).Methods("GET")

	apis.HandleFunc("/users/new", handlers.CreateUser).Methods("POST")
	apis.HandleFunc("/users/login", handlers.Authenticate).Methods("POST")
	apis.HandleFunc("/users", handlers.GetUsers).Methods("GET")
	apis.HandleFunc("/users", handlers.GetUsers).Queries("limit", "{limit}", "page", "{page}").Methods("GET")
	apis.HandleFunc("/users/count", handlers.GetUsersCount).Methods("GET")
	apis.HandleFunc("/users/{id:[0-9]+}", handlers.GetUser).Methods("GET")
	apis.HandleFunc("/get-my-info", handlers.GetUser).Methods("GET")
	apis.HandleFunc("/users/{id:[0-9]+}", handlers.UpdateUser).Methods("PATCH")
	apis.HandleFunc("/users/{id:[0-9]+}", handlers.DeleteUserByAdmin).Methods("DELETE")
	apis.HandleFunc("/users/{id:[0-9]+}/changepassword", handlers.ChangeUserPasswordByAdmin).Methods("PATCH")
	apis.HandleFunc("/users/changepassword", handlers.ChangePasswordByUser).Methods("PATCH")
	apis.HandleFunc("/changepassword", handlers.ChangePasswordByToken).Methods("PATCH")
	apis.HandleFunc("/resetpassword", handlers.ResetPassword).Methods("POST")
	apis.HandleFunc("/update-my-info", handlers.UpdateUser).Methods("PATCH")

	apis.HandleFunc("/sessions/new", handlers.CreateSession).Methods("POST")
	apis.HandleFunc("/sessions/next", handlers.GetNextSessions).Methods("GET")
	apis.HandleFunc("/sessions", handlers.GetAllSessions).Methods("GET")
	apis.HandleFunc("/sessions", handlers.GetAllSessions).Queries("limit", "{limit}", "page", "{page}").Methods("GET")

	apis.HandleFunc("/sessions/{id:[0-9]+}", handlers.GetSessionByID).Methods("GET")
	apis.HandleFunc("/sessions/{id:[0-9]+}", handlers.UpdateSession).Methods("PATCH")
	apis.HandleFunc("/sessions/{id:[0-9]+}", handlers.DeleteSession).Methods("DELETE")

	apis.HandleFunc("/sessions/signup", handlers.DoSignup).Methods("POST")
	apis.HandleFunc("/sessions/{id:[0-9]+}/signups", handlers.GetSignupsBySessionID).Methods("GET")
	apis.HandleFunc("/sessions/{id:[0-9]+}/start", handlers.DoAttendence).Methods("PATCH")
	apis.HandleFunc("/get-my-signups", handlers.GetNextSignups).Methods("GET")
	apis.HandleFunc("/sessions/count", handlers.GetSessionCount).Methods("GET")
	// apis.HandleFunc("/get-my-finished-signups", handlers.GetParticipatedSignups).Methods("GET")
	apis.HandleFunc("/signups/{id:[0-9]+}/cancel", handlers.CancelSignup).Methods("DELETE")

	spa := &models.SpaHandler{StaticPath: "ui/dist", IndexPath: "index.html"}
	router.PathPrefix("/").Handler(spa)

	// These are also in case the backend and frontend are in different servers
	// headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Set-Cookie", "Authentication", "Content-Type"})
	// originsOk := handlers.AllowedOrigins([]string{"http://localhost:8080"})
	// methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "PATCH", "OPTIONS"})
	// allowCredentials := handlers.AllowCredentials()

	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	fmt.Println(port)

	log.Fatal(http.ListenAndServe(":"+port, router))
	// These are in case the backend and frontend are in different servers
	// log.Fatal(http.ListenAndServe(":" + port, handlers.CORS(allowCredentials, originsOk, headersOk, methodsOk)(router)))

}
