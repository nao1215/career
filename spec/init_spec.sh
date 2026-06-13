# shellcheck shell=sh
Describe 'career init'
  Include "$SHELLSPEC_SPECDIR/spec_helper.sh"

  setup() { WORK=$(mktemp -d); }
  cleanup() { rm -rf "$WORK"; }
  BeforeEach 'setup'
  AfterEach 'cleanup'

  It 'writes a starter file'
    When run career init "$WORK/resume.yaml"
    The status should be success
    The output should include 'wrote'
    The path "$WORK/resume.yaml" should be exist
  End

  It 'refuses to overwrite without --force'
    career init "$WORK/resume.yaml"
    When run career init "$WORK/resume.yaml"
    The status should be failure
    The stderr should include 'already exists'
  End

  It 'overwrites with --force'
    career init "$WORK/resume.yaml"
    When run career init "$WORK/resume.yaml" --force
    The status should be success
    The output should include 'wrote'
  End

  It 'produces a file that generate accepts'
    career init "$WORK/resume.yaml"
    When run career generate "$WORK/resume.yaml" -t cv -o "$WORK/cv.pdf"
    The status should be success
    The output should include 'wrote'
    The path "$WORK/cv.pdf" should be exist
  End
End
