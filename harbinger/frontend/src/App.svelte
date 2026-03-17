<script>
  import { EventsOn } from '../wailsjs/runtime';
  import { SelectWordlist, StartScan, StopScan } from '../wailsjs/go/main/App';

  // --- SCANNER STATE ---
  let targetUrl = "http://";
  let wordlistPath = "";
  let threads = 50;
  let results = [];
  let isScanning = false;
  let selectedResult = null;
  let previewRequest = "";
  let previewMode = "raw"; // "raw" veya "render" modunu tutar

  // --- EVENT DİNLEYİCİLERİ ---
  EventsOn("found_result", (res) => {
    if (res.Source === "scanner") {
      results = [res, ...results];
    }
  });

  EventsOn("scan_complete", (msg) => {
    isScanning = false;
    alert(msg);
  });

  // --- FONKSİYONLAR ---
  async function pickFile() {
    const path = await SelectWordlist();
    if (path) wordlistPath = path;
  }

  function handleStart() {
    if (!wordlistPath || !targetUrl) return alert("Lütfen hedef URL ve wordlist seçin!");
    results = [];
    selectedResult = null;
    previewRequest = "";
    isScanning = true;
    StartScan(targetUrl, threads, wordlistPath);
  }

  function loadPreview(res) {
    selectedResult = res;
    let host = "";
    try { host = new URL(res.URL).hostname; } catch { host = "localhost"; }
    // İstek önizlemesini oluşturur
    previewRequest = `${res.Method || 'GET'} ${res.URL} HTTP/1.1\nHost: ${host}\nUser-Agent: Harbinger/1.0\nAccept: */*\n\n${res.Body || ''}`;
  }
</script>

<main>
  <div class="top-nav">
    <div class="logo">HARBINGER <span>v1.0</span></div>
    <div class="tabs">
        <span style="color: #58a6ff; font-weight: bold; padding: 10px 20px;">URL SCANNER</span>
    </div>
  </div>

  <div class="module-container">
    <div class="controls">
      <input bind:value={targetUrl} placeholder="Target (http://...)" />
      <button class="btn-sec" on:click={pickFile}>📁 {wordlistPath ? 'Wordlist OK' : 'Wordlist'}</button>
      <input type="number" bind:value={threads} style="width: 60px;" title="Threads" />
      <button class={isScanning ? "btn-stop" : "btn-start"} on:click={isScanning ? StopScan : handleStart}>
        {isScanning ? "STOP" : "START SCAN"}
      </button>
    </div>

    <div class="main-grid">
      <div class="panel list-panel">
        <div class="panel-header">RESULTS ({results.length})</div>
        <div class="scroll-area">
          {#each results as res}
            <div class="row code-{res.StatusCode}" class:selected={selectedResult === res} on:dblclick={() => loadPreview(res)}>
              <span class="status">{res.StatusCode}</span>
              <span class="url">[{res.Method}] {res.URL}</span>
            </div>
          {/each}
        </div>
      </div>

      <div class="panel preview-panel">
        <div class="panel-header">
          <div class="preview-mode-tabs">
            <button class:active={previewMode === 'raw'} on:click={() => previewMode = 'raw'}>RAW</button>
            <button class:active={previewMode === 'render'} on:click={() => previewMode = 'render'}>RENDER</button>
          </div>
          <h3>PREVIEW</h3>
        </div>
        
        <div class="preview-content">
          {#if previewMode === 'raw'}
            <textarea readonly value={previewRequest} placeholder="Double click a result to see details..."></textarea>
          {:else}
            <div class="render-view">
              {#if selectedResult && selectedResult.Body}
                <iframe title="Render" srcdoc={selectedResult.Body} sandbox="allow-scripts"></iframe>
              {:else}
                <div class="empty-msg">Gösterilecek HTML içeriği yok.</div>
              {/if}
            </div>
          {/if}
        </div>
      </div>
    </div>
  </div>
</main>

<style>
  /* Global ve Temel Stiller */
  :global(body) { margin: 0; background: #0d1117; color: #c9d1d9; font-family: 'Consolas', monospace; overflow: hidden; }
  main { height: 100vh; display: flex; flex-direction: column; padding: 10px; box-sizing: border-box; }
  .top-nav { display: flex; justify-content: space-between; border-bottom: 1px solid #30363d; margin-bottom: 10px; }
  .logo { color: #58a6ff; font-weight: bold; font-size: 1.2rem; }
  .module-container { display: flex; flex-direction: column; flex: 1; overflow: hidden; }
  .controls { display: flex; gap: 10px; margin-bottom: 10px; }
  .main-grid { display: grid; grid-template-columns: 1fr 1.6fr; gap: 15px; flex: 1; overflow: hidden; }
  
  /* Panel Tasarımları */
  .panel { background: #161b22; border: 1px solid #30363d; border-radius: 6px; display: flex; flex-direction: column; overflow: hidden; }
  .panel-header { background: #21262d; padding: 8px 15px; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #30363d; }
  .panel-header h3 { margin: 0; font-size: 0.75rem; color: #8b949e; text-transform: uppercase; }

  /* Render Sekmeleri */
  .preview-mode-tabs { display: flex; gap: 4px; }
  .preview-mode-tabs button {
    background: #30363d; color: #c9d1d9; border: none; padding: 4px 10px;
    border-radius: 4px; cursor: pointer; font-size: 0.7rem; font-weight: bold;
  }
  .preview-mode-tabs button.active { background: #58a6ff; color: #0d1117; }

  /* İçerik Alanları */
  .scroll-area { overflow-y: auto; flex: 1; background: #0d1117; }
  .preview-content { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
  .row { display: flex; padding: 8px 15px; border-bottom: 1px solid #21262d; cursor: pointer; font-size: 0.85rem; }
  .row:hover { background: #1c2128; }
  .row.selected { background: #23863622; border-left: 4px solid #238636; }
  .status { width: 45px; font-weight: bold; }
  
  textarea { flex: 1; background: #0d1117; color: #7ee787; border: none; padding: 12px; resize: none; outline: none; font-family: inherit; font-size: 0.9rem; }
  
  /* Render Ekranı */
  .render-view { flex: 1; background: white; border: none; }
  iframe { width: 100%; height: 100%; border: none; }
  .empty-msg { display: flex; align-items: center; justify-content: center; height: 100%; color: #8b949e; font-style: italic; background: #0d1117; }

  /* Butonlar ve Girdiler */
  .btn-start { background: #238636; color: white; border: none; padding: 8px 18px; border-radius: 4px; cursor: pointer; font-weight: bold; }
  .btn-stop { background: #da3633; color: white; border: none; padding: 8px 18px; border-radius: 4px; cursor: pointer; font-weight: bold; }
  .btn-sec { background: #30363d; color: white; border: none; padding: 8px 12px; border-radius: 4px; cursor: pointer; }
  input { background: #0d1117; border: 1px solid #30363d; color: white; padding: 8px; border-radius: 4px; }
  .code-200 { color: #3fb950; }
  .code-404 { color: #f85149; }
</style>