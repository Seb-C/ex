# ex.Terminator is a destructor helper for Go

[![GoDoc](https://godoc.org/github.com/Seb-C/ex?status.svg)](https://pkg.go.dev/github.com/Seb-C/ex)

`ex.Terminator` is an embeddable object providing a destructor mechanism to your objects.

It provides two methods:
- `object.Defer` to defer closing the resources created by the current object (for example `foo.Defer(foo.db.Close)`).
- `object.Close`, once called, executes all of the deferred operations and returns errors if necessary.

Additionally, `ex.Terminate` helps you find leaks by reporting errors if an object gets garbage-collected without having been closed.

`ex.Terminator` has several benefits over maintaining your own `Close` methods:
- Similarly to the `defer` keyword, it is easier to keep track of what is being closed or not, because both the open and close operations always goes together. Meanwhile, it is very easy to forget about it when writing or maintaining a `Close` method.
- Errors in manually maintained `Close` methods are often ignored, and handling it explicitly makes the readability worse. `ex.Terminator` takes care of that for you, and includes helpful error messages.
- `ex.Terminator` takes care of closing the resources in the right order, which is easy to get wrong when manually done.

## Example

https://go.dev/play/p/wFxNeCnYPLd

```go
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

func main() {
	service, err := NewService()
	if err != nil {
		panic(err)
	}
	defer service.Close()

	// Output:
	// "closing DBRepository"
	// "closing FileRepository"
}
```

## Is this based on `runtime.SetFinalizer`?

No. Resources are closed synchronously, meaning that the `Close` methods still must be called, either via `defer`, `.Defer` or manually.

However, `ex.Terminator` uses `runtime.SetFinalizer` to help the developers find mistakes: an error message is printed in the console whenever a non-Closed object gets garbage-collected.
But this only used for this purpose. Closing the objects is never done in `SetFinalizer`.

## Can I use it outside the constructor?

Yes! Although I only provided examples using the constructor (because it is the most common use case), you can use `.Defer` in any method and any time of the life-cycle of your objects.

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
