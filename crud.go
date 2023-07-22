package sf

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"xorm.io/xorm"
)

type ModelData interface {
	GetID() interface{}
}

type IdReq struct {
	ID int `uri:"id" binding:"required"`
}

type Resp struct {
	Status int         `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

type PageData[T any] struct {
	Total int `json:"total"`
	Items []T `json:"items"`
}

type PageReq struct {
	Page     int
	PerPage  int
	OrderBy  string
	OrderDir string
	Keywords string
}

func (r *PageReq) Start() int {
	return (r.Page - 1) * r.PerPage
}

type Crud[T ModelData] struct {
	db *xorm.Engine
}

func NewCrud[T ModelData](db *xorm.Engine) (cd *Crud[T], err error) {
	cd = new(Crud[T])
	cd.db = db
	return
}

func (cd *Crud[T]) Create(ctx *gin.Context, t T) (ret T, err error) {
	_, err = cd.db.Context(ctx).Insert(&t)
	if err != nil {
		return
	}
	ret = t
	return
}

func (cd *Crud[T]) Update(ctx *gin.Context, t T) (ret T, err error) {
	_, err = cd.db.Context(ctx).ID(t.GetID()).Update(&t)
	if err != nil {
		return
	}
	ret = t
	return
}

func (cd *Crud[T]) List(ctx *gin.Context, req PageReq) (datas PageData[T], err error) {
	if req.PerPage == 0 {
		req.PerPage = 20
	}
	var t T
	nCount, err := cd.db.Context(ctx).Table(&t).Count(&t)
	if err != nil {
		return
	}
	sess := cd.db.Context(ctx).Table(&t)
	if req.OrderBy != "" {
		sess = sess.OrderBy(req.OrderBy)
		if req.OrderDir == "desc" {
			sess = sess.Desc(req.OrderBy)
		}
	}
	datas.Total = int(nCount)
	err = sess.Limit(req.PerPage, req.Start()).Find(&datas.Items)
	if err != nil {
		return
	}
	return
}

func (cd *Crud[T]) Get(ctx *gin.Context, id IdReq) (data T, err error) {
	fmt.Println("id:", id.ID)
	_, err = cd.db.Context(ctx).ID(id.ID).Get(&data)
	return
}

func (cd *Crud[T]) Delete(ctx *gin.Context, id IdReq) (err error) {
	var t T
	_, err = cd.db.Context(ctx).Table(&t).ID(id.ID).Delete()
	return
}
