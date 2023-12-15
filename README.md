# ex.Terminator - destructor for your Go objects

[![GoDoc](https://godoc.org/github.com/Seb-C/ex?status.svg)](https://pkg.go.dev/github.com/Seb-C/ex)

`ex.Terminator` is an embeddable object providing a destructor mechanism to your objects.

It provides two methods:
- `object.Defer` to defer closing the resources created by the current object (for example `foo.Defer(foo.db.Close)`).
- `object.Close`, once called, executes all of the deferred operations and returns errors if necessary.

Additionally, ex.Terminate helps you to find leaks. Errors will be displayed if a garbage-collected object was not properly closed.

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
```

## In which order are resources closed?

`ex.Terminator` follows the same convention than the `defer` keyword. That means that the last deferred operation will be executed first:

```
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

## What if a deferred function returns an error?

Even in case of error, all of the deferred functions are always executed. The errors are all returned using `errors.Join`.

## Do I still need to call `.Close()` manually?

Yes. ex.Terminate is design to work like a tree.

The easiest way to use it is to follow this rule: the part of the code that creates a resource must always defer closing it.

In short, if `NewFoo` calls `NewBar`, it should also defer `bar.Close`.

This way, you end-up with a hierarchy of `Close` calls: `main` calls `foo.Close`, `foo.Close` calls `bar.Close` which in turn might close other resources it's responsible for.

## Can I use this struct as a property rather than embedding it?

Yes, it also works, but a little more verbose because of the way Go works.

The benefit of doing this is that the `Defer` method will be encapsulated.

The drawback is that you will manually have to create a `Close` method.

```
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
