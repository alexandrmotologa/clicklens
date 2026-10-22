const puppeteer = require('puppeteer-core');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

async function main() {
  const chromePath = 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
  const outDir = path.resolve(__dirname, '../docs/images');
  const framesDir = path.resolve(__dirname, '../docs/images/frames');
  if (!fs.existsSync(outDir)) fs.mkdirSync(outDir, { recursive: true });
  if (!fs.existsSync(framesDir)) fs.mkdirSync(framesDir, { recursive: true });

  console.log('Launching headless Chrome from:', chromePath);
  const browser = await puppeteer.launch({
    executablePath: chromePath,
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--window-size=1280,820'],
    defaultViewport: { width: 1280, height: 820 }
  });

  const page = await browser.newPage();
  console.log('Navigating to http://127.0.0.1:8092...');
  await page.goto('http://127.0.0.1:8092/', { waitUntil: 'networkidle0' });

  await new Promise(r => setTimeout(r, 1500));

  // 1. Tab: Parts & Partitions Heatmap
  console.log('Capturing Desktop: Parts Heatmap & Age Histogram...');
  const partsPath = path.join(outDir, 'clicklens_parts_heatmap.png');
  await page.screenshot({ path: partsPath });

  // 2. Tab: Background Merges
  console.log('Switching to Background Merges tab...');
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.textContent && el.textContent.includes('Background Merges'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 1200));
  const mergesPath = path.join(outDir, 'clicklens_merges_monitor.png');
  await page.screenshot({ path: mergesPath });

  // 3. Tab: Query Pipeline DAG
  console.log('Switching to Query Pipeline DAG tab...');
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.textContent && el.textContent.includes('Query Pipeline DAG'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 800));
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.textContent && el.textContent.includes('Profile Execution Pipeline'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 1500));
  const pipelinePath = path.join(outDir, 'clicklens_pipeline_dag.png');
  await page.screenshot({ path: pipelinePath });

  // 4. Tab: Historical Query Log
  console.log('Switching to Historical Query Log tab...');
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.textContent && el.textContent.includes('Historical Query Log'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 1200));
  const queryLogPath = path.join(outDir, 'clicklens_query_log.png');
  await page.screenshot({ path: queryLogPath });

  // 5. Tab: Codec & Compression Auditor (Benchmark Simulator & Migration Playbook)
  console.log('Switching to Codec Optimizer tab...');
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.textContent && el.textContent.includes('Codec & Compression Auditor'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 1200));
  const codecsPath = path.join(outDir, 'clicklens_codec_optimizer.png');
  await page.screenshot({ path: codecsPath });

  // Open Benchmark Simulator Modal
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.textContent && el.textContent.includes('Benchmark'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 1500));
  const benchmarkPath = path.join(outDir, 'clicklens_codec_benchmark.png');
  await page.screenshot({ path: benchmarkPath });

  // Close benchmark modal
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.textContent && el.textContent.includes('Close Benchmark'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 500));

  // 6. Tab: Serverless Local Workbench (with CSV to Parquet converter)
  console.log('Switching to Local Workbench tab...');
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.textContent && el.textContent.includes('Serverless Local Workbench'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 800));
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.textContent && el.textContent.includes('Inspect File'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 1200));
  const localPath = path.join(outDir, 'clicklens_local_workbench.png');
  await page.screenshot({ path: localPath });

  // 7. Health Audit Diagnostic Report Modal
  console.log('Opening Diagnostic Health Audit Report...');
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.title && el.title.includes('diagnostic audit')) ||
              buttons.find(el => el.textContent && el.textContent.includes('Report'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 1500));
  const reportPath = path.join(outDir, 'clicklens_diagnostic_report.png');
  await page.screenshot({ path: reportPath });

  // Close report modal
  await page.keyboard.press('Escape');
  await new Promise(r => setTimeout(r, 500));

  // 8. Command Palette (Ctrl+K)
  console.log('Opening Command Palette...');
  await page.keyboard.down('Control');
  await page.keyboard.press('k');
  await page.keyboard.up('Control');
  await new Promise(r => setTimeout(r, 800));
  const palettePath = path.join(outDir, 'clicklens_command_palette.png');
  await page.screenshot({ path: palettePath });

  // Close command palette
  await page.keyboard.press('Escape');
  await new Promise(r => setTimeout(r, 500));

  // 9. Tab: SQL Console & Processes
  console.log('Switching to SQL Console tab...');
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.textContent && el.textContent.includes('SQL Console & Processes'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 800));
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.textContent && el.textContent.includes('Execute Query'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 1200));
  const consolePath = path.join(outDir, 'clicklens_query_console.png');
  await page.screenshot({ path: consolePath });

  // ==========================================
  // MOBILE RESPONSIVENESS PREVIEW (390 x 844)
  // ==========================================
  console.log('Switching viewport to Mobile Responsive (390x844)...');
  await page.setViewport({ width: 390, height: 844, isMobile: true, hasTouch: true });

  // Mobile Parts Tab
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.textContent && el.textContent.includes('Parts & Partitions'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 1200));
  const mobilePartsPath = path.join(outDir, 'clicklens_mobile_parts.png');
  await page.screenshot({ path: mobilePartsPath });
  console.log('Saved Mobile Parts view to:', mobilePartsPath);

  // Mobile Background Merges Tab
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.textContent && el.textContent.includes('Background Merges'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 1000));
  const mobileMergesPath = path.join(outDir, 'clicklens_mobile_merges.png');
  await page.screenshot({ path: mobileMergesPath });
  console.log('Saved Mobile Merges view to:', mobileMergesPath);

  // Mobile Diagnostic Report Modal
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const b = buttons.find(el => el.title && el.title.includes('diagnostic audit')) ||
              buttons.find(el => el.textContent && el.textContent.includes('Report'));
    if (b) b.click();
  });
  await new Promise(r => setTimeout(r, 1200));
  const mobileReportPath = path.join(outDir, 'clicklens_mobile_report.png');
  await page.screenshot({ path: mobileReportPath });
  console.log('Saved Mobile Report view to:', mobileReportPath);

  // Reset viewport back to desktop for clean exit
  await page.setViewport({ width: 1280, height: 820 });

  // Record animated sequence
  console.log('Generating frames for animated demo GIF...');
  const sequence = [
    { name: 'parts', file: partsPath, durationSec: 2 },
    { name: 'merges', file: mergesPath, durationSec: 2 },
    { name: 'pipeline', file: pipelinePath, durationSec: 2 },
    { name: 'query_log', file: queryLogPath, durationSec: 2 },
    { name: 'codecs', file: codecsPath, durationSec: 2 },
    { name: 'benchmark', file: benchmarkPath, durationSec: 2 },
    { name: 'local', file: localPath, durationSec: 2 },
    { name: 'report', file: reportPath, durationSec: 2 },
    { name: 'palette', file: palettePath, durationSec: 2 },
    { name: 'console', file: consolePath, durationSec: 2 }
  ];

  let frameIdx = 0;
  const fps = 2; // 2 frames per second
  for (const item of sequence) {
    const frameCount = item.durationSec * fps;
    for (let f = 0; f < frameCount; f++) {
      const frameFilename = `frame_${String(frameIdx).padStart(4, '0')}.png`;
      fs.copyFileSync(item.file, path.join(framesDir, frameFilename));
      frameIdx++;
    }
  }

  await browser.close();

  console.log('Compiling demo GIF using ffmpeg...');
  const gifPath = path.join(outDir, 'clicklens_demo.gif');
  const ffmpegCmd = `"C:\\Users\\alexander\\AppData\\Local\\Microsoft\\WinGet\\Links\\ffmpeg.exe" -y -framerate 2 -i "${framesDir}\\frame_%04d.png" -vf "fps=2,scale=1024:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" "${gifPath}"`;
  execSync(ffmpegCmd, { stdio: 'inherit' });

  // Cleanup frames
  fs.rmSync(framesDir, { recursive: true, force: true });
  console.log('✓ Successfully created all desktop & mobile screenshots and animated demo GIF at:', gifPath);
}

main().catch(err => {
  console.error('Error during asset capture:', err);
  process.exit(1);
});
