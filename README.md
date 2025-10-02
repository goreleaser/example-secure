# secure-example

This is an example repository showing how to securely release using GoReleaser
and GitHub Actions.

## Workflows

- `build`: runs tests
- `release`: runs goreleaser
- `security`: security scans: grype, govulncheck, codeql, license-check

## How it works

GoReleaser manages the entire thing, basically.

It will:

- build using the Go Mod Proxy as source of truth
- create archives
- call `syft` to create the SBOMs
- create the checksum file
- sign it with `cosign`
- create the github release
- push artifacts there
- create a docker image (with SBOM) using the binary it just built (thus, the binary inside the docker image is the same as the one released)
- sign the docker image with `cosign` as well

## Verifying

```bash
export VERSION="$(gh release list -L 1 -R goreleaser/example-secure --json=tagName -q '.[] | .tagName')"
```

### Checksums

```shell
wget https://github.com/goreleaser/example-secure/releases/download/$VERSION/checksums.txt
cosign verify-blob \
    --certificate-identity 'https://github.com/goreleaser/example-secure/.github/workflows/release.yml@refs/tags/$VERSION' \
    --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
    --cert https://github.com/goreleaser/example-secure/releases/download/$VERSION/checksums.txt.pem \
    --signature https://github.com/goreleaser/example-secure/releases/download/$VERSION/checksums.txt.sig \
    ./checksums.txt
```

You can then download any file you want from the release, and verify it with, for example:

```shell
wget https://github.com/goreleaser/example-secure/releases/download/$VERSION/example_linux_amd64.tar.gz
sha256sum --ignore-missing -c checksums.txt
```

And both should say "OK".

### SBOMs

You can then inspect the `.sbom` file to see the entire dependency tree of the
binary, check for vulnerable dependencies and whatnot.

To get the SBOM of an artifact, you can use the same download URL, adding
`.sbom.json` to the end of the URL:

```shell
wget https://github.com/goreleaser/example-secure/releases/download/$VERSION/example_linux_amd64.tar.gz.sbom.json
sha256sum --ignore-missing -c checksums.txt
grype sbom:example_linux_amd64.tar.gz.sbom.json
```

### Attestations

This example also publishes build attestations.
You can verify any artifact with:

```shell
gh attestation verify \
  --owner goreleaser \
  *.tar.gz
```

### Docker image

```shell
cosign verify \
  --certificate-identity 'https://github.com/goreleaser/example-secure/.github/workflows/release.yml@refs/tags/$VERSION' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  ghcr.io/goreleaser/example-secure:$VERSION
```

The images are also attested:

```shell
gh attestation verify \
  --owner goreleaser \
  oci://ghcr.io/goreleaser/example-secure:$VERSION
```
