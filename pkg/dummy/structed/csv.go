package structed

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/sirupsen/logrus"
)

// CSV generation function using gofakeit
//
// CapacitySize is in GB and generates csv files
// within the entered dummyDir path.
func GenerateRandomCSV(dummyDir string, capacitySize int) error {
	dummyDir = filepath.Join(dummyDir, "csv")
	if err := utils.IsDir(dummyDir); err != nil {
		logrus.WithFields(logrus.Fields{"jobName": "csv create"}).Errorf("IsDir function error : %v", err)
		return err
	}

	countNum := make(chan int, capacitySize*10)
	resultChan := make(chan error, capacitySize*10)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			randomCSVWorker(countNum, dummyDir, resultChan)
		}()
	}

	for i := 0; i < capacitySize*10; i++ {
		countNum <- i
	}
	close(countNum)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for ret := range resultChan {
		if ret != nil {
			logrus.WithFields(logrus.Fields{"jobName": "csv create"}).Errorf("return error : %v", ret)
			return ret
		}
	}

	return nil
}

// csv worker
func randomCSVWorker(countNum chan int, dirPath string, resultChan chan<- error) {
	for cnt := range countNum {
		gofakeit.Seed(0)
		dataGenerators := []func(int, string, int) error{
			generateCSVBook,
			generateCSVCar,
			generateCSVAddress,
			generateCSVCreditCard,
			generateCSVJob,
			generateCSVMovie,
			generateCSVPerson,
		}

		for _, generator := range dataGenerators {
			resultChan <- generator(cnt, dirPath, 121000)
		}
	}
}

// generate book.csv
func generateCSVBook(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("book_%d.csv", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	type Book struct {
		Books []*gofakeit.BookInfo `csv:"book"`
	}

	book := Book{}

	for i := 0; i < count; i++ {
		b := gofakeit.Book()
		if err := gofakeit.Struct(b); err != nil {
			return err
		}
		book.Books = append(book.Books, b)
	}

	csvWriter := csv.NewWriter(file)

	err = csvWriter.Write([]string{"Title", "Author", "Genre"})
	if err != nil {
		return err
	}

	for _, b := range book.Books {
		record := []string{b.Title, b.Author, b.Genre}
		err := csvWriter.Write(record)
		if err != nil {
			return err
		}
	}

	logrus.WithFields(logrus.Fields{"jobName": "csv create"}).Infof("Creation success: %v", file.Name())

	csvWriter.Flush()
	return csvWriter.Error()
}

// generate car.csv
func generateCSVCar(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("car_%d.csv", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	type Car struct {
		Cars []*gofakeit.CarInfo `csv:"car"`
	}

	car := Car{}

	for i := 0; i < count; i++ {
		c := gofakeit.Car()
		if err := gofakeit.Struct(c); err != nil {
			return err
		}
		car.Cars = append(car.Cars, c)
	}

	csvWriter := csv.NewWriter(file)

	err = csvWriter.Write([]string{"Type", "Fuel", "Transmission", "Brand", "Model", "Year"})
	if err != nil {
		return err
	}

	for _, c := range car.Cars {
		record := []string{c.Type, c.Fuel, c.Transmission, c.Brand, c.Model, fmt.Sprint(c.Year)}
		err := csvWriter.Write(record)
		if err != nil {
			return err
		}
	}
	logrus.WithFields(logrus.Fields{"jobName": "csv create"}).Infof("Creation success: %v", file.Name())
	csvWriter.Flush()
	return csvWriter.Error()
}

// generate address.csv
func generateCSVAddress(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("address_%d.csv", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	type Address struct {
		Addresses []*gofakeit.AddressInfo `csv:"address"`
	}

	address := Address{}

	for i := 0; i < count; i++ {
		a := gofakeit.Address()
		if err := gofakeit.Struct(a); err != nil {
			return err
		}
		address.Addresses = append(address.Addresses, a)
	}

	csvWriter := csv.NewWriter(file)

	err = csvWriter.Write([]string{"Street", "City", "State", "Zip", "Country", "Latitude", "Longitude"})
	if err != nil {
		return err
	}

	for _, a := range address.Addresses {
		record := []string{a.Street, a.City, a.State, a.Zip, a.Country, strconv.FormatFloat(a.Latitude, 'f', -1, 64), strconv.FormatFloat(a.Longitude, 'f', -1, 64)}
		err := csvWriter.Write(record)
		if err != nil {
			return err
		}
	}
	logrus.WithFields(logrus.Fields{"jobName": "csv create"}).Infof("Creation success: %v", file.Name())

	csvWriter.Flush()
	return csvWriter.Error()
}

// generate creditcard.csv
func generateCSVCreditCard(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("creditcard_%d.csv", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	type CreditCard struct {
		CreditCards []*gofakeit.CreditCardInfo `csv:"creditcard"`
	}

	creditCard := CreditCard{}

	for i := 0; i < count; i++ {
		c := gofakeit.CreditCard()
		if err := gofakeit.Struct(c); err != nil {
			return err
		}
		creditCard.CreditCards = append(creditCard.CreditCards, c)
	}

	csvWriter := csv.NewWriter(file)

	err = csvWriter.Write([]string{"Type", "Number", "Exp", "Cvv"})
	if err != nil {
		return err
	}

	for _, c := range creditCard.CreditCards {
		record := []string{c.Type, c.Number, c.Exp, c.Cvv}
		err := csvWriter.Write(record)
		if err != nil {
			return err
		}
	}
	logrus.WithFields(logrus.Fields{"jobName": "csv create"}).Infof("Creation success: %v", file.Name())

	csvWriter.Flush()
	return csvWriter.Error()
}

// generate job.csv
func generateCSVJob(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("job_%d.csv", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	type Job struct {
		Jobs []*gofakeit.JobInfo `csv:"job"`
	}

	job := Job{}

	for i := 0; i < count; i++ {
		j := gofakeit.Job()
		if err := gofakeit.Struct(j); err != nil {
			return err
		}
		job.Jobs = append(job.Jobs, j)
	}

	csvWriter := csv.NewWriter(file)

	err = csvWriter.Write([]string{"Company", "Title", "Descriptor", "Level"})
	if err != nil {
		return err
	}

	for _, j := range job.Jobs {
		record := []string{j.Company, j.Title, j.Descriptor, j.Level}
		err := csvWriter.Write(record)
		if err != nil {
			return err
		}
	}
	logrus.WithFields(logrus.Fields{"jobName": "csv create"}).Infof("Creation success: %v", file.Name())

	csvWriter.Flush()
	return csvWriter.Error()
}

// generate movie.csv
func generateCSVMovie(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("movie_%d.csv", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	type Movie struct {
		Movies []*gofakeit.MovieInfo `csv:"movie"`
	}

	movie := Movie{}

	for i := 0; i < count; i++ {
		m := gofakeit.Movie()
		if err := gofakeit.Struct(m); err != nil {
			return err
		}
		movie.Movies = append(movie.Movies, m)
	}

	csvWriter := csv.NewWriter(file)

	err = csvWriter.Write([]string{"Name", "Genre"})
	if err != nil {
		return err
	}

	for _, m := range movie.Movies {
		record := []string{m.Name, m.Genre}
		err := csvWriter.Write(record)
		if err != nil {
			return err
		}
	}
	logrus.WithFields(logrus.Fields{"jobName": "csv create"}).Infof("Creation success: %v", file.Name())

	csvWriter.Flush()
	return csvWriter.Error()
}

// generate person.csv
func generateCSVPerson(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("person_%d.csv", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	type Person struct {
		Persons []*gofakeit.PersonInfo `csv:"person"`
	}

	person := Person{}

	for i := 0; i < count; i++ {
		p := gofakeit.Person()
		if err := gofakeit.Struct(p); err != nil {
			return err
		}
		person.Persons = append(person.Persons, p)
	}

	csvWriter := csv.NewWriter(file)

	err = csvWriter.Write([]string{"FirstName", "LastName", "Gender", "SSN", "Image", "Hobby"})
	if err != nil {
		return err
	}

	for _, p := range person.Persons {
		record := []string{p.FirstName, p.LastName, p.Gender, p.SSN, p.Image, p.Hobby}
		err := csvWriter.Write(record)
		if err != nil {
			return err
		}
	}
	logrus.WithFields(logrus.Fields{"jobName": "csv create"}).Infof("Creation success: %v", file.Name())

	csvWriter.Flush()
	return csvWriter.Error()
}
