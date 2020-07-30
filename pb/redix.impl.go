package pb

import (
	"context"
	"fmt"
	"strconv"

	"github.com/alash3al/redix/db"
	"github.com/golang/protobuf/ptypes/empty"
)

type RedixService struct {
	db *db.DB
}

func (s *RedixService) SetString(c context.Context, e *SetStringRequest) (*empty.Empty, error) {
	putFunc := s.db.Put
	if e.IfNotExists {
		putFunc = s.db.PutIfNotExists
	}

	return &empty.Empty{}, putFunc([]byte(e.Key), []byte(e.Value), e.Expires)
}

func (s *RedixService) SetFloat(c context.Context, e *SetFloatRequest) (*empty.Empty, error) {
	putFunc := s.db.Put
	if e.IfNotExists {
		putFunc = s.db.PutIfNotExists
	}

	return &empty.Empty{}, putFunc([]byte(e.Key), []byte(fmt.Sprintf("%f", e.Value)), e.Expires)
}

func (s *RedixService) GetString(c context.Context, e *GetRequest) (*GetStringResponse, error) {
	val, err := s.db.Get([]byte(e.Key))
	if err != nil {
		return nil, err
	}

	return &GetStringResponse{Value: string(val)}, nil
}

func (s *RedixService) GetFloat(c context.Context, e *GetRequest) (*GetFloatResponse, error) {
	val, err := s.db.Get([]byte(e.Key))
	if err != nil {
		return nil, err
	}

	floatVal, err := strconv.ParseFloat(string(val), 64)
	if err != nil {
		return nil, err
	}

	return &GetFloatResponse{Value: floatVal}, nil
}

func (s *RedixService) Incr(c context.Context, e *IncrRequest) (*IncrResponse, error) {
	val, err := s.db.Incr([]byte(e.Key), e.Delta, e.Expires)
	if err != nil {
		return nil, err
	}

	return &IncrResponse{NewValue: val}, nil
}

func (s *RedixService) Has(c context.Context, e *HasRequest) (*HasResponse, error) {
	exists, err := s.db.Has([]byte(e.Key))
	if err != nil {
		return nil, err
	}

	return &HasResponse{Exists: exists}, nil
}
