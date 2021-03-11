package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// ProjectRequest response struct
//swagger:model ProjectRequest
type ProjectRequest struct {
	ProjectName   string    `json:"project_name"`
	Detail        string    `json:"detail"`
	ProjectImages [][]byte  `json:"project_images"`
	StartDate     time.Time `json:"start_date"`
	FinishDate    time.Time `json:"finish_date"`
}

// ProjectResponse response struct
//swagger:response ProjectResponse
type ProjectResponse struct {
	ID            int       `json:"id"`
	ProjectName   string    `json:"project_name"`
	Detail        string    `json:"detail"`
	ProjectImages [][]byte  `json:"project_images"`
	StartDate     time.Time `json:"start_date"`
	FinishDate    time.Time `json:"finish_date"`
}

// AddProjectItem add new ProjectItem to database and creates a json response of the data
func (app *App) AddProjectItem(writer http.ResponseWriter, req *http.Request) {

	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusBadRequest, err, "Failed to read request body")
		return
	}

	var request ProjectRequest
	err = json.Unmarshal(reqBody, &request)
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusInternalServerError, err, "Failed convert JSON")
		return
	}

	status, message, errorValidate := request.ValidateProject()
	if errorValidate != nil {
		app.RenderErrorResponse(writer, status, errorValidate, message)
		return
	}

	sql := fmt.Sprint("INSERT INTO project(project_name,detail,project_images,start_date,finish_date,created) VALUES($1,$2,$3,$4,$5,$6) returning uid;")
	var lastInsertId int
	err = app.db.QueryRow(sql, request.ProjectName,
		request.Detail,
		request.ProjectImages,
		request.StartDate,
		request.FinishDate,
		time.Now(),
	).Scan(&lastInsertId)

	if err != nil {
		app.RenderErrorResponse(writer, http.StatusInternalServerError, err, "Failed to create project")
		return
	} else {
		app.RenderJson(writer, http.StatusOK, ProjectResponse{ProjectName: request.ProjectName, Detail: request.Detail, ProjectImages: request.ProjectImages, ID: lastInsertId})
	}
}

// GetProjectItems fetches all project items from database and creates a json response of the data
func (app *App) GetProjectItems(writer http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	sql := fmt.Sprintf("SELECT uid,project_name,detail,project_images,start_date,finish_date FROM where start_date >= $1 and finish_date < $2 project")
	rows, err := app.db.Query(sql, params["start_date"], "finish_date")
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusInternalServerError, err, "Failed get projects")
		return
	}

	var project []ProjectResponse
	for rows.Next() {
		u := ProjectResponse{}
		rows.Scan(&u.ID, &u.ProjectName, &u.Detail, &u.ProjectImages, &u.StartDate, &u.FinishDate)
		project = append(project, u)
	}
	app.RenderJson(writer, http.StatusOK, project)
}

// FindProjectItem finds project item from database and creates a json response of the data
func (app *App) FindProjectItem(writer http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	sql := fmt.Sprint("SELECT uid,project_name,detail,project_images,start_date,finish_date FROM project WHERE uid=$1")
	rows, err := app.db.Query(sql, params["id"])

	if err != nil {
		app.RenderErrorResponse(writer, http.StatusBadRequest, err, "Failed find project")
		return
	}
	u := ProjectResponse{}
	for rows.Next() {
		err = rows.Scan(&u.ID, &u.ProjectName, &u.Detail, &u.ProjectImages, &u.StartDate, &u.FinishDate)
	}
	if err != nil || u.ID == 0 {
		app.RenderErrorResponse(writer, http.StatusNotFound, err, fmt.Sprintf("Project [%s] not found", params["id"]))
		return
	}
	app.RenderJson(writer, http.StatusOK, u)
}

// DeleteProjectItem deletes project item from database and creates a json response of the data
func (app *App) DeleteProjectItem(writer http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	sql := fmt.Sprint("DELETE FROM project WHERE uid=$1")
	rows, err := app.db.Query(sql, params["id"])
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusInternalServerError, err, "Failed delete project")
		return
	}

	rows.Next()
	app.RenderJson(writer, http.StatusOK, nil)
}

// ValidateProject validates request
func (request *ProjectRequest) ValidateProject() (int, string, error) {
	if request.ProjectName == "" {
		return http.StatusBadRequest, "Project name not null", fmt.Errorf("Project name is wrong")
	}

	if request.Detail == "" {
		return http.StatusBadRequest, "Detail not null", fmt.Errorf("Detail is wrong")
	}

	return http.StatusOK, "", nil
}
