package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// GetUserByID a method to get user given userID params in URL
func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	nowTime := time.Now()
	defer r.Body.Close()
	userID, err := strconv.ParseInt(param.ByName("userID"), 10, 64)
	if err != nil {
		log.Printf("[internal][GetUserByID] fail to convert user_id to int:%+v\n", err)
		renderJSON(w, []byte(`{
			"error":"user_id is not int"
		}`), http.StatusBadRequest)
		return
	}
	query := fmt.Sprintf("SELECT id, COALESCE(name, '-') FROM users WHERE id = %d", userID)
	rows, err := h.DB.Query(query)
	if err != nil {
		log.Printf("[internal][GetUserByID] failed to select user user_id:%s :%+v\n",
			param.ByName("userID"), err)
		return
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		user := User{}
		err := rows.Scan(
			&user.ID,
			&user.Name,
		)
		if err != nil {
			log.Printf("[internal][GetUserByID] failed to scan user:%+v\n",
				err)
			continue
		}
		users = append(users, user)
	}

	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("[internal][GetUserByID] failed to marshal user:%+v\n",
			err)
		return
	}

	renderJSON(w, bytes, http.StatusOK)

	processTime := time.Since(nowTime).Seconds()
	log.Printf("[internal][GetUserByID] execution time %f\n", processTime)
}

// InsertUser a function to insert user data (id, name) to DB
func (h *Handler) InsertUser(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	// read json body
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Printf("[internal][InsertUser] failed to parse json body user:%s :%+v\n",
			param.ByName("userID"), err)
		renderJSON(w, []byte(`
			message: "Failed to read body"
		`), http.StatusBadRequest)
		return
	}

	// parse json body
	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		log.Printf("[internal][InsertUser] failed to parse json body:%+v\n", err)
		return
	}

	query := fmt.Sprintf("INSERT INTO users (id, name) VALUES (%d, '%s')", user.ID, user.Name)
	_, err = h.DB.Exec(query)
	if err != nil {
		log.Printf("[internal][InsertUser] failed to query insert user:%+v :%+v\n",
			user, err)
		return
	}

	// executing insert query
	renderJSON(w, []byte(`
	{
		status: "success",
		message: "Insert user success!"
	}
	`), http.StatusOK)
}

// EditUserByID a function to change user data (name) in DB with given params (id, name)
func (h *Handler) EditUserByID(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Printf("[internal][EditUserByID] failed to parse body:%+v\n",
			err)
		renderJSON(w, []byte(`
			message: "Failed to read body"
		`), http.StatusBadRequest)
		return
	}

	// parse json body
	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		log.Printf("[internal][EditUserByID] failed to unmarshal body to json:%+v\n",
			err)
		return
	}

	query := fmt.Sprintf("UPDATE users SET name = '%s' WHERE id = %s", user.Name, param.ByName("userID"))
	_, err = h.DB.Exec(query)
	if err != nil {
		log.Printf("[internal][EditUserByID] fail to update user:%+v :%+v\n",
			user, err)
		return
	}

	// executing insert query
	renderJSON(w, []byte(`
	{
		status: "success",
		message: "Put book success!"
	}
	`), http.StatusOK)
}

// DeleteUserByID a function to remove user data from DB given the userID
func (h *Handler) DeleteUserByID(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	// TODO: implement this. Query = DELETE FROM users WHERE id = <userID>
	defer r.Body.Close()
	query := fmt.Sprintf("DELETE FROM users WHERE ID = %s", param.ByName("userID"))
	_, err := h.DB.Exec(query)
	if err != nil {
		log.Printf("[internal][DeleteUserByID] failed to query delete user user_id:%s :%+v\n",
			param.ByName("userID"), err)
		return
	}

	// executing insert query
	renderJSON(w, []byte(`
	{
		status: "success",
		message: "Delete user success!"
	}
	`), http.StatusOK)
}
