# Fish completion script for llc

# Disable file completion for specific options
complete -c llc -s a -l all -d "Show all files including hidden (. and ..)"
complete -c llc -s A -l almost-all -d "Show all files except . and .."
complete -c llc -s 1 -d "Single column output"
complete -c llc -s i -d "Show inode number"
complete -c llc -s d -d "List directory itself instead of contents"
complete -c llc -s h -d "Human-readable sizes"
complete -c llc -s F -d "Add type indicators"
complete -c llc -s L -d "Follow symbolic links"
complete -c llc -s t -d "Sort by modification time"
complete -c llc -s u -d "Sort by access time"
complete -c llc -s U -d "Sort by creation time"
complete -c llc -s S -d "Sort by file size"
complete -c llc -l sort-ext -d "Sort by extension"
complete -c llc -s r -d "Reverse sort order"
complete -c llc -s R -d "Recursive listing"
complete -c llc -l group-directories-first -d "List directories before files"
complete -c llc -l ignore -d "Ignore pattern" -r
complete -c llc -l gitignore -d "Use .gitignore rules"
complete -c llc -l tree -d "Tree format output"
complete -c llc -l json -d "JSON format output"
complete -c llc -l csv -d "CSV format output"
complete -c llc -l color -d "Color output" -x -a "always auto never"
complete -c llc -l no-color -d "Disable color output"
complete -c llc -l time-style -d "Time format" -x -a "default iso long-iso full-iso"
complete -c llc -s e -d "Set file comment" -r
complete -c llc -l comment -d "Comment text" -r
complete -c llc -l version -d "Show version"
complete -c llc -l help -d "Show help"
