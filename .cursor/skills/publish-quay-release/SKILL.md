---
name: publish-quay-release
description: >-
  Compila frontend y BFF, publica las imágenes en Quay, deja esa versión
  actualizada en la documentación Helm de instalación, y solo entonces hace
  push al repo git. Use when the user asks to publish to Quay, release a
  version, subir imágenes, publicar el plugin, actualizar el chart, or
  compile y suba una version a repo quay.
---

# Publish Quay release

Purpose (verbatim): compile y suba una version a repo quay, esa version dejarla actualizada en la documentacion del helm para instalar, posterior a que la imagen o imagenes esten correctamente publicadas y la documentacion con la version actualizada, hacer push al repo git.

Do **not** update Helm docs or push git until both images are confirmed on Quay.

## Defaults

| Item | Value |
|------|--------|
| Registry | `QUAY_REGISTRY` env, else `quay.io/rh-ai-community-plugins` |
| Frontend image | `maas-finops` |
| BFF image | `maas-finops-bff` |
| Platform | `linux/amd64` (always; local Mac is arm64, OpenShift nodes are amd64) |
| Builder | `podman` |

Images to publish every release:

- `$REGISTRY/maas-finops:$VERSION`
- `$REGISTRY/maas-finops-bff:$VERSION`

## Workflow

Copy and track:

```
Release:
- [ ] Version chosen
- [ ] podman logged in to quay.io
- [ ] Both images built linux/amd64
- [ ] Both images pushed
- [ ] Both images verified on Quay
- [ ] Helm install docs + chart metadata set to $VERSION
- [ ] Commit
- [ ] git push (only after verify + docs)
```

### 1. Version

If the user gave a version, use it. Otherwise read `package.json` `version` and bump patch, unless they asked for minor/major.

Must be semver (`0.1.0`, `0.2.0`). Do not prompt interactively (`scripts/build-push.sh` without a version asks `y/N` — never use that path).

Set `VERSION` and `REGISTRY` in the shell before building.

### 2. Preconditions

Stop if any fail:

- `podman` available
- `podman login quay.io` succeeds (use existing credentials; do not invent tokens)
- Working tree is this plugin repo (has `Containerfile` and `bff/Containerfile`)
- User named an explicit version **or** agreed to the computed bump

Do not skip git hooks. Do not `--force` push to `main`/`master`.

### 3. Build and push Quay

Do **not** call `./scripts/build-push.sh` without a version (interactive). Prefer explicit builds so `--platform linux/amd64` is always set:

```bash
REGISTRY="${QUAY_REGISTRY:-quay.io/rh-ai-community-plugins}"

podman build --platform linux/amd64 \
  -t "${REGISTRY}/maas-finops:${VERSION}" -f Containerfile .

podman build --platform linux/amd64 \
  -t "${REGISTRY}/maas-finops-bff:${VERSION}" -f bff/Containerfile bff/

podman push "${REGISTRY}/maas-finops:${VERSION}"
podman push "${REGISTRY}/maas-finops-bff:${VERSION}"
```

If `podman login` fails, stop. Do not update docs. Do not git push.

### 4. Verify published

Both must succeed before any doc or git step:

```bash
skopeo inspect "docker://${REGISTRY}/maas-finops:${VERSION}"
skopeo inspect "docker://${REGISTRY}/maas-finops-bff:${VERSION}"
```

If `skopeo` is missing:

```bash
podman pull --platform linux/amd64 "${REGISTRY}/maas-finops:${VERSION}"
podman pull --platform linux/amd64 "${REGISTRY}/maas-finops-bff:${VERSION}"
```

If either inspect/pull fails, stop and report which image is missing. Leave git untouched.

Optional chart OCI (only if the user asked to publish the chart too):

```bash
helm package chart/
helm push maas-finops-chart-${VERSION}.tgz "oci://${REGISTRY}"
```

Default: **images only**. Helm install docs may keep `helm install maas-finops chart/` (local chart) with `--set image.tag=$VERSION`.

### 5. Update Helm install documentation

Only after step 4 passed. Keep one version string everywhere.

1. Set `"version"` in `package.json` to `$VERSION`.
2. Run `node scripts/sync-chart-version.js` (updates `chart/Chart.yaml`, `plugin.yaml` tags, and `--version` flags in README / `docs/deployment/OPENSHIFT_DEPLOY.md` / `docs/development/BUILD_AND_PUSH.md`).
3. Also set these if they still show the old tag or empty tag:

| File | Fields |
|------|--------|
| `plugin.yaml` | `version`, `image.tag`, `bff_image.tag` |
| `chart/Chart.yaml` | `version`, `appVersion` |
| `chart/values.yaml` | `image.tag`, `bff.image.tag` (use `"$VERSION"`, not empty) |
| `chart/README.md` | `--set image.tag=` and `bff.image.tag=` |
| `README.md` | Helm `--version` and `--set image.tag` examples |
| `docs/deployment/OPENSHIFT_DEPLOY.md` | `helm install ... --version` |
| `CHANGELOG.md` | entry for `$VERSION` |

`image.repository` / `bff.image.repository` stay `$REGISTRY/maas-finops` and `$REGISTRY/maas-finops-bff` unless the user named another org.

Helm install snippet to keep accurate (adjust registry if overridden):

```bash
helm install maas-finops oci://quay.io/rh-ai-community-plugins/maas-finops-chart \
  --version $VERSION \
  --namespace cp-maas-finops \
  --create-namespace \
  --set adminUser=admin
```

Local chart fallback must use the same image tag:

```bash
helm install maas-finops chart/ \
  --namespace cp-maas-finops \
  --create-namespace \
  --set image.tag=$VERSION \
  --set bff.image.tag=$VERSION \
  --set adminUser=admin
```

### 6. Git push (last)

Only after images verified **and** docs updated.

1. If `git remote` has no `origin`, **stop**. Ask for the GitHub URL. Do not `git init` a second repo. Do not guess `rh-ai-community-plugins`.
2. Stage only release files (version + docs + changelog). Do not add secrets, `Dockerfile` copies, or `node_modules`.
3. Commit with HEREDOC (user git protocol). Message: why, e.g. `Release $VERSION to Quay and sync Helm install docs.`
4. `git push -u origin HEAD` (permissions `all`). Do not `--force` on `main`.
5. If push needs a new remote the user just provided: `git remote add origin <url>` then push.

If commit hooks fail, fix and create a **new** commit. Do not `--amend` after a successful push.

## Stop conditions

- Quay login or push failed
- One of the two images did not inspect/pull
- No git `origin` and user did not give a URL
- User asked only to build locally — skip Quay, docs, and push

## Do not

- Update Helm docs before Quay verify
- Push git before docs are updated
- Use interactive `scripts/build-push.sh` without `$VERSION`
- Build without `--platform linux/amd64`
- Skip hooks or force-push `main`
