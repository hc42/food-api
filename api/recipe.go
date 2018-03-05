package api

import (
  "encoding/json"
  "io"
  "log"
  "net/http"
  "strconv"
  "github.com/gorilla/mux"
  "github.com/hc42/food-api/model"
  "github.com/hc42/food-api/library"
)

func ListRecipes(w http.ResponseWriter, r *http.Request) {
  var limit int = 25
  var page int = 1
  params := r.URL.Query()

  if v, ok := params["limit"]; ok && len(v) > 0 {
    if overwriteLimit, err := strconv.Atoi(v[0]); err == nil && overwriteLimit > 0 && overwriteLimit < 1000 {
      limit = overwriteLimit
    }
  }

  if v, ok := params["page"]; ok && len(v) > 0 {
    if overwritePage, err := strconv.Atoi(v[0]); err == nil && overwritePage > 0 {
      page = overwritePage
    }
  }

  db, tx, err := library.CreateTransaction()
  if err != nil {
    log.Println(err)
    InternalError(w, r)
    return
  }
  defer db.Close()
  defer tx.Commit()

  result, err := model.GetListPage(tx, page, limit)
  if err != nil {
    log.Println(err)
    InternalError(w, r)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(*result)
}

func GetRecipe(w http.ResponseWriter, r *http.Request) {
  params := mux.Vars(r)
  id, err := strconv.ParseInt(params["id"], 10, 64)
  if err != nil {
    log.Println(err)
    NotFound(w, r)
    return
  }

  db, tx, err := library.CreateTransaction()
  if err != nil {
    log.Println(err)
    InternalError(w, r)
    return
  }
  defer db.Close()
  defer tx.Commit()

  recipe, err := model.GetRecipeById(tx, id)
  if err != nil {
    log.Println(err)
    NotFound(w, r)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(*recipe)
}

func DeleteRecipe(userId int64, w http.ResponseWriter, r *http.Request) {
  params := mux.Vars(r)
  id, err := strconv.ParseInt(params["id"], 10, 64)

  db, tx, err := library.CreateTransaction()
  if err != nil {
    log.Println(err)
    InternalError(w, r)
    return
  }
  defer db.Close()

  recipe, err := model.GetRecipeById(tx, id)
  if err != nil {
    if err.Error() != "sql: no rows in result set" {
      log.Println(err)
    }
    NotFound(w, r)
    tx.Commit()
    return
  }

  err = recipe.Delete(tx)
  if err != nil {
    tx.Rollback()
    log.Println(err)
    InternalError(w,r)
    return
  }

  err = tx.Commit()
  if err != nil {
    tx.Rollback()
    log.Println(err)
    InternalError(w,r)
  }
}

func CreateRecipe(userId int64, w http.ResponseWriter, r *http.Request) {
  recipe := &model.Recipe{}
  err := json.NewDecoder(r.Body).Decode(recipe)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "Invalid Json: " + err.Error())
    return
  }

  db, tx, err := library.CreateTransaction()
  if err != nil {
    log.Println(err)
    InternalError(w, r)
    return
  }
  defer db.Close()

  err = recipe.Create(tx)
  if err != nil {
    tx.Rollback()
    log.Println(err)
    InternalError(w, r)
    return
  }

  err = tx.Commit()
  if err != nil {
    tx.Rollback()
    log.Println(err)
    InternalError(w, r)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(*recipe)
}

func UpdateRecipe(userId int64, w http.ResponseWriter, r *http.Request) {

  params := mux.Vars(r)
  id, err := strconv.ParseInt(params["id"], 10, 64)
  if err != nil {
    log.Println(err)
    NotFound(w, r)
    return
  }

  recipe := &model.Recipe{}
  err = json.NewDecoder(r.Body).Decode(recipe)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "Invalid Json: " + err.Error())
    return
  }
  recipe.ID = id

  db, tx, err := library.CreateTransaction()
  if err != nil {
    log.Println(err)
    InternalError(w, r)
    return
  }
  defer db.Close()

  err = recipe.Update(tx)
  if err != nil {
    tx.Rollback()
    log.Println(err)
    InternalError(w, r)
    return
  }

  err = tx.Commit()
  if err != nil {
    tx.Rollback()
    log.Println(err)
    InternalError(w, r)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(*recipe)
}
