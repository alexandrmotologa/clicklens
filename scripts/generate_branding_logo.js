const fs = require('fs');
const path = require('path');
const puppeteer = require('puppeteer-core');

function buildLynxLogoSvg() {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="1024" height="1024">
  <defs>
    <clipPath id="squircle-clip">
      <rect x="24" y="24" width="976" height="976" rx="220" />
    </clipPath>

    <!-- ClickHouse Yellow / Amber Gradient -->
    <linearGradient id="ch-amber" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#fbbf24"/>
      <stop offset="60%" stop-color="#f59e0b"/>
      <stop offset="100%" stop-color="#d97706"/>
    </linearGradient>

    <!-- Electric Cyan Telemetry Gradient -->
    <linearGradient id="cyan-glow" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#38bdf8"/>
      <stop offset="100%" stop-color="#0284c7"/>
    </linearGradient>

    <!-- Obsidian Titanium Body Facet Gradients -->
    <linearGradient id="facet-light" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#334155"/>
      <stop offset="100%" stop-color="#1e293b"/>
    </linearGradient>

    <linearGradient id="facet-mid" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#1e293b"/>
      <stop offset="100%" stop-color="#0f172a"/>
    </linearGradient>

    <linearGradient id="facet-dark" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#0f172a"/>
      <stop offset="100%" stop-color="#050811"/>
    </linearGradient>

    <linearGradient id="facet-highlight" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#64748b"/>
      <stop offset="100%" stop-color="#475569"/>
    </linearGradient>

    <filter id="subtle-shadow" x="-10%" y="-10%" width="120%" height="120%">
      <feDropShadow dx="0" dy="16" stdDeviation="20" flood-color="#000000" flood-opacity="0.14" />
    </filter>
  </defs>

  <!-- Luxury White Squircle Container -->
  <rect x="24" y="24" width="976" height="976" rx="220" fill="#ffffff" stroke="#e2e8f0" stroke-width="6" />

  <g clip-path="url(#squircle-clip)">
    <g transform="translate(512, 512)" filter="url(#subtle-shadow)">

      <!-- Hexagonal Architectural Gateway (ClickHouse & Storage Partition Frame) -->
      <polygon points="
        0,-390
        338,-195
        338,195
        0,390
        -338,195
        -338,-195
      " fill="none" stroke="#0f172a" stroke-width="36" stroke-linejoin="round" />

      <!-- Inner Telemetry Laser Accent Frame (Amber to Cyan) -->
      <polygon points="
        0,-355
        307,-177
        307,177
        0,355
        -307,177
        -307,-177
      " fill="none" stroke="url(#ch-amber)" stroke-width="5" opacity="0.6" stroke-dasharray="24, 16" />

      <!-- 100% Watertight Solid Base Silhouette (Prevents background leaks) -->
      <path d="
        M 0,-330
        L 85,-290
        L 135,-370
        L 175,-385
        L 180,-300
        L 230,-190
        L 260,-90
        L 270,30
        L 210,140
        L 150,220
        L 70,275
        L 0,295
        L -70,275
        L -150,220
        L -210,140
        L -270,30
        L -260,-90
        L -230,-190
        L -180,-300
        L -175,-385
        L -135,-370
        L -85,-290
        Z
      " fill="#070a12" stroke="#0f172a" stroke-width="4" stroke-linejoin="round" />

      <!-- LYNX FACET MESH (Centered at 0,0) -->

      <!-- Crown & Forehead Central Facet -->
      <polygon points="0,-330 85,-290 0,-180" fill="url(#facet-light)" />
      <polygon points="0,-330 -85,-290 0,-180" fill="url(#facet-mid)" />

      <!-- Lynx Ear Tufts (The Iconic Sensory Antennas) -->
      <!-- Right Ear Base & Tuft -->
      <polygon points="85,-290 135,-370 180,-300" fill="url(#facet-highlight)" />
      <polygon points="135,-370 175,-385 180,-300" fill="url(#ch-amber)" />
      <polygon points="85,-290 180,-300 130,-220" fill="url(#facet-mid)" />
      <!-- Left Ear Base & Tuft -->
      <polygon points="-85,-290 -135,-370 -180,-300" fill="url(#facet-mid)" />
      <polygon points="-135,-370 -175,-385 -180,-300" fill="url(#ch-amber)" />
      <polygon points="-85,-290 -180,-300 -130,-220" fill="url(#facet-dark)" />

      <!-- Forehead Upper Side Plates -->
      <polygon points="85,-290 130,-220 0,-180" fill="url(#facet-mid)" />
      <polygon points="-85,-290 -130,-220 0,-180" fill="url(#facet-dark)" />

      <!-- Brow Ridge & Central Nose Bridge -->
      <polygon points="0,-180 65,-120 0,-30" fill="url(#facet-light)" />
      <polygon points="0,-180 -65,-120 0,-30" fill="url(#facet-mid)" />

      <!-- Upper Temples & Cheek Flares -->
      <polygon points="180,-300 230,-190 130,-220" fill="url(#facet-dark)" />
      <polygon points="-180,-300 -230,-190 -130,-220" fill="url(#facet-dark)" />

      <polygon points="130,-220 230,-190 170,-110" fill="url(#facet-mid)" />
      <polygon points="-130,-220 -230,-190 -170,-110" fill="url(#facet-dark)" />

      <polygon points="130,-220 170,-110 65,-120" fill="url(#facet-light)" />
      <polygon points="-130,-220 -170,-110 -65,-120" fill="url(#facet-mid)" />

      <!-- PRECISE OPTICAL LENS EYES (The ClickLens Vision) -->
      <!-- Right Eye Socket & Lens -->
      <polygon points="65,-120 145,-105 130,-65 60,-75" fill="#020617" stroke="#00f5ff" stroke-width="2" />
      <polygon points="75,-110 135,-100 125,-75 70,-80" fill="url(#ch-amber)" />
      <polygon points="98,-107 106,-106 102,-72 96,-73" fill="#070a12" /> <!-- Slit Pupil -->
      <circle cx="115" cy="-92" r="3.5" fill="#ffffff" /> <!-- Specular Glint -->

      <!-- Left Eye Socket & Lens -->
      <polygon points="-65,-120 -145,-105 -130,-65 -60,-75" fill="#020617" stroke="#00f5ff" stroke-width="2" />
      <polygon points="-75,-110 -135,-100 -125,-75 -70,-80" fill="url(#ch-amber)" />
      <polygon points="-98,-107 -106,-106 -102,-72 -96,-73" fill="#070a12" /> <!-- Slit Pupil -->
      <circle cx="-115" cy="-92" r="3.5" fill="#ffffff" /> <!-- Specular Glint -->

      <!-- Lower Eye Orbital Plates -->
      <polygon points="60,-75 130,-65 85,-20" fill="url(#facet-mid)" />
      <polygon points="-60,-75 -130,-65 -85,-20" fill="url(#facet-dark)" />

      <!-- Broad Lynx Cheeks & Facial Ruffs (Lateral Expansion) -->
      <polygon points="230,-190 260,-90 170,-110" fill="url(#facet-mid)" />
      <polygon points="-230,-190 -260,-90 -170,-110" fill="url(#facet-dark)" />

      <polygon points="170,-110 260,-90 195,-20" fill="url(#facet-light)" />
      <polygon points="-170,-110 -260,-90 -195,-20" fill="url(#facet-mid)" />

      <polygon points="170,-110 195,-20 130,-65" fill="url(#facet-mid)" />
      <polygon points="-170,-110 -195,-20 -130,-65" fill="url(#facet-dark)" />

      <!-- Outer Bearded Ruff (Lower Cheeks) -->
      <polygon points="260,-90 270,30 195,-20" fill="url(#facet-dark)" />
      <polygon points="-260,-90 -270,30 -195,-20" fill="url(#facet-dark)" />

      <polygon points="195,-20 270,30 210,140" fill="url(#facet-mid)" />
      <polygon points="-195,-20 -270,30 -210,140" fill="url(#facet-dark)" />

      <polygon points="195,-20 210,140 140,70" fill="url(#facet-highlight)" />
      <polygon points="-195,-20 -210,140 -140,70" fill="url(#facet-mid)" />

      <polygon points="130,-65 195,-20 85,-20" fill="url(#facet-dark)" />
      <polygon points="-130,-65 -195,-20 -85,-20" fill="url(#facet-dark)" />

      <!-- Snout & Whisker Pads (Titanium Slate, eliminating buck teeth illusion) -->
      <polygon points="0,-30 65,-120 85,-20 0,60" fill="url(#facet-light)" />
      <polygon points="0,-30 -65,-120 -85,-20 0,60" fill="url(#facet-mid)" />

      <polygon points="85,-20 140,70 55,80" fill="url(#facet-mid)" />
      <polygon points="-85,-20 -140,70 -55,80" fill="url(#facet-dark)" />

      <polygon points="85,-20 55,80 0,60" fill="url(#facet-highlight)" />
      <polygon points="-85,-20 -55,80 0,60" fill="url(#facet-mid)" />

      <!-- Nose Triangle (Geometric Obsidian) -->
      <polygon points="0,60 30,90 -30,90" fill="#090d16" stroke="url(#cyan-glow)" stroke-width="2" />

      <!-- Muzzle & Lower Jaw -->
      <polygon points="0,90 45,150 0,170" fill="url(#facet-mid)" />
      <polygon points="0,90 -45,150 0,170" fill="url(#facet-dark)" />

      <polygon points="30,90 55,80 45,150" fill="url(#facet-light)" />
      <polygon points="-30,90 -55,80 -45,150" fill="url(#facet-mid)" />

      <polygon points="55,80 140,70 120,160 45,150" fill="url(#facet-mid)" />
      <polygon points="-55,80 -140,70 -120,160 -45,150" fill="url(#facet-dark)" />

      <!-- Chin & Neck Armor Plates (MergeTree Data Layer Symbolism) -->
      <polygon points="0,170 45,150 70,220 0,250" fill="url(#facet-light)" />
      <polygon points="0,170 -45,150 -70,220 0,250" fill="url(#facet-mid)" />

      <polygon points="45,150 120,160 150,220 70,220" fill="url(#facet-dark)" />
      <polygon points="-45,150 -120,160 -150,220 -70,220" fill="url(#facet-dark)" />

      <polygon points="120,160 210,140 150,220" fill="url(#facet-mid)" />
      <polygon points="-120,160 -210,140 -150,220" fill="url(#facet-dark)" />

      <!-- Base Neck Facets -->
      <polygon points="0,250 70,220 70,275 0,295" fill="url(#facet-mid)" />
      <polygon points="0,250 -70,220 -70,275 0,295" fill="url(#facet-dark)" />

      <!-- Optical Lens Reticle Line (Subtle Telemetry Horizon) -->
      <line x1="-280" y1="0" x2="-220" y2="0" stroke="url(#cyan-glow)" stroke-width="4" stroke-linecap="round" />
      <line x1="220" y1="0" x2="280" y2="0" stroke="url(#cyan-glow)" stroke-width="4" stroke-linecap="round" />
      <circle cx="0" cy="275" r="5" fill="#00f5ff" />

    </g>
  </g>
</svg>`;
}

async function main() {
  const svg = buildLynxLogoSvg();
  const docsDir = path.resolve(__dirname, '../docs/images');
  const assetsDir = path.resolve(__dirname, '../assets');
  if (!fs.existsSync(docsDir)) fs.mkdirSync(docsDir, { recursive: true });
  if (!fs.existsSync(assetsDir)) fs.mkdirSync(assetsDir, { recursive: true });

  const svgDocPath = path.join(docsDir, 'logo.svg');
  const pngDocPath = path.join(docsDir, 'logo.png');
  const svgAssetPath = path.join(assetsDir, 'logo.svg');

  fs.writeFileSync(svgDocPath, svg, 'utf8');
  fs.writeFileSync(svgAssetPath, svg, 'utf8');
  console.log('Saved vector SVG to:', svgDocPath);

  // Render 1024x1024 PNG with Headless Chrome
  const chromePath = 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
  const browser = await puppeteer.launch({
    executablePath: chromePath,
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--window-size=1024,1024'],
    defaultViewport: { width: 1024, height: 1024, deviceScaleFactor: 1 }
  });

  const page = await browser.newPage();
  const html = `<!DOCTYPE html><html><body style="margin:0;padding:0;background:transparent;overflow:hidden;">${svg}</body></html>`;
  await page.setContent(html, { waitUntil: 'networkidle0' });
  await page.screenshot({ path: pngDocPath, omitBackground: true });
  await browser.close();

  console.log('✓ Successfully generated official mascot logo at:', pngDocPath);
}

main().catch(err => {
  console.error('Error generating logo:', err);
  process.exit(1);
});
