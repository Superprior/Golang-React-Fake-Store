package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"github.com/yigitsadic/fake_store/auth/auth_grpc/auth_grpc"
	"github.com/yigitsadic/fake_store/cart/cart_grpc/cart_grpc"
	"github.com/yigitsadic/fake_store/gateway/auth"
	"github.com/yigitsadic/fake_store/orders/orders_grpc/orders_grpc"
	"github.com/yigitsadic/fake_store/products/product_grpc/product_grpc"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/yigitsadic/fake_store/gateway/graph"
	"github.com/yigitsadic/fake_store/gateway/graph/generated"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3035"
	}

	var authClient auth_grpc.AuthServiceClient
	var productClient product_grpc.ProductServiceClient
	var cartClient cart_grpc.CartServiceClient
	var orderClient orders_grpc.OrdersServiceClient

	authConnection, err := acquireConnection("auth")
	if err != nil {
		authConnection = nil

		log.Println("Cannot obtain auth service connection")
	} else {
		authClient = auth_grpc.NewAuthServiceClient(authConnection)
		defer authConnection.Close()
	}

	productsConnection, err := acquireConnection("products")
	if err != nil {
		productClient = nil

		log.Println("Cannot obtain product service connection")
	} else {
		productClient = product_grpc.NewProductServiceClient(productsConnection)

		defer productsConnection.Close()
	}

	cartConnection, err := acquireConnection("cart")
	if err != nil {
		cartClient = nil

		log.Println("Cannot obtain product service connection")
	} else {
		cartClient = cart_grpc.NewCartServiceClient(cartConnection)

		defer cartConnection.Close()
	}

	ordersConnection, err := acquireConnection("orders")
	if err != nil {
		orderClient = nil

		log.Println("Cannot obtain auth service connection")
	} else {
		orderClient = orders_grpc.NewOrdersServiceClient(ordersConnection)
		defer authConnection.Close()
	}

	resolver := graph.Resolver{
		AuthService:     authClient,
		ProductsService: productClient,
		CartService:     cartClient,
		OrdersService:   orderClient,
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolver}))

	r := chi.NewRouter()

	r.Use(cors.AllowAll().Handler)

	// Parses JWT token in the Authorization key in header and stores it to context with key *userId*
	r.Use(auth.Middleware)

	r.Handle("/", playground.Handler("GraphQL playground", "/query"))
	r.Handle("/query", srv)

	log.Printf("Server is up and running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func acquireConnection(serviceName string) (*grpc.ClientConn, error) {
	dialCtx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, fmt.Sprintf("%s:9000", serviceName), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	return conn, nil
}
