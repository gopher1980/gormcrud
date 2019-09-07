package gormcrud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	"github.com/biezhi/gorm-paginator/pagination"
)

type ErrorCrud struct {
	error
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type LinkStatusCrud struct {
	Message     string `json:"message"`
	Status      string `json:"status"`
	Operation   string `json:"operation"`
	CountAfter  int    `json:"count_after"`
	CountBefore int    `json:"count_before"`
}

// ValidateSave is interface for validate save
type ValidateSave interface {
	CrudValidateSave(db *gorm.DB) error
}

// ValidateDelete is interface for validate delete
type ValidateDelete interface {
	CrudValidateDelete(db *gorm.DB) error
}

// Save entity
func Save(db *gorm.DB, new interface{}) func(w http.ResponseWriter, r *http.Request, id string) {

	return func(w http.ResponseWriter, r *http.Request, id string) {
		db1 := db.Set("gorm:auto_preload", true).
			Set("gorm:association_autoupdate", false).
			Set("gorm:association_autocreate", false)
		w.Header().Set("Content-Type", "application/json")
		entity := reflect.New(reflect.TypeOf(new)).Interface()
		reqBody, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(reqBody, entity)
		if ok, err := entity.(ValidateSave); err {
			errValidation := ok.CrudValidateSave(db1)
			if errValidation != nil {
				json.NewEncoder(w).Encode(errValidation)
				return
			}
		}
		fmt.Println(
			reflect.New(reflect.TypeOf(new)))
		ret := db1.Save(entity)
		if ret != nil {
			if ret.Error != nil {
				json.NewEncoder(w).Encode(ret)
				return
			}
		}
		json.NewEncoder(w).Encode(ret)
	}
}

// All return all entities
func All(db *gorm.DB, elem interface{}) func(w http.ResponseWriter, r *http.Request, id string) {
	return func(w http.ResponseWriter, r *http.Request, id string) {
		db := db.Set("gorm:auto_preload", true).Set("gorm:association_autoupdate", false).Set("gorm:association_autocreate", false)
		w.Header().Set("Content-Type", "application/json")
		entity := reflect.New(reflect.TypeOf(elem)).Interface()
		ret := db.Find(entity)
		if ret.RowsAffected == 0 {
			var a [0]interface{}
			json.NewEncoder(w).Encode(a)
			return
		}
		json.NewEncoder(w).Encode(entity)
	}
}

// Page return pagination 
func Page(db *gorm.DB, elem interface{}) func(w http.ResponseWriter, r *http.Request, id string) {
	return func(w http.ResponseWriter, r *http.Request, id string) {
		db := db.Set("gorm:auto_preload", true).Set("gorm:association_autoupdate", false).Set("gorm:association_autocreate", false)
		w.Header().Set("Content-Type", "application/json")
		entity := reflect.New(reflect.TypeOf(elem)).Interface()
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

		ret := pagination.Paging(&pagination.Param{
			DB:      db,
			Page:    page,
			Limit:   limit,
			OrderBy: []string{"id desc"},
		}, entity)

		json.NewEncoder(w).Encode(ret)
	}
}

// Get return one entity
func Get(db *gorm.DB, elem interface{}) func(w http.ResponseWriter, r *http.Request, id string) {
	return func(w http.ResponseWriter, r *http.Request, id string) {
		db := db.Set("gorm:auto_preload", true).
			Set("gorm:association_autoupdate", false).
			Set("gorm:association_autocreate", false)
		w.Header().Set("Content-Type", "application/json")
		entity := reflect.New(reflect.TypeOf(elem)).Interface()

		key := id
		ret := db.Where("id = ?", key).First(entity)
		if ret != nil {
			if ret.RowsAffected == 0 {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(ErrorCrud{Message: "Status Not Found", Code: http.StatusNotFound})
				return
			}
			if ret.Error != nil {
				json.NewEncoder(w).Encode(ret)
				return
			}
		}
		json.NewEncoder(w).Encode(entity)
	}
}

// Delete is operation for delete entity
func Delete(db *gorm.DB, new interface{}) func(w http.ResponseWriter, r *http.Request, id string) {
	return func(w http.ResponseWriter, r *http.Request, id string) {
		w.Header().Set("Content-Type", "application/json")
		entity := reflect.New(reflect.TypeOf(new)).Interface()
		key := id
		ret := db.Where("id = ?", key).First(entity)
		if ret != nil {
			if ret.RowsAffected == 0 {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(ErrorCrud{Message: "Status Not Found", Code: http.StatusNotFound})
				return
			}
			if ret.Error != nil {
				_ = json.NewEncoder(w).Encode(ret)
				return
			}
		}

		if ok, err := entity.(ValidateDelete); err {
			errValidation := ok.CrudValidateDelete(db)
			if errValidation != nil {
				_ = json.NewEncoder(w).Encode(errValidation)
				return
			}
		}

		db.Delete(entity)
		_ = json.NewEncoder(w).Encode(entity)
		return
	}
}

// Link is operation for link and unlink entities
func Link(db *gorm.DB, root interface{}, op string) func(w http.ResponseWriter, r *http.Request, id string) {
	return func(w http.ResponseWriter, r *http.Request, id string) {
		db := db.Set("gorm:auto_preload", true).Set("gorm:association_autoupdate", false).Set("gorm:association_autocreate", false)
		id1 := id

		result := make(map[string]LinkStatusCrud)
		rootEntity := reflect.New(reflect.TypeOf(root)).Interface()
		w.Header().Set("Content-Type", "application/json")
		ret := db.Where("id = ?", id1).First(rootEntity)
		if ret != nil {
			if ret.RowsAffected == 0 {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(ErrorCrud{Message: "Status Not Found (1)", Code: http.StatusNotFound})
				return
			}
			if ret.Error != nil {
				json.NewEncoder(w).Encode(ret)
				return
			}
		}
		for key, values := range r.URL.Query() {
			field := key
			for _, id2 := range values {
				elem := reflect.ValueOf(rootEntity).Elem()

				var child reflect.Type
				var childCurrentValue reflect.Value
				child = nil
				childFieldIndex := 0
				for i := 0; i < elem.NumField(); i++ {
					name := elem.Type().Field(i).Name
					if strings.EqualFold(name, field) {
						child = elem.Field(i).Type()
						childCurrentValue = elem.Field(i)
						childFieldIndex = i
						break
					}
				}
				if child == nil {
					result[field+id1+id2] = LinkStatusCrud{
						Message:     "ID:" + id1 + " -> " + id2 + "(err)(Field Not Found)",
						Status:      "err",
						Operation:   op,
						CountAfter:  -1,
						CountBefore: -1,
					}
					continue
				}

				childEntity := reflect.New(child).Interface()
				ret = db.Where("id = ?", id2).First(childEntity)
				if ret != nil {
					if ret.RowsAffected == 0 {
						w.WriteHeader(http.StatusNotFound)
						result[field+id1+"_"+id2] = LinkStatusCrud{
							Message:     "ID:" + id1 + " -> " + id2 + "(err)(Status Not Found)",
							Status:      "err",
							Operation:   op,
							CountAfter:  -1,
							CountBefore: -1,
						}
						continue
					}
					if ret.Error != nil {
						result[field+id1+id2] = LinkStatusCrud{
							Message:     "ID:" + id1 + " -> " + id2 + "(err)(" + string(ret.Error.Error()) + ")",
							Status:      "err",
							Operation:   op,
							CountAfter:  -1,
							CountBefore: -1,
						}
						continue
					}
				}

				func() {
					association := db.Model(rootEntity).Association(field)
					countBefore := association.Count()

					if op == "link" {
						defer func() {
							if r := recover(); r != nil {
								result[field+id1+"_"+id2] = LinkStatusCrud{
									Message:     "ID:" + id1 + " -> " + id2 + " (err) " + r.(string) + ".",
									Status:      "err",
									Operation:   op,
									CountBefore: countBefore,
									CountAfter:  -1,
								}
							}
						}()
						value := reflect.ValueOf(childEntity).Elem().Index(0)
						childCurrentValue = reflect.Append(childCurrentValue, value)
						reflect.ValueOf(rootEntity).Elem().Field(childFieldIndex).Set(childCurrentValue)
						db := db.
							Set("gorm:association_autoupdate", true).
							Set("gorm:association_autocreate", true)

						db.Save(rootEntity)
						countAfter := association.Count()
						result[field+id1+"_"+id2] = LinkStatusCrud{
							Message:     "ID:" + id1 + " -> " + id2 + " (ok)",
							Status:      "ok",
							Operation:   "link",
							CountBefore: countBefore,
							CountAfter:  countAfter,
						}
					} else {
						countAfter := db.Model(rootEntity).Association(field).Delete(childEntity).Count()
						result[field+id1+"_"+id2] = LinkStatusCrud{
							Message:     "ID:" + id1 + " -/-> " + id2 + " (ok)",
							Status:      "ok",
							Operation:   "unlink",
							CountBefore: countBefore,
							CountAfter:  countAfter,
						}
					}
				}()

			}
		}

		json.NewEncoder(w).Encode(result)
		return

	}

}

type MapperGormCrud struct {
	R        *mux.Router
	RestBase string
	Db       *gorm.DB
	Entity   interface{}
	Array    interface{}
}

func WrapMux(f func(http.ResponseWriter, *http.Request, string)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		f(w, r, id)
	}
}

// https://tools.ietf.org/html/draft-snell-link-method-12

// MapMux is constructor for mapper of mux
func MapMux(r *mux.Router, db *gorm.DB) MapperGormCrud {
	return MapperGormCrud{R: r, Db: db}
}

func (g MapperGormCrud) NewMap(restBase string, entity interface{}, array interface{}) MapperGormCrud {
	return MapperGormCrud{R: g.R, RestBase: restBase, Db: g.Db, Entity: entity, Array: array}
}

func (g MapperGormCrud) Save() MapperGormCrud {
	g.R.HandleFunc(g.RestBase, WrapMux(Save(g.Db, g.Entity))).Methods(http.MethodPost)
	return g
}

func (g MapperGormCrud) All() MapperGormCrud {
	g.R.HandleFunc(g.RestBase, WrapMux(All(g.Db, g.Array))).Methods(http.MethodGet)
	return g
}

func (g MapperGormCrud) Page() MapperGormCrud {
	g.R.HandleFunc(g.RestBase+".page", WrapMux(Page(g.Db, g.Array))).Methods(http.MethodGet)
	return g
}

func (g MapperGormCrud) Get() MapperGormCrud {
	g.R.HandleFunc(g.RestBase+"/{id}", WrapMux(Get(g.Db, g.Entity))).Methods(http.MethodGet)
	return g
}

func (g MapperGormCrud) Delete() MapperGormCrud {
	g.R.HandleFunc(g.RestBase+"/{id}", WrapMux(Delete(g.Db, g.Entity))).Methods(http.MethodDelete)
	return g
}

func (g MapperGormCrud) LinkMethod() MapperGormCrud {
	g.R.HandleFunc(g.RestBase+"/{id}", WrapMux(Link(g.Db, g.Entity, "link"))).Methods("LINK")
	g.R.HandleFunc(g.RestBase+"/{id}", WrapMux(Link(g.Db, g.Entity, "unlink"))).Methods("UNLINK")
	return g
}

func (g MapperGormCrud) LinkUrl() MapperGormCrud {
	g.R.HandleFunc(g.RestBase+"/{id}/link", WrapMux(Link(g.Db, g.Entity, "link"))).Methods(http.MethodGet)
	g.R.HandleFunc(g.RestBase+"/{id}/unlink", WrapMux(Link(g.Db, g.Entity, "unlink"))).Methods(http.MethodGet)
	return g
}

func (g MapperGormCrud) Base() MapperGormCrud {
	g.
		Delete().
		Get().
		Page().
		Save()
	return g
}

func (g MapperGormCrud) Full() MapperGormCrud {
	g.
		All().
		Delete().
		Get().
		LinkMethod().
		LinkUrl().
		Page().
		Save()
	return g
}

// MapperGinGornCrud is struct of mapper gingonic
type MapperGinGormCrud struct {
	R        *gin.Engine
	RestBase string
	Db       *gorm.DB
	Entity   interface{}
	Array    interface{}
}

// WrapF is a helper function for wrapping http.HandlerFunc and returns a Gin middleware.
func WrapGin(f func(http.ResponseWriter, *http.Request, string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		fmt.Println(id)
		f(c.Writer, c.Request, id)
	}
}

// MapGin is constructor mapper for gingonic
func MapGin(engine *gin.Engine, db *gorm.DB) MapperGinGormCrud {
	g := MapperGinGormCrud{R: engine, Db: db}
	return g
}

// NewMap configuration endpoint
func (g MapperGinGormCrud) NewMap(restBase string, entity interface{}, array interface{}) MapperGinGormCrud {
	return MapperGinGormCrud{R: g.R, RestBase: restBase, Db: g.Db, Entity: entity, Array: array}
}

// Save one entity
func (g MapperGinGormCrud) Save() MapperGinGormCrud {
	g.R.POST(g.RestBase, WrapGin(Save(g.Db, g.Entity)))
	return g
}

// Return all entities
func (g MapperGinGormCrud) All() MapperGinGormCrud {
	g.R.GET(g.RestBase, WrapGin(All(g.Db, g.Array)))
	return g
}

// Page return page with querystring page(number page) and limit (size page) .page?pahe=1&limit=10
func (g MapperGinGormCrud) Page() MapperGinGormCrud {
	g.R.GET(g.RestBase+".page", WrapGin(Page(g.Db, g.Array)))
	return g
}

// Get return one entity for id
func (g MapperGinGormCrud) Get() MapperGinGormCrud {
	g.R.GET(g.RestBase+"/:id", WrapGin(Get(g.Db, g.Entity)))
	return g
}

// Delete map operation delete on method delete 
func (g MapperGinGormCrud) Delete() MapperGinGormCrud {
	g.R.DELETE(g.RestBase+"/:id", WrapGin(Delete(g.Db, g.Entity)))

	return g
}

// LinkMethod map operation link and unlink with indicator in method htpp LINK UNLINK
func (g MapperGinGormCrud) LinkMethod() MapperGinGormCrud {
	g.R.Handle("LINK", g.RestBase+"/:id/link", WrapGin(Link(g.Db, g.Entity, "link")))
	g.R.Handle("UNLINK", g.RestBase+"/:id/unlink", WrapGin(Link(g.Db, g.Entity, "unlink")))
	return g
}

// LinkUrl map operation link and unlink with indicator in url
func (g MapperGinGormCrud) LinkUrl() MapperGinGormCrud {
	g.R.GET(g.RestBase+"/:id/link", WrapGin(Link(g.Db, g.Entity, "link")))
	g.R.GET(g.RestBase+"/:id/unlink", WrapGin(Link(g.Db, g.Entity, "unlink")))
	return g
}

// Base map only Delete, Get , Page and Save
func (g MapperGinGormCrud) Base() MapperGinGormCrud {
	g.
		Delete().
		Get().
		Page().
		Save()
	return g
}

// Full map all apis for entity
func (g MapperGinGormCrud) Full() MapperGinGormCrud {
	g.
		All().
		Delete().
		Get().
		LinkMethod().
		LinkUrl().
		Page().
		Save()
	return g
}
