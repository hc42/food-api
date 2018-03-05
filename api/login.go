package api

import (
  "strconv"
  "net/http"
  "log"
  "github.com/hc42/food-api/library"
  "github.com/hc42/food-api/model"
)

func SetToken(user *model.User, w http.ResponseWriter) bool {
  token, err := library.CreateJwtToken (strconv.FormatInt(user.ID, 10), user.Name)
  if err == nil {
    w.Header().Set("Authorization", "BEARER " + token)
  }
  return err != nil
}

func RequireLogin(handler func(userId int64, w http.ResponseWriter, r *http.Request)) (func(w http.ResponseWriter, r *http.Request)) {
  return func(w http.ResponseWriter, r *http.Request) {

    subject, err := library.ValidateJwtAndGetSubject(r)
    if err != nil {
      notLoggedIn(w)
      log.Println(err)
      return
    }

    id, err := strconv.ParseInt(subject, 10, 64)
    if err != nil {
      log.Println("JWT subject no int")
      log.Println(err)
      notLoggedIn(w)
      return
    }

    db, tx, err := library.CreateTransaction()
    if err != nil {
      log.Println(err)
      InternalError(w, r)
      return
    }

    user, err := model.GetUser(tx, id)
    if err != nil && err.Error() != "sql: no rows in result set" {
      log.Println(err)
      InternalError(w, r)
      db.Close()
      return
    } else if ! user.Enabled {
      notLoggedIn(w)
      db.Close()
      return
    }
    err = tx.Commit()
    if err != nil {
      log.Println(err)
      InternalError(w, r)
      db.Close()
      return
    }
    db.Close()

    // Handle logged in request
    handler(id, w, r)
  }
}
