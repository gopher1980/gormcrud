# Gormcrud

Motivation for this project is to provide the a Golang module for   it can drive all CRUD api of your GORM entities.

Example (gorilla/mux):

```golang
	r := mux.NewRouter()
        gormcrud.MapMux(r, db).
                NewMap("/api/v1/author", Author{}, []Author{}).Full().
                NewMap("/api/v1/category", Category{}, []Category{}).Full().
                NewMap("/api/v1/tag", Tag{}, []Tag{}).Full().
                NewMap("/api/v1/note", Note{}, []Note{}).Full()
        http.Handle("/", r)
	log.Fatal(http.ListenAndServe(addr, nil))
```
full example mux https://github.com/gopher1980/gormcrud/blob/master/mux_example/main.go
Example (Gin Web Framework):

```golang
        r := gin.Default()
        gormcrud.MapGin(r, db).
                NewMap("/api/v1/author", Author{}, []Author{}).Full().
                NewMap("/api/v1/category", Category{}, []Category{}).Full().
                NewMap("/api/v1/tag", Tag{}, []Tag{}).Full().
                NewMap("/api/v1/note", Note{}, []Note{}).Full()

        r.Run(addr)
```

full example gin https://github.com/gopher1980/gormcrud/blob/master/gin_example/main.go
