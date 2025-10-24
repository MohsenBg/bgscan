---
layout: landing
---

<style>
  html:has(#lp-root),
  body:has(#lp-root) {
    --bg: #faf9f6;
    --ink: #24231f;
    --muted: #706d65;
    --line: #e3e0d8;
    --accent: #345b91;
    --code: #f0eee8;

    background: var(--bg) !important;
    color: var(--ink);
    color-scheme: light;
    -webkit-text-size-adjust: 100%;
  }

  html.lp-site-dark:has(#lp-root),
  body.lp-site-dark:has(#lp-root) {
    --bg: #08090b;
    --ink: #e7e7e4;
    --muted: #92938f;
    --line: #25262a;
    --accent: #91b8f5;
    --code: #111214;

    background: var(--bg) !important;
    color: var(--ink);
    color-scheme: dark;
  }

  body:has(#lp-root) main,
  body:has(#lp-root) #content,
  body:has(#lp-root) .book-page,
  body:has(#lp-root) .book-section,
  body:has(#lp-root) .content {
    max-width: none !important;
    background: transparent !important;
  }

  body:has(#lp-root) {
    overflow-x: hidden;
  }

  #lp-root {
    --mono: ui-monospace, "SFMono-Regular", Consolas, monospace;
    --pad: clamp(1rem, 4vw, 3.5rem);

    width: 100%;
    max-width: 1400px;
    margin: 0 auto;
    padding: 0 max(var(--pad), env(safe-area-inset-right)) 3rem max(var(--pad), env(safe-area-inset-left));
    font: 1rem/1.7 system-ui, -apple-system, "Segoe UI", sans-serif;
    color: var(--ink);
    overflow-wrap: break-word;
  }

  #lp-root,
  #lp-root * {
    box-sizing: border-box;
  }

  #lp-root img,
  #lp-root video {
    max-width: 100%;
  }

  #lp-root a {
    color: var(--accent);
    text-decoration: underline;
    text-underline-offset: 3px;
  }

  #lp-root a:hover {
    text-decoration-thickness: 2px;
  }

  #lp-root a:focus-visible,
  #lp-root button:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 3px;
  }

  #lp-root p {
    margin: 0 0 1rem;
  }

  #lp-root .lp-muted {
    color: var(--muted);
  }

  #lp-root .lp-hero {
    max-width: 880px;
  }

  #lp-root h1 {
    margin: 0 0 0.5rem;
    font-size: clamp(2.4rem, 6.5vw, 4.4rem);
    font-weight: 750;
    line-height: 1;
    letter-spacing: -0.05em;
  }

  #lp-root .lp-tagline {
    margin: 0 0 1.1rem;
    color: var(--accent);
    font-size: clamp(1.1rem, 2vw, 1.4rem);
    font-weight: 600;
    line-height: 1.35;
    letter-spacing: -0.02em;
  }

  #lp-root .lp-lead {
    max-width: 42rem;
    margin: 0;
    color: var(--muted);
    font-size: clamp(1rem, 1.5vw, 1.1rem);
    line-height: 1.75;
  }

  #lp-root .lp-actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.7rem 1.4rem;
    margin-top: 1.6rem;
  }

  #lp-root .lp-button,
  #lp-root .lp-button-secondary {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 7.5rem;
    height: 2.5rem;
    padding: 0 1.25rem;
    border-radius: 6px;
    font-size: 0.92rem;
    font-weight: 600;
    text-decoration: none;
    white-space: nowrap;
    transition: opacity 0.15s, transform 0.15s, border-color 0.15s, color 0.15s, background-color 0.15s;
  }

  #lp-root .lp-button {
    border: 1px solid var(--accent);
    background: var(--accent);
    color: var(--bg) !important;
  }

  #lp-root .lp-button:hover {
    opacity: 0.9;
    transform: translateY(-1px);
  }

  #lp-root .lp-button-secondary {
    border: 1px solid var(--line);
    background: transparent;
    color: var(--ink) !important;
  }

  #lp-root .lp-button-secondary:hover {
    border-color: var(--accent);
    color: var(--accent) !important;
    background: color-mix(in srgb, var(--accent) 6%, transparent);
    transform: translateY(-1px);
  }

  #lp-root .lp-button:active,
  #lp-root .lp-button-secondary:active {
    transform: translateY(0);
  }

  #lp-root .lp-shot {
    margin: clamp(2rem, 4vw, 3.2rem) 0 0;
    margin-inline: 0;
  }

  #lp-root .lp-window {
    overflow: hidden;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: #08090b;
  }

  #lp-root .lp-window video {
    display: block;
    width: 100%;
    height: auto;
    aspect-ratio: 1280 / 768;
    object-fit: contain;
  }

  #lp-root .lp-shot figcaption {
    margin-top: 0.65rem;
    color: var(--muted);
    font-size: 0.84rem;
  }

  #lp-root section {
    display: grid;
    grid-template-columns: minmax(160px, 1fr) minmax(0, 3fr);
    gap: 1rem clamp(2rem, 6vw, 5rem);
    margin-top: clamp(3rem, 6vw, 5rem);
    padding-top: 1.6rem;
    border-top: 1px solid var(--line);
    scroll-margin-top: 1.5rem;
  }

  #lp-root section > * {
    min-width: 0;
  }

  #lp-root h2 {
    margin: 0;
    font-size: clamp(1.15rem, 1.8vw, 1.35rem);
    line-height: 1.4;
    font-weight: 650;
    letter-spacing: -0.02em;
  }

  #lp-root .lp-probe-groups {
    display: grid;
    gap: 2rem;
  }

  #lp-root h3 {
    margin: 0 0 0.5rem;
    font-size: 0.8rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--muted);
  }

  #lp-root dl.lp-probes {
    display: grid;
    grid-template-columns: minmax(0, auto) minmax(0, 1fr);
    gap: 0.5rem 1.5rem;
    margin: 0;
  }

  #lp-root dl.lp-probes dt {
    font-family: var(--mono);
    font-size: 0.88rem;
    font-weight: 700;
  }

  #lp-root dl.lp-probes dt small {
    margin-left: 0.4rem;
    color: var(--muted);
    font-weight: 400;
    font-size: 0.72rem;
  }

  #lp-root dl.lp-probes dd {
    margin: 0;
    color: var(--muted);
    font-size: 0.94rem;
  }

  #lp-root .lp-cmd {
    max-width: 820px;
    margin: 0 0 1.2rem;
  }

  #lp-root .lp-cmd:last-child {
    margin-bottom: 0;
  }

  #lp-root .lp-cmd-label {
    margin: 0 0 0.4rem;
    font-size: 0.9rem;
    font-weight: 600;
  }

  #lp-root .lp-cmd-box {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.85rem 0.95rem;
    border: 1px solid var(--line);
    border-radius: 5px;
    background: var(--code);  
}

  #lp-root .lp-cmd-box code {
    flex: 1;
    min-width: 0;
    color: var(--ink);
    font-family: var(--mono);
    font-size: 0.8rem;
    line-height: 1.7;
    overflow-wrap: anywhere;
    padding: 5px 10px;
  }

  #lp-root .lp-copy {
    flex: 0 0 auto;
    padding: 0.2rem 0.65rem;
    border: 1px solid var(--line);
    border-radius: 4px;
    background: var(--bg);
    color: var(--muted);
    font: inherit;
    font-size: 0.78rem;
    cursor: pointer;
  }

  #lp-root .lp-copy:hover {
    color: var(--ink);
  }

  #lp-root .lp-footer {
    margin-top: clamp(4rem, 8vw, 7rem);
    padding-top: 1.6rem;
    border-top: 1px solid var(--line);
    display: flex;
    flex-wrap: wrap;
    gap: 0.6rem 1.5rem;
    align-items: center;
    justify-content: space-between;
    color: var(--muted);
    font-size: 0.88rem;
  }

  #lp-root .lp-footer a {
    color: var(--muted);
  }

  #lp-root .lp-footer a:hover {
    color: var(--accent);
  }

  #lp-root .lp-foot-links {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem 1.4rem;
  }

  @media (max-width: 820px) {
    #lp-root section {
      grid-template-columns: 1fr;
      gap: 0.6rem;
      padding-top: 1.3rem;
    }

    #lp-root dl.lp-probes {
      grid-template-columns: 1fr;
      gap: 0.3rem 0;
    }

    #lp-root dl.lp-probes dt {
      margin-top: 0.6rem;
    }
  }

  @media (max-width: 560px) {
    #lp-root .lp-cmd-box {
      flex-direction: column;
      gap: 0.6rem;
    }

    #lp-root .lp-copy {
      align-self: flex-start;
    }

    #lp-root .lp-actions {
      flex-direction: column;
      align-items: stretch;
    }

    #lp-root .lp-button,
    #lp-root .lp-button-secondary {
      width: 100%;
    }

    #lp-root .lp-footer {
      flex-direction: column;
      align-items: flex-start;
    }
  }

  @media (pointer: coarse) {
    #lp-root .lp-copy,
    #lp-root .lp-button,
    #lp-root .lp-button-secondary {
      min-height: 2.75rem;
    }
  }
</style>


<div id="lp-root">
  <main>
    <header class="lp-hero">
      <h1 id="lp-title">BGScan</h1>
      <p class="lp-tagline">A scanner with chainable probes.</p>
      <p class="lp-lead">BGScan is a multi-protocol scanner written in Go. It supports TCP, ICMP, HTTP, DNS, DNS tunneling protocols, and Xray outbounds. You can run probes independently or chain them together, passing the results of one stage to the next. Scan results can be saved as CSV and reused as input for later scans. Worker counts, timeouts, and probe-specific settings can be configured to fit different scanning tasks.</p>
      <div class="lp-actions">
        <a class="lp-button lp-docs-link" href="/docs/">Read the docs</a>
        <a class="lp-button-secondary" href="https://github.com/MohsenBg/bgscan">GitHub</a>
      </div>
    </header>
    <figure class="lp-shot">
      <div class="lp-window">
        <video autoplay loop muted playsinline preload="auto" width="1280" height="768" aria-label="BGScan running a scan in the terminal">
          <source src="./bgscan-demo.mp4" type="video/mp4">
        </video>
      </div>
      <figcaption>Running a chained scan from the terminal.</figcaption>
    </figure>
    <section aria-labelledby="lp-probes-title">
      <h2 id="lp-probes-title">Probes</h2>
      <div class="lp-body">
        <p class="lp-muted">DNS tunnel probes are listed separately. They test specific tunnel protocols, not ordinary lookups. Workers, timeouts, and probe settings are configurable per stage via TOML.</p>
        <div class="lp-probe-groups">
          <div>
            <h3>Reachability</h3>
            <dl class="lp-probes">
              <dt>ICMP <small>L3</small></dt>
              <dd>Host reachability and ping checks.</dd>
              <dt>TCP <small>L4</small></dt>
              <dd>Connection attempts and TCP handshake validation.</dd>
            </dl>
          </div>
          <div>
            <h3>Web and DNS</h3>
            <dl class="lp-probes">
              <dt>HTTP <small>L7</small></dt>
              <dd>HTTP/1.1, HTTP/2, and HTTP/3 over QUIC.</dd>
              <dt>DNS <small>L7</small></dt>
              <dd>Resolver queries with fallback and anti-hijacking checks.</dd>
            </dl>
          </div>
          <div>
            <h3>DNS tunnels</h3>
            <dl class="lp-probes">
              <dt>DNSTT <small>L7</small></dt>
              <dd>Tunnel validation with SOCKS and SSH base connections.</dd>
              <dt>VayDNS <small>L7</small></dt>
              <dd>Tunnel validation with SOCKS and SSH base connections.</dd>
              <dt>Slipstream <small>L7</small></dt>
              <dd>DNS tunnel checks through the Slipstream client.</dd>
              <dt>MasterDNS <small>L7</small></dt>
              <dd>Tunnel checks with configurable encryption and MTU discovery.</dd>
              <dt>StormDNS <small>L7</small></dt>
              <dd>Tunnel checks with configurable encryption and MTU discovery.</dd>
              <dt>TheFeed <small>L7</small></dt>
              <dd>Encrypted feed tunnel checks over supported transports.</dd>
            </dl>
          </div>
          <div>
            <h3>Outbounds</h3>
            <dl class="lp-probes">
              <dt>Xray <small>L7</small></dt>
              <dd>Validate Xray outbounds and run connectivity or speed tests.</dd>
            </dl>
          </div>
        </div>
      </div>
    </section>
    <section id="lp-install" aria-labelledby="lp-install-title">
      <h2 id="lp-install-title">Install</h2>
      <div class="lp-body">
        <p class="lp-muted">The install script picks the right release for your system. You can also grab an archive from the <a href="https://github.com/MohsenBg/bgscan/releases">releases page</a>.</p>
        <div class="lp-cmd" id="lp-install-unix">
          <p class="lp-cmd-label">Linux and macOS</p>
          <div class="lp-cmd-box">
            <code>curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh</code>
            <button class="lp-copy" type="button">Copy</button>
          </div>
        </div>
        <div class="lp-cmd" id="lp-install-windows">
          <p class="lp-cmd-label">Windows (PowerShell)</p>
          <div class="lp-cmd-box">
            <code>irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.ps1 | iex</code>
            <button class="lp-copy" type="button">Copy</button>
          </div>
        </div>
        <div class="lp-cmd" id="lp-install-android">
          <p class="lp-cmd-label">Android (Termux)</p>
          <div class="lp-cmd-box">
            <code>{ command -v curl >/dev/null 2>&amp;1 || pkg install -y curl; } &amp;&amp; curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh</code>
            <button class="lp-copy" type="button">Copy</button>
          </div>
        </div>
      </div>
    </section>
    <footer class="lp-footer">
      <span>BGScan · Written in Go · GPL-3.0 · v2.12.0</span>
      <nav class="lp-foot-links" aria-label="Links">
        <a class="lp-docs-link" href="/docs/">Docs</a>
        <a href="https://github.com/MohsenBg/bgscan">GitHub</a>
        <a href="https://github.com/MohsenBg/bgscan/releases">Releases</a>
        <a href="https://github.com/MohsenBg/bgscan/issues">Issues</a>
      </nav>
    </footer>
  </main>
</div>

<script>
  (function () {
    var root = document.getElementById('lp-root');
    if (!root) return;
   
   var html = document.documentElement;
    var body = document.body;

    function isDark() {
      var t = (
        html.getAttribute('data-theme') ||
        body.getAttribute('data-theme') ||
        ''
      ).toLowerCase();

      return t === 'dark' ||
        html.classList.contains('dark') ||
        html.classList.contains('dark-theme') ||
        html.classList.contains('dark-mode') ||
        body.classList.contains('dark') ||
        body.classList.contains('dark-mode');
    }

    function sync() {
      var d = isDark();
      html.classList.toggle('lp-site-dark', d);
      body.classList.toggle('lp-site-dark', d);
    }

    sync();

    new MutationObserver(sync).observe(html, {
      attributes: true,
      attributeFilter: ['class', 'data-theme', 'style']
    });

    new MutationObserver(sync).observe(body, {
      attributes: true,
      attributeFilter: ['class', 'data-theme', 'style']
    });

    var isFa = /\/fa(\/|$)/.test(location.pathname);
    var base = location.pathname
      .replace(/\/fa(\/.*)?$/, '')
      .replace(/\/$/, '');

    root.querySelectorAll('.lp-docs-link').forEach(function (a) {
      a.setAttribute('href', base + (isFa ? '/fa' : '') + '/docs/');
    });

    function fallback(text) {
      var ta = document.createElement('textarea');
      ta.value = text;
      ta.setAttribute('readonly', '');
      ta.style.cssText = 'position:fixed;opacity:0';
      body.appendChild(ta);
      ta.select();

      var ok = false;

      try {
        ok = document.execCommand('copy');
      } catch (e) {}

      body.removeChild(ta);
      return ok;
    }

    root.querySelectorAll('.lp-copy').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var code = btn.parentNode.querySelector('code');
        if (!code) return;

        var text = code.textContent.replace(/\s+/g, ' ').trim();

        function done(ok) {
          btn.textContent = ok ? 'Copied' : 'Select it';
          setTimeout(function () {
            btn.textContent = 'Copy';
          }, 1600);
        }

        if (navigator.clipboard && window.isSecureContext) {
          navigator.clipboard.writeText(text).then(
            function () {
              done(true);
            },
            function () {
              done(fallback(text));
            }
          );
        } else {
          done(fallback(text));
        }
      });
    });
  })();
</script>
