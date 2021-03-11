package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// JobRequest struct
//swagger:model JobRequest
type JobRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Department  string `json:"department"`
	PhoneNumber string `json:"phone_number"`
	CvMessage   string `json:"cv_message"`
}

// JobResponse struct
//swagger:response JobResponse
type JobResponse struct {
	ID          int    `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Department  string `json:"department"`
	PhoneNumber string `json:"phone_number"`
	CvMessage   string `json:"cv_message"`
}

// AddJobApplications adds a new job to database and creates a json response of the data
func (app *App) AddJobApplications(writer http.ResponseWriter, req *http.Request) {
	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusBadRequest, err, "Failed to read request body")
		return
	}

	var request JobRequest
	err = json.Unmarshal(reqBody, &request)
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusInternalServerError, err, "Failed to convert the input to json")
	}

	status, message, errValidate := request.ValidateJob()
	if errValidate != nil {
		app.RenderErrorResponse(writer, status, errValidate, message)
		return
	}

	sql := fmt.Sprint("INSERT INTO job_application(first_name,last_name,email,department,phone_number,cv_message,created) VALUES($1,$2,$3,$4,$5,$6,$7) returning uid;")
	var lastInsertId int
	err = app.db.QueryRow(
		sql,
		request.FirstName,
		request.LastName,
		request.Email,
		request.Department,
		request.PhoneNumber,
		request.CvMessage,
		time.Now()).Scan(&lastInsertId)

	if err != nil {
		app.RenderErrorResponse(writer, http.StatusInternalServerError, err, "Failed to write job")
		return
	}

	app.RenderJson(
		writer,
		http.StatusOK,
		JobResponse{
			FirstName:   request.FirstName,
			LastName:    request.LastName,
			Email:       request.Email,
			Department:  request.Department,
			PhoneNumber: request.PhoneNumber,
			CvMessage:   request.CvMessage,
			ID:          lastInsertId,
		},
	)

}

// GetJobApplications gets all job applications from database with id and creates a json response of the data
func (app *App) GetJobApplications(writer http.ResponseWriter, req *http.Request) {
	sql := fmt.Sprintf("SELECT uid,first_name,last_name,email,department,phone_number,cv_message FROM job_application")
	rows, err := app.db.Query(sql)
	if err != nil {
		ex := ErrorResponse{Status: http.StatusInternalServerError, Error: err, Message: "Failed get job application"}
		app.RenderError(writer, ex)
		return
	}

	var jobResponses []JobResponse
	for rows.Next() {
		resp := JobResponse{}
		rows.Scan(&resp.ID, &resp.FirstName, &resp.LastName, &resp.Email, &resp.Department, &resp.PhoneNumber, &resp.CvMessage)
		jobResponses = append(jobResponses, resp)
	}
	app.RenderJson(writer, http.StatusOK, jobResponses)
}

// FindJobApplicationByID finds job applications from database with id and creates a json response of the data
func (app *App) FindJobApplicationByID(writer http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	sql := fmt.Sprint("SELECT uid,first_name,last_name,email,department,phone_number,cv_message FROM job_application WHERE uid=$1")

	rows, err := app.db.Query(sql, params["id"])
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusInternalServerError, err, "Failed find job application")
		return
	}

	resp := JobResponse{}
	rows.Next()
	err = rows.Scan(&resp.ID, &resp.FirstName, &resp.LastName, &resp.Email, &resp.Department, &resp.PhoneNumber, &resp.CvMessage)
	if err != nil || resp.ID == 0 {
		app.RenderErrorResponse(writer, http.StatusInternalServerError, err, "Failed to fetch job application")
		return
	}

	app.RenderJson(writer, http.StatusOK, resp)
}

// DeleteJob deletes job application from database with id and creates a json response of the data
func (app *App) DeleteJob(writer http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	sql := fmt.Sprint("DELETE FROM job_application WHERE uid=$1")

	rows, err := app.db.Query(sql, params["id"])
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusInternalServerError, err, "Failed delete job application")
		return
	}
	rows.Next()
	app.RenderJson(writer, http.StatusOK, nil)

}

// ValidateJob validates request
func (request *JobRequest) ValidateJob() (int, string, error) {

	if request.FirstName == "" {
		return http.StatusBadRequest, "First name not null", fmt.Errorf("First name is wrong")
	}
	if request.LastName == "" {
		return http.StatusBadRequest, "Lastname name not null", fmt.Errorf("Last name is wrong")
	}
	if request.Email == "" {
		return http.StatusBadRequest, "Email not null", fmt.Errorf("Email is wrong")
	}

	if request.PhoneNumber == "" {
		return http.StatusBadRequest, "Phone number not null", fmt.Errorf("Phone number is wrong")
	}

	if request.Department == "" {
		return http.StatusBadRequest, "Department name not null", fmt.Errorf("Department name is wrong")
	}

	return http.StatusOK, "", nil
}
