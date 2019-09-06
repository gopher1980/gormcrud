# Gormcrud

Motivation for this project is to provide the a Golang module for   it can drive all CRUD api of your GORM entities.

Example:

```golang
	gormcrud.Map(r, db).
	NewMap("/api/v1/author", []Author{}).Full().
	NewMap("/api/v1/category", []Category{}).Full().
	NewMap("/api/v1/tag", []Tag{}).Full().
	NewMap("/api/v1/note", []Note{}).Full()
```

full example https://github.com/gopher1980/gormcrud/blob/master/example/main.go

