package server

import "github.com/oxtoacart/bpool"
import hhelper "github.com/m3ng9i/go-utils/http"


// TODO set number of buffer in config
// global buffer pool for ran server
var bufferPool = bpool.NewBufferPool(200)

// a function to generate a 12 characters random request id.
var getRequestId = hhelper.RequestIdGenerator(12)

