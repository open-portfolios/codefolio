//go:build linux

package yoga

// On Linux, libstdc++.so.6 / libpthread.so.0 / libgcc_s.so.1 are system
// libraries and must stay dynamic -- statically linking them causes C++ ABI
// conflicts with the rest of the system and breaks distribution.
//
// #cgo CFLAGS: -I${SRCDIR}/../../install/include -Wno-deprecated-declarations
// #cgo LDFLAGS: -L${SRCDIR}/../../install/lib -lyogacore -lstdc++ -lpthread
import "C"