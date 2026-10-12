# Zensor — Style guide (v1.0.0)

The Zensor design system guide, written for people and for LLMs that generate UI.
If you are an LLM: **read all of section 1 before writing any interface code** and check your result against the checklist in section 13.

UI copy stays in Chilean Spanish. Quoted interface strings below (button labels, states, messages) are the real product text: use them as written.

---

## 0. What Zensor is

An IoT, permaculture and automation platform for the countryside. The first module is **Riego** (irrigation: soil moisture sensors + weather + flow meters so you water only what is needed). Security, Home and Fault detection come next.

**Brand personality:** calm, organic, trustworthy. It is a tool for someone who already has too many things to maintain: the interface should remove work, not add it.

**Logo idea:** an open ensō (zen circle), drawn with a brush stroke of variable width. The circle does not close, and the node in the gap is the sensor. Zensor = *sensor* + *zen*.

## Files

| File | Purpose |
|---|---|
| `tokens.json` | **Source of truth.** Palette, light/dark semantics, typography, spacing, radii, sizes, domain rules. |
| `tokens.css` | `--zs-*` CSS variables derived from `tokens.json`, light/dark theme, type roles (`.zs-text-*`), focus and animations. |
| `elevation.json` | Shadows and z-index. |
| `motion.json` | Durations, easing curves and motion patterns. |
| `tailwind.preset.js` | Tailwind preset that **replaces** the default scale with these tokens. |
| `support.js` | `window.Zensor`: theme, humidity state, es-CL formatting, ensō gauge, drawer. |
| `preview.html` | Live catalog with examples. Open it in the browser. |
| `logos/` | Logo SVGs and PNGs, app icons and favicons. |

---

## 1. Hard rules (read these first)

**ALWAYS**
1. Use **semantic tokens** for color: `--zs-bg-surface`, `--zs-text-primary`, `--zs-status-ok-fill`… (in Tailwind: `bg-surface`, `text-ink`, `bg-ok-fill`). That way they work in both light and dark themes.
2. Use **only values from the scale**: spacing (`--zs-space-*`), radii (`--zs-radius-*`), text sizes (`--zs-text-*`), durations (`--zs-duration-*`).
3. Every interactive element is **at least 44 × 44 px** (`--zs-size-touch-min`). Large action buttons: 52 px.
4. Sensor data in **JetBrains Mono**. Headings and brand in **Zen Maru Gothic**. Everything else in **DM Sans**.
5. Cards are separated with a **thin border** (`--zs-border-default`), not a shadow.
6. Device detail is shown in a **right-side panel (drawer)**, not a new page or a centered modal.
7. Every state is communicated with **text + color** (never color alone).
8. Copy in **neutral Chilean Spanish**, informal "tú", short sentences and concrete verbs. es-CL number format (`8,6`, `31 %`).
9. Respect `prefers-reduced-motion` and keep focus visible (`tokens.css` already does this).

**NEVER**
1. Don't write loose hex values or default Tailwind colors (`blue-500`, `gray-100`, `red-600`). Don't use blue: it is not part of the brand.
2. Don't use the raw palette (`moss-800`, `clay-500`) inside components. If you need a new color, create a semantic token in `tokens.json`.
3. Don't use `<div onClick>` as a button: use `<button>`, `<a href>`, `<input>` with `<label>`.
4. Don't use emojis as icons or decoration. Icons: stroke SVGs (stroke 1.7–1.9, `round`), 20 px in general.
5. Don't use gradients, card shadows, glassmorphism or cards with a colored left border.
6. Don't put UPPERCASE, wide-tracked labels above every heading. Use sentence case.
7. Don't close the logo circle, don't recolor the node green or water, don't animate it decoratively.
8. Don't build long boxes with every data point visible. Essentials first in the list; the rest in the drawer (section 8).
9. Don't invent data, metrics or filler text. If a value is missing, show an empty state or an explicit placeholder.

---

## 2. Brand and logo

Files in `logos/`:

| File | Use |
|---|---|
| `zensor-lockup.svg` | Mark + wordmark on a light background. Headers, documents. |
| `zensor-lockup-light.svg` | Mark + wordmark on a dark background (`--zs-bg-inverse`). |
| `zensor-mark.svg` / `-light.svg` | Mark only, for square spaces (avatar, collapsed sidebar). |
| `zensor-mark-mono-dark.svg` / `-light.svg` | Single ink (print, embroidery, stamps). |
| `zensor-wordmark.svg` | "zensor" only. Text is converted to outlines: no font needed. |
| `zensor-app-icon.svg`, `app-icon-1024.png`, `app-icon-512.png` | App icon (rounded corners). |
| `apple-touch-icon-180.png` | iOS (square, unrounded; iOS rounds it). |
| `zensor-favicon.svg`, `favicon.ico`, `favicon-16/32/48.png` | Favicon (larger node for legibility). |

Rules:
- Clear space: at least the node's width (~16 % of the mark's height) on every side.
- Minimum size: mark 16 px. Below 32 px use `zensor-favicon.svg`.
- On light backgrounds: stroke `--zs-brand-ink`, node `--zs-accent`. On dark: stroke `--zs-color-sand-100`, node `--zs-accent-on-inverse`.
- If the logo must follow the theme, inline it and use `fill="var(--zs-brand-ink)"` for the stroke and `fill="var(--zs-accent)"` for the node (`preview.html` does this).

HTML for `<head>`:

```html
<link rel="icon" href="/logos/zensor-favicon.svg" type="image/svg+xml">
<link rel="icon" href="/logos/favicon.ico" sizes="48x48">
<link rel="apple-touch-icon" href="/logos/apple-touch-icon-180.png">
<meta name="theme-color" content="#2B3A2A">
```

---

## 3. Color

Two layers. Layer 2 is the one you use.

1. **Raw palette** (`--zs-color-<family>-<step>`): `moss` (greens: Musgo 800, Bosque 600, Salvia 400, Liquen 200), `sand` (warm neutrals), `clay` (Tierra, the warm accent), `water` (Agua, the live state).
2. **Semantic** (`--zs-<name>`): these change in dark theme.

| I need… | Token | Tailwind |
|---|---|---|
| Page background | `--zs-bg-canvas` | `bg-canvas` |
| Card, row, drawer, panel | `--zs-bg-surface` | `bg-surface` |
| Sunken background / bar track | `--zs-bg-sunken` / `--zs-track` | `bg-sunken` / `bg-track` |
| Highlighted dark panel | `--zs-bg-inverse` (+ `--zs-text-inverse`) | `bg-inverse text-on-inverse` |
| Primary / supporting text | `--zs-text-primary` / `--zs-text-secondary` | `text-ink` / `text-muted` |
| Card border | `--zs-border-default` | `border-line` |
| Primary button | `--zs-action-primary-bg` + `-fg` | `bg-primary text-on-primary` |
| Highlighted action on dark ("Iniciar riego") | `--zs-accent-on-inverse` + `--zs-accent-on-inverse-fg` | `bg-accent-inverse text-on-accent-inverse` |
| OK state (online, optimal) | `--zs-status-ok-bg/-fg/-fill` | `bg-ok-bg text-ok-fg` / `bg-ok-fill` |
| Attention (offline, below target, destructive) | `--zs-status-attention-*` | `bg-warn-bg text-warn-fg` / `bg-warn-fill` |
| Live (watering, high saturation) | `--zs-status-live-*` | `bg-live-bg text-live-fg` / `bg-live-fill` |
| Focus | `--zs-focus-ring` | `outline-focus` (already global) |

Usage rules:
- **Tierra (`accent`)** is the attention color and the logo node's color. It is the only warm one. Don't use it for small text: for text use `--zs-text-attention` (darker, meets contrast).
- **Agua (`live`)** means "water is moving or saturated". It is not a decorative blue.
- Minimum contrast: 4.5:1 for normal text, 3:1 from 24 px. The `*-bg` / `*-fg` pairs already comply.
- A state is never distinguished by color alone: pair it with text ("Sin conexión") or an icon.
- **Dark theme**: defined in `tokens.css` but marked as *proposed* (not visually validated). Enable it with `data-theme="dark"` on `<html>` or let the system decide.

---

## 4. Typography

Fonts (Google Fonts):

```html
<link href="https://fonts.googleapis.com/css2?family=Zen+Maru+Gothic:wght@500;700&family=DM+Sans:wght@400;500;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
```

| Role | CSS class / Tailwind | Font · size · weight | Where |
|---|---|---|---|
| `display` | `.zs-text-display` / `text-display` | Zen Maru · 64 · 700 | Wordmark, covers |
| `h1` | `text-h1` | Zen Maru · 44 · 700 | Page title (web) |
| `h1-mobile` | `text-h1-mobile` | Zen Maru · 32 · 700 | Page title (mobile) |
| `h2` | `text-h2` | Zen Maru · 28 · 700 | Drawer title |
| `title` | `text-title` | Zen Maru · 22 · 700 | Device name |
| `h3` | `text-h3` | Zen Maru · 18 · 700 | Section title inside a panel |
| `body` | `text-body` | DM Sans · 16 · 400 | Running text |
| `body-sm` | `text-body-sm` | DM Sans · 14 · 400 | Supporting text in rows and cards |
| `label` | `text-label` | DM Sans · 14 · 500 | Labels, navigation |
| `caption` | `text-caption` | DM Sans · 13 · 400 | Metadata |
| `data-lg/md/sm` | `text-data-lg/md/sm` | JetBrains Mono · 28/22/14 · 500 | Sensor values |

- Maximum line length ~68 characters in running text.
- Sentence case in headings, buttons and labels ("Regar ahora", not "REGAR AHORA").
- Units go in a smaller size and `text-secondary` color next to the value: `31 %`, `25,0 °C`, `8,6 L/min`.
- Don't highlight a single word of a heading with another color or italics.

---

## 5. Spacing, layout and shape

- **4 px base.** Scale: `1=4, 2=8, 3=12, 4=16, 5=20, 6=24, 8=32, 10=40, 12=48, 16=64`.
- Web pages: content with `padding: 36px 44px` (`space-8`/`space-10`), max width 1440 px, 260 px sidebar (stacks on narrow screens).
- Mobile: 390 × 844 canvas, 20 px side padding, 84 px bottom bar with 4 destinations (Riego, Seguridad, Hogar, Fallas).
- Card grids with `flex-wrap` and `flex: 1 1 440px` (web) or 1 column (mobile). Don't hard-code px widths.
- **Radii:** `md 16` (tiles, rows), `lg 20` (banners), `xl 24` (device cards), `2xl 28` (brand panels), `pill` (buttons, chips, states), `sm 8` (bars).
- **Elevation:** cards are flat. Only what floats gets a shadow: `overlay` (menus), `drawer` (side panel), `modal` (dialogs, rare). See `elevation.json`.
- **Z-index:** `sticky 10`, `scrim 40`, `drawer 50`, `toast 60`.
- Breakpoints: `sm 640`, `md 768`, `lg 1024`, `xl 1280`, `2xl 1440`.

---

## 6. Motion

Motion **responds to an action** by the person (opening detail, enabling a schedule, changing a value). No decorative entrance animations and no animated hover on every card.

| Pattern | Duration | Curve | Effect |
|---|---|---|---|
| `drawer-enter` | 240 ms | decelerate | `translateX(48px) + fade` → 0 |
| `drawer-exit` | 140 ms | accelerate | leaves faster than it enters |
| `scrim` | 200 ms | standard | scrim fade |
| `press` | 80 ms | standard | `scale(.98)` on buttons and cards |
| `toggle` | 140 ms | standard | switch knob movement |
| `gauge-fill` | 400 ms | standard | arc and color transition |

Ready-made classes in `tokens.css`: `.zs-anim-drawer-in`, `.zs-anim-drawer-out`, `.zs-anim-scrim-in`, `.zs-press`. In Tailwind: `animate-drawer-in`, `animate-scrim-in`, `duration-fast ease-standard`.
With `prefers-reduced-motion: reduce` every duration drops to ~0 and the drawer only fades.

---

## 7. Components (recipes)

The recipes live as CSS in `preview.html` and here as Tailwind (with the preset). Copy the pattern.

### Primary / secondary / destructive button
```html
<button class="min-h-touch rounded-pill bg-primary px-5 text-md font-bold text-on-primary transition duration-instant ease-standard hover:bg-primary-hover active:scale-[.98]">Nuevo programa</button>
<button class="min-h-touch rounded-pill border-control border-brand bg-transparent px-5 text-md font-bold text-ink">Pausar todo</button>
<button class="min-h-touch rounded-pill px-5 text-md font-bold text-attention">Eliminar</button>
```
One primary button per view. The "Iniciar riego" action uses `bg-accent-inverse text-on-accent-inverse` inside a `bg-inverse` panel, 52 px tall (`min-h-control`).

### Filter chip (exclusive)
```html
<div role="group" aria-label="Filtro de dispositivos" class="flex flex-wrap gap-2">
  <button aria-pressed="true"  class="min-h-touch rounded-pill border-control border-brand bg-primary px-4 text-md font-bold text-on-primary">Todos <span class="font-mono font-medium">4</span></button>
  <button aria-pressed="false" class="min-h-touch rounded-pill border-control border-brand px-4 text-md font-bold text-ink">En línea <span class="font-mono font-medium">2</span></button>
</div>
```

### Status pill
```html
<span class="rounded-pill bg-ok-bg px-3 py-1 text-xs font-bold text-ok-fg">En línea</span>
<span class="rounded-pill bg-live-bg px-3 py-1 text-xs font-bold text-live-fg">Regando</span>
<span class="rounded-pill bg-warn-bg px-3 py-1 text-xs font-bold text-warn-fg">Sin conexión</span>
```

### Notice (one line + one action)
```html
<div class="flex flex-wrap items-center gap-4 rounded-lg bg-warn-bg px-5 py-3 text-warn-fg">
  <p class="min-w-0 flex-1">2 dispositivos sin conexión hace más de 170 días.</p>
  <button class="min-h-touch rounded-pill border-control border-warn-fg px-4 text-md font-bold">Ver</button>
</div>
```

### Device card (list)
One card = **one device, the essentials and a single click to the detail**. All controls live in the drawer.
- Header: name (`text-title`), state as text (`En línea · ahora` / `Sin conexión · hace 243 d`) and a pill.
- Body (only when online): humidity bars per sensor, temperature and flow. When offline: "Sin datos recientes" and the card drops to `opacity-offline`.
- Footer: schedule summary ("Diario 01:00 · 20 min" or "Sin programas") and "Detalle ›".
```html
<button class="flex flex-1 basis-[440px] flex-col gap-5 rounded-xl border border-line bg-surface p-5 text-left text-ink">…</button>
```

### Sensor tile
```html
<div class="flex min-w-[190px] flex-1 flex-col gap-2 rounded-md border border-line bg-surface p-4">
  <span class="text-sm text-muted">Humedad #1</span>
  <span class="text-data-lg">95,0<small class="ml-1 text-xs text-muted">%</small></span>
  <span class="h-2 overflow-hidden rounded-sm bg-track"><i class="block h-full bg-live-fill" style="width:95%"></i></span>
</div>
```

### Ensō gauge (soil moisture)
Open 300° arc with the gap at the bottom. The **dark node marks the target**. Use `Zensor.gaugeDash(value)` and `Zensor.gaugeMarker(target)`; the arc color comes from `Zensor.humidityState(value).fill`. Parameters (in `tokens.json → domain.gauge`): radius 70, stroke 14, rotation 120°. See the interactive example in `preview.html`.

### Schedule switch
```html
<button role="switch" aria-checked="true" aria-label="Activar programa" class="min-h-touch w-[56px] …"> … </button>
```
Track 44 × 26, knob 20; on = `--zs-status-ok-fill`, off = `--zs-border-strong`. Full recipe in `preview.html`.

### Empty state
Text in a dashed-border box (`border-dashed border-line-strong`) that explains what is missing and how to fix it: *"Aún no hay programas. Crea uno para que Zensor Garden riegue solo."* No large illustrations and no duplicated buttons: the "Agregar" action is already in the section header.

### Drawer (right-side panel)
- 500 px wide (`w-drawer max-w-full`), full height, `bg-canvas`, `shadow-drawer` shadow, `z-drawer`, enters with `animate-drawer-in`. Scrim `bg-scrim` (`z-scrim`) that closes on click.
- Structure: **fixed header** (name, state, 44 px close button) + **scrolling body** (`overflow-y-auto`).
- Accessibility: `role="dialog" aria-modal="true" aria-labelledby`, focus moves to the close button on open, **Escape closes**, focus returns to the element that opened it. `Zensor.drawer({ panel, scrim })` already implements this.
- Section order in the device drawer: **Regar ahora → Sensores → Programas → Consumo (7 días) → Dispositivo (señal, batería)**.

---

## 8. Experience patterns

1. **Simple list, detail on demand.** The list answers "is everything OK?" in 5 seconds. Every control and detailed value lives in the drawer.
2. **One notice instead of four KPIs.** Problems are summarized in a banner with an action ("Ver" filters by Sin conexión). No giant metric cards.
3. **Filters as chips with a count** (Todos · En línea · Sin conexión). A single exclusive group.
4. **Offline devices:** dimmed card, no invented data, and in the drawer the "Regar ahora" block explains why watering isn't possible. Never a silent grey disabled button.
5. **Regar ahora (water now):** duration (1–60 min, default 5) and a single button. While watering, the same button becomes "Detener riego", the pill becomes "Regando" and the valve "Abierta".
6. **Schedules (Programas):** list with an enable/pause switch, edit and delete. "Agregar" opens a second drawer or a step inside the same one (not a centered modal). Each action keeps its name throughout the flow ("Guardar programa" → "Programa guardado").
7. **Deleting is reversible or asks for confirmation.** Destructive actions use `text-attention`, never the primary button.
8. **Navigation:** sidebar with modules (Dispositivos, Mantención, Medicinas, Energía, Mensajes en vivo) and a tenant selector at the top. On mobile, a bottom bar with 4 destinations.

---

## 9. Domain rules

**Soil moisture (%)** — `Zensor.humidityState(v)`:

| Range | State | Fill token | Text |
|---|---|---|---|
| `< 30` | attention | `--zs-status-attention-fill` | Bajo |
| `30 – 84` | ok | `--zs-status-ok-fill` | Óptimo |
| `≥ 85` | live | `--zs-status-live-fill` | Saturado |

- The moisture target is marked with a line (bars) or the node (gauge). Show "N puntos bajo el objetivo de 40 %".
- A `0 %` reading on a sensor that used to report something else may mean a disconnected sensor: show it as is and don't interpret it as an OK state.
- **Units:** `%`, `°C`, `L/min`, `L`, `min`. With a space between number and unit (`31 %`, use a non-breaking space).
- **Decimals:** decimal comma (`8,6`). Use `Zensor.formatNumber()` / `formatUnit()`. Sensors with 1 decimal (`25,0`), whole liters.
- **Dates and times:** 24 h (`19:30`), relative when it helps ("hace 243 d", "mañana 01:00").
- **Valve / relay:** `Abierta` / `Cerrada` (not "On/Off").

---

## 10. Voice and content

- Chilean Spanish, informal "tú". Sentence case. Active verbs. Short sentences.
- Buttons say exactly what happens: "Iniciar riego", "Guardar programa". Not "Enviar", not "Aceptar".
- Name things the way the person watering understands them, not how they are built: "Programas", not "tareas programadas"; "Sin conexión", not "Offline".
- Errors say what happened and what to do, without apologizing: "No se pudo iniciar el riego: Yurta no responde. Revisa su alimentación y vuelve a intentarlo."
- Empty states invite action; they don't describe the emptiness.
- No emojis. No exclamation marks. No marketing language in the interface.

---

## 11. Accessibility

- AA contrast minimum. Touch targets ≥ 44 px.
- Visible focus: 2 px `--zs-focus-ring` with `outline-offset: 2px` (global in `tokens.css`).
- Gauge and charts: `role="img"` with an `aria-label` that includes the value and the target ("Humedad de suelo 31 por ciento, objetivo 40 por ciento").
- Decorative icons: `aria-hidden="true"`. Icon-only buttons: `aria-label`.
- Switches with `role="switch"` and `aria-checked`. Filters with `aria-pressed`.
- Drawer: `role="dialog"`, `aria-modal`, focus trapped/restored, closes with Escape.
- Respect `prefers-reduced-motion` and `prefers-color-scheme`.

---

## 12. Installation

```css
/* Global CSS */
@import './design-system/tokens.css';
```

```js
// tailwind.config.js
module.exports = {
  presets: [require('./design-system/tailwind.preset.js')],
  content: ['./src/**/*.{html,js,jsx,ts,tsx,vue,svelte}'],
};
```

```html
<script src="/design-system/support.js"></script>
<script>Zensor.theme.set('system'); /* 'light' | 'dark' | 'system' */</script>
```

The preset **replaces** Tailwind's colors, spacing, radii, shadows and sizes. Classes like `bg-blue-500` or `p-7` don't exist on purpose: if an LLM generates them, the build ignores them.

Syncing: edit `tokens.json` → regenerate `tokens.css` and review `tailwind.preset.js`. Don't edit `tokens.css` by hand.

### Pointing an LLM at this system
In `CLAUDE.md` / `AGENTS.md` / project rules:

```md
## Zensor UI
- Read `design-system/STYLE.md` (section 1: hard rules) before creating or editing UI.
- Use only semantic tokens from `tokens.css` / Tailwind preset classes. No loose hex values or default colors.
- Visual reference: `design-system/preview.html`. Logos in `design-system/logos/`.
- Main pattern: list of compact cards + right drawer with the detail.
```

---

## 13. Checklist before shipping UI

- [ ] Zero loose hex values and zero default Tailwind colors? Only semantic tokens?
- [ ] Is everything clickable ≥ 44 px and a real `<button>`/`<a>`/`<input>`?
- [ ] Do cards use a border, not a shadow? Do only the drawer/menus have shadows?
- [ ] Is sensor data in mono, with its unit and a decimal comma?
- [ ] Does every state have text in addition to color?
- [ ] Is the detail in the drawer, with the list showing only the essentials?
- [ ] Do offline and empty states explain what to do?
- [ ] Is focus visible, do icons have `aria-label`, and does Escape close the drawer?
- [ ] Does it work in light and dark themes and with `prefers-reduced-motion`?
- [ ] Is the copy in Chilean Spanish, sentence case, with no emojis or invented data?
