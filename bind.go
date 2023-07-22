package sf

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

type BindOptions struct {
	mapstructure.DecoderConfig
}

type BindOption func(o *BindOptions)

func AutoBind(c *gin.Context, data any, options ...BindOption) (err error) {
	var cfg = BindOptions{DecoderConfig: mapstructure.DecoderConfig{TagName: "query", WeaklyTypedInput: true, Result: data}}
	for _, v := range options {
		v(&cfg)
	}

	dec, err := mapstructure.NewDecoder(&cfg.DecoderConfig)
	if err != nil {
		fmt.Printf("newdecode error: %#v %s %s\n", data, err.Error(), reflect.TypeOf(data))
		return
	}
	// bind query params
	err = bindQuery(c, dec, data)
	if err != nil {

		return
	}

	// bind uri params
	cfg.TagName = "uri"
	err = bindUri(c, dec, data)
	if err != nil {
		return
	}
	if c.Request.ContentLength > 0 {
		err = c.Bind(data)
		if err != nil {
			return
		}
	}
	return
}

func bindUri(c *gin.Context, dec *mapstructure.Decoder, data any) (err error) {
	params := make(map[string]any, len(c.Params))
	for _, v := range c.Params {
		params[v.Key] = v.Value
	}
	err = dec.Decode(params)
	if err != nil {
		err = fmt.Errorf("bind uri param failed: %w", err)
		return
	}
	return
}

func bindQuery(c *gin.Context, dec *mapstructure.Decoder, data any) (err error) {
	params := make(map[string]any, len(c.Request.URL.Query()))
	for k, v := range c.Request.URL.Query() {
		if len(v) == 1 {
			params[k] = v[0]
		} else {
			params[k] = v
		}
	}
	err = dec.Decode(params)
	if err != nil {
		err = fmt.Errorf("bind query param failed: %w", err)
		return
	}
	return
}
