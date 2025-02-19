package main

import (
	graph2 "github.com/aaanger/graphql-test/internal/graph"
	commentRepository "github.com/aaanger/graphql-test/internal/repository/comment"
	postRepository "github.com/aaanger/graphql-test/internal/repository/post"
	UserRepository "github.com/aaanger/graphql-test/internal/repository/user"
	"github.com/aaanger/graphql-test/pkg/db"
	"github.com/aaanger/graphql-test/pkg/middleware"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.Fatalf("Error loading .env file: %s", err)
	}

	db, err := db.Open(db.PostgresConfig{
		Host:     os.Getenv("PSQL_HOST"),
		Port:     os.Getenv("PSQL_PORT"),
		User:     os.Getenv("PSQL_USER"),
		Password: os.Getenv("PSQL_PASSWORD"),
		DBName:   os.Getenv("PSQL_DBNAME"),
		SSLMode:  "disable",
	})
	if err != nil {
		logrus.Fatalf("Error connecting to db: %s", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	userRepo := UserRepository.NewUserRepository(db)
	postRepo := postRepository.NewPostRepository(db)
	commentRepo := commentRepository.NewCommentRepository(db)

	srv := handler.New(graph2.NewExecutableSchema(graph2.Config{Resolvers: &graph2.Resolver{
		UserRepo:    userRepo,
		PostRepo:    postRepo,
		CommentRepo: commentRepo,
	}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", middleware.UserIdentity(srv))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
