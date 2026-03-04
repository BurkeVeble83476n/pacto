# pacto-plugin-schema-infer

A Pacto plugin that reads a service configuration file (JSON, YAML, or TOML) and infers a JSON Schema from it.

## Install

```bash
make install
```

## Usage

```bash
pacto generate schema-infer pacto.yaml --option file=config.yaml
```

This generates a `config.schema.json` file containing the inferred JSON Schema. Reference it in your pacto contract:

```yaml
configuration:
  schema: config.schema.json
```

## How it works

The plugin parses the configuration file and recursively infers types:

| Go type          | JSON Schema type |
|------------------|------------------|
| `string`         | `string`         |
| `float64`        | `number`         |
| `int64` / `int`  | `integer`        |
| `bool`           | `boolean`        |
| `nil`            | `null`           |
| `map[string]any` | `object`         |
| `[]any`          | `array`          |

All keys found in the sample are marked as `required`. The root object sets `additionalProperties: false`.

> **Note:** The generated schema infers types from sample values and marks all present fields as required. Review the output and relax constraints as needed for your use case.

## Build

```bash
make build    # Build the binary
make test     # Run tests
make install  # Install to /usr/local/bin
```
