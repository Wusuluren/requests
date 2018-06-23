package urllib3

import (
	"net/http"
)

type Conn struct {
	http.Client
}

func (c *Conn) init() {

}

func (c *Conn) Urlopen(req *http.Request) (*http.Response, error) {
	resp, err := c.Do(req)
	return resp, err
}
