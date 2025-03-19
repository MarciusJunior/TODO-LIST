package main

import(
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
	"context"
	"os"
	"os/signal"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var rnd *renderer.Render
var db *mgo.Database

const (
	hostname string = "localhost:27017"
	dbName string = "demo_todo"
	collectionName string = "todo"
	port string = ":8080"
)

type(
	todoModel struct {
		ID bson.ObjectId `bson:"_id,omitempty"`
		Title string `bson:"title"`
		Completed bool `bson:"completed"`
		CreatedAt time.Time `bson:"created_at"`
	}

	todo struct {
		ID bson.ObjectId `json:"id"`
		Title string `json:"title"`
		Completed bool `json:"completed"`
		CreatedAt time.Time `json:"created_at"`
	}
)

func init() {
	rnd = renderer.New()
	session, err := mgo.Dial(hostname)
	checkErr(err)
	session.SetMode(mgo.Monotonic, true)
	db = session.DB(dbName)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := rnd.Template(w, http.StatusOK, []string{"/static/home.tpl"}, nil)
	checkErr(err)
}

func fetchHandler(w http.ResponseWriter, r *http.Request) {
	
}

func main() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homeHandler)
	r.Mount("/todo", todoHandlers())

	srv := &http.Server{
		Addr: port,
		Handler: r,
		ReadTimeout: 60*time.Second,
		WriteTimeout: 60*time.Second,
		IdleTimeout: 60*time.Second,
	}
	go func() {
		log.Println("Listening on port", port)
		if err:= srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel(
		log.Println("Server gracefully stopped")
	)
}

func todoHandlers() http.Handler {
	rh := chi.NewRouter()
	rh.Group(func(r chi.Router) {
		r.Get("/", getAllTodo)
		r.Post("/", createTodo)
		r.Get("/{id}", getTodo)
		r.Put("/{id}", updateTodo)
		r.Delete("/{id}", deleteTodo)
	})
	return rh
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}