#!/usr/bin/env python3
"""Explicitly regenerate the fixture using pinned, unmodified upstream code.

Reads a local NTCharts git checkout; writes only ntcanvas's fixture and a
throwaway module. Normal tests and CI require neither the checkout nor this tool.
"""
import argparse
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
BASELINE = 'bed05cf56936231c7c4b44833b7ad0457aedc5d6'
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--upstream', type=Path, required=True)
args = parser.parse_args()
with tempfile.TemporaryDirectory(prefix='ntcanvas-upstream-') as tmp:
    root = Path(tmp).resolve()
    for name in ['go.mod', 'go.sum', 'canvas/canvas.go', 'canvas/options.go',
                 'canvas/updatehandler.go']:
        data = subprocess.check_output(['git', 'show', f'{BASELINE}:{name}'], cwd=args.upstream)
        target = root / name
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_bytes(data)
    scenario = (ROOT / 'testdata/upstream/scenarios.go.txt').read_text()
    scenario = scenario.replace('nimbleterminal.dev/ntcanvas', 'github.com/NimbleMarkets/ntcharts/v2/canvas')
    (root / 'internal/parity').mkdir(parents=True)
    (root / 'internal/parity/scenarios.go').write_text(scenario)
    (root / 'cmd/snapshot').mkdir(parents=True)
    (root / 'cmd/snapshot/main.go').write_text('''package main
import ("os"; "github.com/NimbleMarkets/ntcharts/v2/internal/parity")
func main() { b,err:=parity.Snapshot(); if err!=nil {panic(err)}; os.Stdout.Write(b) }
''')
    data = subprocess.check_output(['go', 'run', '-mod=readonly', './cmd/snapshot'], cwd=root,
                                   env=os.environ | {'GOWORK': 'off', 'GOOS': '', 'GOARCH': ''})
    (ROOT / 'testdata/upstream/snapshot.json').write_bytes(data + b'\n')
