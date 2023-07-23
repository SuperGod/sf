package sf

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
	"xorm.io/xorm"
)

type Server struct {
	*gin.Engine
}

func NewServer(middleware ...gin.HandlerFunc) *Server {
	s := new(Server)
	s.Engine = gin.New()
	s.Engine.Use(middleware...)
	return s
}

func (s *Server) Run(addr string) error {
	return s.Engine.Run(addr)
}

func (s *Server) POST(path string, fn interface{}) (err error) {
	return Register(s.Engine, "POST", path, fn)
}

func (s *Server) GET(path string, fn interface{}) (err error) {
	return Register(s.Engine, "GET", path, fn)
}

func (s *Server) PUT(path string, fn interface{}) (err error) {
	return Register(s.Engine, "PUT", path, fn)
}

func (s *Server) DELETE(path string, fn interface{}) (err error) {
	return Register(s.Engine, "DELETE", path, fn)
}

func HandleCrud[T any](s *Server, path string, db *xorm.Engine) (crud *Crud[T], err error) {
	crud, err = NewCrud[T](db)
	if err != nil {
		return
	}
	var tmp T
	err = db.Sync2(&tmp)
	if err != nil {
		return
	}
	err = s.POST(path, crud.Create)
	if err != nil {
		return
	}
	err = s.GET(path, crud.List)
	if err != nil {
		return
	}
	err = s.PUT(fmt.Sprintf("%s/:id", path), crud.Update)
	if err != nil {
		return
	}
	err = s.POST(fmt.Sprintf("%s/:id", path), crud.Update)
	if err != nil {
		return
	}
	err = s.GET(fmt.Sprintf("%s/:id", path), crud.Get)
	if err != nil {
		return
	}
	err = s.DELETE(fmt.Sprintf("%s/:id", path), crud.Delete)
	if err != nil {
		return
	}
	return
}

func Register(c gin.IRouter, method, path string, fn interface{}) (err error) {
	typ := reflect.TypeOf(fn)
	if typ.Kind() != reflect.Func {
		err = errors.New("handler func must be func")
		return
	}
	if typ.NumIn() < 1 || typ.In(0) != reflect.TypeOf(&gin.Context{}) {
		err = fmt.Errorf("handler func first arg must be context: %v", typ.In(0))
		return
	}
	if typ.NumOut() < 1 {
		err = errors.New("handler func must have atleast one ret")
		return
	}
	var inParam reflect.Type
	hasInParam := false
	hasOutParam := false
	if typ.NumIn() == 2 {
		hasInParam = true
		inParam = typ.In(1)
		if inParam.Kind() == reflect.Pointer {
			inParam = inParam.Elem()
		}
	}
	if typ.NumOut() == 2 {
		hasOutParam = true
	}

	fnValue := reflect.ValueOf(fn)
	errIndex := typ.NumOut() - 1
	c.Handle(method, path, func(c *gin.Context) {
		params := []reflect.Value{reflect.ValueOf(c)}
		if hasInParam {
			in := reflect.New(inParam)
			err := AutoBind(c, in.Interface())
			if err != nil {
				c.JSON(200, Resp{Status: 400, Msg: err.Error()})
				return
			}
			if typ.In(1).Kind() != reflect.Pointer {
				in = in.Elem()
			}
			params = append(params, in)
		}

		rets := fnValue.Call(params)
		if !rets[errIndex].IsNil() {
			err = rets[errIndex].Interface().(error)
			c.JSON(200, Resp{Status: 500, Msg: err.Error()})
			return
		}
		if hasOutParam {
			c.JSON(200, Resp{Status: 0, Data: rets[0].Interface()})
		} else {
			c.JSON(200, Resp{Status: 0})
		}

	})
	return
}
