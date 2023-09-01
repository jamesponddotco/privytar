# Privytar

[![builds.sr.ht status](https://builds.sr.ht/~jamesponddotco/privytar.svg)](https://builds.sr.ht/~jamesponddotco/privytar?)
[![license](https://img.shields.io/badge/license-EUPL_1.2-orange)](LICENSE.md)

**Privytar** is a super simple privacy and caching reverse proxy for the
[Gravatar](https://en.gravatar.com/) service written in Go. It sits in
front of the Gravatar—which is a service that allows you to have a
single avatar image that follows you from site to site—, so
[Automattic](https://automattic.com/) (the company that owns Gravatar)
can't track you.

Inspired by projects like [Nitter](https://github.com/zedeus/nitter),
[Invidious](https://github.com/iv-org/invidious), and
[Scribe](https://sr.ht/~edwardloveall/Scribe/), **Privytar** is easy to
use, deploy, and should be pretty reliable. Use the official instance
[s.privytar.com](https://s.privytar.com) or host the service yourself.

## Usage

Hosting your own instance of **Privytar** is easy:

* [Hosting the service](doc/hosting.md)

Using **Privytar** is even easier:

* [Using the service](doc/using.md)

## Installation

### From source

First install the dependencies:

- Go 1.20 or above.
- make.
- [scdoc](https://git.sr.ht/~sircmpwn/scdoc).

Then compile and install:

```bash
make
sudo make install
```

## Contributing

Anyone can help make `privytar` better. Send patches on the [mailing
list](https://lists.sr.ht/~jamesponddotco/privytar-devel) and report
bugs on the [issue
tracker](https://todo.sr.ht/~jamesponddotco/privytar).

You must sign-off your work using `git commit --signoff`. Follow the
[Linux kernel developer's certificate of
origin](https://www.kernel.org/doc/html/latest/process/submitting-patches.html#sign-your-work-the-developer-s-certificate-of-origin)
for more details.

All contributions are made under [the EUPL license](LICENSE.md).

## Resources

The following resources are available:

- [Support and general discussions](https://lists.sr.ht/~jamesponddotco/privytar-discuss).
- [Patches and development related questions](https://lists.sr.ht/~jamesponddotco/privytar-devel).
- [Instructions on how to prepare patches](https://git-send-email.io/).
- [Feature requests and bug reports](https://todo.sr.ht/~jamesponddotco/privytar).

---

Released under the [EUPL License](LICENSE.md).
