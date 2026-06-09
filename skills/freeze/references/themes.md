# Themes

Reference for freeze's theme system — the custom `charm` theme and all available chroma themes.

## Overview

Freeze uses the [chroma v2](https://github.com/alecthomas/chroma) styles registry for syntax highlighting themes. Any theme in the registry is available via `--theme`. A custom `charm` theme is registered at init time and serves as the default fallback.

## Theme Resolution

```go
s, ok := styles.Registry[strings.ToLower(config.Theme)]
if s == nil || !ok {
    s = charmStyle
}
if !s.Has(chroma.Background) {
    s, err = s.Builder().Add(chroma.Background, "bg:"+config.Background).Build()
}
```

1. Look up theme name (case-insensitive) in chroma registry
2. If not found, fall back to `charmStyle`
3. If the theme doesn't define a background color, add one from `--background`

## Custom `charm` Theme

Defined in `style.go`, registered into the chroma styles registry:

```go
var charmStyle = styles.Register(chroma.MustNewStyle("charm", chroma.StyleEntries{
    chroma.Text:                "#C4C4C4",
    chroma.Error:               "#F1F1F1 bg:#F05B5B",
    chroma.Comment:             "#676767",
    chroma.CommentPreproc:      "#FF875F",
    chroma.Keyword:             "#00AAFF",
    chroma.KeywordReserved:     "#FF48DD",
    chroma.KeywordNamespace:    "#FF5F87",
    chroma.KeywordType:         "#635ADF",
    chroma.Operator:            "#FF7F83",
    chroma.Punctuation:         "#E8E8A8",
    chroma.Name:                "#C4C4C4",
    chroma.NameBuiltin:         "#FF7CDB",
    chroma.NameTag:             "#B083EA",
    chroma.NameAttribute:       "#7A7AE6",
    chroma.NameClass:           "#F1F1F1 underline bold",
    chroma.NameDecorator:       "#FFFF87",
    chroma.NameFunction:        "#00DC7F",
    chroma.LiteralNumber:       "#6EEFC0",
    chroma.LiteralString:       "#E38356",
    chroma.LiteralStringEscape: "#AFFFD7",
    chroma.GenericDeleted:      "#FD5B5B",
    chroma.GenericEmph:         "italic",
    chroma.GenericInserted:     "#00D787",
    chroma.GenericStrong:       "bold",
    chroma.GenericSubheading:   "#777777",
}))
```

## Available Chroma Themes

All 54 built-in chroma themes are available (case-insensitive):

| Theme | Theme | Theme |
|-------|-------|-------|
| `abap` | `algol` | `algol_nu` |
| `arduino` | `ashen` | `aura-theme-dark` |
| `aura-theme-dark-soft` | `autumn` | `average` |
| `base16-snazzy` | `borland` | `bw` |
| `catppuccin-frappe` | `catppuccin-latte` | `catppuccin-macchiato` |
| `catppuccin-mocha` | `colorful` | `darcula` |
| `doom-one` | `doom-one2` | `dracula` |
| `emacs` | `evergarden` | `friendly` |
| `fruity` | `github` | `github-dark` |
| `gruvbox` | `gruvbox-light` | `hr_high_contrast` |
| `hrdark` | `igor` | `kanagawa-dragon` |
| `kanagawa-lotus` | `kanagawa-wave` | `lovelace` |
| `manni` | `modus-operandi` | `modus-vivendi` |
| `monokai` | `monokailight` | `murphy` |
| `native` | `nord` | `nordic` |
| `onedark` | `onesenterprise` | `paraiso-dark` |
| `paraiso-light` | `pastie` | `perldoc` |
| `pygments` | `rainbow_dash` | `rose-pine` |
| `rose-pine-dawn` | `rose-pine-moon` | `rrt` |
| `solarized-dark` | `solarized-dark256` | `solarized-light` |
| `swapoff` | `tango` | `tokyonight-day` |
| `tokyonight-moon` | `tokyonight-night` | `tokyonight-storm` |
| `trac` | `vim` | `vs` |
| `vulcan` | `witchhazel` | `xcode` |
| `xcode-dark` | | |

Plus the custom `charm` theme = **55 total themes**.

## Theme Application to ANSI Input

For ANSI input, themes are only used for:
- Line number styling: `s.Get(chroma.LineNumbers).Colour.String()`
- Background color injection (if theme lacks `chroma.Background`)

The actual text colors come from ANSI escape sequences, not the theme.
