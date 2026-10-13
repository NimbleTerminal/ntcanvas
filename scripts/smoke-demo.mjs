// Exercise one real WASM program through booba's JS bridge in a fresh process.
// This checks startup/input/resize/exit, not browser canvas rendering.
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import vm from 'node:vm';
import { setTimeout as delay } from 'node:timers/promises';
const demo = process.argv[2];
assert.ok(['logo', 'grids', 'lines', 'viewports'].includes(demo));
vm.runInThisContext(await readFile('web/wasm_exec.js', 'utf8'));
const go = new Go();
let exitCode;
go.exit = code => { exitCode = code; };
const { instance } = await WebAssembly.instantiate(await readFile(`web/demos/${demo}.wasm`), go.importObject);
const timeout = setTimeout(() => { console.error(`${demo}: timed out`); process.exit(1); }, 15000);
const running = go.run(instance);
try {
  for (let n = 0; typeof globalThis.bubbletea_read !== 'function' && n < 100; n++) await delay(20);
  assert.equal(typeof globalThis.bubbletea_read, 'function', 'booba bridge installed');
  bubbletea_resize(80, 28);
  let output = '';
  for (let n = 0; n < 100 && !output.toLowerCase().includes('quit'); n++) {
    await delay(25);
    output += bubbletea_read();
  }
  assert.ok(output.toLowerCase().includes('quit'), 'initial terminal view rendered');
  assert.ok(output.includes('\x1b]7501;?'), 'application queries action support');
  bubbletea_write('\x1b]7501;?\x1b\\');
  let report;
  for (let n = 0; n < 100 && !report; n++) {
    await delay(25);
    output += bubbletea_read();
    report = output.match(/\x1b\]7501;state=idle:[^\x1b]*:msg=([A-Za-z0-9+/=]+)\x1b\\/);
  }
  assert.ok(report, 'application publishes OSC action snapshot');
  const payload = Buffer.from(report[1], 'base64').toString('utf8');
  const snapshot = JSON.parse(payload.split(' | booba-actions/v1 ')[1]);
  const quit = snapshot.actions.find(action => action.label === 'Quit');
  assert.ok(quit?.enabled, 'quit action enabled');
  if (demo === 'logo') {
    for (const label of ['move up', 'move down', 'move left', 'move right', 'Switch canvas']) {
      assert.ok(snapshot.actions.some(action => action.label === label && action.enabled), `${label} action`);
    }
    const right = snapshot.actions.find(action => action.label === 'move right');
    assert.equal(right.keys[right.preferred].key, 'right');
  }
  // Status is emitted from the actual Go model through booba, not from a
  // test-only JS hook. Assert exact focus/coordinates after each separate input.
  async function expectSummary(expected) {
    let received = '';
    for (let n = 0; n < 100; n++) {
      await delay(20);
      received += bubbletea_read();
      for (const match of received.matchAll(/\x1b\]7501;state=idle:[^\x1b]*:msg=([A-Za-z0-9+/=]+)\x1b\\/g)) {
        const message = Buffer.from(match[1], 'base64').toString('utf8');
        if (message.split(' | booba-actions/v1 ')[0] === expected) return;
      }
    }
    assert.fail(`Missing viewport state: ${expected}; received ${JSON.stringify(received)}`);
  }
  async function expectViewport(focus, left, right) {
    return expectSummary(`Canvas viewports: focus=${focus} left=(${left}) right=(${right})`);
  }
  async function input(bytes, focus, left, right) {
    bubbletea_read(); // discard reports from the previous step
    bubbletea_write(bytes);
    await expectViewport(focus, left, right);
  }
  if (demo === 'logo') {
  bubbletea_read();
  bubbletea_resize(60, 18);
  // The original demo retains its canvas/viewport dimensions on terminal resize.
  bubbletea_write('\x1b]7501;?\x1b\\');
  await expectViewport('left', '0,0', '0,0');
  await input('\x1b[C', 'left', '1,0', '0,0');
  await input('\r', 'right', '1,0', '0,0');
  await input('\x1b[<65;23;6M', 'right', '1,0', '0,1');
  // SGR mouse coordinates are one-based. Both zones start on row 5;
  // the left zone starts at column 2 and the right zone at column 22.
  bubbletea_write('\x1b[<0;23;6M');
  await input('\x1b[<32;24;7M', 'right', '1,0', '1,2');
  bubbletea_write('\x1b[<0;24;7m');
  await input('\x1b[<0;3;6M', 'left', '1,0', '1,2');
  await input('\x1b[<32;4;7M', 'left', '2,1', '1,2');
  bubbletea_write('\x1b[<0;4;7m');
  } else {
    const checkAction = label => assert.ok(snapshot.actions.some(action => action.label === label && action.enabled), `${label} action`);
    async function send(bytes, summary) {
      bubbletea_read();
      bubbletea_write(bytes);
      await delay(25);
      bubbletea_write('\x1b]7501;?\x1b\\');
      await expectSummary(summary);
    }
    bubbletea_resize(60, 24);
    if (demo === 'grids') {
      checkAction('Toggle single column'); checkAction('Clear wide cell');
      await send('s', 'Grids: single-column=true cleared=false');
      await send('c', 'Grids: single-column=true cleared=true');
      await send('r', 'Grids: single-column=false cleared=false');
    } else if (demo === 'lines') {
      checkAction('Lower endpoint'); checkAction('Toggle glyph');
      await send('\x1b[B', 'Lines: offset=1 blocks=false');
      await send('g', 'Lines: offset=1 blocks=true');
      await send('r', 'Lines: offset=0 blocks=false');
    } else {
      checkAction('Switch viewport'); checkAction('Edit shared grid'); checkAction('Resize grid');
      await send('\x1b[C', 'Viewports: focus=left left=(1,0) right=(12,4) grid=72x18 marks=0');
      await send('\r', 'Viewports: focus=right left=(1,0) right=(12,4) grid=72x18 marks=0');
      await send('\x1b[B', 'Viewports: focus=right left=(1,0) right=(12,5) grid=72x18 marks=0');
      await send('e', 'Viewports: focus=right left=(1,0) right=(12,5) grid=72x18 marks=1');
      await send('g', 'Viewports: focus=right left=(1,0) right=(12,5) grid=28x10 marks=1');
      await send('b', 'Viewports: focus=none left=(1,0) right=(12,5) grid=28x10 marks=1');
      await send('\x1b[C', 'Viewports: focus=none left=(1,0) right=(12,5) grid=28x10 marks=1');
      await send('r', 'Viewports: focus=left left=(0,0) right=(12,4) grid=28x10 marks=1');
    }
  }
  bubbletea_write(quit.keys[quit.preferred].key);
  await running;
  assert.equal(exitCode, 0);
  assert.equal(globalThis.bubbletea_exited, true, 'exit published to adapter');
  console.log(`${demo}: startup, actions, resize, exact demo state transitions, and clean exit passed`);
} finally {
  clearTimeout(timeout);
}
// Like go_js_wasm_exec, terminate the host process after Go exits. The Go
// runtime may still have scheduled timer callbacks in this Node process.
process.exit(0);
