package model

import (
  "database/sql"
)

type Ingredient struct {
  ID int64 `json:"id"`
  Name string `json:"name"`
  Quantity string `json:"quantity"`
}

func GetIngrediants(tx *sql.Tx, recipeId int64) (* []Ingredient, error) {
  list := []Ingredient{}

  rows, err := tx.Query("SELECT `id`, `name`, `quantity` FROM `ingredient` WHERE `recipe` = ?", recipeId)
  if err != nil {
    return nil, err
  }

  defer rows.Close()
  for rows.Next() {
    ingredient := Ingredient{}
    err = rows.Scan(&(ingredient.ID), &(ingredient.Name), &(ingredient.Quantity))
    if err != nil {
      return nil, err
    }
    list= append(list, ingredient)
  }
  err = rows.Err()
  if err != nil {
    return nil, err
  }

  return &list, nil
}

func (ingredient *Ingredient)Delete(tx *sql.Tx) error {
  _, err := tx.Exec("DELETE FROM `ingredient` WHERE `id` = ?", ingredient.ID)
  return err
}

func (ingredient *Ingredient)Create(tx *sql.Tx, recipe *Recipe) error {
  result, err := tx.Exec(
    "INSERT INTO `ingredient` (`name`, `quantity`, `recipe`) VALUES (?,?,?)",
    ingredient.Name, ingredient.Quantity, recipe.ID)
  if err != nil {
    return err
  }
  ingredient.ID, err = result.LastInsertId()
  if err != nil {
    return err
  }
  return nil;
}

func HasIngrediantForRecipe(tx *sql.Tx, id, recipe int64) (bool, error) {
  var count int
  row := tx.QueryRow(
    "SELECT COUNT(*) FROM `ingredient` WHERE `id` = ? AND `recipe` = ?",
    id, recipe)
  err := row.Scan(&count)
  return count == 1, err
}

func (ingredient *Ingredient)Update(tx *sql.Tx) error {
  _, err := tx.Exec(
    "UPDATE `ingredient` SET `name` = ?, `quantity` = ? WHERE `id` = ?",
    ingredient.Name, ingredient.Quantity, ingredient.ID)
  return err
}
