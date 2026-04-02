#compdef llc

# Zsh completion script for llc

_llc() {
    local curcontext="$curcontext" state line
    typeset -A opt_args

    _arguments -C \
        '(-a --all)'{-a,--all}'[Show all files including hidden (. and ..)]' \
        '(-A --almost-all)'{-A,--almost-all}'[Show all files except . and ..]' \
        '-1[Single column output]' \
        '-i[Show inode number]' \
        '-d[List directory itself instead of contents]' \
        '-h[Human-readable sizes]' \
        '-F[Add type indicators]' \
        '-L[Follow symbolic links]' \
        '(-t -u -U -S --sort-ext)'{-t}'[Sort by modification time]' \
        '(-t -u -U -S --sort-ext)'{-u}'[Sort by access time]' \
        '(-t -u -U -S --sort-ext)'{-U}'[Sort by creation time]' \
        '(-t -u -U -S --sort-ext)'{-S}'[Sort by file size]' \
        '(-t -u -U -S --sort-ext)'{--sort-ext}'[Sort by extension]' \
        '-r[Reverse sort order]' \
        '-R[Recursive listing]' \
        '--group-directories-first[List directories before files]' \
        '--ignore=[Ignore pattern]:pattern:' \
        '--gitignore[Use .gitignore rules]' \
        '--tree[Tree format output]' \
        '--json[JSON format output]' \
        '--csv[CSV format output]' \
        '--color=[Color output]:when:(always auto never)' \
        '--no-color[Disable color output]' \
        '--time-style=[Time format]:style:(default iso long-iso full-iso)' \
        '-e[Set file comment]:file:_files' \
        '--comment=[Comment text]:text:' \
        '--version[Show version]' \
        '--help[Show help]' \
        '*:file:_files'
}

_llc "$@"
