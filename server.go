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

func HandleCrud[T ModelData](s *Server, path string, db *xorm.Engine) (err error) {
	crud, err := NewCrud[T](db)
	if err != nil {
		return err
	}
	var tmp T
	err = db.Sync2(&tmp)
	if err != nil {
		return err
	}
	err = s.POST(path, crud.Create)
	if err != nil {
		return err
	}
	err = s.GET(path, crud.List)
	if err != nil {
		return err
	}
	err = s.PUT(fmt.Sprintf("%s/:id", path), crud.Update)
	if err != nil {
		return err
	}
	err = s.POST(fmt.Sprintf("%s/:id", path), crud.Update)
	if err != nil {
		return err
	}
	err = s.GET(fmt.Sprintf("%s/:id", path), crud.Get)
	if err != nil {
		return err
	}
	err = s.DELETE(fmt.Sprintf("%s/:id", path), crud.Delete)
	if err != nil {
		return err
	}
	return
}

func Register(c gin.IRouter, method, path string, fn interface{}) (err error) {
	typ := reflect.TypeOf(fn)
	if typ.Kind() != reflect.Func {
		err = errors.New("handler func must be func")
		return
	}
	if typ.NumIn() != 2 || typ.In(0) != reflect.TypeOf(&gin.Context{}) {
		err = fmt.Errorf("handler func first arg must be context: %v", typ.In(0))
		return
	}
	if typ.NumOut() < 1 {
		err = errors.New("handler func must have atleast one ret")
		return
	}
	inParam := typ.In(1)
	if inParam.Kind() == reflect.Pointer {
		inParam = inParam.Elem()
	}
	fnValue := reflect.ValueOf(fn)
	errIndex := typ.NumOut() - 1
	c.Handle(method, path, func(c *gin.Context) {
		in := reflect.New(inParam)
		err := AutoBind(c, in.Interface())
		if err != nil {
			c.JSON(200, Resp{Status: 400, Msg: err.Error()})
			return
		}
		params := []reflect.Value{reflect.ValueOf(c), in}

		if inParam.Kind() != reflect.Pointer {
			params[1] = in.Elem()
		}
		rets := fnValue.Call(params)
		if !rets[errIndex].IsNil() {
			err = rets[1].Interface().(error)
			c.JSON(200, Resp{Status: 500, Msg: err.Error()})
			return
		}
		if typ.NumOut() == 2 {
			c.JSON(200, Resp{Status: 200, Data: rets[0].Interface()})
		} else {
			c.JSON(200, Resp{Status: 200})
		}

	})
	return
}
