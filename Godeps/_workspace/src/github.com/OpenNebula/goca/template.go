package goca

import (
	"errors"
)

type Template struct {
	XMLResource
	Id   uint
	Name string
}

type TemplatePool struct {
	XMLResource
}

func CreateTemplate(template string) (uint, error) {
	response, err := client.Call("one.template.allocate")
	if err != nil {
		return 0, err
	}

	return uint(response.BodyInt()), nil
}

func NewTemplatePool(args ...int) (*TemplatePool, error) {
	var who, start_id, end_id int

	switch len(args) {
	case 0:
		who = PoolWhoMine
		start_id = -1
		end_id = -1
	case 3:
		who = args[0]
		start_id = args[1]
		end_id = args[2]
	default:
		return nil, errors.New("Wrong number of arguments")
	}

	response, err := client.Call("one.templatepool.info", who, start_id, end_id)
	if err != nil {
		return nil, err
	}

	templatepool := &TemplatePool{XMLResource{body: response.Body()}}

	return templatepool, err

}

func NewTemplate(id uint) *Template {
	return &Template{Id: id}
}

func NewTemplateFromName(name string) (*Template, error) {
	templatePool, err := NewTemplatePool()
	if err != nil {
		return nil, err
	}

	id, err := templatePool.GetIdFromName(name, "/VMTEMPLATE_POOL/VMTEMPLATE")
	if err != nil {
		return nil, err
	}

	return NewTemplate(id), nil
}

func (template *Template) Info() error {
	response, err := client.Call("one.template.info", template.Id)
	template.body = response.Body()
	return err
}

func (template *Template) Delete() error {
	_, err := client.Call("one.template.delete", template.Id)
	return err
}

func (template *Template) Instantiate(name string, pending bool, extra string) (uint, error) {
	response, err := client.Call("one.template.instantiate", template.Id, name, pending, extra)

	if err != nil {
		return 0, err
	}

	return uint(response.BodyInt()), nil
}
