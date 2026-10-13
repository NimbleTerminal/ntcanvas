// Interpret canvas ANSI with the pinned demo emulator; assert screen cells,
// not escape-sequence spelling. No DOM, browser painting, or PTY is involved.
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { Ghostty, CellFlags } from '../web/ghostty-web/ghostty-web.js';

const frames = JSON.parse(await readFile('dist/screen-frames.json', 'utf8'));
const expected = JSON.parse(await readFile('testdata/screen/frames.json', 'utf8'));
const bytes = await readFile('web/ghostty-web/ghostty-vt.wasm');
const { instance } = await WebAssembly.instantiate(bytes, { env: { log() {} } });
const ghostty = new Ghostty(instance);
assert.equal(frames.length, expected.length);
// Reuse the terminal across frames so stale characters/styles are observable.
const terminal = ghostty.createTerminal(frames[0].width, frames[0].height);
try {
  for (const [i, frame] of frames.entries()) {
    assert.equal(frame.name, expected[i].name);
    terminal.resize(frame.width, frame.height);
    // Small chunks exercise partial escape sequences and UTF-8 boundaries.
    const output = new TextEncoder().encode(frame.output);
    for (let n = 0; n < output.length; n += 3) terminal.write(output.subarray(n, n + 3));
    const cells = terminal.getViewport();
    // Ghostty retains pool capacity when shrinking; only the visible prefix counts.
    assert.ok(cells.length >= frame.width * frame.height);
    const rows = [];
    const styles = [];
    const clusters = [];
    for (let y = 0; y < frame.height; y++) {
      let row = '';
      for (let x = 0; x < frame.width; x++) {
        const c = cells[y * frame.width + x];
        row += String.fromCodePoint(c.codepoint || 32);
        if (c.width !== 1 || c.grapheme_len) {
          clusters.push({ x, y, width: c.width, extra: c.grapheme ?? [] });
        }
        assert.equal(c.flags & ~CellFlags.BOLD, 0, `${frame.name} unexpected attributes at ${x},${y}`);
        if (!c.fgIsDefault || !c.bgIsDefault || c.flags) {
          styles.push({ x, y,
            fg: c.fgIsDefault ? null : [c.fg_r, c.fg_g, c.fg_b],
            bg: c.bgIsDefault ? null : [c.bg_r, c.bg_g, c.bg_b],
            bold: Boolean(c.flags & CellFlags.BOLD) });
        }
      }
      rows.push(row);
    }
    const actual = { name: frame.name, rows, styles };
    if (clusters.length) actual.clusters = clusters;
    assert.deepEqual(actual, expected[i], `${frame.name} screen`);
  }
} finally {
  terminal.free();
}
console.log(`screen: ${frames.length} golden frames passed (text, colors, attributes, resize, overwrite)`);
