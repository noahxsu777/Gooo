let deferredInstallPrompt;
let speechEnabled = false;
let paused = false;
let speaking = false;
let pendingResume = false;
const ttsQueue = [];
const blockedUsers = new Set();
let maxLength = 300;
let queueCapacity = 120;

const el = (id) => document.getElementById(id);
const statusEl = el('status');
const latencyEl = el('latency');
const queueEl = el('queue');
const lastEl = el('last');
const eventsEl = el('events');
const voiceEl = el('voice');

function addEvent(text) {
  const li = document.createElement('li');
  li.textContent = `[${new Date().toLocaleTimeString()}] ${text}`;
  eventsEl.prepend(li);
  while (eventsEl.children.length > 100) eventsEl.removeChild(eventsEl.lastChild);
}

function updateState(state) {
  statusEl.textContent = state.status || '-';
  latencyEl.textContent = state.latencyMs ?? '-';
  queueEl.textContent = state.queueSize ?? 0;
  lastEl.textContent = state.lastEvent ? `${state.lastEvent.user}: ${state.lastEvent.text}` : '-';
  if (state.maxTextLength) maxLength = state.maxTextLength;
  if (state.queueCapacity) queueCapacity = state.queueCapacity;
}

function shouldSpeak(msg) {
  if (!msg?.text || msg.text.length > maxLength) return false;
  if (String(msg.text).trim().startsWith('!')) return false;
  if (blockedUsers.has((msg.user || '').toLowerCase())) return false;
  return true;
}

function loadVoices() {
  const voices = speechSynthesis.getVoices();
  voiceEl.innerHTML = '';
  voices.forEach((v) => {
    const opt = document.createElement('option');
    opt.value = v.name;
    opt.textContent = `${v.name} (${v.lang})`;
    voiceEl.appendChild(opt);
  });
}

function speakNext() {
  if (!speechEnabled || paused || speaking) return;
  const msg = ttsQueue.shift();
  queueEl.textContent = String(ttsQueue.length);
  if (!msg) return;
  const utt = new SpeechSynthesisUtterance(`${msg.user} dice: ${msg.text}`);
  utt.volume = Number(el('volume').value);
  utt.rate = Number(el('rate').value);
  utt.pitch = Number(el('pitch').value);
  utt.lang = el('lang').value || 'es-ES';
  const selected = Array.from(speechSynthesis.getVoices()).find(v => v.name === voiceEl.value);
  if (selected) utt.voice = selected;
  speaking = true;
  const finish = () => {
    speaking = false;
    if (document.hidden) {
      pendingResume = true;
      return;
    }
    speakNext();
  };
  utt.onend = finish;
  utt.onerror = finish;
  speechSynthesis.speak(utt);
}

function enqueue(msg) {
  if (!shouldSpeak(msg)) return;
  if (ttsQueue.length >= queueCapacity) {
    ttsQueue.shift();
  }
  ttsQueue.push(msg);
  queueEl.textContent = String(ttsQueue.length);
  speakNext();
}

el('enableAudio').onclick = () => {
  speechEnabled = true;
  speechSynthesis.resume();
  addEvent('Audio habilitado por interacción explícita del usuario.');
  speakNext();
};
el('pause').onclick = () => { paused = true; speechSynthesis.pause(); };
el('resume').onclick = () => { paused = false; speechSynthesis.resume(); speakNext(); };
el('skip').onclick = () => { speechSynthesis.cancel(); speaking = false; speakNext(); };
el('flush').onclick = async () => {
  ttsQueue.length = 0;
  queueEl.textContent = '0';
  await fetch('/api/control', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({action:'flush'})});
};

window.addEventListener('beforeinstallprompt', (e) => {
  e.preventDefault();
  deferredInstallPrompt = e;
  el('install').hidden = false;
});
el('install').onclick = async () => {
  if (!deferredInstallPrompt) return;
  deferredInstallPrompt.prompt();
  await deferredInstallPrompt.userChoice;
  deferredInstallPrompt = null;
  el('install').hidden = true;
};

if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('/sw.js').catch((err) => addEvent(`SW error: ${err.message}`));
}
if ('mediaSession' in navigator) {
  navigator.mediaSession.setActionHandler('play', () => { paused = false; speechSynthesis.resume(); speakNext(); });
  navigator.mediaSession.setActionHandler('pause', () => { paused = true; speechSynthesis.pause(); });
}

document.addEventListener('visibilitychange', () => {
  if (!document.hidden && pendingResume) {
    pendingResume = false;
    speechSynthesis.resume();
    speakNext();
  }
});

loadVoices();
speechSynthesis.onvoiceschanged = loadVoices;

const es = new EventSource('/events');
es.addEventListener('chat', (e) => {
  const msg = JSON.parse(e.data);
  addEvent(`${msg.user}: ${msg.text}`);
  enqueue(msg);
});
es.addEventListener('state', (e) => {
  const data = JSON.parse(e.data);
  updateState(data.state || data);
});
es.onerror = () => addEvent('SSE desconectado; el navegador reintentará automáticamente.');

fetch('/api/state').then(r => r.json()).then(updateState).catch(() => {});
