# shellcheck shell=sh
Describe 'career CLI'
  Include "$SHELLSPEC_SPECDIR/spec_helper.sh"

  Describe 'no arguments'
    It 'prints root help'
      When run career
      The status should be success
      The output should include 'Usage:'
      The output should include 'generate'
    End
  End

  Describe 'version'
    It 'prints the version'
      When run career version
      The status should be success
      The output should include 'career'
    End
  End

  Describe 'templates'
    It 'lists the available templates'
      When run career templates
      The status should be success
      The output should include 'cv'
      The output should include 'japanese-resume'
      The output should include 'career-history'
    End
  End

  Describe 'help'
    It 'describes the generate command'
      When run career help generate
      The status should be success
      The output should include '--template'
    End
  End

  Describe 'unknown command'
    It 'fails with a helpful message'
      When run career frobnicate
      The status should be failure
      The stderr should include 'unknown command'
    End
  End
End
