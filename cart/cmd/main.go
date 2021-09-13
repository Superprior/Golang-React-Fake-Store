package main

import (
	"fmt"
	"github.com/yigitsadic/fake_store/cart/cart_grpc/cart_grpc"
	"google.golang.org/grpc"
	"log"
	"net"
)

type CartItem struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float32 `json:"price"`
	Image       string  `json:"image"`
}

type CartDatabase struct {
	Storage map[string][]CartItem
}

func NewCartDatabase() *CartDatabase {
	return &CartDatabase{
		Storage: make(map[string][]CartItem),
	}
}

func (d *CartDatabase) formatCartItemsToGrpcCompatible(items []CartItem) []*cart_grpc.CartItem {
	var buildItems []*cart_grpc.CartItem

	for _, item := range items {
		buildItems = append(buildItems, &cart_grpc.CartItem{
			Id:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			Price:       item.Price,
			Image:       item.Image,
		})
	}

	return buildItems
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9000))
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	grpcServer := grpc.NewServer()
	s := server{Database: NewCartDatabase()}

	cart_grpc.RegisterCartServiceServer(grpcServer, &s)

	log.Println("Started to serve cart grpc")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve due to %s\n", err)
	}
}
