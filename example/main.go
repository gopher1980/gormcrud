package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gopher1980/gormcrud"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Author is the one entity
type Author struct {
	ID        uint       `gorm:"primary_key" json:"id" `
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
	Name      string     `json:"name"`
}

// Category is the one entity
type Category struct {
	ID         uint       `gorm:"primary_key" json:"id" `
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at"`
	Title      string     `json:"title"`
	CategoryID *uint      `json:"category_id" sql:"type:integer REFERENCES category(id)"`
	Categories []Category `json:"categories"`
	Notes      []Note     `json:"notes"`
}

//CrudValidateSave is Validate
func (elem Category) CrudValidateSave(db *gorm.DB) error {
	if elem.CategoryID == nil {
		return gormcrud.ErrorCrud{Message: "CategoryID can't not be null ", Code: 500}
	}
	return nil
}

// CrudValidateDelete is Validate
func (elem Category) CrudValidateDelete(db *gorm.DB) error {
	if elem.CategoryID == nil {
		return gormcrud.ErrorCrud{Message: "CategoryID can't not delete root category ", Code: 500}
	}
	return nil
}

// Tag is the one entity
type Tag struct {
	ID        uint       `gorm:"primary_key" json:"id" `
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
	Name      string     `json:"title"`
	Notes     []Note     `json:"notes" gorm:"many2many:tag_note;"`
}

// Note is the one entity
type Note struct {
	ID          uint       `gorm:"primary_key" json:"id" `
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Text        string     `json:"text"`
	Version     string     `json:"version"`
	AuthorID    uint       `json:"author_id" sql:"type:integer REFERENCES author(id)"`
	CategoryID  uint       `json:"category_id" sql:"type:integer REFERENCES category(id)"`
	Tags        []Tag      `json:"tags"  gorm:"many2many:tag_note;"`
}

// TagNote is not the one entity, it is the relation many to many. I map for gorm makeing  REFERENCES key
type TagNote struct {
	TagID  uint `sql:"type:integer REFERENCES tag(id)"`
	NoteID uint `sql:"type:integer REFERENCES note(id)"`
}

var (
	db *gorm.DB
)

func main() {
	var uridb string
	var drivedb string
	var addr string
	var debug bool

	flag.StringVar(&drivedb, "drive", "sqlite3", "[sqlite3, mysql, postgres ]")
	flag.StringVar(&uridb, "db", "test.db", "Uri of db")
	flag.StringVar(&addr, "addr", ":9090", "this is addr of this ListenAndServe")
	flag.BoolVar(&debug, "debug", false, "sql debug")

	flag.Parse()
	var err error
	db, err = gorm.Open(drivedb, uridb)
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
	db.SingularTable(true)
	db.AutoMigrate(&Author{})
	db.AutoMigrate(&Category{})
	db.AutoMigrate(&TagNote{})
	db.AutoMigrate(&Tag{})
	db.AutoMigrate(&Note{})
	if drivedb == "sqlite3" {
		db.Exec("PRAGMA foreign_keys = ON;")
	}

	db.Save(Category{ID: 1, Title: "root", CategoryID: nil})

	r := mux.NewRouter()

	gormcrud.Map(r, db).
	NewMap("/api/v1/author", []Author{}).Full().
	NewMap("/api/v1/category", []Category{}).Full().
	NewMap("/api/v1/tag", []Tag{}).Full().
	NewMap("/api/v1/note", []Note{}).Full()

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(addr, nil))
}
