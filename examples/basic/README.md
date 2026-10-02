# DWARF example

This example exercises the public `ecosystem::dwarf` API with a
small DWARF4 compilation unit, including both section parsing and `ReadAt`
loading. It checks a resource-limit error through the public interface.

This example shares the library root manifest and its dependencies. From the library root, run `goml verify --example basic` to build and test it as an independent downstream module.
