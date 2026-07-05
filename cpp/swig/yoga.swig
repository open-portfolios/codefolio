%module yoga

%{
#include <yoga/Yoga.h>
%}

// Order matters: include the macros header first so that YG_ENUM_DECL and
// friends are defined, then neutralize GCC-attribute macros that swig's
// plain C parser cannot handle.
//
// swig 4.x plain C parser does not support:
//   - __attribute__((deprecated("msg"))) with a string-literal argument
//   - __attribute__((visibility("default"))) and other GCC function attributes
//
// We redefine YG_EXPORT / YG_EXTERN_C_BEGIN / END / YG_DEPRECATED to empty
// tokens so they vanish from the code seen by swig.

%include "yoga/YGMacros.h"

#undef YG_EXPORT
#define YG_EXPORT

#undef YG_EXTERN_C_BEGIN
#define YG_EXTERN_C_BEGIN

#undef YG_EXTERN_C_END
#define YG_EXTERN_C_END

// swig %define accepts variadic args, so we can swallow the whole
// YG_DEPRECATED("msg") call site.
#undef YG_DEPRECATED
%define YG_DEPRECATED(...)
%enddef

// Dependency order: bottom macros/enums/structs first, top-level APIs last.
%include "yoga/YGEnums.h"
%include "yoga/YGValue.h"
%include "yoga/YGConfig.h"
%include "yoga/YGNode.h"
%include "yoga/YGNodeStyle.h"
%include "yoga/YGNodeLayout.h"
%include "yoga/YGPixelGrid.h"