package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"golang.org/x/net/html/charset"
	"log"
)

type RealType struct {
	XMLName    xml.Name `xml:"Real"`
	StartValue float64  `xml:"startValue"`
}
type IntegerType struct {
	XMLName    xml.Name `xml:"Integer"`
	StartValue int      `xml:"startValue"`
}
type BooleanType struct {
	XMLName    xml.Name `xml:"Boolean"`
	StartValue bool     `xml:"startValue"`
}
type StringType struct {
	XMLName    xml.Name `xml:"String"`
	StartValue string   `xml:"startValue"`
}

type ScalarVariable struct {
	XMLName        xml.Name    `xml:"ScalarVariable"`
	Name           string      `xml:"name,attr"`
	ValueReference uint32      `xml:"valueReference,attr"`
	Causality      string      `xml:"causality,attr"`
	RealType       RealType    `xml:"Real"`
	IntegerType    IntegerType `xml:"Integer"`
	BooleanType    BooleanType `xml:"Boolean"`
	StringType     StringType  `xml:"String"`
}

type ModelVariables struct {
	ScalarVariables []ScalarVariable `xml:"ScalarVariable"`
}

type ModelDescription struct {
	XMLName        xml.Name       `xml:"fmiModelDescription"`
	FmiVersion     string         `xml:"fmiVersion,attr"`
	ModelName      string         `xml:"modelName,attr"`
	ModelVariables ModelVariables `xml:"ModelVariables"`
}

func ReadModelDescription(fmuPath string) {

	// Open a zip archive for reading.
	reader, err := zip.OpenReader(fmuPath)
	if err != nil {
		log.Fatal(`ERROR:`, err)
	}
	defer reader.Close()

	var modelDescription ModelDescription
	for _, file := range reader.File {
		// check if the file matches the name for application portfolio xml
		if file.Name == "modelDescription.xml" {
			rc, err := file.Open()
			if err != nil {
				log.Fatal(`ERROR:`, err)
			}

			// Unmarshal bytes
			decoder := xml.NewDecoder(rc)
			decoder.CharsetReader = charset.NewReaderLabel
			err = decoder.Decode(&modelDescription)

			if err != nil {
				log.Fatal(`ERROR:`, err)
			}

			rc.Close()

			fmt.Println("We have this many variables:", len(modelDescription.ModelVariables.ScalarVariables))
			fmt.Println("We have model name: " + modelDescription.ModelName)
			fmt.Println("We have fmi version: " + modelDescription.FmiVersion)

			for i := 0; i < len(modelDescription.ModelVariables.ScalarVariables); i++ {
				variable := modelDescription.ModelVariables.ScalarVariables[i]
				fmt.Println("Variable Name: " + variable.Name)
				fmt.Println("Variable Causality: " + variable.Causality)
				fmt.Println("Variable ValueReference: ", variable.ValueReference)
				fmt.Println("Real?", variable.RealType.XMLName.Local)
				fmt.Println("Integer?", variable.IntegerType.XMLName.Local)
				fmt.Println("Boolean?", variable.BooleanType.XMLName.Local)
				fmt.Println("String?", variable.StringType.XMLName.Local)
			}
		}
	}
	return
}
