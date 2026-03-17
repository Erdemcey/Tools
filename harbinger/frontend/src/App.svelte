<script>
  import { EventsOn } from '../wailsjs/runtime';
  import { SelectWordlist, StartScan, StopScan } from '../wailsjs/go/main/App';

  // --- STATE ---
  let targetUrl = "http://";
  let wordlistPath = "";
  let threads = 50;
  let results = [];
  let isScanning = false;
  let selectedResult = null;

  let previewContent = "";
  let renderContent = "";
  let previewMode = "raw";

  // 🔥 batch buffer
  let batchResults = [];

  const MAX_RESULTS = 1000;

  // --- EVENTS ---
  EventsOn("found_result", (res) => {
    if (res.Source === "scanner") {
      batchResults.push(res);
    }
  });

  EventsOn("scan_complete", (msg) => {
    isScanning = false;

    // 🔥 kalanları bas
    if (batchResults.length > 0) {
      results.unshift(...batchResults);
      batchResults = [];
    }

    alert(msg);
  });

  // 🔥 HER 100ms UI GÜNCELLE
  setInterval(() => {
    if (batchResults.length > 0) {
      results.unshift(...batchResults);
      batchResults = [];

      // 🔥 performans limiti
      if (results.length > MAX_RESULTS) {
        results = results.slice(0, MAX_RESULTS);
      }

      // 🔥 reactivity tetikle
      results = results;
    }
  }, 100);

  // --- ACTIONS ---
  async function pickFile() {
    const path = await SelectWordlist();
    if (path) wordlistPath = path;
  }

  function handleStart() {
    if (!wordlistPath || !targetUrl) {
      return alert("Lütfen hedef URL ve wordlist seçin!");
    }

    results = [];
    selectedResult = null;
    previewContent = "";
    renderContent = "";
    batchResults = [];

    isScanning = true;
    StartScan(targetUrl, threads, wordlistPath);
  }

  function loadPreview(res) {
    selectedResult = res;

    previewContent = res.Raw || "RAW veri yok";
    renderContent = res.Body || "";
  }
</script>

<main>
  <div class="top-nav">
    <div class="logo">HARBINGER <span>v1.0</span></div>
    <div class="tabs">
      <span class="active-tab">URL SCANNER</span>
    </div>
  </div>

  <div class="module-container">
    <div class="controls">
      <input bind:value={targetUrl} placeholder="Target (http://...)" />

```
  <button class="btn-sec" on:click={pickFile}>
    📁 {wordlistPath ? 'Wordlist OK' : 'Wordlist'}
  </button>

  <input type="number" bind:value={threads} style="width: 60px;" />

  <button
    class={isScanning ? "btn-stop" : "btn-start"}
    on:click={isScanning ? StopScan : handleStart}
  >
    {isScanning ? "STOP" : "START SCAN"}
  </button>
</div>

<div class="main-grid">
  <!-- RESULTS -->
  <div class="panel list-panel">
    <div class="panel-header">
      RESULTS ({results.length})
    </div>

    <div class="scroll-area">
      {#each results as res}
        <div
          class="row code-{res.StatusCode}"
          class:selected={selectedResult === res}
          on:dblclick={() => loadPreview(res)}
        >
          <span class="status">{res.StatusCode}</span>
          <span class="url">[{res.Method}] {res.URL}</span>
        </div>
      {/each}
    </div>
  </div>

  <!-- PREVIEW -->
  <div class="panel preview-panel">
    <div class="panel-header">
      <div class="preview-mode-tabs">
        <button
          class:active={previewMode === 'raw'}
          on:click={() => previewMode = 'raw'}
        >
          RAW
        </button>

        <button
          class:active={previewMode === 'render'}
          on:click={() => previewMode = 'render'}
        >
          RENDER
        </button>
      </div>

      <h3>PREVIEW</h3>
    </div>

    <div class="preview-content">
      {#if previewMode === 'raw'}
        <textarea
          readonly
          bind:value={previewContent}
          placeholder="Double click a result..."
        ></textarea>

      {:else}
        <div class="render-view">
          {#if renderContent}
            <iframe
              title="Render"
              srcdoc={renderContent}
              sandbox="allow-scripts allow-same-origin"
            ></iframe>
          {:else}
            <div class="empty-msg">
              Render edilecek HTML yok.
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</div>
```

  </div>
</main>

<style>
  :global(body) {
    margin: 0;
    background: #0d1117;
    color: #c9d1d9;
    font-family: 'Consolas', monospace;
    overflow: hidden;
  }

  main {
    height: 100vh;
    display: flex;
    flex-direction: column;
    padding: 10px;
  }

  .top-nav {
    display: flex;
    justify-content: space-between;
    border-bottom: 1px solid #30363d;
    margin-bottom: 10px;
  }

  .logo {
    color: #58a6ff;
    font-weight: bold;
  }

  .active-tab {
    color: #58a6ff;
    font-weight: bold;
  }

  .controls {
    display: flex;
    gap: 10px;
    margin-bottom: 10px;
  }

  .main-grid {
    display: grid;
    grid-template-columns: 1fr 1.6fr;
    gap: 15px;
    flex: 1;
  }

  .panel {
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 6px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .panel-header {
    background: #21262d;
    padding: 8px 15px;
    display: flex;
    justify-content: space-between;
  }

  .scroll-area {
    overflow-y: auto;
    flex: 1;
  }

  .row {
    display: flex;
    padding: 8px 15px;
    border-bottom: 1px solid #21262d;
    cursor: pointer;
  }

  .row:hover {
    background: #1c2128;
  }

  .row.selected {
    background: #23863622;
    border-left: 4px solid #238636;
  }

  .status {
    width: 45px;
    font-weight: bold;
  }

  textarea {
    flex: 1;
    background: #0d1117;
    color: #7ee787;
    border: none;
    padding: 12px;
    resize: none;
  }

  .render-view {
    flex: 1;
    background: white;
  }

  iframe {
    width: 100%;
    height: 100%;
    border: none;
  }

  .empty-msg {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #8b949e;
  }

  .btn-start {
    background: #238636;
    color: white;
    border: none;
    padding: 8px 18px;
  }

  .btn-stop {
    background: #da3633;
    color: white;
    border: none;
    padding: 8px 18px;
  }

  .btn-sec {
    background: #30363d;
    color: white;
    border: none;
    padding: 8px 12px;
  }

  input {
    background: #0d1117;
    border: 1px solid #30363d;
    color: white;
    padding: 8px;
  }
  .main-grid {
  display: grid;
  grid-template-columns: 1fr 1.6fr;
  gap: 15px;
  flex: 1;
  height: 100%; /* 🔥 EKLE */
  overflow: hidden;
  }
  .code-200 { color: #3fb950; }
  .code-404 { color: #f85149; }
</style>