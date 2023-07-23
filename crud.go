package sf

import (
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

type CreateFn[T any] func(*gin.Context, T) (T, error)
type UpdateFn[T any] func(*gin.Context, string, T) (T, error)
type GetPreFn[T any] func(*gin.Context, string) (T, error)
type GetPostFn[T any] func(*gin.Context, string, T) (T, error)
type ListPreFn func(*gin.Context, PageReq) error
type ListPostFn[T any] func(*gin.Context, PageReq, PageData[T]) (PageData[T], error)

type DeleteFn func(*gin.Context, string) error

type Crud[T any] struct {
	db         *xorm.Engine
	createPre  CreateFn[T]
	createPost CreateFn[T]
	updatePre  UpdateFn[T]
	updatePost UpdateFn[T]
	getPre     GetPreFn[T]
	getPost    GetPostFn[T]
	listPre    ListPreFn
	listPost   ListPostFn[T]
	deletePre  DeleteFn
	deletePost DeleteFn
}

func NewCrud[T any](db *xorm.Engine) (cd *Crud[T], err error) {
	cd = new(Crud[T])
	cd.db = db
	return
}

func (cd *Crud[T]) SetCreateFn(fnPre, fnPost CreateFn[T]) {
	cd.createPre = fnPre
	cd.createPost = fnPost
}

func (cd *Crud[T]) SetUpdateFn(fnPre, fnPost UpdateFn[T]) {
	cd.updatePre = fnPre
	cd.updatePost = fnPost
}

func (cd *Crud[T]) SetGetFn(fnPre GetPreFn[T], fnPost GetPostFn[T]) {
	cd.getPre = fnPre
	cd.getPost = fnPost
}

func (cd *Crud[T]) SetListFn(fnPre ListPreFn, fnPost ListPostFn[T]) {
	cd.listPre = fnPre
	cd.listPost = fnPost
}

func (cd *Crud[T]) SetDeleteFn(fnPre, fnPost DeleteFn) {
	cd.deletePre = fnPre
	cd.deletePost = fnPost
}

func (cd *Crud[T]) Create(ctx *gin.Context, t T) (ret T, err error) {
	if cd.createPre != nil {
		t, err = cd.createPre(ctx, t)
		if err != nil {
			return
		}
	}
	_, err = cd.db.Context(ctx).Insert(&t)
	if err != nil {
		return
	}
	if cd.createPost != nil {
		t, err = cd.createPost(ctx, t)
		if err != nil {
			return
		}
	}
	ret = t
	return
}

func (cd *Crud[T]) Update(ctx *gin.Context, t T) (ret T, err error) {
	strId := ctx.Param("id")
	if cd.updatePre != nil {
		t, err = cd.updatePre(ctx, strId, t)
		if err != nil {
			return
		}
	}
	_, err = cd.db.Context(ctx).ID(strId).Update(&t)
	if err != nil {
		return
	}
	if cd.updatePost != nil {
		t, err = cd.updatePost(ctx, strId, t)
		if err != nil {
			return
		}
	}
	ret = t
	return
}

func (cd *Crud[T]) List(ctx *gin.Context, req PageReq) (datas PageData[T], err error) {
	if req.PerPage == 0 {
		req.PerPage = 20
	}
	if cd.listPre != nil {
		err = cd.listPre(ctx, req)
		if err != nil {
			return
		}
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
	if cd.listPost != nil {
		datas, err = cd.listPost(ctx, req, datas)
		if err != nil {
			return
		}
	}
	return
}

func (cd *Crud[T]) Get(ctx *gin.Context) (data T, err error) {
	strId := ctx.Param("id")
	if cd.getPre != nil {
		data, err = cd.getPre(ctx, strId)
		if err != nil {
			return
		}
	}
	_, err = cd.db.Context(ctx).ID(strId).Get(&data)
	if cd.getPost != nil {
		data, err = cd.getPost(ctx, strId, data)
		if err != nil {
			return
		}
	}
	return
}

func (cd *Crud[T]) Delete(ctx *gin.Context) (err error) {
	strId := ctx.Param("id")
	if cd.deletePre != nil {
		err = cd.deletePre(ctx, strId)
		if err != nil {
			return
		}
	}
	var t T
	_, err = cd.db.Context(ctx).ID(strId).Delete(&t)
	if cd.deletePost != nil {
		err = cd.deletePost(ctx, strId)
		if err != nil {
			return
		}
	}
	return
}
