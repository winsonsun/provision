package models

import (
	"encoding/json"
	"errors"
	"github.com/xeipuuv/gojsonschema"
	"log"
)

type Params struct {
	Params map[string]json.RawMessage
}

func (p *Params) Validate(err ErrorAdder) {
	for k := range p.Params {
		err.AddError(ValidParamName("Invalid Param Name", k))
	}
}

func(p *Params) Fill() {
	if len(p.Params) == 0 {
		p.Params = map[string]json.RawMessage{}
	}
}

func (p *Params) GetParams() map[string]interface{} {
	res := map[string]interface{}{}
	for k,v := range p.Params {
		var vv interface{}
		if err := json.Unmarshal(v,&vv); err != nil {
			log.Panicf("GetParams unmarshal error %s:%s: %v", k,v,err)
		}
		res[k] = vv
	}
	return res
}

func (p *Params) SetParams(val map[string]interface{}) {
	res := map[string]json.RawMessage{}
	for k, v := range val {
		buf, err := json.Marshal(v)
		if err != nil {
			log.Panicf("SetParams marshal error %s:%v: %v", k, v, err)
		}
		res[k] = buf
	}
	p.Params = res
}

func (p *Params) GetRawParam(key string) json.RawMessage {
	p.Fill()
	return p.Params[key]
}

func (p *Params) SetRawParam(key string, buf json.RawMessage) {
	p.Fill()
	p.Params[key] = buf
}

func (p *Params) DeleteParam(key string) {
	p.Fill()
	delete(p.Params,key)
}

func (p *Params) GetParam(key string, val interface{}) error {
	p.Fill()
	buf,ok := p.Params[key]
	if !ok {
		return errors.New("Not Found")
	}
	return json.Unmarshal(buf,&val)
}

func (p *Params) SetParam(key string, val interface{}) error {
	p.Fill()
	buf, err := json.Marshal(val)
	if err == nil {
		p.Params[key] = buf
	}
	return err
}

// Param represents metadata about a Parameter or a Preference.
// Specifically, it contains a description of what the information
// is for, detailed documentation about the param, and a JSON schema that
// the param must match to be considered valid.
// swagger:model
type Param struct {
	Validation
	Access
	Meta
	Owned
	Bundled
	// Name is the name of the param.  Params must be uniquely named.
	//
	// required: true
	Name string
	// Description is a one-line description of the parameter.
	Description string
	// Documentation details what the parameter does, what values it can
	// take, what it is used for, etc.
	Documentation string
	// Secure implies that any API interactions with this Param
	// will deal with SecureData values.
	//
	// required: true
	Secure bool
	// Schema must be a valid JSONSchema as of draft v4.
	//
	// required: true
	Schema json.RawMessage
}

func (p *Param) GetMeta() Meta {
	return p.Meta
}

func (p *Param) SetMeta(d Meta) {
	p.Meta = d
}

func (p *Param) GetDocumentation() string {
	return p.Documentation
}

func (p *Param) GetDescription() string {
	return p.Description
}

func (p *Param) rawSchemaVal(key string) ([]byte, error) {
	mp := map[string]json.RawMessage{}
	if err := json.Unmarshal(p.Schema,&mp); err != nil {
		return nil, err
	}
	return mp[key], nil
}
func (p *Param) DefaultValue() (interface{}, bool) {
	v, err := p.rawSchemaVal("default")
	if err != nil {
		return nil, false
	}
	var res interface{}
	if err = json.Unmarshal(v,&res); err != nil {
		return nil, false
	}
	return res, true
}

func (p *Param) TypeValue() (interface{}, bool) {
	v, err := p.rawSchemaVal("default")
	if err != nil {
	return nil, false}
	var res interface{}
	if err = json.Unmarshal(v,&res); err != nil {
		return nil, false
	}
	return res, true
}

func (p *Param) Validate() {
	p.AddError(ValidParamName("Invalid Name", p.Name))
	if p.Schema != nil && p.Schema[0] == '{' {
		validator, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(p.Schema))
		if err != nil {
			p.AddError(err)
			return
		}
		dv,err := p.rawSchemaVal("default")
		if err != nil || dv == nil {
			return
		}
		res, err := validator.Validate(gojsonschema.NewBytesLoader(dv))
		if err != nil {
			p.Errorf("Error validating default value: %v", err)
		} else if !res.Valid() {
			for _, e := range res.Errors() {
				p.Errorf("Error in default value: %v", e.String())
			}
		}
	}
}

func (p *Param) SetName(s string) {
	p.Name = s
}

func (p *Param) Prefix() string {
	return "params"
}

func (p *Param) Key() string {
	return p.Name
}

func (p *Param) KeyName() string {
	return "Name"
}

func (p *Param) Fill() {
	if p.Meta == nil {
		p.Meta = Meta{}
	}
	p.Validation.fill(p)
}

func (p *Param) AuthKey() string {
	return p.Key()
}

func (b *Param) SliceOf() interface{} {
	s := []*Param{}
	return &s
}

func (b *Param) ToModels(obj interface{}) []Model {
	items := obj.(*[]*Param)
	res := make([]Model, len(*items))
	for i, item := range *items {
		res[i] = Model(item)
	}
	return res
}
