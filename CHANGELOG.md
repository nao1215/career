# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- Add spacing below the 自己PR section in the work-history so a following section
  (e.g. リンク) is no longer crowded against it.

## [v0.2.2] - 2026-06-13

### Added

- A `career.links` section renders GitHub, blog, and other URLs as a bulleted
  list, placed last in the work-history (リンク) and cv (Links) documents.

### Changed

- The rule between company entries is omitted when the next company starts a new
  page; the page break itself separates the entries, so no rule is left dangling
  at the foot of a page.

## [v0.2.1] - 2026-06-13

### Added

- A short centered rule separates consecutive company entries in the
  `work-history` and `cv` documents.

### Changed

- A blank line between paragraphs in a text field is now kept as a blank line of
  vertical space in the output, instead of collapsing to a plain line break, so
  the paragraph gap written in the source is visible.

## [v0.2.0] - 2026-06-13

### Added

- Japanese line-breaking rules (禁則処理) for wrapped text: a line no longer
  begins with a closing bracket or sentence punctuation, nor ends with an opening
  bracket; the offending character is pushed to the adjacent line.
- Soft line wrapping in text fields: a single newline is treated as a soft wrap
  (CJK lines joined with no space, Latin words with one) and a blank line is an
  intentional paragraph break, so long fields can be wrapped in the source for
  readability without forcing breaks in the output.

### Changed

- Reworked the `work-history` and `cv` layout for readability: entries align on a
  consistent indentation rail, metadata captions (役職 / 役割・規模 / 使用技術 /
  Tech) render in gray in the body face, entry markers are tinted with the accent
  color, and each project is ordered role, tech, then description.
- README: added a preview gallery of the generated documents right after the demo
  GIF (images link to the sample PDFs) and condensed the Templates section into a
  single table.
- The `career init` scaffold comment is now written in English.

## [v0.1.0] - 2026-06-13

Initial release.

### Added

- `generate` renders a resume YAML file into a PDF, selected with `--template`.
  Bundled templates: `cv` (English curriculum vitae), `japanese-resume`
  (JIS-style 履歴書), and `work-history` (職務経歴書). The Japanese names `履歴書` /
  `職務経歴書` and the legacy `career-history` work as aliases.
- Render several documents at once: repeat `--template`, comma-separate names, or
  pass `all`. With no `--template`, `cv` is rendered.
- Multilingual text fields: any text value may be a plain scalar or a
  `{ ja:, en: }` map, so one file backs the English CV and the Japanese documents,
  each in the right language.
- `init` writes a starter resume YAML file (`--force` to overwrite), with today's
  date filled in.
- `templates` lists the available templates and their aliases.
- Portrait support for the 履歴書 via `profile.photo` (resolved relative to the
  YAML file) or `--photo` (relative to the current directory). The image is fitted
  to the 3:4 frame without distortion, warns on a mismatched aspect ratio, and
  falls back to a placeholder when the file is missing.
- `--accent` flag and a `theme.accent` YAML field for the `cv` and `work-history`
  accent color (`none` for monochrome); `japanese-resume` always renders in black.
- Long 履歴書 tables (学歴・職歴, 免許・資格) and free-text fields flow onto additional
  pages instead of dropping rows or overrunning the page; over-long values shrink
  and are ellipsis-truncated within their cell.
- Required-field validation rejects empty or whitespace-only values.
- Embedded IPAex Mincho/Gothic fonts, so PDFs render identically without font
  setup.
- Example resume files, a vhs demo GIF, and a shellspec end-to-end suite.

[Unreleased]: https://github.com/nao1215/career/compare/v0.2.2...HEAD
[v0.2.2]: https://github.com/nao1215/career/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/nao1215/career/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/nao1215/career/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/nao1215/career/releases/tag/v0.1.0
