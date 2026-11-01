package resolver

import "github.com/manatunga/eloque/internal/diagnostic"

// CodeUndefinedName identifies a reference to a name that has no
// visible declaration in the current lexical scope or its ancestors.
const CodeUndefinedName diagnostic.Code = "RESOLVE_UNDEFINED_NAME"

// CodeDuplicateDeclaration identifies a name declared more than once
// within the same lexical scope.
const CodeDuplicateDeclaration diagnostic.Code = "RESOLVE_DUPLICATE_DECLARATION"

// CodeShadowedDeclaration identifies a declaration that hides
// a visible declaration from an enclosing lexical scope.
const CodeShadowedDeclaration diagnostic.Code = "RESOLVE_SHADOWED_DECLARATION"