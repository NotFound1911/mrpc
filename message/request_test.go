package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	testCases := []struct {
		name string
		req  *Request
	}{
		{
			name: "normal",
			req: &Request{
				//HeadLength:  120,
				RequestID:   111,
				Version:     11,
				Compresser:  12,
				Serializer:  13,
				ServiceName: "user-service",
				MethodName:  "GetById",
				Meta: map[string]string{
					"trace-id": "123",
					"test":     "aa",
				},
				Data: []byte("hello world"),
			},
		},
		{
			name: "data with \n ",
			req: &Request{
				//HeadLength:  120,
				RequestID:   111,
				Version:     11,
				Compresser:  12,
				Serializer:  13,
				ServiceName: "user-service",
				MethodName:  "GetById",
				Meta: map[string]string{
					"trace-id": "123",
					"test":     "aa",
				},
				Data: []byte("hello \n world"),
			},
		},
		{
			name: "no meta",
			req: &Request{
				RequestID:   111,
				Version:     11,
				Compresser:  12,
				Serializer:  13,
				ServiceName: "user-service",
				MethodName:  "GetById",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.CalHeaderLen()
			tc.req.CalBodyLen()
			data := EncodeReq(tc.req)
			req := DecodeReq(data)
			assert.Equal(t, tc.req, req)
		})
	}
}
