/**
 * Zensor — support.js (v1.0.0)
 * Helpers sin dependencias. Sirven para preview.html y también para la app.
 *
 *   Zensor.theme.get() / .set('light'|'dark'|'system') / .toggle()
 *   Zensor.humidityState(31)       -> { state: 'attention', label: 'Bajo', fill: 'var(--zs-status-attention-fill)', ... }
 *   Zensor.formatNumber(8.6, 1)    -> "8,6"   (es-CL: coma decimal)
 *   Zensor.formatUnit(31, '%')     -> "31 %"  (espacio fino antes de la unidad)
 *   Zensor.gaugeDash(31)           -> { track: '366.5 73.3', value: '113.6 326.2' }
 *   Zensor.gaugeMarker(40)         -> { x, y } posición del nodo de objetivo (viewBox 180x180)
 *   Zensor.copy(text)              -> Promise<boolean>
 */
(function (global) {
  'use strict';

  var KEY = 'zensor-theme';
  var root = document.documentElement;

  /* ---------- Tema ---------- */
  var theme = {
    get: function () {
      try { return localStorage.getItem(KEY) || 'system'; } catch (e) { return 'system'; }
    },
    set: function (mode) {
      if (mode === 'light' || mode === 'dark') root.setAttribute('data-theme', mode);
      else root.removeAttribute('data-theme');
      try {
        if (mode === 'system') localStorage.removeItem(KEY); else localStorage.setItem(KEY, mode);
      } catch (e) { /* almacenamiento no disponible: se ignora */ }
      document.dispatchEvent(new CustomEvent('zensor:theme', { detail: { mode: mode } }));
    },
    effective: function () {
      var m = theme.get();
      if (m !== 'system') return m;
      return global.matchMedia && global.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    },
    toggle: function () {
      var next = theme.effective() === 'dark' ? 'light' : 'dark';
      theme.set(next);
      return next;
    }
  };
  // aplicar tema guardado lo antes posible
  (function () { var m = theme.get(); if (m !== 'system') root.setAttribute('data-theme', m); })();

  /* ---------- Reglas de dominio: humedad de suelo ---------- */
  // Debe coincidir con tokens.json -> domain.humidityThresholds
  function humidityState(v) {
    var n = Number(v);
    var s = n < 30 ? 'attention' : n >= 85 ? 'live' : 'ok';
    var label = { attention: 'Bajo', ok: 'Óptimo', live: 'Saturado' }[s];
    return {
      state: s,
      label: label,
      fill: 'var(--zs-status-' + s + '-fill)',
      bg: 'var(--zs-status-' + s + '-bg)',
      fg: 'var(--zs-status-' + s + '-fg)'
    };
  }

  /* ---------- Formato (es-CL) ---------- */
  function formatNumber(n, digits) {
    var d = digits == null ? 0 : digits;
    return new Intl.NumberFormat('es-CL', { minimumFractionDigits: d, maximumFractionDigits: d }).format(n);
  }
  function formatUnit(n, unit, digits) {
    return formatNumber(n, digits) + '\u00A0' + unit; // espacio duro: "31 %"
  }

  /* ---------- Gauge ensō (arco abierto 300°, abertura abajo) ---------- */
  var GAUGE = { r: 70, arc: 300, rotate: 120 };
  function gaugeDash(value, r) {
    var radius = r || GAUGE.r;
    var c = 2 * Math.PI * radius;
    var arc = c * (GAUGE.arc / 360);
    var val = arc * Math.max(0, Math.min(100, value)) / 100;
    var f = function (x) { return (Math.round(x * 10) / 10).toString(); };
    return { track: f(arc) + ' ' + f(c - arc), value: f(val) + ' ' + f(c - val), rotate: GAUGE.rotate };
  }
  function gaugeMarker(target, cx, cy, r) {
    var radius = r || GAUGE.r, x0 = cx == null ? 90 : cx, y0 = cy == null ? 90 : cy;
    var ang = (GAUGE.rotate + (GAUGE.arc * Math.max(0, Math.min(100, target)) / 100)) * Math.PI / 180;
    return { x: Math.round((x0 + radius * Math.cos(ang)) * 10) / 10, y: Math.round((y0 + radius * Math.sin(ang)) * 10) / 10 };
  }

  /* ---------- Utilidades ---------- */
  function copy(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(text).then(function () { return true; }, function () { return false; });
    }
    return Promise.resolve(false);
  }
  function prefersReducedMotion() {
    return !!(global.matchMedia && global.matchMedia('(prefers-reduced-motion: reduce)').matches);
  }

  /* ---------- Drawer lateral derecho (comportamiento reutilizable) ---------- */
  // Zensor.drawer({ panel, scrim, onClose }).open(); .close()
  function drawer(opts) {
    var panel = opts.panel, scrim = opts.scrim, last = null;
    function onKey(e) { if (e.key === 'Escape') close(); }
    function open() {
      last = document.activeElement;
      panel.hidden = false; if (scrim) scrim.hidden = false;
      panel.classList.remove('zs-anim-drawer-out'); panel.classList.add('zs-anim-drawer-in');
      if (scrim) scrim.classList.add('zs-anim-scrim-in');
      document.addEventListener('keydown', onKey);
      var f = panel.querySelector('[data-autofocus], button, [href], input'); if (f) f.focus();
    }
    function close() {
      document.removeEventListener('keydown', onKey);
      var done = function () { panel.hidden = true; if (scrim) scrim.hidden = true; if (last && last.focus) last.focus(); if (opts.onClose) opts.onClose(); };
      if (prefersReducedMotion()) return done();
      panel.classList.remove('zs-anim-drawer-in'); panel.classList.add('zs-anim-drawer-out');
      setTimeout(done, 140);
    }
    return { open: open, close: close };
  }

  global.Zensor = {
    version: '1.0.0',
    theme: theme,
    humidityState: humidityState,
    formatNumber: formatNumber,
    formatUnit: formatUnit,
    gaugeDash: gaugeDash,
    gaugeMarker: gaugeMarker,
    copy: copy,
    prefersReducedMotion: prefersReducedMotion,
    drawer: drawer
  };
})(window);
