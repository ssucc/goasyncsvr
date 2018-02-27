package base

import (
	"github.com/astaxie/beego/config"
	_ "github.com/astaxie/beego/config/xml"
)

type Conf struct {
	confparser config.Configer
}

func NewConf(adapterName, filename string) (*Conf, error) {
	configer, _ := config.NewConfig(adapterName, filename)
	return &Conf{
		confparser: configer,
	}, nil
}

func (c *Conf) LoadFromFile(filetype string, filename string) {
	switch filetype {
	case "xml":
		c.LoadFromXMLFile(filename)

	case "ini":
		c.LoadFromINIFile(filename)
	}
}

func (c *Conf) LoadFromXMLFile(filename string) error {
	parser, err := config.NewConfig("xml", filename)
	if err != nil {
		c.confparser = parser
	}

	return err
}

func (c *Conf) LoadFromINIFile(filename string) error {
	parser, err := config.NewConfig("ini", filename)
	if err != nil {
		c.confparser = parser
	}

	return err
}

func (c *Conf) GetFieldStr(key string, dv string) string {
	return c.confparser.DefaultString(key, dv)
}

func (c *Conf) GetFieldInt(key string, dv int64) int64 {
	return c.confparser.DefaultInt64(key, dv)
}

func (c *Conf) GetSection(section string) map[string]string {
	m, err := c.confparser.GetSection(section)
	if err != nil {
		return make(map[string]string)
	}

	return m
}

/*
func (c *ConfigContainer) GetSection(section string) (map[string]string, error) {
	if v, ok := c.data[section]; ok {
		return v.(map[string]string), nil
	}
	return nil, errors.New("not exist setction")
}
*/
