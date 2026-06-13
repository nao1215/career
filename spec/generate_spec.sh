# shellcheck shell=sh
Describe 'career generate'
  Include "$SHELLSPEC_SPECDIR/spec_helper.sh"

  BeforeEach 'make_workdir'
  AfterEach 'remove_workdir'

  Describe 'cv'
    It 'renders a PDF (default template) with an accent color'
      When run career generate "$WORK/resume.yaml" --accent "#2c6e6e" -o "$WORK/cv.pdf"
      The status should be success
      The output should include 'wrote'
      The path "$WORK/cv.pdf" should be exist
    End

    It 'writes a valid PDF header'
      career generate "$WORK/resume.yaml" -t cv -o "$WORK/cv.pdf"
      When call pdf_magic "$WORK/cv.pdf"
      The output should equal '%PDF'
    End
  End

  Describe 'japanese-resume'
    It 'renders a PDF from a positional input'
      When run career generate "$WORK/resume.yaml" -t japanese-resume -o "$WORK/out.pdf"
      The status should be success
      The output should include 'wrote'
      The path "$WORK/out.pdf" should be exist
    End

    It 'accepts the 履歴書 alias'
      When run career generate "$WORK/resume.yaml" -t 履歴書 -o "$WORK/out.pdf"
      The status should be success
      The output should include 'wrote'
      The path "$WORK/out.pdf" should be exist
    End

    It 'embeds the bundled sample portrait passed with --photo'
      When run career generate "$WORK/resume.yaml" -t japanese-resume --photo "$PROJECT_ROOT/image/sample_japanese_man.jpg" -o "$WORK/out.pdf"
      The status should be success
      The output should include 'wrote'
      The path "$WORK/out.pdf" should be exist
    End
  End

  Describe 'work-history'
    It 'renders a PDF using --input and the default output name'
      When run career generate --input "$WORK/resume.yaml" --template work-history --output "$WORK/ch.pdf"
      The status should be success
      The output should include 'wrote'
      The path "$WORK/ch.pdf" should be exist
    End

    It 'still accepts the legacy career-history alias'
      When run career generate "$WORK/resume.yaml" -t career-history -o "$WORK/legacy.pdf"
      The status should be success
      The output should include 'work-history'
      The path "$WORK/legacy.pdf" should be exist
    End

    It 'accepts the 職務経歴書 alias'
      When run career generate "$WORK/resume.yaml" -t 職務経歴書 -o "$WORK/ja.pdf"
      The status should be success
      The output should include 'wrote'
      The path "$WORK/ja.pdf" should be exist
    End
  End

  Describe 'multiple templates'
    It 'renders every template with -t all'
      # Run inside WORK so the default output files land there, not in the repo.
      When run sh -c "cd '$WORK' && '$CAREER_BIN' generate resume.yaml -t all"
      The status should be success
      The output should include 'cv'
      The output should include 'japanese-resume'
      The output should include 'work-history'
      The path "$WORK/cv.pdf" should be exist
      The path "$WORK/japanese-resume.pdf" should be exist
      The path "$WORK/work-history.pdf" should be exist
    End
  End

  Describe 'errors'
    It 'fails when the input file is missing'
      When run career generate "$WORK/nope.yaml" -t cv -o "$WORK/out.pdf"
      The status should be failure
      The stderr should be present
    End

    It 'fails on an unknown template'
      When run career generate "$WORK/resume.yaml" -t bogus -o "$WORK/out.pdf"
      The status should be failure
      The stderr should include 'unknown template'
    End

    It 'fails on an invalid accent color'
      When run career generate "$WORK/resume.yaml" -t cv --accent bogus -o "$WORK/out.pdf"
      The status should be failure
      The stderr should include 'hex color'
    End

    It 'fails when no input is given'
      When run career generate -t cv
      The status should be failure
      The stderr should include 'no input file'
    End
  End
End
