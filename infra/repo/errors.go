package repo

type ErrRepoOp struct {
	Op   string
	Code int
	Err  error
}

func (r ErrRepoOp) Error() string {
	return r.Err.Error()
}

func (r ErrRepoOp) Unwrap() error {
	return r.Err
}

func (r ErrRepoOp) Is(other error) bool {
	_, ok := other.(ErrRepoOp)
	return ok
}
