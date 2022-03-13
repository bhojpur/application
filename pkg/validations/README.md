# Bhojpur Application - Validations Library

The `validations` framework provides a means to [*validate*](https://en.wikipedia.org/wiki/Data_validation)
[Bhojpur ORM](https://github.com/bhojpur/orm) models when creating and updating them.

## Register ORM Callbacks

The `validations` library uses [Bhojpur ORM](https://github.com/bhojpur/orm) callbacks to handle
*validations*, so you will need to register callbacks first:

```go
import (
  orm "github.com/bhojpur/orm/pkg/engine"
  "github.com/bhojpur/applicaiton/pkg/validations"
)

func main() {
  db, err := orm.Open("sqlite3", "demo_db")

  validations.RegisterCallbacks(db)
}
```

### Usage

After `callbacks` have been registered, attempting to create or update any record will trigger the `Validate`
method that you have implemented for your model. If your implementation adds or returns an error, the attempt
will be aborted.

```go
type User struct {
  orm.Model
  Age uint
}

func (user User) Validate(db *orm.DB) {
  if user.Age <= 18 {
    db.AddError(errors.New("age need to be 18+"))
  }
}

db.Create(&User{Age: 10})         // won't insert the record into database, as the `Validate` method will return error

var user User{Age: 20}
db.Create(&user)                  // user with age 20 will be inserted into database
db.Model(&user).Update("age", 10) // user's age won't be updated, will return error `age need to be 18+`

// If you have added more than one error, could get all of them with `db.GetErrors()`
func (user User) Validate(db *orm.DB) {
  if user.Age <= 18 {
    db.AddError(errors.New("age need to be 18+"))
  }
  if user.Name == "" {
    db.AddError(errors.New("name can't be blank"))
  }
}

db.Create(&User{}).GetErrors() // => []error{"age need to be 18+", "name can't be blank"}
```

## [Bhojpur Errors](https://github.com/bhojpur/errors/pkg/validation) integration

The application [validations](https://github.com/bhojpur/application/pkg/validations) supports
[Bhojpur Errors](https://github.com/bhojpur/errors/pkg/validation), so you could add a tag into your
struct for some common *validations*, such as *check required*, *numeric*, *length*, etc.

```go
type User struct {
  orm.Model
  Name           string `valid:"required"`
  Password       string `valid:"length(6|20)"`
  SecurePassword string `valid:"numeric"`
  Email          string `valid:"email"`
}
```

## Customize errors on form field

If you want to display errors for each form field in [Bhojpur CMS](http://github.com/bhojpur/cms),
you could register your errors like this:

```go
func (user User) Validate(db *orm.DB) {
  if user.Age <= 18 {
    db.AddError(validations.NewError(user, "Age", "age need to be 18+"))
  }
}
```

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).