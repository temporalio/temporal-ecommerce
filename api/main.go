package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bojanz/httpx"
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
	r.Handle("/cart", http.HandlerFunc(CreateCartHandler)).Methods("POST")
	r.Handle("/cart/{workflowID}/{runID}", http.HandlerFunc(GetCartHandler)).Methods("GET")
	r.Handle("/cart/{workflowID}/{runID}/add", http.HandlerFunc(AddToCartHandler)).Methods("PUT")
	r.Handle("/cart/{workflowID}/{runID}/remove", http.HandlerFunc(RemoveFromCartHandler)).Methods("PUT")

	r.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	http.Handle("/", r)
	server := httpx.NewServer(":"+HTTPPort, http.DefaultServeMux)
	server.WriteTimeout = time.Second * 240

	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
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
	res["runID"] = we.GetRunID()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func GetCartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	response, err := temporal.QueryWorkflow(context.Background(), vars["workflowID"], vars["runID"], "getCart")
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

	update := app.UpdateCartMessage{Remove: false, Item: item}

	err = temporal.SignalWorkflow(context.Background(), vars["workflowID"], vars["runID"], "updateCart", update)
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

	update := app.UpdateCartMessage{Remove: true, Item: item}

	err = temporal.SignalWorkflow(context.Background(), vars["workflowID"], vars["runID"], "updateCart", update)
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
