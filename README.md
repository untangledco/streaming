[![pkgsite](https://pkg.go.dev/badge/github.com/untangledco/streaming)](https://pkg.go.dev/github.com/untangledco/streaming)
[![0 dependencies!](https://0dependencies.dev/0dependencies.svg)](https://0dependencies.dev)

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

There are `TODO` notes in the source code, too.
[godoc] provides a graphic interface to view these
with the `-notes` flag:

	godoc -notes TODO

Of course `grep` works too:

	git grep -n TODO

[godoc]: https://pkg.go.dev/golang.org/x/tools/cmd/godoc

### Patches

Patches are preferred via email so that we're not too locked in to GitHub.
Post them to the mailing list
[~otl/untangledco@lists.sr.ht] ([archives]).
or to [Oliver Lowe].
See [git-send-email.io] if you're unfamiliar with the workflow.

	git config sendemail.to '~otl/untangledco@lists.sr.ht'

We also accept changes via [pull request].

### Commit messages

Commit messages follow the same format used by the [Go] project (and others).
The commit subject starts with the affected package name then a brief description of the change.
The body may contain an explanation of the change and why it was made.
For example:

	sdp: store attributes as key-value pairs

	This matches what the spec allows, and lets users not worry about
	encoding.

[archives]: https://lists.sr.ht/~otl/untangledco
[~otl/untangledco@lists.sr.ht]: mailto:~otl/untangledco@lists.sr.ht
[git-send-email.io]: https://git-send-email.io
[pull request]: https://github.com/untangledco/streaming/pulls

### Code review

We try to make all code feel familiar to Go programmers
so that it's easier for others to learn from and contribute to in the
future.

In general we follow the guidelines laid out in the following articles:

- [Effective Go]
- [Go Code Review Comments]

Don't worry if you're asked about changing things around!
If there are trivial changes, we may make the changes ourselves.
In this case you still retain all authorship and copyright over your
submission.

If you've read this far and want to join in but feel a little unsure;
I know *exactly* how you feel.
Please feel free to email [Oliver Lowe] and maybe I can help out :)

[Effective Go]: https://go.dev/doc/effective_go
[Go Code Review Comments]: https://go.dev/wiki/CodeReviewComments
[Go]: https://go.dev/doc/contribute#commit_messages
[Oliver Lowe]: mailto:o@olowe.co

## License

Unless otherwise noted, this sotfware is licensed under the ISC License.
See LICENSE.
