---
layout: landing
---

<style>
/* Reset link color bleed from theme vars */
.lp a,
.lp a:link,
.lp a:visited,
.lp a:hover,
.lp a:active {
  color: inherit;
  text-decoration: none;
}

/* ── LIGHT tokens (default) ── */
.lp {
  --ac:    #2563eb;
  --ac-bg: rgba(37,99,235,0.07);
  --ac-br: rgba(37,99,235,0.18);
  --tx:    #0f172a;
  --tx2:   #64748b;
  --bg:    #ffffff;
  --bg2:   #f8fafc;
  --br:    #e2e8f0;
  --br2:   #94a3b8;
  --mono:  'SF Mono','Fira Code','Roboto Mono',monospace;
}

/*
  DARK: target every way the theme might be applied:
  1. html.dark-theme  (your CSS file)
  2. body.dark        (navbar JS applyTheme)
  3. [data-theme="dark"] (some Hugo book themes)
  4. CSS variable check via inline style on :root (handled by JS below)
*/
html.dark-theme .lp,
body.dark .lp,
[data-theme="dark"] .lp,
.lp.lp--dark {
  --ac:    #60a5fa;
  --ac-bg: rgba(96,165,250,0.08);
  --ac-br: rgba(96,165,250,0.2);
  --tx:    #f1f5f9;
  --tx2:   #94a3b8;
  --bg:    #000000;
  --bg2:   #0d1117;
  --br:    #1e293b;
  --br2:   #475569;
}

/* ── KEYFRAMES ── */
@keyframes fadeUp  { from{opacity:0;transform:translateY(20px)} to{opacity:1;transform:translateY(0)} }
@keyframes fadeIn  { from{opacity:0} to{opacity:1} }
@keyframes shimmer { 0%,100%{opacity:1} 50%{opacity:.6} }

/* ── BASE ── */
*, *::before, *::after { box-sizing: border-box; }

.lp {
  max-width: 880px;
  margin: 0 auto;
  padding: 0 1.5rem;
  font-family: system-ui,-apple-system,sans-serif;
  color: var(--tx);
  background: var(--bg);
}

/* ── HERO ── */
.lp .hero {
  text-align: center;
  padding: 2rem 0 3.5rem;
  animation: fadeUp .6s ease both;
}

.lp .hero-badge {
  display: inline-block;
  padding: .28rem .85rem;
  border-radius: 999px;
  border: 1px solid var(--ac-br);
  background: var(--ac-bg);
  color: var(--ac) !important;
  font-size: .72rem;
  font-weight: 600;
  letter-spacing: .08em;
  text-transform: uppercase;
  margin-bottom: 1.8rem;
  animation: fadeIn .5s .1s ease both;
}

.lp .hero h1 {
  font-size: clamp(2.2rem,5.5vw,3.6rem);
  font-weight: 800;
  line-height: 1.1;
  margin: 0 0 1.2rem;
  letter-spacing: -.03em;
  color: var(--tx);
  animation: fadeUp .6s .15s ease both;
}

.lp .hero p {
  font-size: 1.05rem;
  color: var(--tx2);
  max-width: 480px;
  margin: 0 auto 2.4rem;
  line-height: 1.7;
  animation: fadeUp .6s .25s ease both;
}

.lp .hero-actions {
  display: flex;
  gap: .75rem;
  justify-content: center;
  flex-wrap: wrap;
  animation: fadeUp .6s .35s ease both;
}

/* ── BUTTONS ── !important beats --color-link overrides */
.lp .btn-p {
  display: inline-flex !important;
  align-items: center;
  gap: .4rem;
  padding: .75rem 1.6rem;
  background: var(--ac) !important;
  color: #fff !important;
  border-radius: 8px;
  font-weight: 600;
  font-size: .93rem;
  text-decoration: none !important;
  border: none;
  cursor: pointer;
  transition: filter .2s, transform .15s;
  width: auto;
}
.lp .btn-p:hover { filter: brightness(1.12); transform: translateY(-2px); }

.lp .btn-o {
  display: inline-flex !important;
  align-items: center;
  gap: .4rem;
  padding: .75rem 1.6rem;
  background: transparent !important;
  color: var(--tx) !important;
  border-radius: 8px;
  font-weight: 500;
  font-size: .93rem;
  text-decoration: none !important;
  border: 1px solid var(--br2) !important;
  cursor: pointer;
  transition: border-color .2s, color .2s, transform .15s;
  width: auto;
}
.lp .btn-o:hover { border-color: var(--ac) !important; color: var(--ac) !important; transform: translateY(-2px); }

/* ── STAR NUDGE ── */
.lp .star-nudge {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: .8rem 1.1rem;
  border: 1px solid var(--br);
  border-radius: 8px;
  background: var(--bg2);
  margin: 1.5rem 0 0;
  animation: fadeUp .6s .45s ease both;
  transition: border-color .2s;
}
.lp .star-nudge:hover { border-color: var(--ac-br); }

.lp .star-nudge-icon {
  color: var(--ac);
  flex-shrink: 0;
  animation: shimmer 2.5s ease-in-out infinite;
}

.lp .star-nudge p {
  flex: 1;
  margin: 0;
  font-size: .84rem;
  color: var(--tx2);
  line-height: 1.5;
}
.lp .star-nudge p strong { color: var(--tx); font-weight: 600; }

.lp .star-link {
  display: inline-flex !important;
  align-items: center;
  gap: .35rem;
  padding: .42rem .9rem;
  border: 1px solid var(--br2) !important;
  border-radius: 6px;
  background: var(--bg) !important;
  color: var(--tx) !important;
  font-size: .81rem;
  font-weight: 600;
  text-decoration: none !important;
  flex-shrink: 0;
  transition: border-color .2s, color .2s;
}
.lp .star-link:hover { border-color: var(--ac) !important; color: var(--ac) !important; }

/* ── STATS ── */
.lp .stats {
  display: grid;
  grid-template-columns: repeat(4,1fr);
  border: 1px solid var(--br);
  border-radius: 10px;
  overflow: hidden;
  margin: 2rem 0 3.5rem;
}
.lp .stat-item {
  padding: 1.4rem 1rem;
  text-align: center;
  background: var(--bg);
  border-right: 1px solid var(--br);
  transition: background .2s;
}
.lp .stat-item:last-child { border-right: none; }
.lp .stat-item:hover { background: var(--bg2); }
.lp .stat-num {
  display: block;
  font-size: 1.75rem;
  font-weight: 800;
  letter-spacing: -.03em;
  color: var(--ac);
  margin-bottom: .15rem;
}
.lp .stat-label {
  font-size: .7rem;
  color: var(--tx2);
  text-transform: uppercase;
  letter-spacing: .08em;
  font-weight: 500;
}

/* ── SCROLL REVEAL ── */
.lp .reveal {
  opacity: 0;
  transform: translateY(16px);
  transition: opacity .55s ease, transform .55s ease;
}
.lp .reveal.visible { opacity: 1; transform: translateY(0); }

/* ── SECTION TITLE ── */
.lp .stitle { text-align: center; margin-bottom: 1.8rem; }
.lp .stitle h2 {
  font-size: 1.5rem;
  font-weight: 700;
  margin: 0 0 .35rem;
  letter-spacing: -.02em;
  color: var(--tx);
}
.lp .stitle p { color: var(--tx2); font-size: .92rem; margin: 0; }

/* ── FEATURES ── */
.lp .features {
  display: grid;
  grid-template-columns: repeat(auto-fit,minmax(240px,1fr));
  border: 1px solid var(--br);
  border-radius: 10px;
  overflow: hidden;
  margin-bottom: 3.5rem;
}
.lp .feature {
  padding: 1.5rem;
  background: var(--bg);
  border-right: 1px solid var(--br);
  border-bottom: 1px solid var(--br);
  transition: background .2s;
}
.lp .feature:hover { background: var(--bg2); }
.lp .fi {
  width: 34px; height: 34px;
  border-radius: 7px;
  background: var(--ac-bg);
  display: flex; align-items: center; justify-content: center;
  margin-bottom: .9rem;
  font-size: 1rem;
  transition: transform .2s;
}
.lp .feature:hover .fi { transform: scale(1.1); }
.lp .feature h3 { font-size: .9rem; font-weight: 700; margin: 0 0 .35rem; color: var(--tx); }
.lp .feature p  { font-size: .83rem; color: var(--tx2); margin: 0; line-height: 1.6; }

/* ── PROTOCOLS ── */
.lp .proto-section { margin-bottom: 3.5rem; }
.lp .proto-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit,minmax(190px,1fr));
  gap: .55rem;
}
.lp .proto-card {
  display: flex;
  align-items: center;
  gap: .85rem;
  padding: .85rem 1rem;
  border: 1px solid var(--br);
  border-radius: 7px;
  background: var(--bg);
  transition: border-color .2s, transform .15s;
}
.lp .proto-card:hover { border-color: var(--ac-br); transform: translateY(-1px); }
.lp .ptag {
  font-size: .63rem; font-weight: 700;
  padding: .18rem .42rem;
  border-radius: 3px;
  background: var(--ac-bg);
  color: var(--ac) !important;
  white-space: nowrap; flex-shrink: 0;
  letter-spacing: .04em;
}
.lp .pname { font-size: .86rem; font-weight: 600; margin: 0 0 .1rem; color: var(--tx); }
.lp .pdesc { font-size: .74rem; color: var(--tx2); margin: 0; }

/* ── INSTALL ── */
.lp .install-section { margin-bottom: 3.5rem; }
.lp .install-block {
  border: 1px solid var(--br);
  border-radius: 10px;
  overflow: hidden;
  background: var(--bg);
}
.lp .install-tabs {
  display: flex;
  border-bottom: 1px solid var(--br);
  background: var(--bg2);
  padding: 0 .5rem;
}
.lp .tab-btn {
  padding: .65rem 1rem;
  font-size: .81rem; font-weight: 500;
  color: var(--tx2);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  user-select: none;
  background: none;
  border-top: none; border-left: none; border-right: none;
  transition: color .15s, border-color .15s;
  font-family: inherit;
}
.lp .tab-btn.active { color: var(--ac) !important; border-bottom-color: var(--ac); }
.lp .tab-btn:hover:not(.active) { color: var(--tx); }
.lp .tab-pane { display: none; }
.lp .tab-pane.active { display: block; }
.lp .install-cmd {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.1rem 1.3rem;
}
.lp .install-cmd code {
  flex: 1;
  font-size: .8rem;
  color: var(--tx);
  word-break: break-all;
  background: none; border: none; padding: 0;
  font-family: var(--mono);
}
.lp .copy-btn {
  flex-shrink: 0;
  font-size: .74rem;
  padding: .28rem .7rem;
  border: 1px solid var(--br2);
  border-radius: 4px;
  cursor: pointer;
  background: var(--bg2);
  color: var(--tx2);
  font-family: inherit;
  transition: color .15s, border-color .15s;
}
.lp .copy-btn:hover { color: var(--ac); border-color: var(--ac); }
.lp .install-note {
  padding: .55rem 1.3rem .7rem;
  font-size: .77rem;
  color: var(--tx2);
  border-top: 1px solid var(--br);
  background: var(--bg2);
}
.lp .install-note a { color: var(--ac) !important; text-decoration: none; }
.lp .install-note a:hover { text-decoration: underline; }

/* ── FOOTER CTA ── */
.lp .footer-cta {
  text-align: center;
  padding: 3rem 1rem 4rem;
  border-top: 1px solid var(--br);
}
.lp .footer-cta h2 {
  font-size: 1.5rem; font-weight: 700;
  margin: 0 0 .5rem;
  letter-spacing: -.02em;
  color: var(--tx);
}
.lp .footer-cta p { color: var(--tx2); margin: 0 0 2rem; }
.lp .btn-group { display: flex; gap: .8rem; justify-content: center; flex-wrap: wrap; }

/* ── RESPONSIVE ── */
@media(max-width:580px){
  .lp .stats { grid-template-columns: repeat(2,1fr); }
  .lp .stat-item:nth-child(2){ border-right: none; }
  .lp .star-nudge { flex-direction: column; text-align: center; }
}
</style>

<div class="lp" id="lp-root">

<!-- HERO -->
<div class="hero">
  <div class="hero-badge">v2.6.0 · Now available</div>
  <h1>Scan everything.<br>At full speed.</h1>
  <p>bgscan is a blazing-fast, multi-protocol network scanner built in Go — with a modular chain engine and an interactive terminal UI.</p>
  <div class="hero-actions">
    <a href="/docs/" class="btn-p">Get started →</a>
    <a href="https://github.com/MohsenBg/bgscan" class="btn-o">
      <svg width="15" height="15" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/></svg>
      View on GitHub
    </a>
  </div>
</div>

<!-- STAR NUDGE -->
<div class="star-nudge">
  <svg class="star-nudge-icon" width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>
  <p><strong>If bgscan has been useful to you, I would appreciate your support with a star on GitHub.</strong> This helps increase the project's visibility and boosts my motivation for future updates and improvements.</p>
  <a href="https://github.com/MohsenBg/bgscan" class="star-link">Star on GitHub</a>
</div>

<!-- STATS -->
<div class="stats reveal">
  <div class="stat-item"><span class="stat-num">8</span><span class="stat-label">Protocols</span></div>
  <div class="stat-item"><span class="stat-num">100%</span><span class="stat-label">Go</span></div>
  <div class="stat-item"><span class="stat-num">4</span><span class="stat-label">Platforms</span></div>
  <div class="stat-item"><span class="stat-num">MIT</span><span class="stat-label">License</span></div>
</div>

<!-- FEATURES -->
<div class="stitle reveal">
  <h2>Built for speed and control</h2>
  <p>Everything you need for advanced network reconnaissance in one tool.</p>
</div>
<div class="features reveal">
  <div class="feature"><div class="fi">⚡</div><h3>Concurrent scanning</h3><p>Runs multiple protocols in parallel across thousands of targets without breaking a sweat.</p></div>
  <div class="feature"><div class="fi">🔗</div><h3>Chain engine</h3><p>Compose scan stages using Stream, Sequential, or Batch modes to build full detection pipelines.</p></div>
  <div class="feature"><div class="fi">🖥️</div><h3>Interactive TUI</h3><p>A Bubble Tea terminal UI — scan, monitor, and explore results live, no browser needed.</p></div>
  <div class="feature"><div class="fi">💾</div><h3>Save &amp; replay</h3><p>Persist results and re-run new scans against saved data. Full Xray outbound management included.</p></div>
  <div class="feature"><div class="fi">🌐</div><h3>Advanced DNS</h3><p>DNS Tunnel support with a built-in fallback mechanism for complex resolution scenarios.</p></div>
  <div class="feature"><div class="fi">🛰️</div><h3>Xray integration</h3><p>Save, manage, and validate Xray outbounds directly from the scanner.</p></div>
</div>

<!-- PROTOCOLS -->
<div class="proto-section">
  <div class="stitle reveal">
    <h2>8 protocols. One tool.</h2>
    <p>From Layer 3 to Layer 7 — every protocol you need for deep network analysis.</p>
  </div>
  <div class="proto-grid reveal">
    <div class="proto-card"><span class="ptag">L3</span><div><p class="pname">ICMP</p><p class="pdesc">Host discovery &amp; ping</p></div></div>
    <div class="proto-card"><span class="ptag">L4</span><div><p class="pname">TCP</p><p class="pdesc">Connection &amp; handshake</p></div></div>
    <div class="proto-card"><span class="ptag">L7</span><div><p class="pname">HTTP</p><p class="pdesc">HTTP/1.1 · HTTP/2 · QUIC</p></div></div>
    <div class="proto-card"><span class="ptag">L7</span><div><p class="pname">TLS</p><p class="pdesc">TLS 1.0 through 1.3</p></div></div>
    <div class="proto-card"><span class="ptag">L7</span><div><p class="pname">DNS</p><p class="pdesc">Advanced queries + fallback</p></div></div>
    <div class="proto-card"><span class="ptag">L7</span><div><p class="pname">DNSTT</p><p class="pdesc">DNS Tunnel validation</p></div></div>
    <div class="proto-card"><span class="ptag">L7</span><div><p class="pname">Slipstream</p><p class="pdesc">SOCKS-based validation</p></div></div>
    <div class="proto-card"><span class="ptag">L7</span><div><p class="pname">Xray</p><p class="pdesc">Outbound testing</p></div></div>
  </div>
</div>

<!-- INSTALL -->
<div class="install-section">
  <div class="stitle reveal">
    <h2>One command to install</h2>
    <p>Works on Linux, macOS, Windows, and Android (Termux).</p>
  </div>
  <div class="install-block reveal">
    <div class="install-tabs">
      <button class="tab-btn active" onclick="swTab(event,'linux')">Linux / macOS</button>
      <button class="tab-btn" onclick="swTab(event,'windows')">Windows</button>
      <button class="tab-btn" onclick="swTab(event,'android')">Android</button>
    </div>
    <div id="tab-linux" class="tab-pane active">
      <div class="install-cmd">
        <code>curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh</code>
        <button class="copy-btn" onclick="cpCmd(this)">Copy</button>
      </div>
    </div>
    <div id="tab-windows" class="tab-pane">
      <div class="install-cmd">
        <code>irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.ps1 | iex</code>
        <button class="copy-btn" onclick="cpCmd(this)">Copy</button>
      </div>
    </div>
    <div id="tab-android" class="tab-pane">
      <div class="install-cmd">
        <code>pkg update -y && pkg install bash curl unzip -y && curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | bash</code>
        <button class="copy-btn" onclick="cpCmd(this)">Copy</button>
      </div>
    </div>
    <div class="install-note">Pre-built binaries on the <a href="https://github.com/MohsenBg/bgscan/releases">releases page</a>.</div>
  </div>
</div>

<!-- FOOTER CTA -->
<div class="footer-cta reveal">
  <h2>Ready to start scanning?</h2>
  <p>Read the docs, grab the latest release, or drop a star on GitHub if it helps you.</p>
  <div class="btn-group">
    <a href="/docs/" class="btn-p">Get started →</a>
    <a href="https://github.com/MohsenBg/bgscan" class="btn-o">
      <svg width="15" height="15" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/></svg>
      View on GitHub
    </a>
  </div>
</div>

</div>

<script>
(function(){
  var lp = document.getElementById('lp-root');
  if (!lp) return;

  /* ── sync dark class on .lp to match whatever the site is doing ── */
  function syncTheme() {
    var html = document.documentElement;
    var body = document.body;

    var isDark =
      html.classList.contains('dark-theme') ||   // your CSS file selector
      body.classList.contains('dark') ||          // navbar JS applyTheme
      html.classList.contains('dark') ||
      html.getAttribute('data-theme') === 'dark' ||
      body.getAttribute('data-theme') === 'dark' ||
      // fallback: check if body-background inline style is dark
      (function(){
        var bg = getComputedStyle(document.documentElement)
                   .getPropertyValue('--body-background').trim();
        return bg === '#000000' || bg === '#000' || bg === 'rgb(0,0,0)' || bg === 'rgb(0, 0, 0)';
      })();

    lp.classList.toggle('lp--dark', isDark);
  }

  /* run once on load */
  syncTheme();

  /* watch for class changes on <html> and <body> */
  var mo = new MutationObserver(syncTheme);
  mo.observe(document.documentElement, { attributes: true, attributeFilter: ['class','data-theme','style'] });
  mo.observe(document.body,            { attributes: true, attributeFilter: ['class','data-theme','style'] });

  /* ── tab switcher ── */
  window.swTab = function(e, id){
    document.querySelectorAll('.lp .tab-btn').forEach(function(b){ b.classList.remove('active'); });
    document.querySelectorAll('.lp .tab-pane').forEach(function(p){ p.classList.remove('active'); });
    e.target.classList.add('active');
    document.getElementById('tab-'+id).classList.add('active');
  };

  /* ── copy button ── */
  window.cpCmd = function(btn){
    navigator.clipboard.writeText(btn.previousElementSibling.textContent).then(function(){
      btn.textContent = 'Copied!';
      setTimeout(function(){ btn.textContent = 'Copy'; }, 1800);
    });
  };

  /* ── scroll reveal ── */
  var revEls = document.querySelectorAll('.lp .reveal');
  if (revEls.length && 'IntersectionObserver' in window) {
    var io = new IntersectionObserver(function(entries){
      entries.forEach(function(e){
        if (e.isIntersecting){ e.target.classList.add('visible'); io.unobserve(e.target); }
      });
    }, { threshold: 0.12 });
    revEls.forEach(function(el){ io.observe(el); });
  } else {
    revEls.forEach(function(el){ el.classList.add('visible'); });
  }
})();
</script>
