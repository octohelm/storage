package sqlx

import (
	"github.com/octohelm/sqlx/pkg/dberr"
)

func DBErr(err error) *dbErr {
	return &dbErr{
		err: err,
	}
}

type dbErr struct {
	err         error
	errDefault  error
	errNotFound error
	errConflict error
}

func (r dbErr) WithNotFound(err error) *dbErr {
	r.errNotFound = err
	return &r
}

func (r dbErr) WithDefault(err error) *dbErr {
	r.errDefault = err
	return &r
}

func (r dbErr) WithConflict(err error) *dbErr {
	r.errConflict = err
	return &r
}

func (r *dbErr) IsNotFound() bool {
	return dberr.IsNotFound(r.err)
}

func (r *dbErr) IsConflict() bool {
	return dberr.IsConflict(r.err)
}

func (r *dbErr) Err() error {
	if r.err == nil {
		return nil
	}
	if sqlErr, ok := dberr.UnwrapAll(r.err).(*dberr.SqlError); ok {
		switch sqlErr.Type {
		case dberr.SqlErrTypeNotFound:
			if r.errNotFound != nil {
				return r.errNotFound
			}
		case dberr.SqlErrTypeConflict:
			if r.errConflict != nil {
				return r.errConflict
			}
		}
		if r.errDefault != nil {
			return r.errDefault
		}
	}
	return r.err
}
