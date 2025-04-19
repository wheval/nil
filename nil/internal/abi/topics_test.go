// Copyright 2019 The go-ethereum Authors
// (original work)
// Copyright 2024 The Erigon Authors
// (modifications)
// This file is part of Erigon.
//
// Erigon is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Erigon is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Erigon. If not, see <http://www.gnu.org/licenses/>.

package abi

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/types"
)

func TestMakeTopics(t *testing.T) { //nolint:tparallel
	t.Parallel()

	type args struct {
		query [][]any
	}
	tests := []struct {
		name    string
		args    args
		want    [][]common.Hash
		wantErr bool
	}{
		{
			"support fixed byte types, right padded to 32 bytes",
			args{[][]any{{[5]byte{1, 2, 3, 4, 5}}}},
			[][]common.Hash{{common.Hash{1, 2, 3, 4, 5}}},
			false,
		},
		{
			"support common.Hash types in topics",
			args{[][]any{{common.Hash{1, 2, 3, 4, 5}}}},
			[][]common.Hash{{common.Hash{1, 2, 3, 4, 5}}},
			false,
		},
		{
			"support address types in topics",
			args{[][]any{{types.Address{1, 2, 3, 4, 5}}}},
			[][]common.Hash{{common.Hash{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5}}},
			false,
		},
		{
			"support *big.Int types in topics",
			args{[][]any{{big.NewInt(1).Lsh(big.NewInt(2), 254)}}},
			[][]common.Hash{{common.Hash{128}}},
			false,
		},
		{
			"support boolean types in topics",
			args{[][]any{
				{true},
				{false},
			}},
			[][]common.Hash{
				{common.Hash{
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
				}},
				{common.Hash{0}},
			},
			false,
		},
		{
			"support int/uint(8/16/32/64) types in topics",
			args{[][]any{
				{int8(-2)},
				{int16(-3)},
				{int32(-4)},
				{int64(-5)},
				{int8(1)},
				{int16(256)},
				{int32(65536)},
				{int64(4294967296)},
				{uint8(1)},
				{uint16(256)},
				{uint32(65536)},
				{uint64(4294967296)},
			}},
			[][]common.Hash{
				{common.Hash{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 254,
				}},
				{common.Hash{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 253,
				}},
				{common.Hash{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 252,
				}},
				{common.Hash{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 251,
				}},
				{common.Hash{
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
				}},
				{common.Hash{
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0,
				}},
				{common.Hash{
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0,
				}},
				{common.Hash{
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0,
				}},
				{common.Hash{
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
				}},
				{common.Hash{
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0,
				}},
				{common.Hash{
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0,
				}},
				{common.Hash{
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0,
				}},
			},
			false,
		},
		{
			"support string types in topics",
			args{[][]any{{"hello world"}}},
			[][]common.Hash{{common.Keccak256Hash([]byte("hello world"))}},
			false,
		},
		{
			"support byte slice types in topics",
			args{[][]any{{[]byte{1, 2, 3}}}},
			[][]common.Hash{{common.Keccak256Hash([]byte{1, 2, 3})}},
			false,
		},
	}
	for _, tt := range tests { //nolint:paralleltest
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeTopics(tt.args.query...)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeTopics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeTopics() = %v, want %v", got, tt.want)
			}
		})
	}
}

type args struct {
	createObj func() any
	resultObj func() any
	resultMap func() map[string]any
	fields    Arguments
	topics    []common.Hash
}

type bytesStruct struct {
	StaticBytes [5]byte
}
type int8Struct struct {
	Int8Value int8
}
type int256Struct struct {
	Int256Value *big.Int
}

type hashStruct struct {
	HashValue common.Hash
}

type funcStruct struct {
	FuncValue [24]byte
}

type topicTest struct {
	name    string
	args    args
	wantErr bool
}

func setupTopicsTests() []topicTest {
	bytesType, _ := NewType("bytes5", "", nil)
	int8Type, _ := NewType("int8", "", nil)
	int256Type, _ := NewType("int256", "", nil)
	tupleType, _ := NewType("tuple(int256,int8)", "", nil)
	stringType, _ := NewType("string", "", nil)
	funcType, _ := NewType("function", "", nil)

	tests := []topicTest{
		{
			name: "support fixed byte types, right padded to 32 bytes",
			args: args{
				createObj: func() any { return &bytesStruct{} },
				resultObj: func() any { return &bytesStruct{StaticBytes: [5]byte{1, 2, 3, 4, 5}} },
				resultMap: func() map[string]any {
					return map[string]any{"staticBytes": [5]byte{1, 2, 3, 4, 5}}
				},
				fields: Arguments{Argument{
					Name:    "staticBytes",
					Type:    bytesType,
					Indexed: true,
				}},
				topics: []common.Hash{
					{1, 2, 3, 4, 5},
				},
			},
			wantErr: false,
		},
		{
			name: "int8 with negative value",
			args: args{
				createObj: func() any { return &int8Struct{} },
				resultObj: func() any { return &int8Struct{Int8Value: -1} },
				resultMap: func() map[string]any {
					return map[string]any{"int8Value": int8(-1)}
				},
				fields: Arguments{Argument{
					Name:    "int8Value",
					Type:    int8Type,
					Indexed: true,
				}},
				topics: []common.Hash{
					{
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "int256 with negative value",
			args: args{
				createObj: func() any { return &int256Struct{} },
				resultObj: func() any { return &int256Struct{Int256Value: big.NewInt(-1)} },
				resultMap: func() map[string]any {
					return map[string]any{"int256Value": big.NewInt(-1)}
				},
				fields: Arguments{Argument{
					Name:    "int256Value",
					Type:    int256Type,
					Indexed: true,
				}},
				topics: []common.Hash{
					{
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "hash type",
			args: args{
				createObj: func() any { return &hashStruct{} },
				resultObj: func() any { return &hashStruct{common.Keccak256Hash([]byte("stringtopic"))} },
				resultMap: func() map[string]any {
					return map[string]any{"hashValue": common.Keccak256Hash([]byte("stringtopic"))}
				},
				fields: Arguments{Argument{
					Name:    "hashValue",
					Type:    stringType,
					Indexed: true,
				}},
				topics: []common.Hash{
					common.Keccak256Hash([]byte("stringtopic")),
				},
			},
			wantErr: false,
		},
		{
			name: "function type",
			args: args{
				createObj: func() any { return &funcStruct{} },
				resultObj: func() any {
					return &funcStruct{[24]byte{
						255, 255, 255, 255, 255, 255, 255, 255,
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					}}
				},
				resultMap: func() map[string]any {
					return map[string]any{"funcValue": [24]byte{
						255, 255, 255, 255, 255, 255, 255, 255,
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					}}
				},
				fields: Arguments{Argument{
					Name:    "funcValue",
					Type:    funcType,
					Indexed: true,
				}},
				topics: []common.Hash{
					{
						0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 255, 255, 255, 255, 255, 255,
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "error on topic/field count mismatch",
			args: args{
				createObj: func() any { return nil },
				resultObj: func() any { return nil },
				resultMap: func() map[string]any { return make(map[string]any) },
				fields: Arguments{Argument{
					Name:    "tupletype",
					Type:    tupleType,
					Indexed: true,
				}},
				topics: []common.Hash{},
			},
			wantErr: true,
		},
		{
			name: "error on unindexed arguments",
			args: args{
				createObj: func() any { return &int256Struct{} },
				resultObj: func() any { return &int256Struct{} },
				resultMap: func() map[string]any { return make(map[string]any) },
				fields: Arguments{Argument{
					Name:    "int256Value",
					Type:    int256Type,
					Indexed: false,
				}},
				topics: []common.Hash{
					{
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "error on tuple in topic reconstruction",
			args: args{
				createObj: func() any { return &tupleType },
				resultObj: func() any { return &tupleType },
				resultMap: func() map[string]any { return make(map[string]any) },
				fields: Arguments{Argument{
					Name:    "tupletype",
					Type:    tupleType,
					Indexed: true,
				}},
				topics: []common.Hash{{0}},
			},
			wantErr: true,
		},
		{
			name: "error on improper encoded function",
			args: args{
				createObj: func() any { return &funcStruct{} },
				resultObj: func() any { return &funcStruct{} },
				resultMap: func() map[string]any {
					return make(map[string]any)
				},
				fields: Arguments{Argument{
					Name:    "funcValue",
					Type:    funcType,
					Indexed: true,
				}},
				topics: []common.Hash{
					{
						0, 0, 0, 0, 0, 0, 0, 128, 255, 255, 255, 255, 255, 255, 255, 255,
						255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
					},
				},
			},
			wantErr: true,
		},
	}

	return tests
}

func TestParseTopics(t *testing.T) { //nolint:tparallel
	t.Parallel()

	tests := setupTopicsTests()

	for _, tt := range tests { //nolint:paralleltest
		t.Run(tt.name, func(t *testing.T) {
			createObj := tt.args.createObj()
			if err := ParseTopics(createObj, tt.args.fields, tt.args.topics); (err != nil) != tt.wantErr {
				t.Errorf("parseTopics() error = %v, wantErr %v", err, tt.wantErr)
			}
			resultObj := tt.args.resultObj()
			if !reflect.DeepEqual(createObj, resultObj) {
				t.Errorf("parseTopics() = %v, want %v", createObj, resultObj)
			}
		})
	}
}

func TestParseTopicsIntoMap(t *testing.T) { //nolint:tparallel
	t.Parallel()

	tests := setupTopicsTests()

	for _, tt := range tests { //nolint:paralleltest
		t.Run(tt.name, func(t *testing.T) {
			outMap := make(map[string]any)
			if err := ParseTopicsIntoMap(outMap, tt.args.fields, tt.args.topics); (err != nil) != tt.wantErr {
				t.Errorf("parseTopicsIntoMap() error = %v, wantErr %v", err, tt.wantErr)
			}
			resultMap := tt.args.resultMap()
			if !reflect.DeepEqual(outMap, resultMap) {
				t.Errorf("parseTopicsIntoMap() = %v, want %v", outMap, resultMap)
			}
		})
	}
}
