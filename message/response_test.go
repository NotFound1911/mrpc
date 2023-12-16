package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRespEncodeDecode(t *testing.T) {
	testCases := []struct {
		name string
		resp *Response
	}{
		{
			name: "normal",
			resp: &Response{
				RequestID:  111,
				Version:    11,
				Compresser: 12,
				Serializer: 13,
				Error:      []byte("error message"),
				Data:       []byte("hello, world"),
			},
		},
		{
			name: "no data",
			resp: &Response{
				RequestID:  111,
				Version:    11,
				Compresser: 12,
				Serializer: 13,
				Error:      []byte("error message"),
			},
		},
		{
			name: "no error",
			resp: &Response{
				RequestID:  111,
				Version:    11,
				Compresser: 12,
				Serializer: 13,
				Data:       []byte("hello, world"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.resp.CalHeaderLength()
			tc.resp.CalBodyLength()
			data := EncodeResp(tc.resp)
			resp := DecodeResp(data)
			assert.Equal(t, tc.resp, resp)
		})
	}
}
