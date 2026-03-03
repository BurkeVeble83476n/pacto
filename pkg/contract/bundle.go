package contract

import "io/fs"

// Bundle represents a contract bundled with its referenced files.
type Bundle struct {
	Contract *Contract
	RawYAML  []byte // Original YAML bytes; populated for local reads.
	FS       fs.FS
}
