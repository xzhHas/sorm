package sorm

import "database/sql"

// Result 结构体表示数据库操作的结果
type Result struct {
	err error
	res sql.Result
}

// Err 返回 Result 中的错误信息
func (r Result) Err() error {
	return r.err
}

// LastInsertId 返回最后一次插入操作生成的 ID
func (r Result) LastInsertId() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.res.LastInsertId()
}

// RowsAffected 返回受影响的行数
func (r Result) RowsAffected() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.res.RowsAffected()
}
