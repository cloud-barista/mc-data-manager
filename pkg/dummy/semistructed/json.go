package semistructed

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/sirupsen/logrus"
)

// Structures to be used to generate json dummy data
type bookInfo struct {
	BookID string `json:"book_id" fake:"{password:true,true,true,false,false,24}"`
	Title  string `json:"title" xml:"name" fake:"{booktitle}"`
	Author string `json:"author" xml:"author" fake:"{bookauthor}"`
	Genre  string `json:"genre" xml:"genre" fake:"{bookgenre}"`
}

// Structures to be used to generate json dummy data
type carInfo struct {
	CarID        string `json:"car_id" fake:"{password:true,true,true,false,false,24}"`
	Type         string `json:"type" xml:"type" fake:"{cartype}"`
	Fuel         string `json:"fuel" xml:"fuel" fake:"{carfueltype}"`
	Transmission string `json:"transmission" xml:"transmission" fake:"{cartransmissiontype}"`
	Brand        string `json:"brand" xml:"brand" fake:"{carmaker}"`
	Model        string `json:"model" xml:"model" fake:"{carmodel}"`
	Year         int    `json:"year" xml:"year" fake:"{year}"`
}

// Structures to be used to generate json dummy data
type addressInfo struct {
	AddrID     string `json:"addr_id" fake:"{password:true,true,true,false,false,24}"`
	CountryAbr string `json:"countryabr" xml:"countryabr" fake:"{countryabr}"`
	Street     string `json:"street" xml:"street" fake:"{street}"`
	City       string `json:"city" xml:"city" fake:"{city}"`
	State      string `json:"state" xml:"state" fake:"{state}"`
	Zip        string `json:"zip" xml:"zip" fake:"{zip}"`
	Country    string `json:"country" xml:"country" fake:"{country}"`
	Latitude   int    `json:"latitude" xml:"latitude" fake:"{number:-90,90}"`
	Longitude  int    `json:"longitude" xml:"longitude" fake:"{number:-180,180}"`
}

// Structures to be used to generate json dummy data
type movieInfo struct {
	MovID string `json:"mov_id" fake:"{password:true,true,true,false,false,24}"`
	Name  string `json:"name" xml:"name" fake:"{moviename}"`
	Genre string `json:"genre" xml:"genre" fake:"{moviegenre}"`
}

// Structures to be used to generate json dummy data
type creditCardInfo struct {
	CardID string `json:"card_id" fake:"{password:true,true,true,false,false,24}"`
	Type   string `json:"type" xml:"type" fake:"{creditcardtype}"`
	Number string `json:"number" xml:"number" fake:"{creditcardnumber}"`
	Exp    string `json:"exp" xml:"exp" fake:"{creditcardexp}"`
	Cvv    string `json:"cvv" xml:"cvv" fake:"{creditcardcvv}"`
}

// Structures to be used to generate json dummy data
type jobInfo struct {
	JobID      string `json:"job_id" fake:"{password:true,true,true,false,false,24}"`
	Company    string `json:"company" xml:"company" fake:"{company}"`
	Title      string `json:"title" xml:"title" fake:"{jobtitle}"`
	Descriptor string `json:"descriptor" xml:"descriptor" fake:"{jobdescriptor}"`
	Level      string `json:"level" xml:"level" fake:"{joblevel}"`
}

// Structures to be used to generate json dummy data
type personInfo struct {
	PersonID   string                `json:"person_id" fake:"{password:true,true,true,false,false,24}"`
	Name       string                `json:"name" xml:"name" fake:"{name}"`
	FirstName  string                `json:"first_name" xml:"first_name" fake:"{firstname}"`
	LastName   string                `json:"last_name" xml:"last_name" fake:"{lastname}"`
	Gender     string                `json:"gender" xml:"gender" fake:"{gender}"`
	SSN        string                `json:"ssn" xml:"ssn" fake:"{ssn}"`
	Hobby      string                `json:"hobby" xml:"hobby" fake:"{hobby}"`
	Job        *jobInfo              `json:"job" xml:"job"`
	Address    *addressInfo          `json:"address" xml:"address"`
	Contact    *gofakeit.ContactInfo `json:"contact" xml:"contact"`
	CreditCard *creditCardInfo       `json:"credit_card" xml:"credit_card"`
}

// json generation function using gofakeit
//
// CapacitySize is in GB and generates json files
// within the entered dummyDir path.
func GenerateRandomJSON(dummyDir string, capacitySize int) error {
	dummyDir = filepath.Join(dummyDir, "json")
	if err := utils.IsDir(dummyDir); err != nil {
		logrus.Errorf("IsDir function error : %v", err)
		return err
	}

	countNum := make(chan int, capacitySize*1000)
	resultChan := make(chan error, capacitySize*1000)

	var wg sync.WaitGroup
	for i := 0; i < capacitySize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			randomJsonWorker(countNum, dummyDir, resultChan)
		}()
	}

	for i := 0; i < capacitySize*1000; i++ {
		countNum <- i
	}
	close(countNum)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for ret := range resultChan {
		if ret != nil {
			logrus.Errorf("return error : %v", ret)
			return ret
		}
	}

	return nil
}

// json worker
func randomJsonWorker(countNum chan int, dirPath string, resultChan chan<- error) {
	for cnt := range countNum {
		gofakeit.Seed(0)
		dataGenerators := []func(int, string, int) error{
			generateJSONBook,
			generateJSONCar,
			generateJSONAddress,
			generateJSONCreditCard,
			generateJSONJob,
			generateJSONMovie,
			generateJSONPerson,
		}

		for _, generator := range dataGenerators {
			resultChan <- generator(cnt, dirPath, 475)
		}
	}
}

// generate book.json
func generateJSONBook(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("book_%d.json", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	Books := []*bookInfo{}

	for i := 0; i < count; i++ {
		b := &bookInfo{}
		if err := gofakeit.Struct(b); err != nil {
			return err
		}
		Books = append(Books, b)
	}

	data, err := json.MarshalIndent(Books, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err == nil {
		logrus.Infof("Creation success: %v", file.Name())
	}
	return err
}

// generate car.json
func generateJSONCar(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("car_%d.json", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	Cars := []*carInfo{}

	for i := 0; i < count; i++ {
		c := &carInfo{}
		if err := gofakeit.Struct(c); err != nil {
			return err
		}
		Cars = append(Cars, c)
	}

	data, err := json.MarshalIndent(Cars, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err == nil {
		logrus.Infof("Creation success: %v", file.Name())
	}
	return err
}

// generate address.json
func generateJSONAddress(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("address_%d.json", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	Addresses := []*addressInfo{}

	for i := 0; i < count; i++ {
		a := &addressInfo{}
		if err := gofakeit.Struct(a); err != nil {
			return err
		}
		Addresses = append(Addresses, a)
	}

	data, err := json.MarshalIndent(Addresses, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err == nil {
		logrus.Infof("Creation success: %v", file.Name())
	}
	return err
}

// generate creditcard.json
func generateJSONCreditCard(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("creditcard_%d.json", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	CreditCards := []*creditCardInfo{}

	for i := 0; i < count; i++ {
		c := &creditCardInfo{}
		if err := gofakeit.Struct(c); err != nil {
			return err
		}
		CreditCards = append(CreditCards, c)
	}

	data, err := json.MarshalIndent(CreditCards, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err == nil {
		logrus.Infof("Creation success: %v", file.Name())
	}
	return err
}

// generate job.json
func generateJSONJob(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("job_%d.json", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	Jobs := []*jobInfo{}

	for i := 0; i < count; i++ {
		j := &jobInfo{}
		if err := gofakeit.Struct(j); err != nil {
			return err
		}
		Jobs = append(Jobs, j)
	}

	data, err := json.MarshalIndent(Jobs, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err == nil {
		logrus.Infof("Creation success: %v", file.Name())
	}
	return err
}

// generate movie.json
func generateJSONMovie(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("movie_%d.json", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	Movies := []*movieInfo{}

	for i := 0; i < count; i++ {
		m := &movieInfo{}
		if err := gofakeit.Struct(m); err != nil {
			return err
		}
		Movies = append(Movies, m)
	}

	data, err := json.MarshalIndent(Movies, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err == nil {
		logrus.Infof("Creation success: %v", file.Name())
	}
	return err
}

// generate person.json
func generateJSONPerson(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("person_%d.json", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	Persons := []*personInfo{}

	for i := 0; i < count; i++ {
		p := &personInfo{}
		if err := gofakeit.Struct(p); err != nil {
			return err
		}
		Persons = append(Persons, p)
	}

	data, err := json.MarshalIndent(Persons, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err == nil {
		logrus.Infof("Creation success: %v", file.Name())
	}
	return err
}
