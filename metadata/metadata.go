package metadata

import (
	"archive/zip"
	"cse-server-go/structs"
	"encoding/xml"
	"golang.org/x/net/html/charset"
	"log"
)

type RealType struct {
	XMLName    xml.Name `xml:"Real"`
	StartValue float64  `xml:"start"`
}
type IntegerType struct {
	XMLName    xml.Name `xml:"Integer"`
	StartValue int      `xml:"start"`
}
type BooleanType struct {
	XMLName    xml.Name `xml:"Boolean"`
	StartValue bool     `xml:"start"`
}
type StringType struct {
	XMLName    xml.Name `xml:"String"`
	StartValue string   `xml:"start"`
}

type ScalarVariable struct {
	XMLName        xml.Name    `xml:"ScalarVariable"`
	Name           string      `xml:"name,attr"`
	ValueReference int         `xml:"valueReference,attr"`
	Causality      string      `xml:"causality,attr"`
	Variability    string      `xml:"variability,attr"`
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

func getValueType(variable ScalarVariable) string {
	if variable.RealType.XMLName.Local == "Real" {
		return "Real"
	}
	if variable.IntegerType.XMLName.Local == "Integer" {
		return "Integer"
	}
	if variable.BooleanType.XMLName.Local == "Boolean" {
		return "Boolean"
	}
	if variable.StringType.XMLName.Local == "String" {
		return "String"
	}
	return ""
}

func ReadModelDescription(fmuPath string) (fmu structs.FMU) {

	// Open a zip archive for reading.
	reader, err := zip.OpenReader(fmuPath)
	if err != nil {
		log.Fatal(`ERROR:`, err)
	}
	defer reader.Close()

	var modelDescription ModelDescription
	for _, file := range reader.File {
		if file.Name == "modelDescription.xml" {
			rc, err := file.Open()
			if err != nil {
				log.Fatal(`ERROR:`, err)
			}
			decoder := xml.NewDecoder(rc)
			decoder.CharsetReader = charset.NewReaderLabel
			err = decoder.Decode(&modelDescription)
			if err != nil {
				log.Fatal(`ERROR:`, err)
			}
			rc.Close()

			var variables []structs.Variable
			for _, scalarVariable := range modelDescription.ModelVariables.ScalarVariables {
				variables = append(variables, structs.Variable{
					Name:           scalarVariable.Name,
					ValueReference: scalarVariable.ValueReference,
					Causality:      scalarVariable.Causality,
					Variability:    scalarVariable.Variability,
					Type:           getValueType(scalarVariable),
				})
			}
			fmu.Variables = variables
			fmu.Name = modelDescription.ModelName
			return fmu
		}
	}
	return
}
