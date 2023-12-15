# ex.Terminator is a destructor for Go objects

[![GoDoc](https://godoc.org/github.com/Seb-C/ex?status.svg)](https://pkg.go.dev/github.com/Seb-C/ex)

`ex.Terminator` is an embeddable object providing a destructor mechanism to your objects.

It provides two methods:
- `object.Defer` to defer closing the resources created by the current object (for example `foo.Defer(foo.db.Close)`).
- `object.Close`, once called, executes all of the deferred operations and returns errors if necessary.

Additionally, `ex.Terminate` helps you to find leaks by reporting errors if an object gets garbage-collected without having been closed.

## Example

https://go.dev/play/p/yQvnVDz04yE

```go
type FileRepository struct {
	ex.Terminator
	file *os.File
}

func NewFileRepository() *FileRepository {
	fileRepo := &FileRepository{}

	fileRepo.file, _ = os.Create("data.json")
	fileRepo.Defer(fileRepo.file.Close)

	fileRepo.Defer(func() error {
		fmt.Println("closing FileRepository")
		return nil
	})

	return fileRepo
}

type DBRepository struct {
	ex.Terminator
	db *sql.DB
}

func NewDBRepository() *DBRepository {
	dbRepo := &DBRepository{}

	dbRepo.db, _ = sql.Open("sqlite3", ":memory:")
	dbRepo.Defer(dbRepo.db.Close)

	dbRepo.Defer(func() error {
		fmt.Println("closing DBRepository")
		return nil
	})

	return dbRepo
}

type Service struct {
	ex.Terminator
	fileRepo *FileRepository
	dbRepo   *DBRepository
}

func NewService() *Service {
	service := &Service{}

	service.fileRepo = NewFileRepository()
	service.Defer(service.fileRepo.Close)

	service.dbRepo = NewDBRepository()
	service.Defer(service.dbRepo.Close)

	return service
}

func main() {
	service := NewService()
	defer service.Close()
}

// closing DBRepository
// closing FileRepository
```

## In which order are resources closed?

`ex.Terminator` follows the same convention than the `defer` keyword: the last deferred operation is executed first:

```go
fileRepo.Defer(func() error { fmt.Println("closing A"); return nil })
fileRepo.Defer(func() error { fmt.Println("closing B"); return nil })
fileRepo.Defer(func() error { fmt.Println("closing C"); return nil })
```

Result:

```
closing C
closing B
closing A
```

## What happens if a deferred Close returns an error?

Even in case of error, all of the deferred functions are always executed. The errors are then returned using `errors.Join`.

## Do I still need to call `.Close()` manually?

No, but it has to be explicitly deferred instead.

`ex.Terminate` is designed to keep the concerns strictly separated.

The best way to use it is to follow this rule: the code which creates a resource is always responsible for closing it.

In short, if `main` calls `NewFoo` which calls `NewBar`, then `main` should defer `foo.Close`, and `foo` should defer `bar.Close`.

This way, you end-up with a hierarchy of `Close` calls: `main` calls `foo.Close`, `foo.Close` calls `bar.Close` which in turn might close other resources it's responsible for.

## Can I use this struct as a property rather than embedding it?

While I recommend embedding it because I think that it is less error-prone and requires less boilerplate, it is very much possible to encapsulate it instead.

However, it is a trade off because of constraints inherent to the Go language:
- The benefit is that the `Defer` method will be encapsulated.
- The drawback is that you will manually have to create a `Close` method to close the terminator.

```go
type Foo struct {
	terminator ex.Terminator
	file       *os.File
}

func NewFoo() *Foo {
	var terminator ex.Terminator

	file, _ := os.Create("data.json")
	terminator.Defer(file.Close)

	return &Foo{
		terminator: terminator,
		file:       file,
	}
}

func (foo *Foo) Close() error {
	return foo.terminator.Close()
}
```
