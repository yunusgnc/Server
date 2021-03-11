package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/sirupsen/logrus"
)

type Node struct {
	Name     string
	Type     string
	Children []*Node
}

// ReadModels reads vehicle models from json file
func ReadModels(modelsFile string) (*[]Node, error) {

	logrus.WithFields(logrus.Fields{
		"modelsFile": modelsFile,
	}).Info("Reading models file")

	content, err := ioutil.ReadFile(modelsFile)
	if err != nil {
		return nil, err
	}

	var nodes []Node
	err = json.Unmarshal(content, &nodes)
	if err != nil {
		return nil, err
	}
	return &nodes, nil
}

func main() {
	model, err := ReadModels("./models/models.json")
	if err != nil {
		log.Fatalf("Failed to read models file: %+v", err)
	}
	logrus.Debug(model)
}
