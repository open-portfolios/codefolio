package domain

import "github.com/google/wire"

var Set = wire.NewSet(NewSession)
