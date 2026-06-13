# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial release of `career`.
- `generate` command rendering a YAML resume into a JIS-style 履歴書 or a
  flowing 職務経歴書 PDF, selected with `--template`.
- `templates` command listing the available document templates.
- Embedded IPAex Mincho/Gothic fonts so PDFs render identically without font
  setup.
- Example resume files and shellspec end-to-end tests.
