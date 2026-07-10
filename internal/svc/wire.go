package svc

import "github.com/google/wire"

var Set = wire.NewSet(NewAgent)
