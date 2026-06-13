# shellcheck shell=sh
Describe 'career generate'
  Include "$SHELLSPEC_SPECDIR/spec_helper.sh"

  BeforeEach 'make_workdir'
  AfterEach 'remove_workdir'

  Describe 'rirekisho'
    It 'renders a PDF from a positional input and short flags'
      When run career generate "$WORK/resume.yaml" -t rireki -o "$WORK/out.pdf"
      The status should be success
      The output should include 'wrote'
      The path "$WORK/out.pdf" should be exist
    End

    It 'writes a valid PDF header'
      career generate "$WORK/resume.yaml" -t rirekisho -o "$WORK/out.pdf"
      When call pdf_magic "$WORK/out.pdf"
      The output should equal '%PDF'
    End
  End

  Describe 'shokumukeirekisho'
    It 'renders a PDF using --input and an explicit output'
      When run career generate --input "$WORK/resume.yaml" --template shokumukeirekisho --output "$WORK/cv.pdf"
      The status should be success
      The output should include 'wrote'
      The path "$WORK/cv.pdf" should be exist
    End

    It 'accepts the cv alias'
      When run career generate "$WORK/resume.yaml" -t cv -o "$WORK/cv.pdf"
      The status should be success
      The output should include 'wrote'
      The path "$WORK/cv.pdf" should be exist
    End
  End

  Describe 'errors'
    It 'fails when the input file is missing'
      When run career generate "$WORK/nope.yaml" -t rireki -o "$WORK/out.pdf"
      The status should be failure
      The stderr should be present
    End

    It 'fails on an unknown template'
      When run career generate "$WORK/resume.yaml" -t bogus -o "$WORK/out.pdf"
      The status should be failure
      The stderr should include 'unknown template'
    End

    It 'fails when no input is given'
      When run career generate -t rireki
      The status should be failure
      The stderr should include 'no input file'
    End
  End
End
