#!/bin/bash
# Bash completion script for llc

_llc_complete() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Long and short options
    opts="-a -A -1 -i -d -h -F -L -t -u -U -S -r -R --group-directories-first
          --ignore= --gitignore --tree --json --csv --color= --no-color
          --time-style= --sort-ext --version --help -e --comment="

    case "$prev" in
        --color)
            COMPREPLY=( $(compgen -W "always auto never" -- "$cur") )
            return 0
            ;;
        --time-style)
            COMPREPLY=( $(compgen -W "default iso long-iso full-iso" -- "$cur") )
            return 0
            ;;
        -e|--comment)
            # File completion
            COMPREPLY=( $(compgen -f -- "$cur") )
            return 0
            ;;
    esac

    # If current word starts with -, complete with options
    if [[ "$cur" == -* ]]; then
        COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
        return 0
    fi

    # Default: file/directory completion
    COMPREPLY=( $(compgen -f -- "$cur") )
}

complete -F _llc_complete llc
