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

func UserLogin(w http.ResponseWriter, r *http.Request) {
  params := r.URL.Query()
  name, ok := params["name"]
  if ! ok || len(name) == 0 {
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "missing parameter name")
    return
  }
  password, ok := params["password"]
  if ! ok || len(password) == 0 {
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "missing parameter password")
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

  user, err := model.GetUserByName(tx, name[0])
  if err != nil && err.Error() != "sql: no rows in result set" {
    log.Println(err)
  } else if user.CheckPassword(password[0]) && user.Enabled {
    SetToken(user, w)
    return
  }
  NotFound(w, r)
}

func GetSelf(userId int64, w http.ResponseWriter, r *http.Request) {

  db, tx, err := library.CreateTransaction()
  if err != nil {
    log.Println(err)
    InternalError(w, r)
    return
  }
  defer db.Close()
  defer tx.Commit()

  user, err := model.GetUser(tx, userId)
  if err != nil {
    if err.Error() == "sql: no rows in result set" {
      NotFound(w, r)
      return
    } else {
      log.Println(err)
      return
    }
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(user)
}

func UpdateSelf(userId int64, w http.ResponseWriter, r *http.Request) {

  user := &model.User{}
  err := json.NewDecoder(r.Body).Decode(user)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "Invalid Json: " + err.Error())
    return
  }

  if userId != user.ID {
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "Selected user must be self")
    return
  }

  if ! user.Enabled {
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "Can't disable own account")
    return
  }

  db, tx, err := library.CreateTransaction()
  defer db.Close()

  oldUser, err := model.GetUser(tx, userId)
  if err != nil {
    if err.Error() == "sql: no rows in result set" {
      NotFound(w, r)
      return
    } else {
      log.Println(err)
      return
    }
  }

  oldUser.Name = user.Name

  err = oldUser.Update(tx)
  if err != nil {
    if err.Error() == "UNIQUE constraint failed: user.name" {
      w.WriteHeader(http.StatusBadRequest)
      io.WriteString(w, "name already in use")
    } else {
      log.Println(err)
      InternalError(w, r)
    }
    tx.Rollback()
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
  json.NewEncoder(w).Encode(*oldUser)
}


func SetPassword(userId int64, w http.ResponseWriter, r *http.Request) {

  var passwords struct {
    OldPassword string `json:"oldPassword"`
    NewPassword string `json:"newPassword"`
  }

  err := json.NewDecoder(r.Body).Decode(&passwords)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "Invalid Json: " + err.Error())
    return
  }

  db, tx, err := library.CreateTransaction()
  defer db.Close()

  user, err := model.GetUser(tx, userId)
  if err != nil {
    if err.Error() == "sql: no rows in result set" {
      NotFound(w, r)
      return
    } else {
      log.Println(err)
      return
    }
  }

  if ! user.CheckPassword(passwords.OldPassword) {
    tx.Rollback();
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "Invalid old password")
    return
  }

  if ! user.SetPassword(passwords.NewPassword) {
    tx.Rollback();
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "Invalid new password")
    return
  }

  err = user.Update(tx)
  if err != nil {
    tx.Rollback();
    log.Println(err)
    InternalError(w, r)
    return
  }

  err = tx.Commit()
  if err != nil {
    tx.Rollback();
    log.Println(err)
    InternalError(w, r)
    return
  }
}

func CreateUser(userId int64, w http.ResponseWriter, r *http.Request) {
  var newUser struct {
    *model.User
    Password string `json:"password"`
  }

  err := json.NewDecoder(r.Body).Decode(&newUser)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "Invalid Json: " + err.Error())
    return
  }

  db, tx, err := library.CreateTransaction()
  defer db.Close()

  if ! newUser.User.SetPassword(newUser.Password) {
    tx.Rollback();
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "Invalid new password")
    return
  }

  err = newUser.User.Create(tx)
  if err != nil {
    if err.Error() == "UNIQUE constraint failed: user.name" {
      w.WriteHeader(http.StatusBadRequest)
      io.WriteString(w, "name already in use")
    } else {
      log.Println(err)
      InternalError(w, r)
    }
    tx.Rollback();
    return
  }

  err = tx.Commit()
  if err != nil {
    log.Println(err)
    InternalError(w, r)
    tx.Rollback();
    return
  }
}

func ListUsers(userId int64, w http.ResponseWriter, r *http.Request) {
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

  result, err := model.GetUserPage(tx, page, limit)
  if err != nil {
    log.Println(err)
    InternalError(w, r)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(*result)
}

func GetUser(userId int64, w http.ResponseWriter, r *http.Request) {
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

  user, err := model.GetUser(tx, id)
  if err != nil {
    if err.Error() == "sql: no rows in result set" {
      NotFound(w, r)
    } else {
      log.Println(err)
      InternalError(w, r)
    }
    return
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(*user)
}

func DeleteUser(userId int64, w http.ResponseWriter, r *http.Request) {
  params := mux.Vars(r)
  id, err := strconv.ParseInt(params["id"], 10, 64)
  if err != nil {
    log.Println(err)
    NotFound(w, r)
    return
  }

  if userId == id {
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "Can't delete self")
    return
  }

  db, tx, err := library.CreateTransaction()
  if err != nil {
    log.Println(err)
    InternalError(w, r)
    return
  }
  defer db.Close()

  user, err := model.GetUser(tx, id)
  if err != nil {
    if err.Error() == "sql: no rows in result set" {
      NotFound(w, r)
    } else {
      log.Println(err)
      InternalError(w, r)
    }
    tx.Rollback()
    return
  }

  err = user.Delete(tx)
  if err != nil {
    log.Println(err)
    InternalError(w, r)
    tx.Rollback()
    return
  }

  err = tx.Commit()
  if err != nil {
    log.Println(err)
    InternalError(w, r)
    tx.Rollback()
    return
  }
}

func UpdateUser(userId int64, w http.ResponseWriter, r *http.Request) {

  params := mux.Vars(r)
  id, err := strconv.ParseInt(params["id"], 10, 64)
  if err != nil {
    log.Println(err)
    NotFound(w, r)
    return
  }

  if userId == id {
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "Can't overwrite self, use /self for that")
    return
  }

  user := &model.User{}
  err = json.NewDecoder(r.Body).Decode(user)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "Invalid Json: " + err.Error())
    return
  }

  db, tx, err := library.CreateTransaction()
  defer db.Close()

  oldUser, err := model.GetUser(tx, id)
  if err != nil {
    if err.Error() == "sql: no rows in result set" {
      NotFound(w, r)
      return
    } else {
      log.Println(err)
      return
    }
  }

  oldUser.Name = user.Name
  oldUser.Enabled = user.Enabled

  err = oldUser.Update(tx)
  if err != nil {
    if err.Error() == "UNIQUE constraint failed: user.name" {
      w.WriteHeader(http.StatusBadRequest)
      io.WriteString(w, "name already in use")
    } else {
      log.Println(err)
      InternalError(w, r)
    }
    tx.Rollback()
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
  json.NewEncoder(w).Encode(*oldUser)
}
