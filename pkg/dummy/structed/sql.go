package structed

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/sirupsen/logrus"
)

type sqlData struct {
	DBName        string
	Books         []books
	Members       []members
	BorrowedBooks []borrowedBooks
}

type books struct {
	BookID          uint8  `fake: "{uint8}"`
	Title           string `fake: "{booktitle}"`
	Author          string `fake: "{bookauthor}"`
	PublicationYear uint8  `fake: "{year}"`
	Publisher       string `fake: "{bookGenre}"`
	Quantity        uint8  `fake: "{number:1,100}"`
}

type members struct {
	MemberID   uint8     `fake: "{uint8}"`
	Name       string    `fake: "{name}"`
	Address    string    `fake: "{country}"`
	PhoneNo    string    `fake: "{phone}"`
	Email      string    `fake: "{email}"`
	JoinedDate time.Time `fake: "{date}"`
	ExpiryDate time.Time `fake: "{date}"`
	IsActive   bool
}

type borrowedBooks struct {
	BorrowID     uint8      `fake: "{uint8}"`
	MemberID     uint8      `fake: "{uint8}"`
	BookID       uint8      `fake: "{uint8}"`
	BorrowedDate time.Time  `fake: "{date}"`
	DueDate      time.Time  `fake: "{date}"`
	ReturnedDate *time.Time `fake: "{date}"`
	FinePaid     uint8      `fake: "{number:0,100}"`
}

const createSql string = `
CREATE DATABASE IF NOT EXISTS {{ .DBName }};

USE {{ .DBName }};

DROP TABLE IF EXISTS Books;

CREATE TABLE Books (
	BookID INT AUTO_INCREMENT,
	Title VARCHAR(255),
	Author VARCHAR(255),
	PublicationYear INT,
	Publisher VARCHAR(255),
	Quantity INT,
	PRIMARY KEY (BookID)
);
{{range .Books}}
INSERT INTO Books (Title, Author, PublicationYear, Publisher, Quantity) VALUES ("{{.Title}}", "{{.Author}}", {{.PublicationYear}}, "{{.Publisher}}", {{.Quantity}});
{{end}}

DROP TABLE IF EXISTS Members;

CREATE TABLE Members (
	MemberID INT AUTO_INCREMENT,
	Name VARCHAR(255),
	Address VARCHAR(255),
	PhoneNo VARCHAR(20),
	Email VARCHAR(50),
	JoinedDate DATE,
	ExpiryDate DATE,
	IsActive BOOLEAN DEFAULT 1, 
	PRIMARY KEY (MemberID)
);
{{range .Members}}
INSERT INTO Members (Name, Address, PhoneNo ,Email ,JoinedDate ,ExpiryDate ,IsActive ) VALUES ("{{.Name}}", "{{.Address}}", "{{.PhoneNo}}" ,"{{.Email}}" ,"{{formatTime .JoinedDate}}" ,"{{formatTime .ExpiryDate}}" , {{if .IsActive }}1 {{else}}0 {{end}});
{{end}}

DROP TABLE IF EXISTS BorrowedBooks;

CREATE TABLE BorrowedBooks (
	BorrowID INT AUTO_INCREMENT, 
	MemberID INT, 
	BookID INT, 
	BorrowedDate DATE,  
	DueDate DATE,  
	ReturnedDate DATE NULL DEFAULT NULL,
	FinePaid DECIMAL(5,2) DEFAULT 0.00,
	PRIMARY KEY (BorrowID)  
);
{{range .BorrowedBooks}}
INSERT INTO BorrowedBooks (MemberID, BookID, BorrowedDate, DueDate, ReturnedDate, FinePaid) VALUES ({{.MemberID}}, {{.BookID}}, "{{formatTime .BorrowedDate}}", "{{formatTime .DueDate}}", "{{formatTime .ReturnedDate}}", {{.FinePaid}});
{{end}}
`

// SQL generation function using gofakeit
//
// CapacitySize is in GB and generates sql files
// within the entered dummyDir path.
func GenerateRandomSQL(dummyDir string, capacitySize int) error {
	dummyDir = filepath.Join(dummyDir, "sql")
	if err := utils.IsDir(dummyDir); err != nil {
		logrus.Errorf("IsDir function error : %v", err)
		return err
	}

	size := capacitySize * 1000

	countNum := make(chan int, size)
	resultChan := make(chan error, size)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			randomSQLWorker(countNum, dummyDir, resultChan)
		}()
	}

	for i := 0; i < size; i++ {
		countNum <- i
	}
	close(countNum)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for ret := range resultChan {
		if ret != nil {
			logrus.WithFields(logrus.Fields{"jobName": "sql create"}).Errorf("result error : %v", ret)
			return ret
		}
	}

	return nil
}

// sql worker
func randomSQLWorker(countNum chan int, dirPath string, resultChan chan<- error) {
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
	}

	tmpl, err := template.New("mysqlData").Funcs(funcMap).Parse(createSql)
	if err != nil {
		resultChan <- err
	}

	for num := range countNum {

		data := sqlData{}
		data.DBName = fmt.Sprintf("LibraryManagement_%d", num)

		for i := 0; i < 2350; i++ {
			book := books{}
			gofakeit.Struct(&book)
			data.Books = append(data.Books, book)

			members := members{}
			gofakeit.Struct(&members)
			data.Members = append(data.Members, members)

			borrowedBooks := borrowedBooks{}
			gofakeit.Struct(&borrowedBooks)
			data.BorrowedBooks = append(data.BorrowedBooks, borrowedBooks)
		}

		var buffer bytes.Buffer
		if err := tmpl.Execute(&buffer, data); err != nil {
			resultChan <- err
			continue
		}

		file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("LibraryManagement_%d.sql", num)))
		if err != nil {
			resultChan <- err
			continue
		}
		defer file.Close()

		if _, err := io.Copy(file, &buffer); err != nil {
			resultChan <- err
			continue
		}

		logrus.Infof("Creation success: %v", file.Name())
		file.Close()

		resultChan <- nil
	}
}
