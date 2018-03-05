package library

import (
  "database/sql"
  "fmt"
  "log"
  "math/rand"
  "time"
  _ "github.com/mattn/go-sqlite3"
  "github.com/hc42/food-api/model"
)

func InitDb() error {
  log.Println("Initialize DB schema")
  db, err := ConnectDb()
  if err != nil {
    return err
  }
  defer db.Close()
  err = createSchema(db)
  if err != nil {
    return err
  }
  err = createFirstUser(db)
  return err
}

func createFirstUser(db *sql.DB) error {
  count := -1
  err := db.QueryRow("SELECT COUNT(*) FROM `user`").Scan(&count)
  if err != nil {
    return err
  }
  if (count > 0) {
    log.Printf("%d users found in db\n", count)
    return nil
  }
  rand.Seed(time.Now().UTC().UnixNano())
  newPassword := fmt.Sprintf("admin%d", rand.Intn(1000000))
  newUser := model.User{Name:"admin", Enabled:true}
  newUser.SetPassword(newPassword)
  tx, err := db.Begin()
  if err != nil {
    return err
  }
  err = newUser.Create(tx)
  if err == nil {
    err = tx.Commit()
  }
  if err != nil {
    tx.Rollback()
    return err
  }
  log.Printf(
    "New first user created with username %s and password %s\nChange password after first login\n",
     newUser.Name, newPassword)
  return nil
}


func ConnectDb() (*sql.DB, error) {
  return sql.Open("sqlite3", "food.db")
}

func CreateTransaction() (*sql.DB, *sql.Tx, error) {
  db, err := sql.Open("sqlite3", "food.db")
  if err != nil {
    return nil, nil, err
  }
  tx, err := db.Begin()
  if err != nil {
    db.Close()
    return nil, nil, err
  }
  return db, tx, nil
}

func createSchema(db *sql.DB) error {

  var tables []string

  tables = append(tables, "CREATE TABLE IF NOT EXISTS `user` (" +
    "`id` INTEGER NOT NULL PRIMARY KEY," +
    "`name` VARCHAR(255) NOT NULL UNIQUE," +
    "`enabled` BOOL NOT NULL," +
    "`password` VARCHAR(255) NULL)")

  tables = append(tables, "CREATE TABLE IF NOT EXISTS `recipe` (" +
    "`id` INTEGER NOT NULL PRIMARY KEY," +
    "`title` VARCHAR(255) NOT NULL," +
    "`description` TEXT NOT NULL)")

  tables = append(tables, "CREATE TABLE IF NOT EXISTS `ingredient` (" +
    "`id` INTEGER NOT NULL PRIMARY KEY," +
    "`name` VARCHAR(255) NOT NULL," +
    "`quantity` VARCHAR(255) NULL," +
    "`recipe` INTEGER NOT NULL," +
    "FOREIGN KEY(recipe) REFERENCES reipce(id))")

  for _, table := range tables {
    _, err := db.Exec(table)
    if err != nil {
      return err
    }
  }
  return nil
}
