package main 

import (
	"strconv"
	"encoding/json"
	"io"
	"os"
	"github.com/xuri/excelize/v2"
)

type Table struct {
	file        *excelize.File
	rowsCounter  int
	path         string
}

func NewTable(path string) (t *Table, err error) {
	t = &Table{}
	t.path = path
	t.file, err = excelize.OpenFile(t.path)
	if err == nil {
		var rows [][]string
		rows, err = t.file.GetRows("Links")
		if err == nil {
			t.rowsCounter = len(rows)
		}		
	}

	return 
}

func (t *Table) Close() (err error) {
	err = t.file.Close()
	return
}

func (t *Table) AddRecord(firstname, secondname, username, link string) (err error) {
	err = t.file.SetSheetRow(
		"Links",
		"A" + strconv.Itoa(t.rowsCounter + 1), 
		&[]interface{}{
			firstname,
			secondname,
			username,
			link,
		},
	)
	if err == nil {
		err = t.Save()
		if err == nil {
			t.rowsCounter++
		}
	}
	return
}

func (t *Table) Save() error {
	err := t.file.SaveAs(t.path)
	return err
}

type Config struct {
	Token     string `json:"token"`
	Channel   int64  `json:"channel"`
	Admins  []int64  `json:"admins"`
}

func OpenConfig(filePath string) (c *Config, err error) {
	jsonFile, err := os.Open(filePath)

	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()
	
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(byteValue, &c)

	return c, err
}