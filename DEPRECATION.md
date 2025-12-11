Write helpers disabled
======================

Summary
-------

The package's built-in file-write helpers (`MarshalToFile`, `MarshalToWriter`) are intentionally internalized and no longer exported. Public mutating methods such as `UpdateVariable`, `UpdateConnectionString`, `UpdateExpression`, and `UpdateProperty` have been removed from the public API (internal versions remain), so programmatic mutation should be done by editing package structs directly or by using internal APIs.

Rationale
---------

Temporarily disabling write capabilities reduces accidental changes to on-disk DTSX packages when the library is used for analysis, parsing, or inspection.

Migration / How to persist or mutate packages
--------------------------------------------

- To write a package to disk, serialize with `dtsx.Marshal(pkg)` then write the bytes with the standard library:

```go
data, err := dtsx.Marshal(pkg)
if err != nil { /* handle */ }
err = os.WriteFile("output.dtsx", data, 0644)
```

- To modify a package programmatically, mutate the package structs directly (e.g., update `pkg.Property`, `pkg.Variables.Variable[...]`, or `pkg.ConnectionManagers.ConnectionManager[...]`) and then call `dtsx.Marshal` to serialize and persist.

Notes for maintainers
---------------------

- Internally, original implementations have been moved to unexported functions (e.g., `marshalToFile`, `updateVariable`); these can be re-exposed or guarded behind a feature flag if write capabilities need to be restored.
// Exported wrappers were removed to reflect the new focus on read/analysis-only API. Use internal functions or direct struct modification where necessary.
