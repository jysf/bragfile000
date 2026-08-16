#!/usr/bin/env sh
# spark-chart.sh — render `brag spark --format json` as a self-contained HTML chart.
#
# Reads the DEC-014 spark envelope on stdin, writes standalone HTML on stdout.
# A stacked bar per bucket: height = total entries, segments coloured by project.
#
# Requires jq only — the same dependency the brag plugin hooks already assume.
# No network, no build step, nothing added to the `brag` binary: DEC-031 named an
# external-plotter pipe as the layer above in-terminal glyphs, and the `--format
# json` envelope is the seam. That is why this is a recipe and not a subcommand.
set -eu

usage() {
    cat <<'HELPTEXT'
spark-chart.sh — chart `brag spark` as standalone HTML

USAGE
    brag spark [--week|--month|--quarter] --format json | scripts/spark-chart.sh > pulse.html
    open pulse.html

EXAMPLES
    # last 13 weeks, one bar per week
    brag spark --quarter --format json | scripts/spark-chart.sh > pulse.html

    # last 7 days, one bar per day
    brag spark --week --format json | scripts/spark-chart.sh > week.html

    # one project against the total
    brag spark --quarter --project bragfile --format json | scripts/spark-chart.sh > p.html

OUTPUT
    One HTML document on stdout: a stacked bar chart, a legend, and the same
    data as a table. Theme-aware (light/dark), no external assets, no network.

NOTES
    `brag spark` reports only the top 8 projects in `by_project`, so the
    remainder is folded into a neutral "Other" segment — bar height always
    equals `total.series`, and nothing is silently dropped.

REQUIRES
    jq
HELPTEXT
}

case "${1:-}" in
    -h|--help) usage; exit 0 ;;
    "") ;;
    *) printf 'spark-chart: unknown argument %s\n\n' "$1" >&2; usage >&2; exit 1 ;;
esac

if [ -t 0 ]; then
    printf 'spark-chart: reads the output of: brag spark --format json\n\n' >&2
    usage >&2
    exit 1
fi

command -v jq >/dev/null 2>&1 || { echo "spark-chart: jq is required" >&2; exit 1; }

# One jq pass: envelope in, HTML out.
#
# Geometry is in px against a fixed 940x380 viewBox that scales to its container.
# The eight-slot categorical palette is in fixed order, so colour follows the
# PROJECT and not its rank — a project keeps its hue as the window changes.
# `brag spark` already tops out at 8 project rows, so a 9th hue never arises.
jq -r '
  ( 940 ) as $W | ( 380 ) as $H
  | ( 56 ) as $padL | ( 16 ) as $padR | ( 20 ) as $padT | ( 52 ) as $padB
  | ( $W - $padL - $padR ) as $plotW
  | ( $H - $padT - $padB ) as $plotH
  | .window as $win
  | ( $win.buckets ) as $n
  | ( $win.bucket_width_days ) as $wd
  | ( $win.start | strptime("%Y-%m-%dT%H:%M:%SZ") | mktime ) as $t0
  | .by_project as $projects
  | ( .total.series ) as $totSeries
  | ( [ .total.series[] ] | max ) as $rawPeak
  | ( if $rawPeak <= 0 then 1 else $rawPeak end ) as $peak
  | ( ($peak / 4) | ceil ) as $step
  | ( $step * 4 ) as $yMax
  | ( $plotW / $n ) as $slot
  | ( if ($slot - 8) < 4 then 4 else ($slot - 8) end ) as $barW
  | ( ["#2a78d6","#eb6834","#1baf7a","#eda100","#e87ba4","#008300","#4a3aa7","#e34948"] ) as $LH
  | ( ["#3987e5","#d95926","#199e70","#c98500","#d55181","#008300","#9085e9","#e66767"] ) as $DH
  | ( [ range(0; $n) | ($t0 + (. * $wd * 86400)) | strftime("%b %-d") ] ) as $labels
  | ( .scope ) as $scope
  | ( .generated_at ) as $gen
  # The remainder past the top-8 rows, per bucket and in total.
  | ( [ range(0; $n) as $i
        | $totSeries[$i] - ( [ $projects[].series[$i] ] | add ) ] ) as $otherSeries
  | ( .total.count - ( [ $projects[].count ] | add ) ) as $otherCount

  # Everything below is one comma-joined stream of lines; jq -r prints each.
  | ( "<style>"
  , ".viz{color-scheme:light;--surface:#fcfcfb;--plane:#f9f9f7;--ink:#0b0b0b;--ink2:#52514e;--muted:#898781;--grid:#e1e0d9;--axis:#c3c2b7;--other:#b4b2aa;"
  , ( [ range(0; ($LH|length)) | "--s\(.+1):\($LH[.]);" ] | join("") )
  , "}"
  , "@media (prefers-color-scheme:dark){:root:where(:not([data-theme=\"light\"])) .viz{color-scheme:dark;--surface:#1a1a19;--plane:#0d0d0d;--ink:#fff;--ink2:#c3c2b7;--muted:#898781;--grid:#2c2c2a;--axis:#383835;--other:#5c5b56;"
  , ( [ range(0; ($DH|length)) | "--s\(.+1):\($DH[.]);" ] | join("") )
  , "}}"
  , ":root[data-theme=\"dark\"] .viz{color-scheme:dark;--surface:#1a1a19;--plane:#0d0d0d;--ink:#fff;--ink2:#c3c2b7;--muted:#898781;--grid:#2c2c2a;--axis:#383835;--other:#5c5b56;"
  , ( [ range(0; ($DH|length)) | "--s\(.+1):\($DH[.]);" ] | join("") )
  , "}"
  , "body{margin:0;background:var(--plane);color:var(--ink);font:14px/1.5 ui-sans-serif,-apple-system,system-ui,sans-serif;}"
  , ".viz{background:var(--plane);padding:28px 24px 36px;}"
  , ".wrap{max-width:1000px;margin:0 auto;}"
  , "h1{font-size:19px;margin:0 0 2px;letter-spacing:-.01em;}"
  , ".sub{color:var(--ink2);font-size:13px;margin:0 0 22px;}"
  , ".card{background:var(--surface);border:1px solid var(--grid);border-radius:10px;padding:18px 16px 10px;overflow-x:auto;}"
  , ".legend{display:flex;flex-wrap:wrap;gap:6px 18px;margin:16px 2px 0;padding:0;list-style:none;}"
  , ".legend li{display:flex;align-items:center;gap:7px;font-size:13px;color:var(--ink2);}"
  , ".sw{width:11px;height:11px;border-radius:3px;flex:none;}"
  , ".legend b{color:var(--ink);font-weight:600;font-variant-numeric:tabular-nums;}"
  , "table{border-collapse:collapse;width:100%;margin-top:26px;font-size:12.5px;font-variant-numeric:tabular-nums;}"
  , "th,td{padding:5px 8px;text-align:right;border-bottom:1px solid var(--grid);white-space:nowrap;}"
  , "th:first-child,td:first-child{text-align:left;}"
  , "thead th{color:var(--muted);font-weight:500;}"
  , "tbody tr:last-child td{border-bottom:none;font-weight:600;}"
  , "td.z{color:var(--muted);}"
  , "rect.seg{transition:opacity .12s;}"
  , "g.bar:hover rect.seg{opacity:.45;}"
  , "g.bar:hover rect.seg:hover{opacity:1;}"
  , "figcaption{color:var(--muted);font-size:12px;margin-top:14px;}"
  , "</style>"

  , "<div class=\"viz\"><div class=\"wrap\">"
  , "<h1>Brag pulse — \($scope), by project</h1>"
  , "<p class=\"sub\">\($n) buckets of \($wd) days · \([$projects[].count]|add) entries across \($projects|length) projects · generated \($gen)</p>"
  , "<figure style=\"margin:0\"><div class=\"card\">"
  , "<svg viewBox=\"0 0 \($W) \($H)\" width=\"100%\" role=\"img\" aria-label=\"Stacked bars of brag entries per bucket, coloured by project\">"

  # gridlines + y ticks
  , ( [ range(0; 5) as $t
        | ( $padT + $plotH - ($t / 4 * $plotH) ) as $y
        | "<line x1=\"\($padL)\" x2=\"\($W - $padR)\" y1=\"\($y)\" y2=\"\($y)\" stroke=\"var(--grid)\" stroke-width=\"1\"/>"
          + "<text x=\"\($padL - 10)\" y=\"\($y + 4)\" text-anchor=\"end\" font-size=\"11\" fill=\"var(--muted)\">\($step * $t)</text>"
      ] | join("") )

  # stacked bars, 2px gap between segments, 4px rounded top on the stack
  , ( [ range(0; $n) as $i
        | ( $padL + ($i * $slot) + (($slot - $barW) / 2) ) as $x
        # `brag spark` shows the top 8 projects; anything past that is real
        # activity the bars would otherwise silently drop. Fold it into a
        # neutral "Other" segment so bar height always equals total.series.
        | ( reduce range(0; ($projects|length)) as $p
              ({ acc: 0, out: [] };
                ( $projects[$p].series[$i] ) as $v
                | if $v > 0 then
                    { acc: (.acc + $v),
                      out: (.out + [{ p: $p, v: $v, base: .acc }]) }
                  else . end)
            | ( $totSeries[$i] - .acc ) as $rest
            | if $rest > 0
              then .out + [{ p: -1, v: $rest, base: .acc }]
              else .out end ) as $segs
        | ( $totSeries[$i] ) as $tot
        | "<g class=\"bar\">"
          + ( [ $segs[]
                | ( .v / $yMax * $plotH ) as $h
                | ( $padT + $plotH - ((.base + .v) / $yMax * $plotH) ) as $y
                | ( if $h > 2 then $h - 2 else $h end ) as $hh
                # p == -1 is the folded "Other" bucket; jq indexes -1 from the
                # END of an array, so it must be branched on, never looked up.
                | ( if .p < 0 then "Other" else $projects[.p].project end ) as $name
                | ( if .p < 0 then "var(--other)" else "var(--s\(.p + 1))" end ) as $fill
                | "<rect class=\"seg\" x=\"\($x)\" y=\"\($y)\" width=\"\($barW)\" height=\"\($hh)\" rx=\"2\" fill=\"\($fill)\">"
                  + "<title>\($name) · \($labels[$i]) · \(.v)</title></rect>"
              ] | join("") )
          + ( if $tot > 0
              then "<text x=\"\($x + $barW/2)\" y=\"\($padT + $plotH - ($tot / $yMax * $plotH) - 6)\" text-anchor=\"middle\" font-size=\"10.5\" fill=\"var(--ink2)\">\($tot)</text>"
              else "" end )
          + "<text x=\"\($x + $barW/2)\" y=\"\($padT + $plotH + 18)\" text-anchor=\"middle\" font-size=\"10.5\" fill=\"var(--muted)\">\($labels[$i])</text>"
          + "</g>"
      ] | join("") )

  , "<line x1=\"\($padL)\" x2=\"\($W - $padR)\" y1=\"\($padT + $plotH)\" y2=\"\($padT + $plotH)\" stroke=\"var(--axis)\" stroke-width=\"1\"/>"
  , "</svg>"
  , "</div>"

  # legend — identity is never colour-alone
  , "<ul class=\"legend\">"
  , ( [ range(0; ($projects|length))
        | "<li><span class=\"sw\" style=\"background:var(--s\(.+1))\"></span>\($projects[.].project) <b>\($projects[.].count)</b></li>" ]
      | join("") )
  , ( if $otherCount > 0
      then "<li><span class=\"sw\" style=\"background:var(--other)\"></span>Other <b>\($otherCount)</b></li>"
      else "" end )
  , "</ul>"
  , "<figcaption>Bar height is entries per bucket; segments are projects in fixed colour order. Hover a segment for its count.</figcaption>"
  , "</figure>"

  # table view — the relief for sub-3:1 hues on the light surface
  , "<table><caption class=\"sub\" style=\"text-align:left;margin:26px 0 6px\">Same data, as a table</caption>"
  , "<thead><tr><th>Project</th>"
  , ( [ $labels[] | "<th>\(.)</th>" ] | join("") )
  , "<th>Total</th></tr></thead><tbody>"
  , ( [ range(0; ($projects|length)) as $p
        | "<tr><td>\($projects[$p].project)</td>"
          + ( [ $projects[$p].series[] | if . == 0 then "<td class=\"z\">·</td>" else "<td>\(.)</td>" end ] | join("") )
          + "<td>\($projects[$p].count)</td></tr>" ] | join("") )
  , ( if $otherCount > 0
      then "<tr><td>Other</td>"
           + ( [ $otherSeries[] | if . == 0 then "<td class=\"z\">·</td>" else "<td>\(.)</td>" end ] | join("") )
           + "<td>\($otherCount)</td></tr>"
      else "" end )
  , "<tr><td>Total</td>"
  , ( [ .total.series[] | if . == 0 then "<td class=\"z\">·</td>" else "<td>\(.)</td>" end ] | join("") )
  , "<td>\(.total.count)</td></tr>"
  , "</tbody></table>"
  , "</div></div>"
  )
'
