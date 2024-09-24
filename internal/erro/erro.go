package erro

import (
	"errors"

)

var (
	ErrNotFound 		= errors.New("item not found")
	ErrInsert 			= errors.New("insert data error")
	ErrUpdate			= errors.New("update data error")
	ErrDelete 			= errors.New("delete data error")
	ErrUnmarshal 		= errors.New("unmarshal json error")
	ErrUnauthorized 	= errors.New("not authorized")
	ErrServer		 	= errors.New("server identified error")
	ErrHTTPForbiden		= errors.New("forbiden request")
	ErrInvalid			= errors.New("invalid data")
	ErrTransInvalid		= errors.New("transaction invalid")
	ErrInvalidAmount	= errors.New("invalid amount for this transaction type")
)