# shellcheck shell=sh
# Runs the representative commands shown in the README against the bundled
# examples/resume.yaml so the documented examples cannot silently rot.
Describe 'README examples'
  Include "$SHELLSPEC_SPECDIR/spec_helper.sh"

  setup() {
    WORK=$(mktemp -d)
    cp "$EXAMPLES/resume.yaml" "$WORK/resume.yaml"
  }
  cleanup() { rm -rf "$WORK"; }
  BeforeEach 'setup'
  AfterEach 'cleanup'

  It 'cv example: career generate resume.yaml -t cv -o cv.pdf'
    When run sh -c "cd '$WORK' && '$CAREER_BIN' generate resume.yaml -t cv -o cv.pdf"
    The status should be success
    The output should include 'wrote'
    The path "$WORK/cv.pdf" should be exist
  End

  It 'japanese-resume example: -t japanese-resume -o rirekisho.pdf'
    When run sh -c "cd '$WORK' && '$CAREER_BIN' generate resume.yaml -t japanese-resume -o rirekisho.pdf"
    The status should be success
    The output should include 'wrote'
    The path "$WORK/rirekisho.pdf" should be exist
  End

  It 'work-history example: -t work-history -o shokumukeirekisho.pdf'
    When run sh -c "cd '$WORK' && '$CAREER_BIN' generate resume.yaml -t work-history -o shokumukeirekisho.pdf"
    The status should be success
    The output should include 'wrote'
    The path "$WORK/shokumukeirekisho.pdf" should be exist
  End

  It 'all example: -t all writes the three default file names'
    When run sh -c "cd '$WORK' && '$CAREER_BIN' generate resume.yaml -t all"
    The status should be success
    The output should include 'wrote'
    The path "$WORK/cv.pdf" should be exist
    The path "$WORK/japanese-resume.pdf" should be exist
    The path "$WORK/work-history.pdf" should be exist
  End
End
