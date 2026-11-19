# Releasing Silicon DB

Silicon DB is a mobile app: it ships on Google Play and the Apple App Store.
Publishing a GitHub release with a tag like `v1.0.0` runs
`.github/workflows/release.yml`. The tag sets the version and the run number
the build number. The workflow checks the data and runs the tests, builds every
package and attaches them to the release. The workflow is written by the Giode
CLI from `giode.toml`: after changing `[ci]`, `app.platforms` or a store
section, run `giode ci -f` instead of editing it, then redo the two hand edits
noted above `[ci]` (the pinned CLI version and `go-version-file: go.mod`).

| Where | What the release gets | Needs |
| --- | --- | --- |
| GitHub, Android | an `.apk` to sideload | nothing; the keystore secrets make it a release build that can update earlier installs |
| Google Play | an `.aab` (and `native-debug-symbols.zip`) to upload in Play Console | a Play Console account and the upload keystore secrets |
| Apple App Store (iPhone and iPad) | an `.ipa` to upload with Transporter | an Apple Developer account and the iOS certificate and profile secrets; without them the job only checks the simulator build |

There are no desktop store packages (Microsoft Store, Mac App Store): the app
targets only Android and iOS (`app.platforms`).

## Repository secrets

Set these in GitHub under Settings → Secrets and variables → Actions. Each
`*_BASE64` secret is the file encoded with `base64 -i file | pbcopy` on macOS,
or `base64 -w0 file` on Linux. Without a secret, that step is skipped and the
rest of the release still builds.

| Secret | What it is |
| --- | --- |
| `ANDROID_KEYSTORE_BASE64` | The upload keystore made by `giode keystore` (`release.keystore`): signs the APK and the AAB |
| `GIODE_KEYSTORE_PASS` | Its password |
| `IOS_CERT_BASE64`, `IOS_CERT_PASS` | The "Apple Distribution" certificate as a `.p12`, and its password |
| `IOS_PROFILE_BASE64` | The "App Store Connect" distribution provisioning profile (`.mobileprovision`) for `com.mehran9675.silicondb` |
| `IOS_TEAM_ID` | Your Apple team ID (10 characters, from the Membership page) |

Never commit the keystore, the `.p12` or the profile: `.gitignore` already
leaves out `*.keystore` and `*.jks`.

## Google Play

1. Open a Google Play Console developer account.
2. Create the upload keystore once, on your own computer, and keep both it and
   its password somewhere safe (a password manager and a backup):
   ```sh
   GIODE_KEYSTORE_PASS='<a strong password>' giode keystore
   ```
   It writes `release.keystore` (already named by `android.keystore` in
   `giode.toml`). Add it as `ANDROID_KEYSTORE_BASE64` and the password as
   `GIODE_KEYSTORE_PASS`.
3. In Play Console, create the app: name "Silicon DB", default language
   English (United States), App, Free.
4. Publish a GitHub release. Upload the `.aab` from it to an internal
   testing release first. Accept **Play App Signing** when asked: Google keeps
   the app signing key and your keystore becomes the upload key. If the upload
   key is ever lost, Play support can reset it; without Play App Signing it
   could not be replaced.
5. Upload `native-debug-symbols.zip` from the same release on the bundle's
   "App bundle explorer" page, so crashes in the Go code show symbols.
6. Fill in App content and the store listing (copy below), then promote the
   build to production. New personal developer accounts must first run a
   closed test with testers for 14 days before production access.

Two settings in `[android]` matter for Play:

- `target_sdk = 36`: Google Play requires Android 16 (API 36) for new apps
  and updates since 31 Aug 2026. The CLI defaults to 35, which Play now
  rejects; raise it again when Play's requirement moves.
- `keystore = "release.keystore"`: `giode build -release` fails without
  `android.keystore`, so CI signing with `GIODE_KEYSTORE_PASS` needs it. The
  workflow writes the file from `ANDROID_KEYSTORE_BASE64`; `.gitignore` keeps
  it out of the repository.

The package name is `com.mehran9675.silicondb` (`app.id`). It can never change
once the app is published.

## Apple App Store

1. Join the Apple Developer Program.
2. In Certificates, Identifiers & Profiles:
   - Register the App ID `com.mehran9675.silicondb` (the `app.id` in
     `giode.toml`), with no extra capabilities.
   - Create an "Apple Distribution" certificate and export it from Keychain
     Access as a `.p12` with a password.
   - Create an "App Store Connect" (distribution) provisioning profile for the
     App ID with that certificate.
3. In App Store Connect, create the app: platform iOS, name "Silicon DB",
   primary language English (U.S.), the bundle ID above, SKU `silicondb`.
4. Add the five iOS secrets above and publish a GitHub release. The release
   gets the `.ipa`; upload it with the Transporter app (or
   `xcrun altool --upload-app`). It appears in TestFlight after processing.
5. Fill in the listing (copy below) and submit the build for review.

What the build already answers:

- **Export compliance:** `ITSAppUsesNonExemptEncryption = false` in
  `[ios.plist]`. The app uses only HTTPS through the system and Go's standard
  library, so App Store Connect asks no encryption questions.
- **Privacy manifest:** `PrivacyInfo.xcprivacy` from `[privacy]`: no tracking,
  no collected data, and the required-reason APIs Go itself uses.
- **Permissions:** none. The app needs no usage descriptions (`[ios.usage]`).
- **Devices:** iPhone and iPad, all orientations, iOS 15 and later.

## Store listing

Accurate to the app as it ships. Paste as is.

**Name:** Silicon DB

**Subtitle (Apple, ≤30):** Specs of every chip

**Short description (Google Play, ≤80):**
Specs of CPUs, GPUs, SoCs and AI chips from every maker and era, offline.

**Promotional text (Apple, optional, ≤170):**
Look up, search and compare the specifications of over 1,300 processors, graphics chips, SoCs and AI accelerators, from the Intel 8008 to today.

**Full description:**

```
Silicon DB is a free, open-source reference for the specifications of processing chips: CPUs, GPUs, APUs, SoCs, TPUs and NPUs and other AI accelerators, microcontrollers, DSPs and FPGAs, from every manufacturer and every era, from the Intel 8008 of 1972 on.

BROWSE
Go from category to company, segment, generation and chip: CPU, Intel, Desktop, Core 14th Gen, Core i5-14600K.

SEARCH AND FILTER
Search by name, family, codename or architecture. Filter by category, decade and manufacturer, and sort by date, name or speed.

EVERY SPEC, GROUPED
Each chip page starts with its headline figures, then lists every spec: fabrication, cores and clocks, cache, graphics, memory, performance, power and I/O.

BENCHMARKS
Geekbench 6 (CPU and GPU compute) and Cinebench 2024, R23 and R15 scores, drawn as bars against the fastest chip in the database.

COMPARE
Put up to four chips side by side, with the best value in each row highlighted.

MANUFACTURERS
More than 70 companies, each with its history and its chips by category.

OFFLINE FIRST
The database is kept on your device after the first download, and updates fetch only the files that changed. New chips arrive without an app update.

25 LANGUAGES
English and 24 other European languages, with localized numbers and dates. Dark and light themes, phone and tablet layouts.

PRIVATE
No account, no ads, no analytics, no tracking. The app only downloads the public database.

The data is open: it lives as JSON in the project's GitHub repository, where anyone can suggest corrections.
```

**Keywords (Apple, ≤100):**
`cpu,gpu,processor,chip,soc,specs,benchmark,compare,intel,amd,nvidia,npu,fpga,silicon,hardware`

**Category:** Apple: Reference (secondary: Utilities). Google Play: Books & Reference.

**Support URL / website:** https://xmhrnx.github.io/silicondb/

**Privacy policy URL:** https://xmhrnx.github.io/silicondb/#privacy

**Copyright:** 2026 Mehran Zahiroddini

**Age rating:**
- Apple: answer "None" to every content question, and "No" to unrestricted
  web access (the app shows only its own database; it has no browser), user
  generated content, chat and gambling: rated 4+.
- Google Play (IARC questionnaire): category "Reference, News, or
  Educational"; no violence, sexuality, language, controlled substances,
  gambling, user interaction, sharing of location or personal data, or
  purchases: rated Everyone / PEGI 3.
- Google Play target audience: 13 and over. The app is suitable for children,
  but choosing under-13 groups enrols it in the Families policy for no benefit.
- Ads: No. In-app purchases: none. App access: everything works without
  login.

**App Privacy (Apple):** "Data Not Collected". The app sends nothing to its
developer or to third parties; it only downloads the public database files,
with no identifiers.

**Data safety (Google Play):**
- Does your app collect or share any of the required user data types? **No.**
- Is all of the user data collected by your app encrypted in transit? Not
  asked once the answer above is No. (Downloads use HTTPS.)
- Do you provide a way for users to request that their data be deleted? Not
  asked; nothing is collected. Local data goes with an uninstall.

**Government / news / health / financial declarations (Play):** not
applicable.

## Screenshots

There are none in the repository yet. Take them from the running app in the
dark theme (the default), after the database has downloaded:

1. Home: the database at a glance and latest releases.
2. A chip page with its headline figures and benchmark bars (for example
   GeForce RTX 5090 or Core i5-14600K).
3. Compare with three or four chips.
4. Explore with a search and filters applied.
5. A manufacturer page (Intel or AMD).
6. Browsing: category → company → segment.

Sizes:

- **Apple:** iPhone 6.9" (1320 × 2868) or 6.5" (1284 × 2778), and iPad 13"
  (2064 × 2752): the app runs on iPad, so iPad screenshots are required.
  The iOS Simulator's File → Save Screenshot gives the right size.
- **Google Play:** at least 2 phone screenshots (1080 × 1920 or larger, 9:16),
  and 7" and 10" tablet screenshots for tablet visibility; a 512 × 512 icon
  (`assets/icon.png` scaled down) and a 1024 × 500 feature graphic (the site's
  hero, the "Si" tile on the dark background, works).

## Known limitations

- The first launch needs a connection to download the database; after that the
  app works offline.
- Settings → Data source can point the app at any URL. The stores' reviewers
  see the default (GitHub); nothing else needs explaining.
- No Windows, macOS or Linux store packages: the app is built for Android and
  iOS only.
