# career

[![Build](https://github.com/nao1215/career/actions/workflows/build.yml/badge.svg)](https://github.com/nao1215/career/actions/workflows/build.yml)
[![MultiPlatformUnitTest](https://github.com/nao1215/career/actions/workflows/unit_test.yml/badge.svg)](https://github.com/nao1215/career/actions/workflows/unit_test.yml)
[![E2E](https://github.com/nao1215/career/actions/workflows/e2e_test.yml/badge.svg)](https://github.com/nao1215/career/actions/workflows/e2e_test.yml)
[![reviewdog](https://github.com/nao1215/career/actions/workflows/reviewdog.yml/badge.svg)](https://github.com/nao1215/career/actions/workflows/reviewdog.yml)

career generates résumé PDFs from a single YAML file. Write your career once in
plain text, keep it under version control, and render a polished PDF with one
command.

It ships with three templates and a small template registry that is designed to
grow:

- `cv` — an English curriculum vitae / résumé.
- `japanese-resume` — a JIS-style Japanese 履歴書.
- `career-history` — a Japanese 職務経歴書 (work history).

Adding a new format (another layout, paper size, or language) means registering
one more template, so the tool is not limited to these three.

![demo](./image/demo.gif)

## Install

```bash
go install github.com/nao1215/career@latest
```

Or build from source:

```bash
git clone https://github.com/nao1215/career.git
cd career
make build   # produces ./career
```

## Usage

```bash
career generate INPUT.yaml --template NAME [--output OUT.pdf] [--accent COLOR]
```

The input file may be the first argument or `--input`, and short flags exist
(`-t`, `-i`, `-o`). `career templates` lists every template.

```bash
career generate examples/cv.yaml -t cv -o cv.pdf
```

| Command | Description |
| :--- | :--- |
| `career generate` | Render a resume YAML file into a PDF |
| `career templates` | List the available document templates |
| `career version` | Print the version |
| `career help [command]` | Show help |

## Templates

| Name | Aliases | Output |
| :--- | :--- | :--- |
| `cv` | | English curriculum vitae / résumé |
| `japanese-resume` | `履歴書` | JIS-style Japanese 履歴書 (A4, 2 pages) |
| `career-history` | `職務経歴書` | Japanese 職務経歴書 (work history) |

### cv

An English résumé: name and contact header, then Summary, Skills, Experience,
Education, Certifications, and Publications. See
[`examples/cv.yaml`](./examples/cv.yaml).

| Page 1 | Page 2 |
| :---: | :---: |
| ![cv page 1](./image/cv-p-1.png) | ![cv page 2](./image/cv-p-2.png) |

Download: [`image/cv-sample.pdf`](./image/cv-sample.pdf)

### japanese-resume (履歴書)

The conventional two-page A4 履歴書: photo box, personal block, and the
学歴・職歴 and 免許・資格 tables. This template always renders in black, as a
formal Japanese form should.

| Page 1 | Page 2 |
| :---: | :---: |
| ![japanese-resume page 1](./image/japanese-resume-p-1.png) | ![japanese-resume page 2](./image/japanese-resume-p-2.png) |

Download: [`image/japanese-resume-sample.pdf`](./image/japanese-resume-sample.pdf)

### career-history (職務経歴書)

A flowing 職務経歴書: 職務要約, skills, per-company project history, 資格,
出版, and 自己PR, with automatic page breaks.

| Page 1 | Page 2 |
| :---: | :---: |
| ![career-history page 1](./image/career-history-p-1.png) | ![career-history page 2](./image/career-history-p-2.png) |

Download: [`image/career-history-sample.pdf`](./image/career-history-sample.pdf)

The Japanese examples are rendered from [`examples/resume.yaml`](./examples/resume.yaml).

## Accent color

The `cv` and `career-history` templates use a single accent color for headings.
Set it in YAML or override it on the command line; `japanese-resume` ignores it
and stays black.

```yaml
theme:
  accent: "#1f4e79"   # "" = default slate blue, "none" = monochrome, or any #rrggbb
```

```bash
career generate examples/cv.yaml -t cv --accent "#2c6e6e"   # custom
career generate examples/cv.yaml -t cv --accent none        # monochrome
```

## Writing your resume

Start from [`examples/minimal.yaml`](./examples/minimal.yaml), or
[`examples/resume.yaml`](./examples/resume.yaml) for a fully-commented template.

```yaml
profile:
  name: Taro Mihon
  email: taro.mihon@example.com
  phone: "+81 90-1234-5678"
  address:
    text: Tokyo, Japan

education:
  - { year: 2014, month: 4, value: "B.Eng., Example University" }

career:
  summary: |
    Software engineer focused on backend and cloud infrastructure.
  skills:
    - Backend development in Go
  histories:
    - company: Example Inc.
      period: 2018 - Present
      role: Software Engineer
      projects:
        - title: Platform API
          description: Designed and built the service API.
          tech: [Go, AWS]
```

`year` and `month` accept both numbers (`2018`) and strings (`"20XX"`).
Multi-line fields use YAML block scalars (`|`).

The Japanese 履歴書 also reads a `profile` (with optional `photo`, `gender`,
`birth_date`) and a `rireki` section (`hobby`, `motivation`, `request`, and so
on). See `examples/resume.yaml` for the full set of fields.

## Development

```bash
make tools     # install golangci-lint, octocov, shellspec
make test      # unit tests with coverage
make lint      # golangci-lint
make test-e2e  # shellspec end-to-end tests against the built binary
make build     # build ./career
make demo      # regenerate image/demo.gif (needs vhs)
```

## Fonts and license

career embeds the [IPAex fonts](https://moji.or.jp/ipafont/) (IPAex Mincho and
IPAex Gothic), distributed under the IPA Font License Agreement v1.0. The license
text ships with the fonts under
[`internal/font/assets`](./internal/font/assets).

The career source code is released under the [MIT License](./LICENSE).

## Acknowledgements

The 履歴書 layout is inspired by
[kaityo256/yaml_cv](https://github.com/kaityo256/yaml_cv), a YAML-driven résumé
generator written in Ruby.
