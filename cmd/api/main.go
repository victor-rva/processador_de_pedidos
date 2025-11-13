package main

import (
	// "encoding/json"
	"net/http"

	// "github.com/go-chi/chi/v5"
	// "github.com/go-chi/chi/v5/middleware"
	"github.com/victor-rva/processador_de_pedidos/internal/entity"
	"github.com/labstack/echo/v4"
)

func main() {
	// //chi
	// r := chi.NewRouter()
	// //http.HandleFunc("/order", OrderHandler)
	// r.Use(middleware.Logger)
	// r.Get("/order", OrderHandler)
	// http.ListenAndServe(":8888", r) // servidor http
	e := echo.New()
	e.GET("/order", OrderHandler)
	e.Logger.Fatal(e.Start(":8888"))
}

func OrderHandler(c echo.Context) error {
	order, _ := entity.NewOrder("1", 10, 1)
	err := order.CalculateFinalPrice()
	if err != nil{
		return c.String(http.StatusInternalServerError, err.Error())
	} 
	return c.JSON(http.StatusOK, order)
}

// func OrderHandler(w http.ResponseWriter, r *http.Request){
// 	order, _ := entity.NewOrder("1", 10, 1)
// 	err := order.CalculateFinalPrice()
// 	if err != nil{
// 		w.WriteHeader(http.StatusInternalServerError)
// 	}
// 	// result := json.NewEncoder(w).Encode(order)
// 	// if result != nil{
// 	// 	w.WriteHeader(http.StatusInternalServerError)
// 	// }
// 	json.NewEncoder(w).Encode(order)
// }