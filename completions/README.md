# Shell Completions

Shell completion scripts for llc.

## Bash

Copy to your bash completions directory:

```bash
# Linux
sudo cp completions/llc.bash /etc/bash_completion.d/llc

# macOS (with Homebrew)
cp completions/llc.bash /usr/local/etc/bash_completion.d/llc
```

Or add to your `~/.bashrc`:
```bash
source /path/to/completions/llc.bash
```

## Zsh

Copy to your zsh completions directory:

```bash
mkdir -p ~/.zsh/completions
cp completions/llc.zsh ~/.zsh/completions/_llc
```

Add to your `~/.zshrc`:
```bash
fpath+=~/.zsh/completions
autoload -U compinit && compinit
```

## Fish

Copy to your fish completions directory:

```bash
mkdir -p ~/.config/fish/completions
cp completions/llc.fish ~/.config/fish/completions/llc.fish
```

Or install via Fisher:
```bash
fisher install kkw-h/llc
```
