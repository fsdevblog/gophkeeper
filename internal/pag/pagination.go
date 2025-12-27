package pag

const (
	DefaultPerPage     = 10
	DefaultCurrentPage = 1
)

type Options struct {
	PerPage     int32
	CurrentPage int32
}

type Pagination struct {
	offset int32
	limit  int32
}

func New(opts ...func(*Options)) *Pagination {
	options := new(Options)
	for _, opt := range opts {
		opt(options)
	}

	if options.PerPage < 1 {
		options.PerPage = DefaultPerPage
	}

	if options.CurrentPage < 1 {
		options.CurrentPage = DefaultCurrentPage
	}

	return &Pagination{
		offset: (options.CurrentPage - 1) * options.PerPage,
		limit:  options.PerPage,
	}
}

func (p *Pagination) Offset() int32 {
	return p.offset
}

func (p *Pagination) Limit() int32 {
	return p.limit
}

func (p *Pagination) Page() int32 {
	return p.offset/p.limit + 1
}

func (p *Pagination) PerPage() int32 {
	return p.limit
}
