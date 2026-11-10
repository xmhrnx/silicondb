# Silicon DB data

The app downloads everything it shows from this folder at runtime. Nothing about chips is
compiled into the app, so editing these files and pushing them updates every installed copy
on its next sync.

```
data/
  index.json                         generated: lists every shard with its revision
  schema.json                        categories, spec groups, labels, units, highlights
  manufacturers/<id>/
    manufacturer.json                the company
    <category>.json                  one shard per category the company has chips in
```

The app fetches `index.json` first, then only the shards (and `schema.json`) whose revision
differs from its cached copy. A typo fix in `manufacturers/zilog/cpu.json` therefore costs
users one small download, not the whole database.

## Updating

1. Edit or add files under `manufacturers/`.
2. Run `go run ./cmd/indexgen` from the repository root. It validates every file and rewrites
   `index.json` with fresh revisions and counts.
3. Commit and push.

### Filling in missing specs

`cmd/enrich` adds specs a chip is missing from public sources, so a new entry doesn't have to
be complete by hand:

```sh
go run ./cmd/enrich -dry            # report what would be added, write nothing
go run ./cmd/enrich -only intel     # one manufacturer
go run ./cmd/enrich -category cpu -since 2015
go run ./cmd/enrich                 # every shard
```

It only fills specs that are absent and removes them from the chip's `na` list, so values
written by hand always win and re-running it is safe. Afterwards run `cmd/indexgen` to refresh
`index.json`.

Sources, all of which allow automated access through their robots.txt:

| Source | Supplies |
| --- | --- |
| [NanoReview](https://nanoreview.net) | Cores and clocks, caches, process, transistors, socket, peak temperature, memory type and speed, ECC, PCIe version and lanes, integrated-graphics figures, and the Geekbench 6 and Cinebench 2024 scores. Also covers phone SoCs. |
| [Wikipedia](https://en.wikipedia.org) graphics-processor lists | The FP16, FP32 and FP64 figures, plus the dense tensor-core FP8/FP16/FP4 throughput of recent cards. |
| [CPU-Monkey](https://www.cpu-monkey.com) | Geekbench 6 and Cinebench R23/2024 scores. Off by default (`-sources ...,cpumonkey`) because it rate-limits hard. |

Downloaded pages are cached under `.cache/enrich`, so a second run costs nothing; pass
`-offline` to work from the cache alone. The tool prints warnings for pages it couldn't read
and a per-key count of what it added.

## manufacturer.json

```json
{
  "id": "intel",
  "name": "Intel",
  "country": "United States",
  "founded": 1968,
  "defunct": 0,
  "website": "https://www.intel.com"
}
```

`id` matches the folder name: lowercase ASCII letters, digits and hyphens. `defunct` is the
year the company ceased to exist or was absorbed, 0 if it still exists.

## Shards: manufacturers/&lt;id&gt;/&lt;category&gt;.json

`<category>` is one of the category ids in `schema.json`: `cpu`, `gpu`, `apu`, `soc`, `ai`
(TPUs, NPUs and other AI accelerators), `mcu`, `dsp`, `fpga`.

```json
{
  "manufacturer": "intel",
  "category": "cpu",
  "chips": [
    {
      "id": "intel-8008",
      "name": "8008",
      "family": "MCS-8",
      "released": "1972-04",
      "specs": {
        "process_nm": 10000,
        "transistors": 3500,
        "data_width_bits": 8,
        "address_width_bits": 14,
        "base_clock_mhz": 0.5,
        "cores": 1,
        "package": "18-pin DIP"
      }
    }
  ]
}
```

| Field | Required | Notes |
| --- | --- | --- |
| `id` | yes | Unique across the whole database: `<manufacturer>-<model>`, lowercase, hyphens. |
| `name` | yes | Model name without the manufacturer (`"Core i9-14900K"`, `"GeForce RTX 4090"`). |
| `family` | yes | Product line used for grouping (`"Core i9"`, `"GeForce 40"`, `"Snapdragon 8"`). |
| `released` | yes | `YYYY`, `YYYY-MM` or `YYYY-MM-DD`. |
| `specs` | yes | Flat object of spec keys to values; see below. |

### Spec values

Keys come from `schema.json` (`groups[].fields[].key`). Values are numbers, strings, or
arrays of strings. Use the unit the key names:

- `*_mhz` in MHz (a 3.2 GHz clock is `3200`), `*_kb` in KiB, `*_mb` in MiB (a 24 GB card is
  `24576`), `process_nm` in nanometres (10 µm is `10000`), `tdp_w` in watts,
  `transistors` as a plain integer, `die_size_mm2` in mm², `*_tflops` in TFLOPS (dense, no
  sparsity), `*_tops` in TOPS, `launch_price_usd` in US dollars.
- `memory_speed_mts` is the rated memory speed in MT/s (`DDR5-6400` is `6400`);
  `memory_bandwidth_gbs` is the theoretical figure for the chip's own memory controller.
- `pcie_version` is a decimal string (`"5.0"`), `pcie_lanes` the lane count.
- `peak_temperature_c` in °C. `unlocked_multiplier` and `ecc_memory` are booleans; use `na`
  when a chip never supported them.
- The `*_tensor_tflops` keys hold matrix-core throughput (`fp8_tensor_tflops`,
  `fp16_tensor_tflops`, `bf16_tensor_tflops`, `fp4_tensor_tflops`, …), while `fp8_tflops`
  and `fp4_tflops` hold the same dense figure for accelerators that have no separate
  non-tensor path.
- `l1_cache_kb`, `l2_cache_kb`, `l3_cache_kb` are totals for the chip.
- `instruction_sets`, `graphics_api` and `connectivity` are arrays of strings.
- Benchmark keys (`geekbench6_*`, `cinebench_*`) hold the typical stock score as the
  benchmark reports it.
- Leave a spec out rather than guess. Omitted specs simply aren't shown.

### Completeness

`schema.json`'s `expected` lists, per category, the specs a complete entry has, and
`expected_since` adds specs (benchmarks) for chips released in or after a year. Each chip
lists, under `"na"`, the expected keys that don't apply to it or were never published:

```json
{ "id": "intel-80486dx", "specs": { ... }, "na": ["l3_cache_kb", "boost_clock_mhz"] }
```

`go run ./cmd/indexgen -report` lists every chip's missing specs; `-only <id prefix>` narrows
it to one manufacturer.

Keys not in `schema.json` are still shown, under "Other", with a label derived from the key;
adding them to `schema.json` later gives them a proper label, unit and group without an app
update.
