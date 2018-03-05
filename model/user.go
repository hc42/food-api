package model

import (
  "database/sql"
  "golang.org/x/crypto/bcrypt"
)

type User struct {
  ID int64 `json:"id"`
  Name string `json:"name"`
  Enabled bool `json:"enabled"`
  password string `json:"-"`
}

type UserListPage struct {
  List []User `json:"list"`
  Limit int `json:"limit"`
  Page int `json:"page"`
}

func (u *User) SetPassword(passwd string) bool {
  if len(passwd) < 5 {
    return false
  }
  hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
  if err != nil {
    return false
  }
  u.password = string(hashedPassword)
  return true
}

func (u *User) CheckPassword(passwd string) bool {
 err := bcrypt.CompareHashAndPassword([]byte(u.password), []byte(passwd))
 return err == nil
}

func (u *User) GetPasswordHash() string {
  return u.password
}

func (user *User) Create(tx *sql.Tx) error {
  result, err := tx.Exec(
    "INSERT INTO `user` (`name`, `enabled`, `password`) VALUES (?, ?, ?)",
    user.Name, user.Enabled, user.password)
  if err != nil {
    return err
  }
  user.ID, err = result.LastInsertId()
  return nil
}

func (user *User) Delete(tx *sql.Tx) error {
  _, err := tx.Exec("DELETE FROM `user` WHERE `id` = ?", user.ID)
  return err
}

func GetUser(tx *sql.Tx, id int64) (*User, error) {
  user := &User{}
  row := tx.QueryRow("SELECT `id`, `name`, `enabled`, `password` FROM `user` WHERE `id` = ?", id)
  err := row.Scan(&(user.ID), &(user.Name), &(user.Enabled), &(user.password))
  return user, err
}

func GetUserByName(tx *sql.Tx, name string) (*User, error) {
  user := &User{}
  row := tx.QueryRow("SELECT `id`, `name`, `enabled`, `password` FROM `user` WHERE `name` = ?", name)
  err := row.Scan(&(user.ID), &(user.Name), &(user.Enabled), &(user.password))
  return user, err
}

func (user *User)Update(tx *sql.Tx) error {
  _, err := tx.Exec(
    "UPDATE `user` SET `name` = ?, `password` = ?, `enabled` = ? WHERE `id` = ?",
    user.Name, user.password, user.Enabled, user.ID)
  return err
}

func GetUserPage(tx *sql.Tx, page, limit int) (*UserListPage, error) {

  list := UserListPage{Limit: limit, Page: page, List: []User{}}

  rows, err := tx.Query("SELECT `id`, `name`, `enabled` FROM `user` LIMIT ?,?", limit * (page - 1), limit)
  if err != nil {
    return nil, err
  }
  defer rows.Close()
  for rows.Next() {
    user := User{}
    err = rows.Scan(&(user.ID), &(user.Name), &(user.Enabled))
    if err != nil {
      return nil, err
    }
    list.List = append(list.List, user)
  }
  err = rows.Err()
  if err != nil {
    return nil, err
  }

  return &list, nil
}
