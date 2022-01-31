# hlb-parser
Playground to make changes to parser without worrying about checker/codegen.

# Proposal

Goals:
- Address inconsistency across builtin effects and user defined `as ( ... )` binding.
- Address inconsistency between all decls except function declaration (starting with keyword).
- Address inconsistency between `CallStmt` and `CallExpr`.
- Support if else, for loops, composite types, unary / binary operators (with operator precendence and associativity).
- Treat type association `::` as first class rather than simply legal characters in `Ident` symbol.
