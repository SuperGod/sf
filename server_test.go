package sf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func TestRegister(t *testing.T) {
	type Req struct {
		Name string `query:"name" json:"name" binding:"required"`
		Age  int    `json:"age" binding:"required"`
	}
	type Rsp struct {
		Info string
	}
	fn := func(c *gin.Context, r *Req) (resp *Rsp, err error) {
		fmt.Println("req:", *r)
		resp = &Rsp{
			Info: fmt.Sprintf("%s/%d", r.Name, r.Age),
		}
		return
	}
	r := gin.Default()
	err := Register(r, "POST", "/", fn)
	if err != nil {
		t.Fatal(err.Error())
	}
	go func() {
		r.Run("localhost:8885")
	}()
	buf := bytes.NewBuffer([]byte(`{"age": 12, "c":3.0}`))
	resp, err := http.Post("http://localhost:8885/?name=name", "application/json", buf)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	t.Log(string(b))
	time.Sleep(time.Second * 2)
}

type TestCrudData struct {
	ID   int64  `uri:"id" json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
	Addr string `json:"addr"`
}

func (d TestCrudData) GetID() any {
	return d.ID
}

func TestCrud(t *testing.T) {
	srv := NewServer()
	db, err := xorm.NewEngine("sqlite", "./test.db")
	if err != nil {
		t.Fatal(err.Error())
	}
	db.ShowSQL(true)
	err = HandleCrud[TestCrudData](srv, "/data", db)
	if err != nil {
		t.Fatal(err.Error())
	}
	go func() {
		err := srv.Run("localhost:8886")
		if err != nil {
			t.Log(err)
		}
	}()
	time.Sleep(time.Second * 2)

	// Create data
	var buf = bytes.NewBuffer([]byte(`{"name": "a", "age":10,"addr":"Home"}`))
	resp, err := http.Post("http://localhost:8886/data", "application/json", buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	ret, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(string(ret))

	var testData TestCrudData
	var r = Resp{Data: &testData}
	err = json.Unmarshal(ret, &r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("data ret:", testData)

	// Get one data
	resp, err = http.Get(fmt.Sprintf("http://localhost:8886/data/%d", testData.ID))
	if err != nil {
		t.Fatal(err.Error())
	}
	ret, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(string(ret))

	// Update one data
	buf = bytes.NewBuffer([]byte(`{"name": "ab", "age":10,"addr":"Home"}`))
	resp, err = http.Post(fmt.Sprintf("http://localhost:8886/data/%d", testData.ID), "application/json", buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	ret, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log("update resp:", string(ret))

	// List datas
	resp, err = http.Get("http://localhost:8886/data")
	if err != nil {
		t.Fatal(err.Error())
	}
	ret, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(string(ret))

}
