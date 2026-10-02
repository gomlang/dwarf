# Bounded DWARF reader

`ecosystem::dwarf` reads `.debug_info`, `.debug_abbrev`, `.debug_line`,
`.debug_str`, and `.debug_line_str` from bounded section data. Call
`load_at(source, spans, file_size, limits)` with a standard `std::io::ReadAt`
source, then `parse_sections(sections, endian, limits)`. The loader checks each
span against the declared source size and limits, then uses exact random-access
reads; short reads return an `Io` error. Section bytes can also be supplied
directly in `Sections`. No ELF decoder or runtime debugger is required.

The parser supports DWARF32 compilation and partial units in versions 2–5,
abbreviation tables, flattened DIEs with depth, and common address, constant,
string, flag, block, expression-byte, reference, section-offset, indirect,
and `implicit_const` forms. It resolves `.debug_str` and `.debug_line_str`
strings. For a unit with `DW_AT_stmt_list`, it parses the corresponding v2–v5
line header, file table, and line-program rows. Expression payloads are returned
as bytes, not evaluated. Limits cover source/section bytes, unit/abbreviation/
DIE/attribute counts, per-string bytes, cumulative decoded text bytes, and line
files/rows. Every materialized string is charged, including repeated `strp` or
`line_strp` references and each joined directory/file path. The budget is
checked before decoding or joining. Errors include the
section and byte offset; malformed data returns an error instead of terminating.

This is a defined subset of the [DWARF 5 specification](https://dwarfstd.org/doc/DWARF5.pdf).
`parse_sections_detailed(sections, endian, limits)` returns `DetailedData` with
the same units and `DetailedLineTable` values. Each `DetailedLineRow` contains
the original six-field `LineRow` in `row`, plus `op_index`, `basic_block`,
`prologue_end`, `epilogue_begin`, `isa`, and `discriminator`. The position is
`detail.row.address` together with `detail.op_index`; operation indexes distinguish
multiple operations at one instruction address. Discriminators distinguish
blocks sharing a source position. Transient flags and the discriminator reset
after each row; ISA persists until the sequence ends. Both parse entry points
share the same state machine, limits and error handling. `load_at` produces
`Sections` accepted by either entry point. Existing `parse_sections`, `LineRow`,
`LineTable` and `Data` retain their original shapes and behavior.

DWARF64, type/skeleton/split units, supplementary objects, indexed
`strx`/`addrx`/`loclistx`/`rnglistx` forms, `.debug_str_offsets`, `.debug_addr`,
range/location list evaluation, CFI, macro tables, accelerator tables,
segmented addresses, and `DW_LNE_define_file` in v5 are not supported. Unknown
forms and unit types return `Unsupported`. Paths are joined lexically; the
reader does not canonicalize a filesystem path or resolve a source file.

`tests/data/source.c` is compiled by `tests/data/generate.go` with GCC 13.3.0
to produce fixed v2, v4, and v5 section fixtures. That generator uses Go 1.26.0
[`debug/dwarf`](https://pkg.go.dev/debug/dwarf) to emit `reference.tsv`.
The GoML test compares unit versions, DIE names/counts, file names, and every
line row to the reference. Regeneration requires GCC and Go 1.26; ordinary
tests use only checked-in binary fixtures and the reference table.

## Development and examples

Requires GoML 0.1.56 or newer. The `examples/basic/` example shares the root manifest and its dependencies. From the library root, run:

```sh
goml run --example basic
goml test
goml verify --timeout 300s
```

`goml test` builds the example and runs its tests. `goml verify` repeats the example checks as an independent module against an isolated registry snapshot. `(cd ../verification && just ecosystem-test dwarf)` also retains the library-specific smoke and compatibility checks.
