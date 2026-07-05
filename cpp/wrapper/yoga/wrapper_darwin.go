//go:build darwin

package yoga

// macOS ships libstdc++/libc++ and pthread in libSystem; there is no separate
// libpthread or libgcc. Yoga's C++ runtime comes from libc++ on Apple platforms.
//
// #cgo CFLAGS: -I${SRCDIR}/../../install/include -Wno-deprecated-declarations
// #cgo LDFLAGS: -L${SRCDIR}/../../install/lib -lyogacore -lc++
import "C"