package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// RelationType struct
type RelationType struct {
	ID   int
	Name string
}

// Relation struct
type Relation struct {
	ID       int
	Name     string
	TypeID   int
	ParentID int
	Path     string
}

// RelationRequest request struct
type RelationRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type_name"`
	ParentID int    `json:"parent_id"`
}

// RelationResponse response struct
type RelationResponse struct {
	ID       []uint8    `json:"id"`
	Name     string `json:"name"`
	TypeID   int64  `json:"type_id"`
	ParentID int    `json:"parent_id"`
}

// CreateRelationType creating new relation type
func (app *App) CreateRelationType(relationType string) (int64, *ErrorResponse) {
	var typeID int64
	sql := fmt.Sprint("SELECT id FROM relation_type WHERE name=$1")
	rows, err := app.db.Query(sql, relationType)
	if err != nil {
		return 0, &ErrorResponse{Status: http.StatusInternalServerError, Error: err, Message: "Failed to fetch data from DB"}
	}
	if rows.Next() {
		err = rows.Scan(&typeID)
		if err != nil {
			return 0, &ErrorResponse{Status: http.StatusBadRequest, Error: err, Message: "Failed to save data"}
		}
		return typeID, nil
	}

	sql = fmt.Sprintf("INSERT INTO relation_type(name,created) VALUES($1,$2) returning id;")
	err = app.db.QueryRow(sql, relationType, time.Now()).Scan(&typeID)
	if err != nil {
		return 0, &ErrorResponse{Status: http.StatusBadRequest, Error: err, Message: "Failed to save data"}
	}
	logrus.Warning("Type id not found. New type created :", relationType)
	return typeID, nil
}

// AddRelation adds relation to database and creates a json response of the data
func (app *App) AddRelation(writer http.ResponseWriter, req *http.Request) {
	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusBadRequest, err, "Failed to read request body")
		return
	}

	var request RelationRequest
	err = json.Unmarshal(reqBody, &request)
	if err != nil {
		app.RenderErrorResponse(writer, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}

	typeID, errRelation := app.CreateRelationType(request.Type)

	if errRelation != nil {
		app.RenderError(writer, *errRelation)
		return
	}

	var parent Relation

	if request.ParentID != 0 {
		sqlRelation := fmt.Sprintf("SELECT id,name,type_id,path_id FROM relation WHERE id = $1;")
		var rows *sql.Rows
		rows, err = app.db.Query(sqlRelation, request.ParentID)
		if err != nil {
			app.RenderErrorResponse(writer, http.StatusNotFound, err, "Sequence not found"+strconv.Itoa(request.ParentID))
			return
		}
		err = rows.Scan(&parent.ID, &parent.Name, &parent.TypeID, &parent.ParentID)
	}

	relationSQL, errorResponse := app.GetRelationSQL(writer, request.Name, typeID, parent, 0)

	if relationSQL != nil {
		app.RenderErrorResponse(writer, errorResponse.Status, errorResponse.Error, errorResponse.Message)
	}

	var path string
	var nextSequence int
	if relationSQL != nil {
		nextSequence, err = app.NextForSequence(writer, "relation_path_id_seq")

		if err != nil {
			app.RenderErrorResponse(writer, http.StatusInternalServerError, err, ""+strconv.Itoa(request.ParentID))
			return
		}

		if request.ParentID != 0 {
			path = parent.Path + "." + strconv.Itoa(nextSequence)
		}
	}

	sql := fmt.Sprintf("INSERT INTO relation(name,type_id,path,path_id,created) VALUES($1,$2,$3,$4,$5) returning id;")
	var lastInsertID []uint8 //UUID
	err = app.db.QueryRow(sql, request.Name, typeID, path, nextSequence, time.Now()).Scan(&lastInsertID)

	if err != nil {
		app.RenderErrorResponse(writer, http.StatusBadRequest, err, "Failed to save data")
		return
	}

	app.RenderJson(writer, http.StatusOK, RelationResponse{Name: request.Name, TypeID: typeID, ID: lastInsertID})
}

// GetRelationSQL fetches relation path if parent not exist adds its relation path.
func (app *App) GetRelationSQL(writer http.ResponseWriter, name string, typeID int64, parent Relation, level int) (*sql.Rows, ErrorResponse) {

	sqlVar := fmt.Sprintf("SELECT * FROM relation WHERE name=$1 AND type_id=$2")

	var rows *sql.Rows
	var err error
	if parent.ID != 0 {
		sqlVar += fmt.Sprintf(" AND path operator(public.<@) $3")
		rows, err = app.db.Query(sqlVar, name, typeID, parent.Path)
		if err != nil {
			return nil, ErrorResponse{Status: http.StatusInternalServerError, Error: err, Message: "Failed save path"}
		}
	} else {
		rows, err = app.db.Query(sqlVar, name, typeID)
		if err != nil {
			return nil, ErrorResponse{Status: http.StatusInternalServerError, Error: err, Message: "Failed save path"}
		}
	}

	return rows, ErrorResponse{}
}

//NextForSequence finds next sequence
func (app *App) NextForSequence(writer http.ResponseWriter, sequenceName string) (int, error) {
	sql := fmt.Sprintf("SELECT NEXTVAL($1) as id")

	row := app.db.QueryRow(sql, sequenceName)

	var nextValue int

	return nextValue, row.Scan(&nextValue)
}
