# Shell Integration Testing — Tape Conventions

Lessons from writing golden-file tapes that drive shell completion in
headless terminals. Each shell has its own line editor, startup behavior,
and configuration discovery; these conventions keep tapes deterministic
across machines and containers.

## Shell startup

- **Always delay before the first `Type`.** Shells start at wildly
  different speeds (bash ~10ms, fish ~1s, pwsh 3-5s). Typing before the
  shell is ready makes early characters race the startup output (banner,
  capability queries) and the screen becomes machine-dependent.
  Use `Sleep 2000ms` or more; pwsh may need 3000ms.
- **Wait for a stable marker** instead of a fixed delay where possible
  (e.g. `Wait+Screen /PowerShell/` for the pwsh banner). Then a short
  settle sleep before typing.

## Prompt and banner control

- **Set the prompt explicitly** in the tape (`PS1`, `$PROMPT`,
  `function prompt { "> " }`, `$env.PROMPT_COMMAND`, ...). Inherited
  prompt hooks (starship, powerlevel10k) otherwise leak into the screen
  and differ per machine.
- **Clear `PROMPT_COMMAND`** for bash: the daemon/CI environment may
  export a prompt hook that resets `PS1` after every command.
- **Suppress welcome banners** where the shell supports it:
  `xonsh --no-rc -DXONSH_SUPPRESS_WELCOME=True` (xonsh >= ~0.14 only —
  older versions don't know the variable; use a `clear` after setup
  instead). xonsh's banner picks a **random tagline per startup**, so it
  can never be part of a golden.
- **`clear` after setup, before the test**: removes banners and default
  prompts that may contain the container hostname (random per run).
  Place it *after* the prompt setup — before it, it only clears an
  empty screen.
- elvish's `Ctrl+L` opens the **location browser**, not clear-screen;
  use the external `clear` command instead.

## Isolation

- Isolate `XDG_CONFIG_HOME` and `XDG_DATA_HOME` per run: shell history
  (fish autosuggestions!) and user configuration leak into the terminal
  output otherwise. Note that `--env` appends to the environment and
  **first occurrence wins** for duplicate keys — the terminal inherits
  the daemon's environment, so pre-existing variables cannot be
  overridden this way; set them in the tape or the runner instead.

## Completion assertions

- **Use a single-match prefix** (e.g. `$TE` → `$TERM`, `ech` → `echo`):
  a single candidate is inserted directly, with no selection menu. Menu
  content (candidate lists, footers) depends on the machine's
  executables and environment.
- If the menu is unavoidable, wait for a stable marker (the `SELECT`
  header of the tabdance TUI, or a tag all candidates share like
  `shell variables`) — never for the footer tag of the last entry
  (environment-dependent).
- **Prefer `Wait+Screen` over `Wait+Line`**: `Wait+Line` matches the
  content row at the cursor's Y index, which misaligns when the shell
  renders extra rows (fish's shell-integration markers, prompt
  redraws).
- Some shells append a trailing space after completion (fish, zsh —
  their snippets don't handle the nospace flag): drop the `$` end
  anchor from the wait pattern.
- xonsh quotes `${`-values for insertion (`'${TERM'`), which
  prompt-toolkit does not apply — use the plain `$VAR` form.

## Golden capture

- End the tape with `Ctrl+L` (readline shells) or `clear` so the golden
  contains only the final prompt line — machine-independent. Add a
  trailing `Sleep 300ms`: the executor's last command may not be
  flushed before capture otherwise.
- **Goldens are container-canonical** when the container's shell
  version differs from the host (nushell 0.116 vs 0.115, xonsh 0.13 vs
  0.24): generate with `--update` in the same container that CI verifies
  in.
- The shell services mount the repository read-write (needed for
  `--update` inside the container).

## Known pitfalls

- Shells that query terminal capabilities at startup (DA, cursor
  position, terminfo) **hang if the responses don't arrive**: the
  terminal's responses must be written to the PTY synchronously and the
  write loop must not exit on an empty-buffer EOF.
- `xonsh` 0.13 (Debian bookworm's apt version) is ancient: no
  `-DXONSH_SUPPRESS_WELCOME`, different completion behavior. Prefer a
  newer xonsh in the container image.
- xonsh's contextual completer only completes arguments of *known*
  commands (`context.command` must be in the registered list).
