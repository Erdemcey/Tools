<script>
  import { EventsOn } from '../wailsjs/runtime';
  import { SelectWordlist, StartScan, StopScan } from '../wailsjs/go/main/App';

  // Genel Navigasyon
  let activeTab = "scanner"; 

  // --- SCANNER STATE ---
  let targetUrl = "http://";
  let wordlistPath = "";
  let threads = 50;
  let results = [];
  let isScanning = false;

  // --- REPEATER STATE ---
  let repeaterRequest = "GET http://testphp.vulnweb.com/ HTTP/1.1\nHost: testphp.vulnweb.com\nUser-Agent: Harbinger/1.0";
  let repeaterResponse = "";
  let isSending = false;

  // --- INTRUDER STATE ---
  let intruderRequest = "";
  let intruderPayloadType = "wordlist"; // "wordlist" veya "manual"
  let manualPayload = "";
  let intruderResults = [];
  let isAttacking = false;
  function copyToRepeater(res) {
    // Burp stili raw request oluşturuyoruz
    repeaterRequest = `${res.Method || 'GET'} ${res.URL} HTTP/1.1\nHost: ${new URL(res.URL).hostname}\nUser-Agent: Harbinger/1.0\nAccept: */*\n\n${res.Body || ''}`;
    activeTab = "repeater";
  }
  // Global Event Dinleyicileri
  EventsOn("found_result", (res) => {
    if (activeTab === "scanner") results = [res, ...results];
    if (activeTab === "intruder") intruderResults = [res, ...intruderResults];
  });

  EventsOn("scan_complete", (msg) => {
    isScanning = false;
    isAttacking = false;
    alert(msg);
  });

  // Dosya Seçimi
  async function pickFile() {
    const path = await SelectWordlist();
    if (path) wordlistPath = path;
  }

  // Scanner Kontrolü
  function handleStart() {
    if (!wordlistPath || !targetUrl) { alert("Eksik alan!"); return; }
    results = [];
    isScanning = true;
    StartScan(targetUrl, threads, wordlistPath);
  }

  // Repeater Kontrolü
  async function handleRepeaterSend() {
    isSending = true;
    repeaterResponse = "Waiting for response...";
    try {
      const result = await window.go.main.App.SendRepeater(repeaterRequest);
      repeaterResponse = result;
    } catch (err) {
      repeaterResponse = "Error: " + err;
    } finally {
      isSending = false;
    }
  }

  // Repeater -> Intruder Aktarımı
  function sendToIntruder() {
    intruderRequest = repeaterRequest;
    activeTab = "intruder";
  }

  // Intruder İşaretleme (Add §)
  function addPlaceholder() {
    const textarea = document.getElementById('intruder-area');
    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const text = intruderRequest;
    if (start === end) return; // Seçili alan yoksa işlem yapma
    intruderRequest = text.substring(0, start) + "§" + text.substring(start, end) + "§" + text.substring(end);
  }

  // Intruder Saldırısını Başlat
  async function handleIntruderAttack() {
    if (intruderPayloadType === "wordlist" && !wordlistPath) {
      alert("Lütfen bir wordlist seçin!");
      return;
    }
    intruderResults = [];
    isAttacking = true;
    // Not: Go tarafında StartIntruder fonksiyonunu bir sonraki adımda yazacağız
    await window.go.main.App.StartIntruder(intruderRequest, intruderPayloadType, manualPayload, wordlistPath, threads);
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
      <div class="header">
        <div class="controls">
          <input bind:value={targetUrl} placeholder="Target (http://...)" class="input-url" />
          <button class="btn-secondary" on:click={pickFile}>📁 Wordlist</button>
          <input type="number" bind:value={threads} class="input-threads" />
          <button class={isScanning ? "btn-stop" : "btn-start"} on:click={isScanning ? StopScan : handleStart}>
            {isScanning ? "STOP" : "START SCAN"}
          </button>
        </div>
      </div>
      <div class="results-container">
        <div class="table-header"><span>Status</span><span>URL</span><span>Length</span></div>
        <div class="table-body">
          {#each results as res}
            <div class="row code-{res.StatusCode}" on:dblclick={() => copyToRepeater(res)} title="Double click to send to Repeater">
                <span class="status">{res.StatusCode}</span>
                <span class="url">[{res.Method || 'GET'}] {res.URL}</span>
                <span class="len">{res.ContentLen} B</span>
            </div>
          {/each}
        </div>
      </div>
    </div>

  {:else if activeTab === 'repeater'}
    <div class="repeater-grid">
      <div class="panel">
        <div class="panel-header">
          <h3>REQUEST</h3>
          <div class="actions">
            <button class="btn-secondary" on:click={sendToIntruder}>➡ TO INTRUDER</button>
            <button class="btn-start" on:click={handleRepeaterSend} disabled={isSending}>SEND</button>
          </div>
        </div>
        <textarea bind:value={repeaterRequest} spellcheck="false"></textarea>
      </div>
      <div class="panel">
        <div class="panel-header"><h3>RESPONSE</h3></div>
        <div class="response-view">{repeaterResponse}</div>
      </div>
    </div>

  {:else if activeTab === 'intruder'}
    <div class="intruder-layout">
      <div class="intruder-top">
        <div class="panel">
          <div class="panel-header">
            <h3>POSITIONS</h3>
            <button class="btn-secondary" on:click={addPlaceholder}>ADD §</button>
          </div>
          <textarea id="intruder-area" bind:value={intruderRequest} spellcheck="false"></textarea>
        </div>
        <div class="panel">
          <div class="panel-header"><h3>ATTACK SETTINGS</h3></div>
          <div class="settings-box">
            <select bind:value={intruderPayloadType}>
              <option value="wordlist">Wordlist Attack</option>
              <option value="manual">Manual Payload</option>
            </select>
            {#if intruderPayloadType === "wordlist"}
              <button class="btn-secondary" on:click={pickFile}>{wordlistPath ? "Wordlist Loaded" : "Select List"}</button>
            {:else}
              <input bind:value={manualPayload} placeholder="Enter payload..." />
            {/if}
            <button class="btn-start" on:click={handleIntruderAttack} disabled={isAttacking}>
              {isAttacking ? "ATTACKING..." : "START ATTACK"}
            </button>
          </div>
        </div>
      </div>
      <div class="results-container">
        <div class="table-header"><span>Status</span><span>Payload</span><span>Length</span></div>
        <div class="table-body">
          {#each intruderResults as res}
            <div class="row code-{res.StatusCode}"><span>{res.StatusCode}</span><span>{res.URL}</span><span>{res.ContentLen} B</span></div>
          {/each}
        </div>
      </div>
    </div>
  {/if}
</main>

<style>
  :global(body) { margin: 0; background: #0d1117; color: #c9d1d9; font-family: monospace; }
  main { padding: 10px; height: 100vh; display: flex; flex-direction: column; box-sizing: border-box; }
  .top-nav { display: flex; justify-content: space-between; border-bottom: 1px solid #30363d; padding-bottom: 5px; }
  .logo { color: #58a6ff; font-weight: bold; }
  .tabs button { background: none; border: none; color: #8b949e; padding: 5px 15px; cursor: pointer; }
  .tabs button.active { color: #58a6ff; border-bottom: 2px solid #58a6ff; }
  .controls, .actions, .settings-box { display: flex; gap: 10px; align-items: center; padding: 10px; }
  .module-container, .intruder-layout { display: flex; flex-direction: column; flex-grow: 1; overflow: hidden; }
  .repeater-grid, .intruder-top { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; height: 50%; margin-top: 10px; }
  .panel { display: flex; flex-direction: column; background: #161b22; border: 1px solid #30363d; border-radius: 4px; }
  .panel-header { display: flex; justify-content: space-between; align-items: center; padding: 5px 10px; background: #21262d; }
  .panel-header h3 { font-size: 0.7rem; margin: 0; }
  textarea { flex-grow: 1; background: transparent; color: #3fb950; border: none; padding: 10px; resize: none; outline: none; }
  .response-view { flex-grow: 1; padding: 10px; white-space: pre-wrap; overflow-y: auto; font-size: 0.8rem; }
  .results-container { flex-grow: 1; background: #161b22; border: 1px solid #30363d; margin-top: 10px; overflow: hidden; display: flex; flex-direction: column; }
  .table-header { display: flex; background: #21262d; padding: 5px; font-weight: bold; font-size: 0.8rem; }
  .table-body { overflow-y: auto; flex: 1; }
  .row { display: flex; padding: 4px 10px; border-bottom: 1px solid #21262d; font-size: 0.8rem; }
  .table-header span, .row span { flex: 1; overflow: hidden; }
  .btn-start { background: #238636; color: white; padding: 5px 10px; border-radius: 4px; }
  .btn-stop { background: #da3633; color: white; padding: 5px 10px; border-radius: 4px; }
  .btn-secondary { background: #30363d; color: white; padding: 5px 10px; border-radius: 4px; }
  input, select { background: #0d1117; border: 1px solid #30363d; color: white; padding: 5px; }
  .code-200 { color: #3fb950; }
  .code-404 { color: #f85149; }
</style>