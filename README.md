# heap-string

Reads a file's contents using unix syscalls, such as mmap, mlock, madvise, returning a buffer without string allocations.

File must be mode 400 and owned by the caller.

## Exports

- FromFile
- Get
- GetRaw
- Free
- Wipe

<hr>

### `FromFile(path string) (error, *Buffer)`
**Description:** Reads the file and returns the buffer.

**Args:**
- `path` - path to file

**Returns:**
- `error`
- `*Buffer`

<hr>

### `(*Buffer) Get() []byte`
**Description:** Returns the bytes of the buffer without a trailing `\n`.

**Returns:**
- `[]byte`

<hr>

### `(*Buffer) GetRaw() []byte`
**Description:** Returns the raw bytes of the buffer.

**Returns:**
- `[]byte`

<hr>

### `(*Buffer) Free() error`
**Description:** Wipes the buffer and frees allocations.

**Returns:**
- `error`

<hr>

### `(*Buffer) Wipe()`
**Description:** Wipes the buffer.
