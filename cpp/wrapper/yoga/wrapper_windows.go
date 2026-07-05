//go:build windows

package yoga

// MinGW runtime libraries. These flags work with any MinGW-style toolchain on
// Windows (MinGW-w64 GCC variants, LLVM/Clang with --target=x86_64-w64-windows-gnu,
// etc.) because they all delegate to GNU ld or lld which accept -Wl,-Bstatic.
//
// Yoga is built as a C++ static lib (libyogacore.a), so it needs libstdc++ at
// link time. Yoga itself does not use threading, but libstdc++ may pull in
// pthread symbols; we link all three runtime libs statically so the resulting
// Go exe does not require libstdc++-6.dll / libgcc_s_seh-1.dll /
// libwinpthread-1.dll next to it.
//
// ntdll / KERNEL32 / KERNELBASE / ucrtbase are Windows OS DLLs and cannot be
// statically linked into the user exe -- they will always appear in `ldd`.

// #cgo CFLAGS: -I${SRCDIR}/../../install/include -Wno-deprecated-declarations
// #cgo LDFLAGS: -L${SRCDIR}/../../install/lib -lyogacore -Wl,-Bstatic -lstdc++ -lpthread -lgcc -Wl,-Bdynamic
import "C"