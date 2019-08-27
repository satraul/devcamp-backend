package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

// GetBookByID a function to get a single book given it's ID
func (h *Handler) GetBookByID(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	query := fmt.Sprintf("SELECT * FROM books WHERE id = %s", param.ByName("bookID"))
	rows, err := h.DB.Query(query)
	if err != nil {
		log.Printf("[internal][GetUserByID] fail select user user_id:%s :%+v\n",
			param.ByName("userID"), err)
		return
	}
	defer rows.Close()

	var books []Book

	for rows.Next() {
		book := Book{}
		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Author,
			&book.ISBN,
			&book.Stock,
		)
		if err != nil {
			log.Println(err)
			continue
		}
		books = append(books, book)
	}

	bytes, err := json.Marshal(books)
	if err != nil {
		log.Println(err)
		return
	}

	renderJSON(w, bytes, http.StatusOK)
}

// InsertBook a function to insert book to DB
func (h *Handler) InsertBook(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	// read json body
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		renderJSON(w, []byte(`
			message: "Failed to read body"
		`), http.StatusBadRequest)
		return
	}

	// parse json body
	var book Book
	err = json.Unmarshal(body, &book)
	if err != nil {
		log.Println(err)
		return
	}

	query := fmt.Sprintf("INSERT INTO books (id, title, author, isbn, stock) VALUES (%d, '%s', '%s', '%s', %d)", book.ID, *book.Title, *book.Author, *book.ISBN, book.Stock)
	_, err = h.DB.Exec(query)
	if err != nil {
		log.Println(err)
		return
	}

	// executing insert query
	renderJSON(w, []byte(`
	{
		status: "success",
		message: "Insert book success!"
	}
	`), http.StatusOK)
}

// EditBook a function to change book data in DB, with given params
func (h *Handler) EditBook(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		renderJSON(w, []byte(`
			message: "Failed to read body"
		`), http.StatusBadRequest)
		return
	}

	// parse json body
	var book Book
	err = json.Unmarshal(body, &book)
	if err != nil {
		log.Println(err)
		return
	}

	query := fmt.Sprintf("UPDATE books SET title = '%s', author = '%s', isbn = '%s', stock = %d WHERE id = %s", *book.Title, *book.Author, *book.ISBN, book.Stock, param.ByName("bookID"))
	_, err = h.DB.Exec(query)
	if err != nil {
		log.Println(err)
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

// DeleteBookByID a function to remove book data from DB, given bookID
func (h *Handler) DeleteBookByID(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	defer r.Body.Close()
	query := fmt.Sprintf("DELETE FROM books WHERE ID = %s", param.ByName("bookID"))
	_, err := h.DB.Exec(query)
	if err != nil {
		log.Println(err)
		return
	}

	// executing insert query
	renderJSON(w, []byte(`
	{
		status: "success",
		message: "Delete book success!"
	}
	`), http.StatusOK)
}

// InsertMultipleBooks a function to insert multiple book data, given file of books data
func (h *Handler) InsertMultipleBooks(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	var buffer bytes.Buffer

	file, header, err := r.FormFile("books")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	// get file name
	name := strings.Split(header.Filename, ".")
	if name[1] != "csv" {
		log.Println("File format not supported")
		return
	}
	log.Printf("Received a file with name = %s\n", name[0])

	// copy file to buffer
	io.Copy(&buffer, file)

	// TODO: uncomment this when implementing
	contents := buffer.String()

	// Split contents to rows
	// TODO: uncomment this when implementing
	rows := strings.Split(contents, "\n")

	// TODO: iterate csv rows here.
	for _, value := range rows[1:] {
		columns := strings.Split(value, ",")

		/** converting the str1 variable into an int using Atoi method */
		id, err := strconv.Atoi(columns[0])
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("inserting %s", columns[1])
		query := fmt.Sprintf("INSERT INTO books (id, title, author, isbn, stock) VALUES (%d, '%s', '%s', '%s', %s)", id+2007, columns[1], columns[2], columns[3], columns[4])
		_, err = h.DB.Exec(query)
		if err != nil {
			log.Println(err)
			continue
		}
	}

	buffer.Reset()

	renderJSON(w, []byte(`
	{
		status: "success",
		message: "Insert book success!"
	}
	`), http.StatusOK)
}

// LendBook a function to record book lending in DB and update book stock in book tables
func (h *Handler) LendBook(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	nowTime := time.Now()
	body, err := ioutil.ReadAll(r.Body)

	// parse json body
	var lendrequest LendRequest
	err = json.Unmarshal(body, &lendrequest)
	if err != nil {
		log.Println(err)
		return
	}

	// Read userID
	bookID := lendrequest.BookID
	userID := lendrequest.UserID

	// Get book stock from DB
	query := fmt.Sprintf("SELECT stock FROM books WHERE id = %d", bookID)
	rows, err := h.DB.Query(query)
	if err != nil {
		log.Println(err)
		return
	}
	rows.Close()

	var stock int
	for rows.Next() {
		err := rows.Scan(
			&stock,
		)
		if err != nil {
			log.Println(err)
			continue
		}
	}

	// Insert Book to Lend tables
	query = fmt.Sprintf("INSERT INTO lend (user_id, book_id) VALUES (%d, %d)", userID, bookID)
	_, err = h.DB.Exec(query)
	if err != nil {
		log.Println(err)
		return
	}

	// Update Book stock query
	query = fmt.Sprintf("UPDATE books SET stock = %d WHERE id = %d", stock-1, bookID)
	_, err = h.DB.Exec(query)
	if err != nil {
		log.Println(err)
		return
	}

	renderJSON(w, []byte(`
	{
		status: "success",
		message: "Lend book success!"
	}
	`), http.StatusOK)
	processTime := time.Since(nowTime).Seconds()
	log.Printf("[internal][LendBook] execution time %f\n", processTime)
}
