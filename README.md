[![pkgsite](https://pkg.go.dev/badge/github.com/untangledco/streaming)](https://pkg.go.dev/github.com/untangledco/streaming)

This repository contains packages for developing media streaming systems in Go.
Watch it being developed live!

- [twitch.tv/untangledco]
- [youtube.com/@untangledco]

We use these packages to self-host the livestream at [olowe.co/live].

[twitch.tv/untangledco]: https://twitch.tv/untangledco
[youtube.com/@untangledco]: https://www.youtube.com/@untangledco
[olowe.co/live]: https://olowe.co/live

## Contributing

### Stuff to do

Larger, fleshed-out tasks are managed in
[issues](https://github.com/untangledco/streaming/issues).

We can put `TODO` notes in the source code, too.
[godoc] provides a graphic interface to view these
with the `-notes` flag:

	godoc -notes TODO

Of course `grep` works too:

	git grep -n TODO

[godoc]: https://pkg.go.dev/golang.org/x/tools/cmd/godoc

### Patches

Patches are preferred via email so that we're not too locked in to GitHub.
Post them to the mailing list
[~otl/untangledco@lists.sr.ht](mailto:~otl/untangledco@lists.sr.ht),
or to [Oliver Lowe](mailto:o@olowe.co).
See [git-send-email.io](https://git-send-email.io) if you're unfamiliar with the workflow.

	git config sendemail.to '~otl/untangledco@lists.sr.ht'

We also accept changes via [pull requests](https://github.com/untangledco/streaming/pulls).

### Commit messages

Commit messages follow the same format used by the [Go] project (among others).
The commit subject starts with the affected package name then a brief description of the change.
The body may contain an explanation of the change and why it was made.

[Go]: https://go.dev/doc/contribute#commit_messages

## License

Unless otherwise noted, this sotfware is licensed under the ISC License.
See LICENSE.
