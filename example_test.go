package ex_test

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/Seb-C/ex"
	_ "github.com/mattn/go-sqlite3"
)

type FileRepository struct {
	ex.Terminator
	file *os.File
}

func NewFileRepository() (fileRepo *FileRepository, err error) {
	fileRepo = &FileRepository{}

	fileRepo.file, err = os.Create("data.json")
	if err != nil {
		return nil, err
	}
	fileRepo.Defer(fileRepo.file.Close)

	fileRepo.Defer(func() error {
		fmt.Println("closing FileRepository")
		return nil
	})

	return fileRepo, nil
}

type DBRepository struct {
	ex.Terminator
	db *sql.DB
}

func NewDBRepository() (dbRepo *DBRepository, err error) {
	dbRepo = &DBRepository{}

	dbRepo.db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}
	dbRepo.Defer(dbRepo.db.Close)

	dbRepo.Defer(func() error {
		fmt.Println("closing DBRepository")
		return nil
	})

	return dbRepo, nil
}

type Service struct {
	ex.Terminator
	fileRepo *FileRepository
	dbRepo   *DBRepository
}

func NewService() (service *Service, err error) {
	service = &Service{}

	service.fileRepo, err = NewFileRepository()
	if err != nil {
		return nil, err
	}
	service.Defer(service.fileRepo.Close)

	service.dbRepo, err = NewDBRepository()
	if err != nil {
		return nil, err
	}
	service.Defer(service.dbRepo.Close)

	return service, nil
}

func Example() {
	service, err := NewService()
	if err != nil {
		panic(err)
	}
	defer service.Close()
}
