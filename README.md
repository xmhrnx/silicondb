# Silicon DB

A mobile app with the specifications of processing chips — CPUs, GPUs, APUs, SoCs,
TPUs/NPUs and other AI accelerators, microcontrollers, DSPs and FPGAs — from every
manufacturer and every era. Built in Go with
[giode](https://github.com/mehran9675/giode) and packaged with the
[giode CLI](https://github.com/mehran9675/giode-cli).

## Features

- **Home**: the database at a glance, categories, latest releases, the earliest chips,
  top manufacturers and recently viewed chips.
- **Explore**: search by name, family, codename or architecture; filter by category,
  decade and manufacturer; sort by date or name.
- **Chip pages**: headline figures and every spec, grouped (fabrication, cores and clocks,
  cache, graphics, memory, performance, power and I/O).
- **Manufacturers**: every company with its history and its chips by category.
- **Compare**: up to four chips side by side, with the best value in each row highlighted.
- **Offline first**: the last download is kept on the device, and updates fetch only the
  files that changed.
- **Benchmarks**: Geekbench 6 (CPU and GPU compute) and Cinebench 2024, R23 and R15 scores,
  shown as bars against the fastest chip in the database, and a "Fastest" sort.
- **25 languages**: English and 24 other European languages, following the system language
  or chosen in Settings, with localized numbers and dates.
- Dark and light themes, phone and tablet layouts, open-source licenses screen.

## Data

Nothing about chips is compiled into the app. It downloads the database from
[`data/`](data/README.md) at runtime:

```
data/index.json                          table of contents with a revision per file
data/schema.json                         categories, spec groups, labels and units
data/manufacturers/<id>/manufacturer.json
data/manufacturers/<id>/<category>.json  one shard per manufacturer and category
```

The app compares each file's revision in `index.json` with its offline copy and downloads
only what changed. Spec labels, units and groups live in `schema.json`, so new fields
need no app update either. See [data/README.md](data/README.md) for the format and how to
edit it; after editing, run:

```sh
go run ./cmd/indexgen        # validates everything and rewrites data/index.json
```

Missing specs can be filled from public sources with `go run ./cmd/enrich`; see
[data/README.md](data/README.md#filling-in-missing-specs).

By default the app reads `https://raw.githubusercontent.com/xmhrnx/silicondb/main/data/`.
Change it at build time with
`-ldflags "-X github.com/mehran9675/silicondb/internal/config.DataURL=<url>"` (or
`app.ldflags` in `giode.toml`), or at runtime in Settings → Data source.

## Building

Requires Go 1.26+ and the giode CLI:

```sh
go install github.com/mehran9675/giode-cli/cmd/giode@latest
giode doctor                 # what each target needs
giode run android            # build, install and launch on a connected phone
giode build android          # dist/android/Silicon DB.apk
giode build ios              # on macOS, with a provisioning profile in giode.toml
giode run                    # on this computer
```

### Releases

`.github/workflows/release.yml` is written by `giode ci` from `giode.toml` (`app.platforms`
and `[ci]`); rerun `giode ci -f` after changing those rather than editing it. Publishing a
GitHub release builds the Android APK and AAB and the iOS IPA, with the release tag as the
version and the run number as the build number, and attaches them to the release. Nothing is
built on ordinary pushes. Signing uses repository secrets:

| Secret | For |
| --- | --- |
| `ANDROID_KEYSTORE_BASE64`, `GIODE_KEYSTORE_PASS` | A release-signed Android build (create the keystore with `giode keystore`). Without them the APK uses a debug key and can't update an earlier install. |
| `IOS_CERT_BASE64`, `IOS_CERT_PASS`, `IOS_PROFILE_BASE64` (and optionally `IOS_TEAM_ID`) | The Apple distribution certificate as .p12, its password, and the provisioning profile for `com.mehran9675.silicondb`. Without them the workflow only checks the simulator build. |

Encode a file for a secret with `base64 -i file | pbcopy` (macOS) or `base64 -w0 file` (Linux).

The app icon is drawn by `go run ./cmd/icongen`, which writes `assets/icon.png`.

## Project layout

```
main.go                     entry point
assets/                     embedded fonts (Inter), SVG icons, app icon, translations
                            (locales/) and open-source licenses (licenses.json)
cmd/indexgen/               data validator and index generator
cmd/enrich/                 fills missing specs from public sources
cmd/icongen/                draws assets/icon.png
data/                       the database (downloaded by the app, not embedded)
internal/
  config/                   build-time settings (data URL, version)
  model/                    chips, manufacturers, shards, index, schema, catalog
  data/                     HTTP client, offline cache, incremental sync
  format/                   spec values, units and dates for display
  store/                    app state: catalog, sync status, queries, preferences
  ui/
    theme/                  palettes, type scale, fonts, icons
    tr/                     translation helpers and the language list
    nav/                    tab and stack navigation
    env/                    what screens share
    components/             reusable widgets, one per file
    screens/                home, explore, detail, makers, maker, compare, settings, licenses
    app/                    assembles the screens into the app
```

## Translations

UI text lives in `assets/locales/<language>.json`, one file per language, read with giode's
`Intl`. Spec, group and category labels come from `data/schema.json` in English and are
translated under `spec.<key>`, `group.<id>` and `category.<id>.name`/`.title`; a field added
to the schema later shows its English label until a translation is added. To add a language,
copy `en.json`, translate it (plus those schema keys) and add it to
`internal/ui/tr/languages.go`.

## Tests

```sh
go test ./internal/...
```

## License

Apache License 2.0; see [LICENSE](LICENSE). Inter is under the SIL Open Font License
([assets/fonts/OFL.txt](assets/fonts/OFL.txt)).
