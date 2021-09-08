package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bojanz/httpx"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.temporal.io/sdk/client"
	"log"
	"net/http"
	"os"
	"temporal-ecommerce/app"
	"time"
)

type (
	ErrorResponse struct {
		Message string
	}

	UpdateEmailRequest struct {
		Email string
	}

	CheckoutRequest struct {
		Email string
	}
)

var (
	HTTPPort = os.Getenv("PORT")
	temporal client.Client
)

func main() {
	var err error
	temporal, err = client.NewClient(client.Options{})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}

	r := mux.NewRouter()
	r.Handle("/products", http.HandlerFunc(GetProductsHandler)).Methods("GET")
	r.Handle("/cart", http.HandlerFunc(CreateCartHandler)).Methods("POST")
	r.Handle("/cart/{workflowID}", http.HandlerFunc(GetCartHandler)).Methods("GET")
	r.Handle("/cart/{workflowID}/add", http.HandlerFunc(AddToCartHandler)).Methods("PUT")
	r.Handle("/cart/{workflowID}/remove", http.HandlerFunc(RemoveFromCartHandler)).Methods("PUT")
	r.Handle("/cart/{workflowID}/checkout", http.HandlerFunc(CheckoutHandler)).Methods("PUT")
	r.Handle("/cart/{workflowID}/email", http.HandlerFunc(UpdateEmailHandler)).Methods("PUT")

	r.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	var cors = handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}), handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}), handlers.AllowedOrigins([]string{"*"}))

	http.Handle("/", cors(r))
	server := httpx.NewServer(":"+HTTPPort, http.DefaultServeMux)
	server.WriteTimeout = time.Second * 240

	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	res := make(map[string]interface{})
	res["products"] = app.Products

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func CreateCartHandler(w http.ResponseWriter, r *http.Request) {
	workflowID := "CART-" + fmt.Sprintf("%d", time.Now().Unix())

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "CART_TASK_QUEUE",
	}

	cart := app.CartState{Items: make([]app.CartItem, 0)}
	we, err := temporal.ExecuteWorkflow(context.Background(), options, app.CartWorkflow, cart)
	if err != nil {
		WriteError(w, err)
		return
	}

	res := make(map[string]interface{})
	res["cart"] = cart
	res["workflowID"] = we.GetID()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func GetCartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	response, err := temporal.QueryWorkflow(context.Background(), vars["workflowID"], "", "getCart")
	if err != nil {
		WriteError(w, err)
		return
	}
	var res interface{}
	if err := response.Get(&res); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var item app.CartItem
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		WriteError(w, err)
		return
	}

	update := app.AddToCartSignal{Route: app.RouteTypes.ADD_TO_CART, Item: item}

	err = temporal.SignalWorkflow(context.Background(), vars["workflowID"], "", "ADD_TO_CART_CHANNEL", update)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res := make(map[string]interface{})
	res["ok"] = 1
	json.NewEncoder(w).Encode(res)
}

func RemoveFromCartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var item app.CartItem
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		WriteError(w, err)
		return
	}

	update := app.RemoveFromCartSignal{Route: app.RouteTypes.REMOVE_FROM_CART, Item: item}

	err = temporal.SignalWorkflow(context.Background(), vars["workflowID"], "", "REMOVE_FROM_CART_CHANNEL", update)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res := make(map[string]interface{})
	res["ok"] = 1
	json.NewEncoder(w).Encode(res)
}

func UpdateEmailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var body UpdateEmailRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		WriteError(w, err)
		return
	}

	updateEmail := app.UpdateEmailSignal{Route: app.RouteTypes.UPDATE_EMAIL, Email: body.Email}

	err = temporal.SignalWorkflow(context.Background(), vars["workflowID"], "", "UPDATE_CART_CHANNEL", updateEmail)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res := make(map[string]interface{})
	res["ok"] = 1
	json.NewEncoder(w).Encode(res)
}

func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var body CheckoutRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		WriteError(w, err)
		return
	}

	checkout := app.CheckoutSignal{Route: app.RouteTypes.CHECKOUT, Email: body.Email}

	err = temporal.SignalWorkflow(context.Background(), vars["workflowID"], "", "CHECKOUT_CHANNEL", checkout)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res := make(map[string]interface{})
	res["ok"] = 1
	json.NewEncoder(w).Encode(res)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	res := ErrorResponse{Message: "Endpoint not found"}
	json.NewEncoder(w).Encode(res)
}

func WriteError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	res := ErrorResponse{Message: err.Error()}
	json.NewEncoder(w).Encode(res)
}
