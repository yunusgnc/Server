package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// NewsRequest request struct
//swagger:model NewsRequest
type NewsRequest struct {
	Title     string `json:"title"`
	Detail    string `json:"detail"`
	NewsImage []byte `json:"news_image"`
}

// NewsResponse response struct
//swagger:response NewsResponse
type NewsResponse struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Detail    string `json:"detail"`
	NewsImage []byte `json:"news_image"`
}

// AddNewsItem adds news item to database and creates a json response of the data
func (app *App) AddNewsItem(writer http.ResponseWriter, req *http.Request) {
	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusBadRequest, err, "Failed to read request body")
		return
	}

	var request NewsRequest
	err = json.Unmarshal(reqBody, &request)
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusInternalServerError, err, "Failed to convert json code")
		return
	}

	status, message, errValidate := request.ValidateNews()
	if errValidate != nil {
		app.RenderErrorResponse(writer, status, errValidate, message)
		return
	}

	sql := fmt.Sprint("INSERT INTO news_item(news_title,detail,news_image,created) VALUES($1,$2,$3,$4) returning uid;")
	var lastInsertId int
	err = app.db.QueryRow(sql, request.Title, request.Detail, request.NewsImage, time.Now()).Scan(&lastInsertId)
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusInternalServerError, err, "Failed to write news")
		return
	}

	app.RenderJson(writer, http.StatusOK, NewsResponse{Title: request.Title, Detail: request.Detail, NewsImage: request.NewsImage, ID: lastInsertId})

}

// GetNewsItems gets all news items from database and creates a json response of the data
func (app *App) GetNewsItems(writer http.ResponseWriter, req *http.Request) {
	sql := fmt.Sprintf("SELECT uid,news_title,detail,news_image FROM news_item")
	rows, err := app.db.Query(sql)
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusInternalServerError, err, "Failed get news items")
		return
	}

	var news []NewsResponse
	for rows.Next() {
		response := NewsResponse{}
		rows.Scan(&response.ID, &response.Title, &response.Detail, &response.NewsImage)
		news = append(news, response)
	}

	app.RenderJson(writer, http.StatusOK, news)
}

// FindNewsItem finds news item from database with id and creates a json response of the data
func (app *App) FindNewsItem(writer http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	sql := fmt.Sprint("SELECT uid,news_title,detail,news_image FROM news_item WHERE uid=$1")
	rows, err := app.db.Query(sql, params["id"])
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusInternalServerError, err, "Failed find news")
		return
	}

	response := NewsResponse{}
	rows.Next()
	err = rows.Scan(&response.ID, &response.Title, &response.Detail, &response.NewsImage)
	if err != nil || response.ID == 0 {
		app.RenderErrorResponse(writer, http.StatusNotFound, err, fmt.Sprintf("News item [%s] not found", params["id"]))
		return
	}
	app.RenderJson(writer, http.StatusOK, response)
}

// DeleteNewsItem deletes news item from database with id and creates a json response of the data
func (app *App) DeleteNewsItem(writer http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	sql := fmt.Sprint("DELETE FROM news_item WHERE uid=$1")
	rows, err := app.db.Query(sql, params["id"])
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusInternalServerError, err, "Failed delete news")
		return
	}

	rows.Next()
	app.RenderJson(writer, http.StatusOK, "Delete news successful")

}

// ValidateNews validates request
func (request *NewsRequest) ValidateNews() (int, string, error) {

	if request.Title == "" {
		return http.StatusBadRequest, "Title not null", fmt.Errorf("Title is wrong")
	}
	if request.Detail == "" {
		return http.StatusBadRequest, "Detail not null", fmt.Errorf("Detail is wrong")
	}

	return http.StatusOK, "", nil

}
