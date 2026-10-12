# Zensor — Guía de estilo (v1.0.0)

Guía del sistema de diseño de Zensor, escrita para personas y para LLMs que generan UI.
Si eres una LLM: **lee la sección 1 completa antes de escribir cualquier código de interfaz** y valida tu resultado con el checklist de la sección 12.

---

## 0. Qué es Zensor

Plataforma de IoT, permacultura y automatización para el campo. El primer módulo es **Riego** (sensores de humedad de suelo + clima + flujómetros para regar solo lo necesario). Después vienen Seguridad, Hogar y Detección de fallas.

**Personalidad de la marca:** calma, orgánica, confiable. Es una herramienta para alguien que ya tiene demasiadas cosas que mantener: la interfaz debe quitar trabajo, no sumar.

**Idea del logo:** un ensō (círculo zen) abierto, dibujado con un trazo de pincel de ancho variable. El círculo no se cierra, y el nodo en la abertura es el sensor. Zensor = *sensor* + *zen*.

## Archivos

| Archivo | Para qué sirve |
|---|---|
| `tokens.json` | **Fuente de verdad.** Paleta, semánticos claro/oscuro, tipografía, espacio, radios, tamaños, reglas de dominio. |
| `tokens.css` | Variables CSS `--zs-*` derivadas de `tokens.json`, tema claro/oscuro, roles tipográficos (`.zs-text-*`), foco y animaciones. |
| `elevation.json` | Sombras y z-index. |
| `motion.json` | Duraciones, curvas y patrones de movimiento. |
| `tailwind.preset.js` | Preset de Tailwind que **reemplaza** la escala por defecto con estos tokens. |
| `support.js` | `window.Zensor`: tema, estado de humedad, formato es-CL, gauge ensō, drawer. |
| `preview.html` | Catálogo vivo con ejemplos. Ábrelo en el navegador. |
| `logos/` | SVG y PNG del logo, íconos de app y favicons. |

---

## 1. Reglas duras (léelas primero)

**SIEMPRE**
1. Usa **tokens semánticos** para color: `--zs-bg-surface`, `--zs-text-primary`, `--zs-status-ok-fill`… (en Tailwind: `bg-surface`, `text-ink`, `bg-ok-fill`). Así funcionan en tema claro y oscuro.
2. Usa **solo valores de la escala**: espacio (`--zs-space-*`), radios (`--zs-radius-*`), tamaños de texto (`--zs-text-*`), duraciones (`--zs-duration-*`).
3. Cada elemento interactivo mide **al menos 44 × 44 px** (`--zs-size-touch-min`). Botones grandes de acción: 52 px.
4. Datos de sensores en **JetBrains Mono**. Títulos y marca en **Zen Maru Gothic**. Todo lo demás en **DM Sans**.
5. Las tarjetas se separan con **borde fino** (`--zs-border-default`), no con sombra.
6. El detalle de un dispositivo se muestra en un **panel lateral derecho (drawer)**, no en una página nueva ni en un modal centrado.
7. Todo estado se comunica con **texto + color** (nunca solo color).
8. Texto en **español neutro de Chile**, tuteo, frases cortas y verbos concretos. Formato numérico es-CL (`8,6`, `31 %`).
9. Respeta `prefers-reduced-motion` y deja el foco visible (ya lo hace `tokens.css`).

**NUNCA**
1. No escribas hexadecimales sueltos ni colores de Tailwind por defecto (`blue-500`, `gray-100`, `red-600`). No uses azul: no es parte de la marca.
2. No uses la paleta cruda (`moss-800`, `clay-500`) dentro de componentes. Si necesitas un color nuevo, crea un token semántico en `tokens.json`.
3. No uses `<div onClick>` como botón: usa `<button>`, `<a href>`, `<input>` con `<label>`.
4. No uses emojis como iconos ni como decoración. Iconos: SVG de trazo (stroke 1,7–1,9, `round`), 20 px en general.
5. No uses degradados, sombras en tarjetas, glassmorphism ni cards con borde izquierdo de color.
6. No pongas etiquetas en MAYÚSCULAS con tracking ancho encima de cada título. Usa sentence case.
7. No cierres el círculo del logo, no recolorees el nodo con verde o agua, no lo animes decorativamente.
8. No hagas cajas largas con todos los datos a la vista. Primero lo esencial en la lista; el resto, en el drawer (sección 8).
9. No inventes datos, métricas ni textos de relleno. Si falta un dato, muestra un estado vacío o un marcador explícito.

---

## 2. Marca y logo

Archivos en `logos/`:

| Archivo | Uso |
|---|---|
| `zensor-lockup.svg` | Marca + wordmark sobre fondo claro. Cabeceras, documentos. |
| `zensor-lockup-light.svg` | Marca + wordmark sobre fondo oscuro (`--zs-bg-inverse`). |
| `zensor-mark.svg` / `-light.svg` | Solo marca, para espacios cuadrados (avatar, sidebar colapsado). |
| `zensor-mark-mono-dark.svg` / `-light.svg` | Una sola tinta (impresión, bordado, sellos). |
| `zensor-wordmark.svg` | Solo "zensor". El texto está convertido a curvas: no requiere la fuente. |
| `zensor-app-icon.svg`, `app-icon-1024.png`, `app-icon-512.png` | Ícono de app (esquinas redondeadas). |
| `apple-touch-icon-180.png` | iOS (cuadrado sin redondeo; iOS lo redondea). |
| `zensor-favicon.svg`, `favicon.ico`, `favicon-16/32/48.png` | Favicon (nodo más grande para legibilidad). |

Reglas:
- Área de respeto: al menos el ancho del nodo (~16 % del alto de la marca) en todos los lados.
- Tamaño mínimo: marca 16 px. Bajo 32 px usa `zensor-favicon.svg`.
- Sobre fondo claro: trazo `--zs-brand-ink`, nodo `--zs-accent`. Sobre oscuro: trazo `--zs-color-sand-100`, nodo `--zs-accent-on-inverse`.
- Si necesitas que el logo cambie con el tema, incrústalo inline y usa `fill="var(--zs-brand-ink)"` para el trazo y `fill="var(--zs-accent)"` para el nodo (`preview.html` lo hace así).

HTML para `<head>`:

```html
<link rel="icon" href="/logos/zensor-favicon.svg" type="image/svg+xml">
<link rel="icon" href="/logos/favicon.ico" sizes="48x48">
<link rel="apple-touch-icon" href="/logos/apple-touch-icon-180.png">
<meta name="theme-color" content="#2B3A2A">
```

---

## 3. Color

Dos capas. La capa 2 es la que se usa.

1. **Paleta cruda** (`--zs-color-<familia>-<paso>`): `moss` (verdes: Musgo 800, Bosque 600, Salvia 400, Liquen 200), `sand` (neutros cálidos), `clay` (Tierra, acento cálido), `water` (Agua, estado vivo).
2. **Semánticos** (`--zs-<nombre>`): cambian en tema oscuro.

| Necesito… | Token | Tailwind |
|---|---|---|
| Fondo de página | `--zs-bg-canvas` | `bg-canvas` |
| Tarjeta, fila, drawer, panel | `--zs-bg-surface` | `bg-surface` |
| Fondo hundido / pista de barra | `--zs-bg-sunken` / `--zs-track` | `bg-sunken` / `bg-track` |
| Panel oscuro destacado | `--zs-bg-inverse` (+ `--zs-text-inverse`) | `bg-inverse text-on-inverse` |
| Texto principal / de apoyo | `--zs-text-primary` / `--zs-text-secondary` | `text-ink` / `text-muted` |
| Borde de tarjeta | `--zs-border-default` | `border-line` |
| Botón primario | `--zs-action-primary-bg` + `-fg` | `bg-primary text-on-primary` |
| Acción destacada sobre oscuro (Iniciar riego) | `--zs-accent-on-inverse` + `--zs-accent-on-inverse-fg` | `bg-accent-inverse text-on-accent-inverse` |
| Estado ok (en línea, óptimo) | `--zs-status-ok-bg/-fg/-fill` | `bg-ok-bg text-ok-fg` / `bg-ok-fill` |
| Atención (sin conexión, bajo objetivo, destructivo) | `--zs-status-attention-*` | `bg-warn-bg text-warn-fg` / `bg-warn-fill` |
| Vivo (regando, saturación alta) | `--zs-status-live-*` | `bg-live-bg text-live-fg` / `bg-live-fill` |
| Foco | `--zs-focus-ring` | `outline-focus` (ya global) |

Reglas de uso:
- **Tierra (`accent`)** es el color de atención y del nodo del logo. Es el único cálido. No lo uses para texto pequeño: para texto usa `--zs-text-attention` (más oscuro, cumple contraste).
- **Agua (`live`)** significa "hay agua en movimiento o saturada". No es un azul decorativo.
- Contraste mínimo: 4,5:1 para texto normal, 3:1 desde 24 px. Los pares `*-bg` / `*-fg` ya cumplen.
- Un estado nunca se distingue solo por color: acompáñalo de texto ("Sin conexión") o icono.
- **Tema oscuro**: definido en `tokens.css` pero marcado como *propuesto* (no validado visualmente). Actívalo con `data-theme="dark"` en `<html>` o deja que lo decida el sistema.

---

## 4. Tipografía

Fuentes (Google Fonts):

```html
<link href="https://fonts.googleapis.com/css2?family=Zen+Maru+Gothic:wght@500;700&family=DM+Sans:wght@400;500;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
```

| Rol | Clase CSS / Tailwind | Fuente · tamaño · peso | Dónde |
|---|---|---|---|
| `display` | `.zs-text-display` / `text-display` | Zen Maru · 64 · 700 | Wordmark, portadas |
| `h1` | `text-h1` | Zen Maru · 44 · 700 | Título de página (web) |
| `h1-mobile` | `text-h1-mobile` | Zen Maru · 32 · 700 | Título de página (móvil) |
| `h2` | `text-h2` | Zen Maru · 28 · 700 | Título del drawer |
| `title` | `text-title` | Zen Maru · 22 · 700 | Nombre de dispositivo |
| `h3` | `text-h3` | Zen Maru · 18 · 700 | Título de sección dentro de un panel |
| `body` | `text-body` | DM Sans · 16 · 400 | Texto corrido |
| `body-sm` | `text-body-sm` | DM Sans · 14 · 400 | Apoyo en filas y tarjetas |
| `label` | `text-label` | DM Sans · 14 · 500 | Etiquetas, navegación |
| `caption` | `text-caption` | DM Sans · 13 · 400 | Metadatos |
| `data-lg/md/sm` | `text-data-lg/md/sm` | JetBrains Mono · 28/22/14 · 500 | Valores de sensores |

- Largo de línea máximo ~68 caracteres en texto corrido.
- Sentence case en títulos, botones y etiquetas ("Regar ahora", no "REGAR AHORA").
- Las unidades van en tamaño menor y color `text-secondary` junto al valor: `31 %`, `25,0 °C`, `8,6 L/min`.
- No resaltes una sola palabra de un título con otro color o cursiva.

---

## 5. Espacio, layout y forma

- **Base 4 px.** Escala: `1=4, 2=8, 3=12, 4=16, 5=20, 6=24, 8=32, 10=40, 12=48, 16=64`.
- Páginas web: contenido con `padding: 36px 44px` (`space-8`/`space-10`), ancho máximo 1440 px, sidebar de 260 px (se apila en pantallas angostas).
- Móvil: lienzo de 390 × 844, padding lateral 20 px, barra inferior de 84 px con 4 destinos (Riego, Seguridad, Hogar, Fallas).
- Grillas de tarjetas con `flex-wrap` y `flex: 1 1 440px` (web) o 1 columna (móvil). No fijes anchos en px.
- **Radios:** `md 16` (tiles, filas), `lg 20` (banners), `xl 24` (tarjetas de dispositivo), `2xl 28` (paneles de marca), `pill` (botones, chips, estados), `sm 8` (barras).
- **Elevación:** las tarjetas son planas. Solo flota lo que flota: `overlay` (menús), `drawer` (panel lateral), `modal` (diálogos, poco frecuentes). Ver `elevation.json`.
- **Z-index:** `sticky 10`, `scrim 40`, `drawer 50`, `toast 60`.
- Breakpoints: `sm 640`, `md 768`, `lg 1024`, `xl 1280`, `2xl 1440`.

---

## 6. Movimiento

El movimiento **responde a una acción** de la persona (abrir detalle, activar programa, cambiar un valor). No hay animaciones de entrada decorativas ni hover animados en cada tarjeta.

| Patrón | Duración | Curva | Efecto |
|---|---|---|---|
| `drawer-enter` | 240 ms | decelerate | `translateX(48px) + fade` → 0 |
| `drawer-exit` | 140 ms | accelerate | sale más rápido de lo que entra |
| `scrim` | 200 ms | standard | fade del velo |
| `press` | 80 ms | standard | `scale(.98)` en botones y tarjetas |
| `toggle` | 140 ms | standard | movimiento de la perilla del switch |
| `gauge-fill` | 400 ms | standard | transición del arco y del color |

Clases listas en `tokens.css`: `.zs-anim-drawer-in`, `.zs-anim-drawer-out`, `.zs-anim-scrim-in`, `.zs-press`. En Tailwind: `animate-drawer-in`, `animate-scrim-in`, `duration-fast ease-standard`.
Con `prefers-reduced-motion: reduce` todas las duraciones pasan a ~0 y el drawer solo hace fade.

---

## 7. Componentes (recetas)

Las recetas están en CSS en `preview.html` y aquí en Tailwind (con el preset). Copia el patrón.

### Botón primario / secundario / destructivo
```html
<button class="min-h-touch rounded-pill bg-primary px-5 text-md font-bold text-on-primary transition duration-instant ease-standard hover:bg-primary-hover active:scale-[.98]">Nuevo programa</button>
<button class="min-h-touch rounded-pill border-control border-brand bg-transparent px-5 text-md font-bold text-ink">Pausar todo</button>
<button class="min-h-touch rounded-pill px-5 text-md font-bold text-attention">Eliminar</button>
```
Un solo botón primario por vista. La acción "Iniciar riego" va en `bg-accent-inverse text-on-accent-inverse` dentro de un panel `bg-inverse`, altura 52 px (`min-h-control`).

### Chip de filtro (exclusivo)
```html
<div role="group" aria-label="Filtro de dispositivos" class="flex flex-wrap gap-2">
  <button aria-pressed="true"  class="min-h-touch rounded-pill border-control border-brand bg-primary px-4 text-md font-bold text-on-primary">Todos <span class="font-mono font-medium">4</span></button>
  <button aria-pressed="false" class="min-h-touch rounded-pill border-control border-brand px-4 text-md font-bold text-ink">En línea <span class="font-mono font-medium">2</span></button>
</div>
```

### Píldora de estado
```html
<span class="rounded-pill bg-ok-bg px-3 py-1 text-xs font-bold text-ok-fg">En línea</span>
<span class="rounded-pill bg-live-bg px-3 py-1 text-xs font-bold text-live-fg">Regando</span>
<span class="rounded-pill bg-warn-bg px-3 py-1 text-xs font-bold text-warn-fg">Sin conexión</span>
```

### Aviso (una sola línea + una acción)
```html
<div class="flex flex-wrap items-center gap-4 rounded-lg bg-warn-bg px-5 py-3 text-warn-fg">
  <p class="min-w-0 flex-1">2 dispositivos sin conexión hace más de 170 días.</p>
  <button class="min-h-touch rounded-pill border-control border-warn-fg px-4 text-md font-bold">Ver</button>
</div>
```

### Tarjeta de dispositivo (lista)
Una tarjeta = **un dispositivo, lo esencial y un solo clic para el detalle**. Todo el control vive en el drawer.
- Encabezado: nombre (`text-title`), estado en texto (`En línea · ahora` / `Sin conexión · hace 243 d`) y píldora.
- Cuerpo (solo si está en línea): barras de humedad por sensor, temperatura y caudal. Si está sin conexión: "Sin datos recientes" y la tarjeta baja a `opacity-offline`.
- Pie: resumen del programa ("Diario 01:00 · 20 min" o "Sin programas") y "Detalle ›".
```html
<button class="flex flex-1 basis-[440px] flex-col gap-5 rounded-xl border border-line bg-surface p-5 text-left text-ink">…</button>
```

### Tile de sensor
```html
<div class="flex min-w-[190px] flex-1 flex-col gap-2 rounded-md border border-line bg-surface p-4">
  <span class="text-sm text-muted">Humedad #1</span>
  <span class="text-data-lg">95,0<small class="ml-1 text-xs text-muted">%</small></span>
  <span class="h-2 overflow-hidden rounded-sm bg-track"><i class="block h-full bg-live-fill" style="width:95%"></i></span>
</div>
```

### Gauge ensō (humedad de suelo)
Arco abierto de 300° con la abertura abajo. El **nodo oscuro marca el objetivo**. Usa `Zensor.gaugeDash(valor)` y `Zensor.gaugeMarker(objetivo)`; el color del arco lo decide `Zensor.humidityState(valor).fill`. Parámetros (en `tokens.json → domain.gauge`): radio 70, trazo 14, rotación 120°. Ver el ejemplo interactivo en `preview.html`.

### Switch de programa
```html
<button role="switch" aria-checked="true" aria-label="Activar programa" class="min-h-touch w-[56px] …"> … </button>
```
Pista 44 × 26, perilla 20; encendido = `--zs-status-ok-fill`, apagado = `--zs-border-strong`. Receta completa en `preview.html`.

### Estado vacío
Texto en una caja con borde discontinuo (`border-dashed border-line-strong`), explica qué falta y cómo resolverlo: *"Aún no hay programas. Crea uno para que Zensor Garden riegue solo."* Sin ilustraciones grandes ni botones duplicados: la acción "Agregar" ya está en el encabezado de la sección.

### Drawer (panel lateral derecho)
- Ancho 500 px (`w-drawer max-w-full`), alto completo, `bg-canvas`, sombra `shadow-drawer`, `z-drawer`, entra con `animate-drawer-in`. Velo `bg-scrim` (`z-scrim`) que cierra al hacer clic.
- Estructura: **encabezado fijo** (nombre, estado, botón cerrar de 44 px) + **cuerpo con scroll** (`overflow-y-auto`).
- Accesibilidad: `role="dialog" aria-modal="true" aria-labelledby`, foco al abrir en el botón cerrar, **Escape cierra**, el foco vuelve al elemento que lo abrió. `Zensor.drawer({ panel, scrim })` ya implementa esto.
- Orden de secciones del drawer de dispositivo: **Regar ahora → Sensores → Programas → Consumo (7 días) → Dispositivo (señal, batería)**.

---

## 8. Patrones de experiencia

1. **Lista simple, detalle bajo demanda.** La lista responde "¿está todo bien?" en 5 segundos. Todo control y dato detallado vive en el drawer.
2. **Un aviso en lugar de cuatro KPI.** Los problemas se resumen en un banner con acción ("Ver" filtra por Sin conexión). No hay tarjetas de métricas gigantes.
3. **Filtros como chips con contador** (Todos · En línea · Sin conexión). Un solo grupo exclusivo.
4. **Dispositivos sin conexión:** tarjeta atenuada, sin datos inventados, y en el drawer el bloque "Regar ahora" explica por qué no se puede regar. No se deshabilita con un botón gris mudo.
5. **Regar ahora:** duración (1–60 min, por defecto 5) y un solo botón. Mientras riega, el mismo botón pasa a "Detener riego", la píldora a "Regando" y la válvula a "Abierta".
6. **Programas:** lista con switch activar/pausar, editar y eliminar. "Agregar" abre un segundo drawer o un paso dentro del mismo (no un modal centrado). Cada acción conserva su nombre en todo el flujo ("Guardar programa" → "Programa guardado").
7. **Eliminar es reversible o confirma.** Las acciones destructivas usan `text-attention`, nunca el botón primario.
8. **Navegación:** sidebar con módulos (Dispositivos, Mantención, Medicinas, Energía, Mensajes en vivo) y selector de tenant arriba. En móvil, barra inferior de 4 destinos.

---

## 9. Reglas de dominio

**Humedad de suelo (%)** — `Zensor.humidityState(v)`:

| Rango | Estado | Token de relleno | Texto |
|---|---|---|---|
| `< 30` | attention | `--zs-status-attention-fill` | Bajo |
| `30 – 84` | ok | `--zs-status-ok-fill` | Óptimo |
| `≥ 85` | live | `--zs-status-live-fill` | Saturado |

- El objetivo de humedad se marca con una línea (barras) o con el nodo (gauge). Se muestra "N puntos bajo el objetivo de 40 %".
- Una lectura de `0 %` en un sensor que antes medía otra cosa puede indicar un sensor desconectado: muéstrala tal cual y no la interpretes como estado ok.
- **Unidades:** `%`, `°C`, `L/min`, `L`, `min`. Con espacio entre número y unidad (`31 %`, usa espacio duro).
- **Decimales:** coma decimal (`8,6`). Usa `Zensor.formatNumber()` / `formatUnit()`. Sensores con 1 decimal (`25,0`), litros enteros.
- **Fechas y horas:** 24 h (`19:30`), relativas cuando ayuda ("hace 243 d", "mañana 01:00").
- **Válvula / relay:** `Abierta` / `Cerrada` (no "On/Off").

---

## 10. Voz y contenido

- Español de Chile, tuteo. Sentence case. Verbos activos. Frases cortas.
- Los botones dicen exactamente qué pasa: "Iniciar riego", "Guardar programa". No "Enviar", no "Aceptar".
- Nombra las cosas como las entiende quien riega, no como están construidas: "Programas", no "tareas programadas"; "Sin conexión", no "Offline".
- Los errores dicen qué pasó y qué hacer, sin disculpas: "No se pudo iniciar el riego: Yurta no responde. Revisa su alimentación y vuelve a intentarlo."
- Los estados vacíos invitan a actuar, no describen el vacío.
- Sin emojis. Sin signos de exclamación. Sin lenguaje de marketing en la interfaz.

---

## 11. Accesibilidad

- Contraste AA mínimo. Objetivos táctiles ≥ 44 px.
- Foco visible: anillo de 2 px `--zs-focus-ring` con `outline-offset: 2px` (global en `tokens.css`).
- Gauge y gráficos: `role="img"` con `aria-label` que incluya el valor y el objetivo ("Humedad de suelo 31 por ciento, objetivo 40 por ciento").
- Iconos decorativos: `aria-hidden="true"`. Botones solo-icono: `aria-label`.
- Switches con `role="switch"` y `aria-checked`. Filtros con `aria-pressed`.
- Drawer: `role="dialog"`, `aria-modal`, foco atrapado/restaurado, cierre con Escape.
- Respeta `prefers-reduced-motion` y `prefers-color-scheme`.

---

## 12. Instalación

```css
/* CSS global */
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

El preset **reemplaza** colores, espaciado, radios, sombras y tamaños de Tailwind. Clases como `bg-blue-500` o `p-7` no existen a propósito: si una LLM las genera, el build las ignora.

Sincronización: edita `tokens.json` → regenera `tokens.css` y revisa `tailwind.preset.js`. No edites `tokens.css` a mano.

### Cómo apuntar a una LLM a este sistema
En `CLAUDE.md` / `AGENTS.md` / reglas del proyecto:

```md
## UI de Zensor
- Lee `design-system/STYLEGUIDE.md` (sección 1: reglas duras) antes de crear o editar UI.
- Usa solo tokens semánticos de `tokens.css` / clases del preset de Tailwind. Nada de hex sueltos ni colores por defecto.
- Referencia visual: `design-system/preview.html`. Logos en `design-system/logos/`.
- Patrón principal: lista de tarjetas compactas + drawer derecho con el detalle.
```

---

## 13. Checklist antes de entregar UI

- [ ] ¿Cero hex sueltos y cero colores por defecto de Tailwind? ¿Solo tokens semánticos?
- [ ] ¿Todo lo clicable mide ≥ 44 px y es un `<button>`/`<a>`/`<input>` real?
- [ ] ¿Las tarjetas usan borde, no sombra? ¿Solo el drawer/menús llevan sombra?
- [ ] ¿Los datos de sensores van en mono, con unidad y coma decimal?
- [ ] ¿Cada estado tiene texto además de color?
- [ ] ¿El detalle está en el drawer, y la lista muestra solo lo esencial?
- [ ] ¿Los estados sin conexión y vacíos explican qué hacer?
- [ ] ¿Hay foco visible, `aria-label` en iconos y Escape cierra el drawer?
- [ ] ¿Funciona en tema claro y oscuro y con `prefers-reduced-motion`?
- [ ] ¿Copy en español de Chile, sentence case, sin emojis ni datos inventados?
