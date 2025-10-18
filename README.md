# secure-example

This example outlines how to securely release using GoReleaser and GitHub
Actions.

## Components

We'll go over a few things: GitHub settings, GoReleaser configuration, and
GitHub actions.

### GitHub Settings

These are some things I recommend you do:

1. General > Require contributors to sign off on web-based commits Loading
1. General > Enable release immutability
1. Actions > General > Require approval for all external contributors
1. Actions > General > Read repository contents and packages permissions
1. Rules > New Ruleset > Import a ruleset > [this file](./.github/ruleset.json)
1. Advanced Security > Private vulnerability reporting
1. Advanced Security > Dependency graph
1. Advanced Security > Automatic dependency submission
1. Advanced Security > Dependabot security updates

There's much more you can change, these are the things I usually do.

### GoReleaser Configuration

The [provided configuration](./.goreleaser.yaml) is commented out and each
section links to the relevant documentation, but here's a rundown:

1. We build for a couple of platforms using the Go mod proxy;
1. We create archives for both the binaries as well as for the source;
1. We create and sign a checksums file (using [Cosign][cosign]);
1. We create [Software Bill of Materials (SBOMs)](https://www.cisa.gov/sbom)
   for all the archives (using [Syft][syft]);
1. all these files are uploaded to the GitHub release;
1. We create a Docker image manifest, which also includes SBOMs;
1. We then sign the image.

### GitHub Actions

We have 3 workflows set up, let's go over them.

#### Build

The [build workflow](./.github/workflows/build.yml) doesn't do much: it checks
out the code, installs Go, and runs `go test`.

#### Security

The [security workflow](./.github/workflows/security.yml) does a lot more, as it
has a couple of jobs:

1. `codeql`: as the name implies, runs the recommended [CodeQL][codeql] queries
   for Go and Actions;
1. `grype`: runs [Grype][], which scans for known vulnerabilities;
1. `govulncheck`: runs the standard [Go vulnerability checker][govulncheck];
1. `dependency-review`: runs only on pull requests, and checks if any
   dependencies being added or updated are allowed.

All these jobs report their status using
[Static Analysis Results Interchange Format (SARIF)](https://sarifweb.azurewebsites.net/),
so any findings will show as security alerts in the
[Security > Code scanning](/security/code-scanning) tab.

#### Release

Finally, the [release workflow](./.github/workflows/release.yml).
Its main job is to, well, release our software, and it uses [GoReleaser][] for
that (surprise!).
But to do that, we first set up [Docker][docker], [Cosign][cosign], and
[Syft][syft].
Then, we run the glorious [goreleaser-action][], which does all the heavy
lifting.
Then, after all is said and done, we attest our build artifacts.

#### Why all the SHA1s?

As you may have noticed, all the actions are pinned to the SHA1 of their
respective tags.

This is recommended, as an attacker might take over an action and re-publish
malicious code under the same tags.

Using only the major versions (like `@v4`) is also not so good, as you are then
even more clueless about what is actually being run.

If you want to pin all the actions in your repositories, I recommend using
[pinata][].

## Releasing

To create a new release, create and push a new tag.
You can get the next semantic version using [svu][]:

```bash
git tag -s $(svu n)
git push --tags
```

And then go over to [the actions tab](/actions/workflows/release.yml) and wait
for the release to finish.

## Verifying the artifacts

Your users will need to know how to verify the artifacts, and this is what this
section is all about.

The first thing we need to do, is get the current latest version:

```bash
export VERSION="$(gh release list -L 1 -R goreleaser/example-secure --json=tagName -q '.[] | .tagName')"
```

Then, we download the `checksums.txt` and the signature bundle
(`checksums.txt.sigstore.json`) files, and then verify them:

```bash
wget https://github.com/goreleaser/example-secure/releases/download/$VERSION/checksums.txt
wget https://github.com/goreleaser/example-secure/releases/download/$VERSION/checksums.txt.sigstore.json
cosign verify-blob \
    --certificate-identity "https://github.com/goreleaser/example-secure/.github/workflows/release.yml@refs/tags/$VERSION" \
    --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
    --bundle "checksums.txt.sigstore.json" \
    ./checksums.txt
```

This should succeed - which means that we can from now on verify any artifact
from the release with this checksum file!

You can then download any file you want from the release, and verify it with, for example:

```bash
wget "https://github.com/goreleaser/example-secure/releases/download/$VERSION/example_linux_amd64.tar.gz"
sha256sum --ignore-missing -c checksums.txt
```

Which should, ideally, say "OK".

You can then inspect the SBOM file to see the entire dependency tree of the
binary, check for vulnerable dependencies and whatnot.

To get the SBOM of an artifact, you can use the same download URL, adding
`.sbom.json` to the end of the URL, and we can then check it out with `grype`:

```bash
wget "https://github.com/goreleaser/example-secure/releases/download/$VERSION/example_linux_amd64.tar.gz.sbom.json"
sha256sum --ignore-missing -c checksums.txt
grype sbom:example_linux_amd64.tar.gz.sbom.json
```

Finally, we can also use the `gh` CLI to verify the attestations:

```bash
gh attestation verify \
  --owner goreleaser \
  *.tar.gz
```

Docker images are a bit simpler, you can verify them with [Cosign][cosign]
and [Grype][grype] directly, and check the attestations as well.

Signature:

```bash
cosign verify \
  --certificate-identity "https://github.com/goreleaser/example-secure/.github/workflows/release.yml@refs/tags/$VERSION" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  "ghcr.io/goreleaser/example-secure:$VERSION"
```

Vulnerabilities:

```bash
grype "docker:ghcr.io/goreleaser/example-secure:$VERSION"
```

Attestations:

```bash
gh attestation verify \
  --owner goreleaser \
  "oci://ghcr.io/goreleaser/example-secure:$VERSION"
```

If all these checks are OK, you have a pretty good indication that everything
is good.

---

I really hope this helps - and please feel free to open PRs improving things you
think need improving, or issues to discuss any concerns you might have.

Thanks for reading!

[syft]: https://github.com/anchore/syft
[cosign]: https://github.com/sigstore/cosign
[grype]: https://github.com/anchore/grype
[codeql]: https://codeql.github.com/
[govulncheck]: https://go.dev/blog/govulncheck
[goreleaser]: https://goreleaser.com
[docker]: https://docker.io
[pinata]: https://github.com/caarlos0/pinata
[svu]: https://github.com/caarlos0/svu
[goreleaser-action]: https://github.com/goreleaser/goreleaser-action
