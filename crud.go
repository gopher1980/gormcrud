package gormcrud

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

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

// ValidateSave is
type ValidateSave interface {
	CrudValidateSave(db *gorm.DB) error
}

// ValidateDelete is
type ValidateDelete interface {
	CrudValidateDelete(db *gorm.DB) error
}

// Save is
func Save(db *gorm.DB, new interface{}) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
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

// All is
func All(db *gorm.DB, elem interface{}) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

// Page is
func Page(db *gorm.DB, elem interface{}) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

// Get is
func Get(db *gorm.DB, elem interface{}) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := db.Set("gorm:auto_preload", true).
			Set("gorm:association_autoupdate", false).
			Set("gorm:association_autocreate", false)
		w.Header().Set("Content-Type", "application/json")
		entity := reflect.New(reflect.TypeOf(elem)).Interface()

		vars := mux.Vars(r)
		key := vars["id"]
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

// Delete is
func Delete(db *gorm.DB, new interface{}) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		entity := reflect.New(reflect.TypeOf(new)).Interface()
		vars := mux.Vars(r)
		key := vars["id"]

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

// Link is
func Link(db *gorm.DB, root interface{}, op string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db := db.Set("gorm:auto_preload", true).Set("gorm:association_autoupdate", false).Set("gorm:association_autocreate", false)
		vars := mux.Vars(r)
		id1 := vars["id"]

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

// https://tools.ietf.org/id/draft-snell-link-method-01.html#rfc.section.5
// https://tools.ietf.org/html/draft-snell-link-method-12
func Map(r *mux.Router, db *gorm.DB) MapperGormCrud {
	return MapperGormCrud{R: r, Db: db}
}

func (g MapperGormCrud) NewMap(restBase string, array interface{}) MapperGormCrud {
	return MapperGormCrud{R: g.R, RestBase: restBase, Db: g.Db, Entity: reflect.TypeOf(array).Elem(), Array: array}
}

func (g MapperGormCrud) Save() MapperGormCrud {
	g.R.HandleFunc(g.RestBase, Save(g.Db, g.Entity)).Methods(http.MethodPost)
	return g
}

func (g MapperGormCrud) All() MapperGormCrud {
	g.R.HandleFunc(g.RestBase, All(g.Db, g.Array)).Methods(http.MethodGet)
	return g
}

func (g MapperGormCrud) Page() MapperGormCrud {
	g.R.HandleFunc(g.RestBase+".page", Page(g.Db, g.Array)).Methods(http.MethodGet)
	return g
}

func (g MapperGormCrud) Get() MapperGormCrud {
	g.R.HandleFunc(g.RestBase+"/{id}", Get(g.Db, g.Entity)).Methods(http.MethodGet)
	return g
}

func (g MapperGormCrud) Delete() MapperGormCrud {
	g.R.HandleFunc(g.RestBase+"/{id}", Delete(g.Db, g.Entity)).Methods(http.MethodDelete)
	return g
}

func (g MapperGormCrud) LinkMethod() MapperGormCrud {
	g.R.HandleFunc(g.RestBase+"/{id}", Link(g.Db, g.Entity, "link")).Methods("LINK")
	g.R.HandleFunc(g.RestBase+"/{id}", Link(g.Db, g.Entity, "unlink")).Methods("UNLINK")
	return g
}

func (g MapperGormCrud) LinkUrl() MapperGormCrud {
	g.R.HandleFunc(g.RestBase+"/{id}/link", Link(g.Db, g.Entity, "link")).Methods(http.MethodGet)
	g.R.HandleFunc(g.RestBase+"/{id}/unlink", Link(g.Db, g.Entity, "unlink")).Methods(http.MethodGet)
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
