# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- The 履歴書 no longer silently drops 学歴・職歴 or 免許・資格 rows that exceed one
  page; the tables now flow onto additional pages, and the free-text fields can
  no longer be pushed off the page.
- Required-field validation now rejects whitespace-only values (e.g. a
  `profile.name` of only spaces).
- Long personal-block values (name, address, email, phone) shrink to fit their
  cell instead of overflowing or overlapping.
- `career init` fills the date with today's date instead of a fixed value.

### Added

- Initial release of `career`.
- `generate` command rendering a YAML resume into a PDF, selected with
  `--template`. Three templates are bundled: `cv` (English curriculum vitae),
  `japanese-resume` (JIS-style 履歴書), and `career-history` (職務経歴書). The
  Japanese names `履歴書` and `職務経歴書` work as aliases.
- `--accent` flag and a `theme.accent` YAML field to set the accent color of the
  `cv` and `career-history` templates (`none` for monochrome). `japanese-resume`
  always renders in black.
- `generate` renders the `cv` template by default and can render several
  documents at once: repeat `--template`, comma-separate names, or pass `all`.
- Portrait support for the 履歴書: `profile.photo` (resolved relative to the YAML
  file) or `--photo` (relative to the current directory). The image is fitted to
  the 3:4 frame without distortion, warns on a mismatched aspect ratio, and falls
  back to the placeholder when the file is missing.
- `init` command that writes a starter resume YAML file (`--force` to overwrite).
- Multilingual text fields: any text value may be a plain scalar or a
  `{ ja:, en: }` map, so a single file renders the English CV and the Japanese
  documents each in the right language.
- `templates` command listing the available document templates.
- Embedded IPAex Mincho/Gothic fonts so PDFs render identically without font
  setup.
- Example resume files, a vhs demo GIF, and shellspec end-to-end tests.
