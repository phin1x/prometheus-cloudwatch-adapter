// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package handlers

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"sync"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/go-logr/logr"
)

var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(ioutil.Discard)
	},
}

func NewRequestCompressionHandler(logger logr.Logger) request.NamedHandler {
	return request.NamedHandler{
		Name: "RequestCompressionHandler",
		Fn: func(req *request.Request) {
			buf := new(bytes.Buffer)
			g := gzipPool.Get().(*gzip.Writer)
			g.Reset(buf)
			size, err := io.Copy(g, req.GetBody())
			if err != nil {
				logger.Error(err,"I! Error occurred when trying to compress payload for operation %v, uncompressed request is sent")
				req.ResetBody()
				return
			}
			g.Close()
			compressedSize := int64(buf.Len())

			if size <= compressedSize {
				logger.Info("D! The payload is not compressed. original payload size: %v, compressed payload size: %v.", size, compressedSize)
				req.ResetBody()
				return
			}

			req.SetBufferBody(buf.Bytes())
			gzipPool.Put(g)
			req.HTTPRequest.ContentLength = compressedSize
			req.HTTPRequest.Header.Set("Content-Length", fmt.Sprintf("%d", compressedSize))
			req.HTTPRequest.Header.Set("Content-Encoding", "gzip")
		},
	}
}