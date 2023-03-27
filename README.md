# rwiir - prose editor

`rwiir` (which stands for **R**e**W**rite **I**t **I**n **R**ust and is
pronounced "rewire") is a prose editor - which is to say a specialized word
processor for writing prose texts. Prior art in this genre includes
[Scrivener][scriv], from which a lot of inspiration is taken.

You may be asking - "Why should I write my next story in `rwiir`?" Here's some
of the design decisions that went into creating it:

* Saves to a human-readable format - plain UTF-8 text with occasional intuitive
  use of escape characters. If you don't like the program, you still have your
  work in a format that can easily be migrated.
* One save file contains many individual 'buffers'.
* Minimalist, no-fuss interface that word-wraps text to a comfortable 80
  characters.
* Uses the familiar and powerful Emacs keybindings, or at your option the less
  powerful but more widespread CUA/Windows bindings
* Exports to HTML, Markdown, and TiddlyWiki, with more output formats (including
  ebook and PDF) planned.
* Free - libre and gratis - software with no telemetry (if you care about that
  sort of thing). No choice between exorbitant license fees, corporate
  surveillance, and sketchy piracy. It's free and always will be.

[scriv]: https://www.literatureandlatte.com/scrivener/overview

## Usage

`rwiir` explains itself using its on-line help system, which can be accessed by
pressing `F1` once the program is launched. You will see a "cheat sheet" of
keybindings.

`rwiir` can export to single files, or directories with one file per buffer. Any
buffer whose name contains a hash (`#`, ASCII 0x23) will not be exported unless
explicitly selected for export.

## Roadmap/Future Plans

* A corkboard/outliner/other features not in MVP
* More element types
  - Tables
  - Lists
* More intuitive UI in a few places:
  - New buffer
  - Selecting output file
* More pages of help

## Copying

`rwiir` is Copyright (C) 2023 japanoise, and is free software licensed under the
terms of the GNU Public License (GPL) version 3, or any later version.
