# Nutty's shell integration for zsh.
#
# Nutty starts zsh with ZDOTDIR set to the directory that holds this file
# as its .zshenv, so zsh reads it before the user's own startup files. It
# puts ZDOTDIR back, reads the user's .zshenv, and in an interactive shell
# adds a hook that tells the terminal the working directory (OSC 7), which
# the tab, the title and the info panel follow. The stock zsh of macOS
# reports its directory to Terminal.app only.

# Put ZDOTDIR back as the user had it: zsh reads .zprofile, .zshrc and
# .zlogin from where it points once this file ends.
if [[ -n ${NUTTY_ZSH_ZDOTDIR+X} ]]; then
    builtin export ZDOTDIR=$NUTTY_ZSH_ZDOTDIR
    builtin unset NUTTY_ZSH_ZDOTDIR
else
    builtin unset ZDOTDIR
fi

{
    # The user's own .zshenv, which zsh would have read in place of this
    # file. It may move ZDOTDIR itself.
    builtin typeset _nutty_file=${ZDOTDIR-$HOME}/.zshenv
    [[ ! -r $_nutty_file ]] || builtin source -- $_nutty_file
} always {
    builtin unset _nutty_file
    if [[ -o interactive ]]; then
        _nutty_report_cwd() {
            builtin emulate -L zsh
            # To the terminal itself: a prompt may have redirected stdout
            # while it draws, and a shell without a terminal has no one to
            # tell.
            [[ -w $TTY ]] || builtin return 0
            # Percent-encode the path byte by byte, as a file URL wants it.
            builtin local LC_ALL=C i ch url=
            for (( i = 1; i <= ${#PWD}; i++ )); do
                ch=$PWD[i]
                if [[ $ch == [/._~A-Za-z0-9-] ]]; then
                    url+=$ch
                else
                    url+=%${(l:2::0:)$(( [##16] #ch ))}
                fi
            done
            builtin printf '\e]7;file://%s%s\a' "$HOST" "$url" >> $TTY
        }
        # Before each prompt, and at each cd, so the directory follows a
        # cd in the middle of a command line too.
        precmd_functions+=(_nutty_report_cwd)
        chpwd_functions+=(_nutty_report_cwd)
    fi
}
