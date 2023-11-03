package semistructed

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/cloud-barista/cm-data-mold/pkg/utils"
)

// xml generation function using gofakeit
//
// CapacitySize is in GB and generates xml files
// within the entered dummyDir path.
func GenerateRandomXML(dummyDir string, capacitySize int) error {
	dummyDir = filepath.Join(dummyDir, "xml")
	if err := utils.IsDir(dummyDir); err != nil {
		return err
	}

	size := capacitySize * 10
	countNum := make(chan int, size)
	resultChan := make(chan error, size)

	var wg sync.WaitGroup
	for i := 0; i < capacitySize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			randomXMLWorker(countNum, dummyDir, resultChan)
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
			return ret
		}
	}

	return nil
}

// xml worker
func randomXMLWorker(countNum chan int, dirPath string, resultChan chan<- error) {
	for cnt := range countNum {
		gofakeit.Seed(0)

		dataGenerators := []func(int, string, int) error{
			generateXMLBook,
			generateXMLCar,
			generateXMLAddress,
			generateXMLCreditCard,
			generateXMLJob,
			generateXMLMovie,
			generateXMLPerson,
		}

		for _, generator := range dataGenerators {
			resultChan <- generator(cnt, dirPath, 49350)
		}
	}
}

// generate book.xml
func generateXMLBook(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("book_%d.xml", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	type Book struct {
		XMLName xml.Name             `xml:"catalog"`
		Books   []*gofakeit.BookInfo `xml:"book"`
	}

	book := Book{}

	for i := 0; i < count; i++ {
		b := gofakeit.Book()
		if err := gofakeit.Struct(b); err != nil {
			return err
		}
		book.Books = append(book.Books, b)
	}

	data, err := xml.MarshalIndent(book, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

// generate car.xml
func generateXMLCar(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("car_%d.xml", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	type Car struct {
		XMLName xml.Name            `xml:"catalog"`
		Cars    []*gofakeit.CarInfo `xml:"car"`
	}

	car := Car{}

	for i := 0; i < count; i++ {
		c := gofakeit.Car()
		if err := gofakeit.Struct(c); err != nil {
			return err
		}
		car.Cars = append(car.Cars, c)
	}

	data, err := xml.MarshalIndent(car, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

// generate address.xml
func generateXMLAddress(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("address_%d.xml", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	type Address struct {
		XMLName   xml.Name                `xml:"catalog"`
		Addresses []*gofakeit.AddressInfo `xml:"address"`
	}

	address := Address{}

	for i := 0; i < count; i++ {
		a := gofakeit.Address()
		if err := gofakeit.Struct(a); err != nil {
			return err
		}
		address.Addresses = append(address.Addresses, a)
	}

	data, err := xml.MarshalIndent(address, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

// generate creditcard.xml
func generateXMLCreditCard(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("creditcard_%d.xml", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	type CreditCard struct {
		XMLName     xml.Name                   `xml:"catalog"`
		CreditCards []*gofakeit.CreditCardInfo `xml:"creditcard"`
	}

	creditCard := CreditCard{}

	for i := 0; i < count; i++ {
		c := gofakeit.CreditCard()
		if err := gofakeit.Struct(c); err != nil {
			return err
		}
		creditCard.CreditCards = append(creditCard.CreditCards, c)
	}

	data, err := xml.MarshalIndent(creditCard, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

// generate job.xml
func generateXMLJob(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("job_%d.xml", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	type Job struct {
		XMLName xml.Name            `xml:"catalog"`
		Jobs    []*gofakeit.JobInfo `xml:"job"`
	}

	job := Job{}

	for i := 0; i < count; i++ {
		j := gofakeit.Job()
		if err := gofakeit.Struct(j); err != nil {
			return err
		}
		job.Jobs = append(job.Jobs, j)
	}

	data, err := xml.MarshalIndent(job, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

// generate movie.xml
func generateXMLMovie(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("movie_%d.xml", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	type Movie struct {
		XMLName xml.Name              `xml:"catalog"`
		Movies  []*gofakeit.MovieInfo `xml:"movie"`
	}

	movie := Movie{}

	for i := 0; i < count; i++ {
		m := gofakeit.Movie()
		if err := gofakeit.Struct(m); err != nil {
			return err
		}
		movie.Movies = append(movie.Movies, m)
	}

	data, err := xml.MarshalIndent(movie, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

// generate person.xml
func generateXMLPerson(cnt int, dirPath string, count int) error {
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("person_%d.xml", cnt)))
	if err != nil {
		return err
	}
	defer file.Close()

	type Person struct {
		XMLName xml.Name               `xml:"catalog"`
		Persons []*gofakeit.PersonInfo `xml:"person"`
	}

	person := Person{}

	for i := 0; i < count; i++ {
		p := gofakeit.Person()
		if err := gofakeit.Struct(p); err != nil {
			return err
		}
		person.Persons = append(person.Persons, p)
	}

	data, err := xml.MarshalIndent(person, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}
