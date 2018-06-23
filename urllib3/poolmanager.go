package urllib3

type PoolManager struct {
	pool map[string]*Conn
}

func (p *PoolManager) Init() {
	p.pool = make(map[string]*Conn)
}

func (p *PoolManager) ConnectionFromUrl(Url string) *Conn {
	if p.pool == nil {
		p.Init()
	}
	if conn, ok := p.pool[Url]; ok {
		return conn
	}
	conn := &Conn{}
	conn.init()
	p.pool[Url] = conn
	return conn
}
