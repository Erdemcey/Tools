<script>
  import { EventsOn } from '../wailsjs/runtime';
  import { SelectWordlist, StartScan, StopScan ,StopIntruder } from '../wailsjs/go/main/App';

  // --- NAVİGASYON ---
  let activeTab = "scanner";

  // --- SCANNER STATE ---
  let targetUrl = "http://";
  let wordlistPath = "";
  let threads = 50;
  let results = [];
  let isScanning = false;
  let selectedResult = null;
  let previewRequest = "";

  // --- REPEATER STATE ---
  let repeaterRequest = "";
  let repeaterResponse = "";
  let isSending = false;

  // --- INTRUDER STATE ---
  let intruderRequest = "";
  let intruderResults = [];
  let isAttacking = false;
  // Manuel payload yerine artık "Paste List" alanı kullanıyoruz
  let intruderPasteList = ""; 
  let intruderSourceType = "wordlist"; // "wordlist" veya "paste"

  // --- EVENT DİNLEYİCİLERİ ---
  EventsOn("found_result", (res) => {
    if (res.Source === "scanner") {
      results = [res, ...results];
    } else if (res.Source === "intruder") {
      intruderResults = [res, ...intruderResults];
    }
  });

  async function handleStopIntruder() {
      try {
          await StopIntruder(); // Go tarafındaki fonksiyon
          isAttacking = false;   // Arayüzü güncelle
      } catch (err) {
          console.error("Durdurma hatası:", err);
      }
  }

  EventsOn("scan_complete", (msg) => {
    isScanning = false;
    isAttacking = false;
    alert(msg);
  });

  // --- GENEL FONKSİYONLAR ---
  async function pickFile() {
    const path = await SelectWordlist();
    if (path) wordlistPath = path;
  }

  // --- SCANNER FONKSİYONLARI ---
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
    previewRequest = `${res.Method || 'GET'} ${res.URL} HTTP/1.1\nHost: ${host}\nUser-Agent: Harbinger/1.0\nAccept: */*\n\n${res.Body || ''}`;
  }

  // --- AKTARIM FONKSİYONLARI ---
  function sendToRepeater() {
    repeaterRequest = previewRequest;
    activeTab = "repeater";
  }

  function sendToIntruder() {
    intruderRequest = previewRequest;
    activeTab = "intruder";
  }

  // --- REPEATER AKSİYONLARI ---
  async function handleRepeaterSend() {
    isSending = true;
    repeaterResponse = "Waiting for response...";
    const result = await window.go.main.App.SendRepeater(repeaterRequest);
    repeaterResponse = result;
    isSending = false;
  }

  // --- INTRUDER AKSİYONLARI ---
  function addIntruderPlaceholder() {
    const area = document.getElementById('intruder-area');
    const start = area.selectionStart;
    const end = area.selectionEnd;
    if (start === end) return;
    intruderRequest = intruderRequest.substring(0, start) + "§" + intruderRequest.substring(start, end) + "§" + intruderRequest.substring(end);
  }

  async function handleIntruderAttack() {
    if (intruderSourceType === "wordlist" && !wordlistPath) return alert("Lütfen bir wordlist dosyası seçin!");
    if (intruderSourceType === "paste" && !intruderPasteList.trim()) return alert("Lütfen bir liste yapıştırın!");
    
    intruderResults = [];
    isAttacking = true;
    
    // Go tarafındaki StartIntruder'a verileri gönderiyoruz
    // Go tarafı manualPayload kısmını artık kopyalanan liste olarak kullanacak
    await window.go.main.App.StartIntruder(
      intruderRequest, 
      intruderSourceType === "wordlist" ? "wordlist" : "manual", 
      intruderPasteList, 
      wordlistPath, 
      threads
    );
  }
</script>

<main>
  <div class="top-nav">
    <div class="logo">HARBINGER <span>v1.0</span></div>
    <div class="tabs">
      <button class:active={activeTab === 'scanner'} on:click={() => activeTab = 'scanner'}>SCANNER</button>
      <button class:active={activeTab === 'repeater'} on:click={() => activeTab = 'repeater'}>REPEATER</button>
      <button class:active={activeTab === 'intruder'} on:click={() => activeTab = 'intruder'}>INTRUDER</button>
    </div>
  </div>

  {#if activeTab === 'scanner'}
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
            <h3>HTTP REQUEST PREVIEW</h3>
            {#if previewRequest}
              <div class="actions">
                <button class="btn-mini" on:click={sendToRepeater}>TO REPEATER</button>
                <button class="btn-mini" on:click={sendToIntruder}>TO INTRUDER</button>
              </div>
            {/if}
          </div>
          <textarea readonly value={previewRequest} placeholder="Double click a result to see details..."></textarea>
        </div>
      </div>
    </div>

  {:else if activeTab === 'repeater'}
    <div class="module-container">
      <div class="main-grid">
        <div class="panel">
          <div class="panel-header">
            <h3>REQUEST</h3>
            <button class="btn-start" on:click={handleRepeaterSend} disabled={isSending}>SEND</button>
          </div>
          <textarea bind:value={repeaterRequest} spellcheck="false"></textarea>
        </div>
        <div class="panel">
          <div class="panel-header"><h3>RESPONSE</h3></div>
          <div class="response-view">{repeaterResponse}</div>
        </div>
      </div>
    </div>

  {:else if activeTab === 'intruder'}
    <div class="module-container">
      <div class="intruder-top-grid">
        <div class="panel">
          <div class="panel-header">
            <h3>1. POSITIONS (Select text and click ADD)</h3>
            <button class="btn-mini" on:click={addIntruderPlaceholder}>ADD §</button>
          </div>
          <textarea id="intruder-area" bind:value={intruderRequest} spellcheck="false" placeholder="Paste request and mark variables with §"></textarea>
        </div>

        <div class="panel">
          <div class="panel-header"><h3>2. PAYLOAD SETTINGS</h3></div>
          <div class="settings-box">
            <div class="source-selector">
              <label>
                <input type="radio" bind:group={intruderSourceType} value="wordlist" /> Wordlist File
              </label>
              <label>
                <input type="radio" bind:group={intruderSourceType} value="paste" /> Paste List
              </label>
            </div>

            {#if intruderSourceType === "wordlist"}
              <div class="file-box">
                <button class="btn-sec" on:click={pickFile}>📁 {wordlistPath ? 'File Loaded' : 'Select Wordlist'}</button>
                <small>{wordlistPath || 'No file selected'}</small>
              </div>
            {:else}
              <div class="paste-box">
                <textarea bind:value={intruderPasteList} placeholder="Paste your payloads here (one per line)..." style="height: 120px; font-size: 0.8rem;"></textarea>
              </div>
            {/if}

            <button class="btn-start" style="margin-top: 10px;" on:click={handleIntruderAttack} disabled={isAttacking}>
              {isAttacking ? "ATTACKING..." : "START ATTACK"}
            </button>

            <button 
                class={isAttacking ? "btn-stop" : "btn-start"} 
                style="margin-top: 10px;" 
                on:click={isAttacking ? handleStopIntruder : handleIntruderAttack}
            >
                {isAttacking ? "STOP ATTACK" : "START ATTACK"}
            </button>
          </div>
        </div>
      </div>
      
      <div class="panel list-panel" style="margin-top: 10px; flex: 0.7;">
        <div class="panel-header">ATTACK RESULTS ({intruderResults.length})</div>
        <div class="scroll-area">
          {#each intruderResults as res}
            <div class="row code-{res.StatusCode}">
              <span class="status">{res.StatusCode}</span>
              <span class="url">{res.URL} — {res.ContentLen} bytes</span>
            </div>
          {/each}
        </div>
      </div>
    </div>
  {/if}
</main>

<style>
  :global(body) { margin: 0; background: #0d1117; color: #c9d1d9; font-family: 'Consolas', monospace; overflow: hidden; }
  main { height: 100vh; display: flex; flex-direction: column; padding: 10px; box-sizing: border-box; }
  
  .top-nav { display: flex; justify-content: space-between; border-bottom: 1px solid #30363d; margin-bottom: 10px; }
  .logo { color: #58a6ff; font-weight: bold; font-size: 1.2rem; }
  .tabs button { background: none; border: none; color: #8b949e; padding: 10px 20px; cursor: pointer; }
  .tabs button.active { color: #58a6ff; border-bottom: 2px solid #58a6ff; }
  
  .module-container { display: flex; flex-direction: column; flex: 1; overflow: hidden; }
  .controls { display: flex; gap: 10px; margin-bottom: 10px; }
  
  .main-grid { display: grid; grid-template-columns: 1fr 1.6fr; gap: 15px; flex: 1; overflow: hidden; }
  .intruder-top-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 15px; flex: 1; overflow: hidden; }
  
  .panel { background: #161b22; border: 1px solid #30363d; border-radius: 6px; display: flex; flex-direction: column; overflow: hidden; }
  .panel-header { background: #21262d; padding: 10px 15px; display: flex; justify-content: space-between; align-items: center; font-size: 0.85rem; border-bottom: 1px solid #30363d; }
  .panel-header h3 { margin: 0; font-size: 0.8rem; color: #8b949e; }

  .scroll-area { overflow-y: auto; flex: 1; background: #0d1117; }
  .row { display: flex; padding: 8px 15px; border-bottom: 1px solid #21262d; cursor: pointer; font-size: 0.85rem; }
  .row:hover { background: #1c2128; }
  .row.selected { background: #23863622; border-left: 4px solid #238636; }
  .status { width: 45px; font-weight: bold; }
  
  textarea { flex: 1; background: #0d1117; color: #7ee787; border: none; padding: 12px; resize: none; outline: none; font-family: inherit; font-size: 0.9rem; }
  .response-view { flex: 1; padding: 15px; white-space: pre-wrap; overflow-y: auto; color: #3fb950; font-size: 0.85rem; background: #0d1117; }
  
  .settings-box { padding: 15px; display: flex; flex-direction: column; gap: 15px; flex: 1; background: #0d1117; }
  .source-selector { display: flex; gap: 20px; border-bottom: 1px solid #30363d; padding-bottom: 10px; }
  .source-selector label { cursor: pointer; font-size: 0.9rem; }
  .file-box { display: flex; flex-direction: column; gap: 5px; }
  .file-box small { color: #8b949e; word-break: break-all; }
  .paste-box { display: flex; flex-grow: 1; }

  .btn-start { background: #238636; color: white; border: none; padding: 8px 18px; border-radius: 4px; cursor: pointer; font-weight: bold; }
  .btn-stop { background: #da3633; color: white; border: none; padding: 8px 18px; border-radius: 4px; cursor: pointer; font-weight: bold; }
  .btn-sec { background: #30363d; color: white; border: none; padding: 8px 12px; border-radius: 4px; cursor: pointer; }
  .btn-mini { background: #58a6ff; color: #0d1117; border: none; padding: 4px 10px; border-radius: 4px; font-size: 0.75rem; font-weight: bold; cursor: pointer; }

  input[type="number"] { width: 60px; background: #0d1117; border: 1px solid #30363d; color: white; padding: 8px; border-radius: 4px; }
  input[placeholder^="Target"] { flex: 1; background: #0d1117; border: 1px solid #30363d; color: white; padding: 8px; border-radius: 4px; }
  .code-200 { color: #3fb950; }
  .code-404 { color: #f85149; }
</style>