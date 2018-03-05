package main

import (
  "log"
  "net/http"
  "github.com/gorilla/mux"
  "github.com/hc42/food-api/api"
  "github.com/hc42/food-api/library"
)

func Init() {
  err := library.InitJwtKeys()
  if err != nil {
    log.Fatal(err)
  }
  err = library.InitDb()
  if err != nil {
    log.Fatal(err)
  }
}

func main() {
  Init()

  router := mux.NewRouter()
  router.HandleFunc("/recipes", api.ListRecipes).Methods("GET")
  router.HandleFunc("/recipes/{id:[0-9]+}", api.GetRecipe).Methods("GET")
  router.HandleFunc("/recipes/{id:[0-9]+}", api.RequireLogin(api.DeleteRecipe)).Methods("DELETE")
  router.HandleFunc("/recipes/{id:[0-9]+}", api.RequireLogin(api.UpdateRecipe)).Methods("PUT")
  router.HandleFunc("/recipes", api.RequireLogin(api.CreateRecipe)).Methods("POST")
  router.HandleFunc("/login", api.UserLogin).Methods("GET")
  router.HandleFunc("/self", api.RequireLogin(api.GetSelf)).Methods("GET")
  router.HandleFunc("/self", api.RequireLogin(api.UpdateSelf)).Methods("PUT")
  router.HandleFunc("/self/setPassword", api.RequireLogin(api.SetPassword)).Methods("POST")
  router.HandleFunc("/user", api.RequireLogin(api.CreateUser)).Methods("POST")
  router.HandleFunc("/user", api.RequireLogin(api.ListUsers)).Methods("GET")
  router.HandleFunc("/user/{id:[0-9]+}", api.RequireLogin(api.GetUser)).Methods("GET")
  router.HandleFunc("/user/{id:[0-9]+}", api.RequireLogin(api.DeleteUser)).Methods("DELETE")
  router.HandleFunc("/user/{id:[0-9]+}", api.RequireLogin(api.UpdateUser)).Methods("PUT")
  router.NotFoundHandler = http.HandlerFunc(api.NotFound)
  log.Fatal(http.ListenAndServe(":8000", router))
}
