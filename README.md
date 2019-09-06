# Gormcrud

Motivation for this project is to provide the a Golang module for   it can drive all CRUD api of your GORM entities.

Example:
```
gormcrud.Mapping(r, "/api/v1/author", db, Author{}, []Author{})
gormcrud.Mapping(r, "/api/v1/category", db, Category{}, []Category{})
gormcrud.Mapping(r, "/api/v1/tag", db, Tag{}, []Tag{})
gormcrud.Mapping(r, "/api/v1/note", db, Note{}, []Note{})
```

