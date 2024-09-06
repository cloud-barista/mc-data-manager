/*
Copyright 2023 The Cloud-Barista Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package rdbc_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/rdbms/mysql"
	"github.com/cloud-barista/mc-data-manager/service/rdbc"
)

func TestMain(m *testing.M) {
	// aws example
	// 	RDBCInfo(models.AWS,"admin","datamoldPassword","127.0.0.1","3306")
	// gcp example
	// 	RDBCInfo(models.GCP,"root","datamoldPassword","127.0.0.1","3306")
	// ncp example
	// 	RDBCInfo(models.NCP,"root","datamoldPassword","127.0.0.1","3306")
	Srdbc := RDBCInfo("your-source-provider", "your-source-mysql-username", "your-source-mysql-password", "your-source-mysql-host", "your-source-mysql-port")
	Drdbc := RDBCInfo("your-target-provider", "your-target-mysql-username", "your-target-mysql-password", "your-target-mysql-host", "your-target-mysql-port")

	// Srdbc에 LibraryManagement.sql import
	if err := Srdbc.Put(testSQL); err != nil {
		panic(err)
	}

	// Srdbc의 LibraryManagement Export
	var dstSQL string
	if err := Srdbc.Get("LibraryManagement", &dstSQL); err != nil {
		panic(err)
	}

	// Srdbc db들을 Drdbc로 이전
	if err := Srdbc.Copy(Drdbc); err != nil {
		panic(err)
	}
}

func RDBCInfo(providerType models.Provider, username, password, host, port string) *rdbc.RDBController {
	sqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, host, port))
	if err != nil {
		panic(err)
	}

	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}

	rdbController, err := rdbc.New(mysql.New(providerType, sqlDB))
	if err != nil {
		panic(err)
	}

	return rdbController
}

const testSQL string = `CREATE DATABASE IF NOT EXISTS LibraryManagement;

USE LibraryManagement;

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

INSERT INTO Books (Title, Author, PublicationYear, Publisher, Quantity) VALUES ("tCEzLpdeW", "NseXeqlBh", 224, "MgOap", 254);

INSERT INTO Books (Title, Author, PublicationYear, Publisher, Quantity) VALUES ("jEWuQzfTY", "YgUjwtNpHU", 220, "QSLCdwY", 157);

INSERT INTO Books (Title, Author, PublicationYear, Publisher, Quantity) VALUES ("iMes", "ihodtYg", 211, "zjpjyMKlF", 228);

INSERT INTO Books (Title, Author, PublicationYear, Publisher, Quantity) VALUES ("wSbEMUU", "iLmPOEas", 55, "TzbA", 62);

INSERT INTO Books (Title, Author, PublicationYear, Publisher, Quantity) VALUES ("DNKqE", "Dajh", 214, "DjEUpBg", 255);

INSERT INTO Books (Title, Author, PublicationYear, Publisher, Quantity) VALUES ("HuqfKx", "FqWSBOQs", 147, "pHZOJuB", 11);

INSERT INTO Books (Title, Author, PublicationYear, Publisher, Quantity) VALUES ("TQVNDEdZ", "fWaE", 171, "EMhFiAHf", 245);

INSERT INTO Books (Title, Author, PublicationYear, Publisher, Quantity) VALUES ("BTMvof", "YLsyBWXQhv", 170, "PMoIFLg", 22);

INSERT INTO Books (Title, Author, PublicationYear, Publisher, Quantity) VALUES ("TsbXTGvs", "yvkJgFKTU", 14, "TivqAzaGAl", 215);

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

INSERT INTO Members (Name, Address, PhoneNo ,Email ,JoinedDate ,ExpiryDate ,IsActive ) VALUES ("YSzcg", "ZWgffcBi", "EcCnDjq" ,"GJMNruf" ,"1996-01-04" ,"2022-03-21" , 0 );

INSERT INTO Members (Name, Address, PhoneNo ,Email ,JoinedDate ,ExpiryDate ,IsActive ) VALUES ("qMmHPvGXtj", "qVNJrmdG", "XSArhl" ,"mDFFwsQJg" ,"1966-10-08" ,"1997-10-14" , 0 );

INSERT INTO Members (Name, Address, PhoneNo ,Email ,JoinedDate ,ExpiryDate ,IsActive ) VALUES ("SIbQzF", "QEfwEpVA", "xjAGI" ,"IpcaUF" ,"1931-09-28" ,"1902-10-23" , 0 );

INSERT INTO Members (Name, Address, PhoneNo ,Email ,JoinedDate ,ExpiryDate ,IsActive ) VALUES ("vxnpMWCNz", "Cnpd", "FRjSjXsHLy" ,"DyVpKicQ" ,"1992-10-07" ,"1917-09-29" , 1 );

INSERT INTO Members (Name, Address, PhoneNo ,Email ,JoinedDate ,ExpiryDate ,IsActive ) VALUES ("fISmWpzwun", "VznGPW", "pnWIj" ,"CzBfTjMHQ" ,"1972-04-25" ,"1922-07-30" , 1 );

INSERT INTO Members (Name, Address, PhoneNo ,Email ,JoinedDate ,ExpiryDate ,IsActive ) VALUES ("gdyeW", "xzYkFQFNE", "GJaAPzRS" ,"jbsXGHBnP" ,"1968-12-29" ,"1907-08-22" , 1 );

INSERT INTO Members (Name, Address, PhoneNo ,Email ,JoinedDate ,ExpiryDate ,IsActive ) VALUES ("pGHvPVrqS", "CGRZRIxYF", "ybitYUs" ,"HTOJUEGEmy" ,"1964-05-30" ,"1992-03-30" , 1 );

INSERT INTO Members (Name, Address, PhoneNo ,Email ,JoinedDate ,ExpiryDate ,IsActive ) VALUES ("eFeyV", "qWJQAPYU", "uPOv" ,"RhCYyjv" ,"2023-08-28" ,"1981-05-24" , 0 );

INSERT INTO Members (Name, Address, PhoneNo ,Email ,JoinedDate ,ExpiryDate ,IsActive ) VALUES ("XtusB", "wtAMte", "iFPRhNT" ,"yeuFBB" ,"2006-01-19" ,"1943-02-22" , 1 );

INSERT INTO Members (Name, Address, PhoneNo ,Email ,JoinedDate ,ExpiryDate ,IsActive ) VALUES ("rdtWz", "cqqSDTp", "cxxhl" ,"FpVOi" ,"1909-04-17" ,"1960-10-08" , 1 );

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

INSERT INTO BorrowedBooks (MemberID, BookID, BorrowedDate, DueDate, ReturnedDate, FinePaid) VALUES (157, 82, "1935-09-26", "2013-07-07", "2022-04-19", 8);

INSERT INTO BorrowedBooks (MemberID, BookID, BorrowedDate, DueDate, ReturnedDate, FinePaid) VALUES (219, 195, "1980-01-24", "1909-05-23", "1907-04-04", 110);

INSERT INTO BorrowedBooks (MemberID, BookID, BorrowedDate, DueDate, ReturnedDate, FinePaid) VALUES (218, 127, "2019-02-06", "2023-12-01", "1986-09-03", 146);

INSERT INTO BorrowedBooks (MemberID, BookID, BorrowedDate, DueDate, ReturnedDate, FinePaid) VALUES (239, 18, "1930-04-08", "1919-03-06", "1982-03-10", 184);

INSERT INTO BorrowedBooks (MemberID, BookID, BorrowedDate, DueDate, ReturnedDate, FinePaid) VALUES (231, 31, "1957-07-31", "1975-04-04", "1988-12-07", 120);

INSERT INTO BorrowedBooks (MemberID, BookID, BorrowedDate, DueDate, ReturnedDate, FinePaid) VALUES (228, 25, "1945-07-11", "1997-01-29", "1995-12-12", 105);

INSERT INTO BorrowedBooks (MemberID, BookID, BorrowedDate, DueDate, ReturnedDate, FinePaid) VALUES (141, 162, "1901-12-05", "2020-11-14", "1912-10-05", 121);

INSERT INTO BorrowedBooks (MemberID, BookID, BorrowedDate, DueDate, ReturnedDate, FinePaid) VALUES (157, 53, "2006-07-28", "2023-01-11", "1942-11-09", 14);

INSERT INTO BorrowedBooks (MemberID, BookID, BorrowedDate, DueDate, ReturnedDate, FinePaid) VALUES (39, 111, "2015-07-15", "1977-12-22", "1957-08-11", 169);
`
