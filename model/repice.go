package model

import (
  "database/sql"
)

type Recipe struct {
  ID int64 `json:"id"`
  Title string `json:"title"`
  Description string `json:"description"`
  Ingredients []Ingredient `json:"ingredients"`
}

type RecipeListPage struct {
  List []Recipe `json:"list"`
  Limit int `json:"limit"`
  Page int `json:"page"`
}

func GetListPage(tx *sql.Tx, page, limit int) (* RecipeListPage, error) {

  list := RecipeListPage{Limit: limit, Page: page, List: []Recipe{}}

  rows, err := tx.Query("SELECT `id`, `title`, `description` FROM `recipe` LIMIT ?,?", limit * (page - 1), limit)
  if err != nil {
    return nil, err
  }
  defer rows.Close()
  for rows.Next() {
    recipe := Recipe{}
    err = rows.Scan(&(recipe.ID), &(recipe.Title), &(recipe.Description))
    if err != nil {
      return nil, err
    }
    ingredients, err := GetIngrediants(tx, recipe.ID)
    if err != nil {
      return nil, err
    }
    recipe.Ingredients = *ingredients
    list.List = append(list.List, recipe)
  }
  err = rows.Err()
  if err != nil {
    return nil, err
  }

  return &list, nil
}

func GetRecipeById(tx *sql.Tx, id int64) (*Recipe, error) {
  recipe := &Recipe{}
  row := tx.QueryRow("SELECT `id`, `title`, `description` FROM `recipe` WHERE `id` = ?", id)
  err := row.Scan(&(recipe.ID), &(recipe.Title), &(recipe.Description))
  if err != nil {
    return nil, err
  }
  ingredients, err := GetIngrediants(tx, id)
  if err != nil {
    return nil, err
  }
  recipe.Ingredients = *ingredients
  return recipe, nil
}

func (recipe *Recipe)Delete(tx *sql.Tx) error {

  for _, ingredient := range recipe.Ingredients {
    err := ingredient.Delete(tx)
    if err != nil {
      return err
    }
  }

  _, err := tx.Exec("DELETE FROM `recipe` WHERE `id` = ?", recipe.ID)
  return err
}

func (recipe *Recipe)Create(tx *sql.Tx) error {
  result, err := tx.Exec(
    "INSERT INTO `recipe` (`title`, `description`) VALUES (?,?)",
    recipe.Title, recipe.Description)
  if err != nil {
    return err
  }
  recipe.ID, err = result.LastInsertId()
  if err != nil {
    return err
  }

  for idx, ingredient := range recipe.Ingredients {
    err := ingredient.Create(tx, recipe)
    if err != nil {
      return err
    }
    recipe.Ingredients[idx] = ingredient
  }
  return nil
}

func (recipe *Recipe)Update(tx *sql.Tx) error {
  _, err := tx.Exec(
    "UPDATE `recipe` SET `title` = ?, `description` = ? WHERE `id` = ?",
    recipe.Title, recipe.Description, recipe.ID)
  if err != nil {
    return err
  }

  ingredients, err := GetIngrediants(tx, recipe.ID)
  if err != nil {
    return err
  }
  for _, ingredient := range *ingredients {
    found := false
    for _, ingredientNew := range recipe.Ingredients {
      if ingredient.ID == ingredientNew.ID {
        found = true
        break
      }
    }
    if ! found {
      err := ingredient.Delete(tx)
      if err != nil {
        return err
      }
    }
  }

  for idx, ingredient := range recipe.Ingredients {
    found, err := HasIngrediantForRecipe(tx, ingredient.ID, recipe.ID)
    if err != nil {
      return err
    }
    if found {
      err = ingredient.Update(tx)
    } else {
      err = ingredient.Create(tx, recipe)
      recipe.Ingredients[idx] = ingredient
    }
    if err != nil {
      return err
    }
  }
  return nil
}
