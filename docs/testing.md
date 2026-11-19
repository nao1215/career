# Testing

Run everything from the repository root:

```sh
make lint
make test
make e2e
make bench
```

## Unit tests

The `*_test.go` files next to each source file check the YAML loader and validation (`internal/resume`), the template registry, the accent color, the photo handling and the rendered PDFs (`internal/pdf`), and the command line (`cmd`). CI runs them on Linux, macOS and Windows, on the go.mod floor and on the newest Go (`.github/workflows/unit_test.yml`).

## Generated properties

[`internal/resume/properties_test.go`](../internal/resume/properties_test.go) uses [rapid](https://pkg.go.dev/pgregory.net/rapid) to generate language maps and scalars, writes them as YAML, decodes them into `Text` and checks `For` against an independently written copy of the documented fallback order (the asked language, then ja, then en, then the rest in key order). The properties run in the ordinary `go test`; a failure prints the seed and a `-rapid.failfile` to replay it.

## Metamorphic relations

[`internal/pdf/properties_test.go`](../internal/pdf/properties_test.go) relates two layout calls whose results must agree without knowing either one: folding two paragraphs separately and joining them with a blank line equals folding them joined, a wider column never needs more lines than a narrower one, and every line `wrap` returns stays one line when wrapped again at the same width. The generator mixes Latin words, kana, ideographs, the punctuation and brackets of 禁則処理, bullets and newlines.

## Fuzzing

`FuzzParse` (`internal/resume`) feeds arbitrary YAML to the loader. `FuzzParseHexColor`, `FuzzFold` and `FuzzWrap` (`internal/pdf`) check the accent color against `encoding/hex`, that fold is idempotent and keeps every visible character, and that each wrapped line fits the column and keeps the text in order. `.github/workflows/fuzz.yml` runs each target nightly.

## CLI end-to-end tests

`e2e/atago/*.atago.yaml` runs the built `career` binary with [atago](https://github.com/nao1215/atago) on Linux, macOS and Windows: subcommands, flags, exit codes, the files `generate` and `init` write, and the commands shown in the README (`make e2e`, `.github/workflows/e2e_test.yml`).

## Paired CI benchmarks

`bench/himorime.yaml` measures `career version` and `career generate` on generated résumés with [himorime](https://github.com/nao1215/himorime). On a pull request, `.github/workflows/bench.yml` builds the base and the head and measures them in alternating rounds on the same runner. See [bench/README.md](../bench/README.md).
